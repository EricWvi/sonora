package store

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"time"

	"sonora/internal/loudness"
	"sonora/internal/media"
)

// Loudness normalization brings every track to TargetLoudness by its own
// loudness, whatever album it is on: the library is mostly heard shuffled,
// where neighbouring tracks matching matters more than an album's own
// dynamics. A gain may cut freely but boost only as far as the true peak
// stays at PeakCeiling; there is no limiter.
const (
	TargetLoudness = -16.0 // LUFS
	PeakCeiling    = -1.0  // dBTP
)

// loudnessVersion identifies the measurement. Tracks measured by an older
// one are measured again; bump it when the meter or a decoder changes
// what it reads.
const loudnessVersion = 1

// trackGain is the normalization gain in dB for a track's loudness and
// true peak; nil when the track has no loudness.
func trackGain(level, peak sql.NullFloat64) *float64 {
	if !level.Valid {
		return nil
	}
	g := normalizationGain(level.Float64, peak.Float64)
	return &g
}

// normalizationGain brings lufs to TargetLoudness, boosting no further
// than the true peak allows, rounded to 0.01 dB. A programme whose peak is
// already past the ceiling is left alone rather than cut for it.
func normalizationGain(lufs, truePeak float64) float64 {
	g := TargetLoudness - lufs
	if g > 0 {
		g = max(0, min(g, PeakCeiling-truePeak))
	}
	// Adding zero turns a rounded -0 into 0, which JSON writes as such.
	return math.Round(g*100)/100 + 0
}

// RunLoudness measures, until ctx ends, the tracks not yet measured by
// the current loudnessVersion, newest first. It starts with whatever the
// library holds and is woken by every ingest, so it never delays one.
func (s *Store) RunLoudness(ctx context.Context) {
	var pending int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tracks WHERE loudness_version < ?`, loudnessVersion).Scan(&pending); err == nil && pending > 0 {
		slog.Info("measuring loudness", "tracks", pending)
	}
	// failed holds tracks whose file could not be read this run, so that
	// a missing file is not retried in a loop; it is retried next start.
	failed := map[int64]bool{}
	start, measured := time.Now(), 0
	for {
		did, err := s.measureNext(ctx, failed)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			slog.Error("loudness", "err", err)
		}
		if did {
			measured++
			continue
		}
		if measured > 0 {
			slog.Info("loudness measured", "tracks", measured, "took", time.Since(start).Round(time.Second))
			measured = 0
		}
		select {
		case <-ctx.Done():
			return
		case <-s.loudnessWake:
			start = time.Now()
		}
	}
}

// wakeLoudness tells RunLoudness there is something new to measure.
func (s *Store) wakeLoudness() {
	select {
	case s.loudnessWake <- struct{}{}:
	default:
	}
}

// measureNext measures the newest track that needs it; did is false when
// none does.
func (s *Store) measureNext(ctx context.Context, failed map[int64]bool) (did bool, err error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, file, format FROM tracks WHERE loudness_version < ? ORDER BY id DESC`, loudnessVersion)
	if err != nil {
		return false, err
	}
	var id int64
	var rel, format string
	found := false
	for rows.Next() {
		if err := rows.Scan(&id, &rel, &format); err != nil {
			rows.Close()
			return false, err
		}
		if !failed[id] {
			found = true
			break
		}
	}
	rows.Close()
	if !found {
		return false, rows.Err()
	}

	res, err := measureFile(ctx, filepath.Join(s.mediaDir, rel), media.Format(format))
	switch {
	case ctx.Err() != nil:
		return false, ctx.Err()
	case errors.Is(err, media.ErrUnsupported), errors.Is(err, media.ErrCorrupt):
		// Recorded as measured without a loudness: playback leaves the
		// track as it is.
		slog.Warn("cannot measure loudness", "track", id, "err", err)
		_, err = s.db.ExecContext(ctx, `UPDATE tracks SET loudness = NULL, true_peak = NULL,
			loudness_version = ? WHERE id = ? AND file = ?`, loudnessVersion, id, rel)
		return true, err
	case err != nil:
		slog.Error("measure loudness", "track", id, "err", err)
		failed[id] = true
		return true, nil
	}
	return true, s.saveTrackLoudness(ctx, id, rel, res)
}

func measureFile(ctx context.Context, path string, format media.Format) (loudness.Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return loudness.Result{}, err
	}
	defer f.Close()
	return media.Loudness(ctxReader{ctx, f}, format)
}

// ctxReader stops a decode when ctx ends.
type ctxReader struct {
	ctx context.Context
	io.ReadSeeker
}

func (r ctxReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.ReadSeeker.Read(p)
}

// saveTrackLoudness records the measurement of the track's file, unless
// the track is gone. Silence has no loudness.
func (s *Store) saveTrackLoudness(ctx context.Context, id int64, rel string, res loudness.Result) error {
	var level, peak sql.NullFloat64
	if l, ok := res.Integrated(); ok {
		level = sql.NullFloat64{Float64: l, Valid: true}
		peak = sql.NullFloat64{Float64: res.TruePeakDB(), Valid: true}
	}
	_, err := s.db.ExecContext(ctx, `UPDATE tracks SET loudness = ?, true_peak = ?,
		loudness_version = ? WHERE id = ? AND file = ?`, level, peak, loudnessVersion, id, rel)
	return err
}
