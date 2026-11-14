package store

import (
	"context"
	"database/sql"
	"math"
	"testing"

	"sonora/internal/loudness"
)

// measureAll does what RunLoudness does until there is nothing left.
func measureAll(t *testing.T, s *Store) {
	t.Helper()
	ctx := context.Background()
	failed := map[int64]bool{}
	for {
		did, err := s.measureNext(ctx, failed)
		if err != nil {
			t.Fatal(err)
		}
		if !did {
			return
		}
	}
}

func gainOf(t *testing.T, s *Store, id int64) *float64 {
	t.Helper()
	d, err := s.Track(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return d.GainDB
}

func nearGain(t *testing.T, name string, got *float64, want float64) {
	t.Helper()
	if got == nil || math.Abs(*got-want) > 0.15 {
		t.Errorf("%s gain = %v, want %.2f", name, deref(got), want)
	}
}

func deref(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}

func TestNormalizationGain(t *testing.T) {
	for _, c := range []struct{ lufs, peak, want float64 }{
		{-8, 0.5, -8},    // cut freely, peak or not
		{-16, -1, 0},     // already on target
		{-20, -10, 4},    // boost with headroom to spare
		{-20, -3, 2},     // boost stops at the ceiling
		{-20, -1, 0},     // no headroom
		{-20, 0.3, 0},    // past the ceiling: not cut for it either
		{-16.004, -9, 0}, // rounded to 0.01 dB
		{-23.456, -4, 3}, // ceiling again
	} {
		if got := normalizationGain(c.lufs, c.peak); got != c.want {
			t.Errorf("normalizationGain(%v, %v) = %v, want %v", c.lufs, c.peak, got, c.want)
		}
	}
}

// The samples are a 440 Hz tone, mono, reading about -19.2 LUFS (MP3)
// and -18.8 LUFS (FLAC) as dual mono, with true peaks near -18 dBTP.
func TestIngestedTracksGetMeasured(t *testing.T) {
	s := openTest(t)
	mp3 := ingest(t, s, sample(t, "tagged.mp3"), "a.mp3", nil).Track
	flac := ingest(t, s, sample(t, "tagged.flac"), "b.flac", nil).Track
	m4a := ingest(t, s, sample(t, "tagged.m4a"), "c.m4a", nil).Track
	if mp3.GainDB != nil {
		t.Fatalf("gain before measuring = %v", *mp3.GainDB)
	}
	select {
	case <-s.loudnessWake:
	default:
		t.Fatal("ingest did not wake the loudness worker")
	}
	measureAll(t, s.Store)

	nearGain(t, "mp3", gainOf(t, s.Store, mp3.ID), 19.2-16)
	nearGain(t, "flac", gainOf(t, s.Store, flac.ID), 18.8-16)
	// The M4A cannot be decoded: measured, but plays as it is.
	if g := gainOf(t, s.Store, m4a.ID); g != nil {
		t.Errorf("m4a gain = %v, want none", *g)
	}
	var version int
	var level sql.NullFloat64
	s.db.QueryRow(`SELECT loudness_version, loudness FROM tracks WHERE id = ?`, m4a.ID).Scan(&version, &level)
	if version != loudnessVersion || level.Valid {
		t.Errorf("m4a: version %d, loudness %v", version, level)
	}
}

// setLoudness records a measurement of a 1 kHz tone at level dBFS for 10 s
// instead of decoding the track's file.
func setLoudness(t *testing.T, s *Store, id int64, level float64) {
	t.Helper()
	m := loudness.NewMeter(48000, []float64{loudness.Front, loudness.Front})
	amp := math.Pow(10, level/20)
	buf := make([]float64, 0, 2*48000*10)
	for n := range 48000 * 10 {
		v := amp * math.Sin(2*math.Pi*1000*float64(n)/48000)
		buf = append(buf, v, v)
	}
	m.WriteFloat64(buf)
	var rel string
	if err := s.db.QueryRow(`SELECT file FROM tracks WHERE id = ?`, id).Scan(&rel); err != nil {
		t.Fatal(err)
	}
	if err := s.saveTrackLoudness(context.Background(), id, rel, m.Result()); err != nil {
		t.Fatal(err)
	}
}

// Each track keeps its own gain on any album, single or not, and when it
// moves to another.
func TestGainIsTheTracksOwn(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	loud := ingest(t, s, sample(t, "tagged.mp3"), "a.mp3", nil).Track
	quiet := ingest(t, s, sample(t, "tagged.flac"), "b.flac", nil).Track
	if loud.Album == nil || quiet.Album == nil || loud.Album.ID != quiet.Album.ID {
		t.Fatal("samples are not on one album")
	}
	setLoudness(t, s.Store, loud.ID, -8)   // -8 LUFS, peak -8 dBTP
	setLoudness(t, s.Store, quiet.ID, -22) // -22 LUFS, peak -22 dBTP
	nearGain(t, "loud", gainOf(t, s.Store, loud.ID), -8)
	nearGain(t, "quiet", gainOf(t, s.Store, quiet.ID), 6)

	yes := true
	if _, err := s.UpdateAlbum(ctx, loud.Album.ID, AlbumInput{Title: loud.Album.Title, Single: &yes}); err != nil {
		t.Fatal(err)
	}
	nearGain(t, "loud on a single", gainOf(t, s.Store, loud.ID), -8)
	if _, err := s.UpdateTrack(ctx, loud.ID, TrackInput{Title: "Loud", Artists: []string{"x"}, Album: "Elsewhere"}); err != nil {
		t.Fatal(err)
	}
	nearGain(t, "loud elsewhere", gainOf(t, s.Store, loud.ID), -8)
	nearGain(t, "quiet left behind", gainOf(t, s.Store, quiet.ID), 6)
}

// A newer measurement version remeasures the library on the next run.
func TestOlderMeasurementsAreRedone(t *testing.T) {
	s := openTest(t)
	a := ingest(t, s, sample(t, "tagged.mp3"), "a.mp3", nil).Track
	measureAll(t, s.Store)
	if _, err := s.db.Exec(`UPDATE tracks SET loudness_version = ?, loudness = -5`, loudnessVersion-1); err != nil {
		t.Fatal(err)
	}
	measureAll(t, s.Store)
	nearGain(t, "remeasured", gainOf(t, s.Store, a.ID), 19.2-16)
}
