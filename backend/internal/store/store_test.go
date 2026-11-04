package store

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"

	"sonora/internal/media"
)

type testStore struct {
	*Store
	dataDir, mediaDir string
}

func openTest(t *testing.T) *testStore {
	t.Helper()
	ts := &testStore{dataDir: t.TempDir(), mediaDir: t.TempDir()}
	s, err := Open(ts.dataDir, ts.mediaDir)
	if err != nil {
		t.Fatal(err)
	}
	ts.Store = s
	t.Cleanup(func() { ts.Store.Close() })
	return ts
}

func sample(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "media", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func ingest(t *testing.T, s *testStore, data []byte, name string, lyrics []byte) IngestResult {
	t.Helper()
	u, err := s.Receive(bytes.NewReader(data), name)
	if err != nil {
		t.Fatalf("receive %s: %v", name, err)
	}
	res, err := s.Ingest(context.Background(), u, lyrics)
	if err != nil {
		t.Fatalf("ingest %s: %v", name, err)
	}
	return res
}

// mediaFiles lists the audio files in the media volume.
func mediaFiles(t *testing.T, s *testStore) []string {
	t.Helper()
	var files []string
	filepath.WalkDir(s.mediaDir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			files = append(files, path)
		}
		return err
	})
	return files
}

func names(refs []Ref) []string {
	out := []string{}
	for _, r := range refs {
		out = append(out, r.Name)
	}
	return out
}

func coverFiles(t *testing.T, s *testStore) int {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(s.dataDir, "covers"))
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() {
			n++
		}
	}
	return n
}

func TestCreateFirstAdminOnlyOnce(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	results := make(chan error, 8)
	for i := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.CreateFirstAdmin(ctx, "admin"+string(rune('a'+i)), "hash")
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	ok := 0
	for err := range results {
		switch {
		case err == nil:
			ok++
		case !errors.Is(err, ErrConflict):
			t.Fatalf("unexpected error %v", err)
		}
	}
	if ok != 1 {
		t.Fatalf("%d setups succeeded, want 1", ok)
	}
	if has, _ := s.HasAdmin(ctx); !has {
		t.Fatal("HasAdmin = false")
	}
}

func TestAdminSessions(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	a, err := s.CreateFirstAdmin(ctx, "Root", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if got, hash, err := s.AdminByUsername(ctx, "root"); err != nil || got != a || hash != "hash" {
		t.Fatalf("lookup is not case-insensitive: %+v %q %v", got, hash, err)
	}
	now := time.Unix(1_800_000_000, 0)
	token, err := s.CreateAdminSession(ctx, a.ID, now, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	sess, err := s.LookupAdminSession(ctx, token)
	if err != nil || sess.Admin != a || !sess.ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Fatalf("session = %+v, %v", sess, err)
	}
	if err := s.DeleteExpiredAdminSessions(ctx, now.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.LookupAdminSession(ctx, token); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expired session still present: %v", err)
	}
}

func TestIngestTaggedFile(t *testing.T) {
	s := openTest(t)
	res := ingest(t, s, sample(t, "tagged.flac"), "C:\\Music\\01 晴天.flac", nil)
	if res.Status != Created {
		t.Fatalf("status = %s", res.Status)
	}
	tr := res.Track
	if tr.Title != "晴天" || !slices.Equal(names(tr.Artists), []string{"周杰伦"}) || tr.Album == nil || tr.Album.Title != "叶惠美" {
		t.Fatalf("track = %+v", tr)
	}
	if !slices.Equal(tr.Album.Artists, tr.Artists) {
		t.Fatalf("album artists = %+v", tr.Album.Artists)
	}
	if tr.TrackNo != 3 || tr.DiscNo != 1 || tr.Year != 2003 || tr.Format != "flac" || tr.DurationMs != 3000 {
		t.Fatalf("fields = %+v", tr)
	}
	if tr.OriginalName != "01 晴天.flac" {
		t.Fatalf("original name = %q", tr.OriginalName)
	}
	// The embedded picture becomes the album cover, not a track cover.
	if tr.Album.Cover == nil || tr.Cover != nil {
		t.Fatalf("album cover = %v, track cover = %v", tr.Album.Cover, tr.Cover)
	}
	if !tr.HasLyrics || tr.Lyrics == "" {
		t.Fatal("embedded lyrics missing")
	}
	if files := mediaFiles(t, s); len(files) != 1 || filepath.Ext(files[0]) != ".flac" {
		t.Fatalf("media files = %v", files)
	}
}

func TestIngestDuplicateKeepsOneFile(t *testing.T) {
	s := openTest(t)
	data := sample(t, "tagged.mp3")
	first := ingest(t, s, data, "a.mp3", nil)
	second := ingest(t, s, data, "copy of a.mp3", nil)
	if second.Status != Duplicate || second.Track.ID != first.Track.ID {
		t.Fatalf("second = %s %d, first = %d", second.Status, second.Track.ID, first.Track.ID)
	}
	if files := mediaFiles(t, s); len(files) != 1 {
		t.Fatalf("media files = %v", files)
	}
}

func TestConcurrentDuplicateUploads(t *testing.T) {
	s := openTest(t)
	data := sample(t, "tagged.ogg")
	var wg sync.WaitGroup
	statuses := make(chan IngestStatus, 4)
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u, err := s.Receive(bytes.NewReader(data), "x.ogg")
			if err != nil {
				t.Error(err)
				return
			}
			res, err := s.Ingest(context.Background(), u, nil)
			if err != nil {
				t.Error(err)
				return
			}
			statuses <- res.Status
		}()
	}
	wg.Wait()
	close(statuses)
	created := 0
	for st := range statuses {
		if st == Created {
			created++
		}
	}
	if created != 1 || len(mediaFiles(t, s)) != 1 {
		t.Fatalf("created = %d, files = %v", created, mediaFiles(t, s))
	}
}

func TestReceiveRejectsBrokenFiles(t *testing.T) {
	s := openTest(t)
	if _, err := s.Receive(bytes.NewReader(sample(t, "truncated.flac")), "bad.flac"); !errors.Is(err, media.ErrCorrupt) {
		t.Fatalf("truncated: err = %v", err)
	}
	if _, err := s.Receive(bytes.NewReader(sample(t, "cover.jpg")), "song.mp3"); !errors.Is(err, media.ErrUnsupported) {
		t.Fatalf("jpeg named mp3: err = %v", err)
	}
	if files := mediaFiles(t, s); len(files) != 0 {
		t.Fatalf("rejected uploads left files: %v", files)
	}
}

func TestIngestUntaggedUsesFileNameAndCountsAsMissing(t *testing.T) {
	s := openTest(t)
	res := ingest(t, s, sample(t, "untagged.mp3"), "无名之歌.mp3", nil)
	if res.Track.Title != "无名之歌" || len(res.Track.Artists) != 0 || res.Track.Album != nil {
		t.Fatalf("track = %+v", res.Track)
	}
	ctx := context.Background()
	page, err := s.Tracks(ctx, TrackQuery{Missing: true})
	if err != nil || page.Total != 1 {
		t.Fatalf("missing = %+v, %v", page, err)
	}
	o, err := s.Overview(ctx)
	if err != nil || o.Tracks != 1 || o.MissingTags != 1 || o.MediaBytes != int64(len(sample(t, "untagged.mp3"))) {
		t.Fatalf("overview = %+v, %v", o, err)
	}
}

func TestSidecarLyricsInGB18030(t *testing.T) {
	s := openTest(t)
	lrc := "[00:00.00]刮风这天\r\n[00:02.00]我试过握着你手"
	gbk, err := simplifiedchinese.GB18030.NewEncoder().Bytes([]byte(lrc))
	if err != nil {
		t.Fatal(err)
	}
	res := ingest(t, s, sample(t, "tagged.mp3"), "晴天.mp3", gbk)
	if res.Track.Lyrics != "[00:00.00]刮风这天\n[00:02.00]我试过握着你手" {
		t.Fatalf("lyrics = %q", res.Track.Lyrics)
	}
}

func TestSameAlbumAcrossFormatsSharesAlbumAndCover(t *testing.T) {
	s := openTest(t)
	a := ingest(t, s, sample(t, "tagged.mp3"), "a.mp3", nil).Track
	b := ingest(t, s, sample(t, "tagged.m4a"), "b.m4a", nil).Track
	if a.Album.ID != b.Album.ID || !slices.Equal(a.Artists, b.Artists) {
		t.Fatalf("albums %d/%d, artists %v/%v", a.Album.ID, b.Album.ID, a.Artists, b.Artists)
	}
	if b.Cover != nil {
		t.Fatalf("identical art stored as track cover %v", *b.Cover)
	}
	if n := coverFiles(t, s); n != 1 {
		t.Fatalf("cover files = %d", n)
	}
}

func TestUpdateTrackMovesItAndPrunesEmptyAlbum(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	tr := ingest(t, s, sample(t, "tagged.flac"), "a.flac", nil).Track
	d, err := s.UpdateTrack(ctx, tr.ID, TrackInput{Title: "晴天 (Live)", Artists: []string{"Jay Chou"}, Album: "Live"})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(names(d.Artists), []string{"Jay Chou"}) || d.Album.Title != "Live" ||
		!slices.Equal(names(d.Album.Artists), []string{"Jay Chou"}) {
		t.Fatalf("updated = %+v", d)
	}
	albums, _ := s.Albums(ctx, AlbumQuery{})
	artists, _ := s.Artists(ctx, "", Window{})
	if albums.Total != 1 || artists.Total != 1 {
		t.Fatalf("albums = %d, artists = %d; old ones not pruned", albums.Total, artists.Total)
	}
	// The old album and its cover went away with it.
	if n := coverFiles(t, s); n != 0 {
		t.Fatalf("cover files = %d", n)
	}
	if _, err := s.UpdateTrack(ctx, tr.ID, TrackInput{Title: "  "}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("empty title: %v", err)
	}
	if _, err := s.UpdateTrack(ctx, 999, TrackInput{Title: "x"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing track: %v", err)
	}
}

func TestMergeArtistsMergesSameNamedAlbums(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	a := ingest(t, s, sample(t, "tagged.mp3"), "a.mp3", nil).Track
	b := ingest(t, s, sample(t, "tagged.ogg"), "b.ogg", nil).Track
	// Spell the second artist differently, as messy tags often do.
	b2, err := s.UpdateTrack(ctx, b.ID, TrackInput{Title: b.Title, Artists: []string{"Jay Chou"}, Album: "叶惠美"})
	if err != nil {
		t.Fatal(err)
	}
	if b2.Album.ID == a.Album.ID {
		t.Fatal("different album artists must give different albums")
	}
	merged, err := s.MergeArtists(ctx, a.Artists[0].ID, []int64{b2.Artists[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	if merged.TrackCount != 2 || merged.AlbumCount != 1 {
		t.Fatalf("merged = %+v", merged)
	}
	if _, err := s.Artist(ctx, b2.Artists[0].ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("source artist remains: %v", err)
	}
	album, err := s.Album(ctx, a.Album.ID)
	if err != nil || album.TrackCount != 2 {
		t.Fatalf("album = %+v, %v", album, err)
	}
	if _, err := s.MergeArtists(ctx, a.Artists[0].ID, []int64{a.Artists[0].ID}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("self merge: %v", err)
	}
}

func TestMergeAlbumsAndRenameConflicts(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	a := ingest(t, s, sample(t, "tagged.mp3"), "a.mp3", nil).Track
	b := ingest(t, s, sample(t, "untagged.mp3"), "b.mp3", nil).Track
	b2, err := s.UpdateTrack(ctx, b.ID, TrackInput{Title: "b", Artists: []string{"周杰伦"}, Album: "叶惠美 (Deluxe)"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateAlbum(ctx, b2.Album.ID, AlbumInput{Title: "叶惠美", Artists: []string{"周杰伦"}}); !errors.Is(err, ErrConflict) {
		t.Fatalf("rename onto existing album: %v", err)
	}
	album, err := s.MergeAlbums(ctx, a.Album.ID, []int64{b2.Album.ID}, false)
	if err != nil || album.TrackCount != 2 || album.Cover == nil {
		t.Fatalf("merged = %+v, %v", album, err)
	}
	page, _ := s.Tracks(ctx, TrackQuery{AlbumID: a.Album.ID})
	if page.Total != 2 {
		t.Fatalf("album tracks = %d", page.Total)
	}
	other, err := s.UpdateTrack(ctx, b.ID, TrackInput{Title: "b", Artists: []string{"Someone"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.RenameArtist(ctx, other.Artists[0].ID, "周杰伦"); !errors.Is(err, ErrConflict) {
		t.Fatalf("rename onto existing artist: %v", err)
	}
}

// newAlbumTrack ingests file as the only track of album "原声带" by artist.
func newAlbumTrack(t *testing.T, s *testStore, file, artist string) Track {
	t.Helper()
	tr := ingest(t, s, sample(t, file), file, nil).Track
	d, err := s.UpdateTrack(context.Background(), tr.ID, TrackInput{Title: file, Artists: []string{artist}, Album: "原声带"})
	if err != nil {
		t.Fatal(err)
	}
	return d.Track
}

func TestMergeAlbumsKeepingCovers(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	target := newAlbumTrack(t, s, "tagged.mp3", "甲")
	diff := newAlbumTrack(t, s, "tagged.flac", "乙")
	same := newAlbumTrack(t, s, "tagged.m4a", "丙")
	bare := newAlbumTrack(t, s, "untagged.mp3", "丁")
	own := newAlbumTrack(t, s, "tagged.ogg", "戊")
	targetAlbum, err := s.SetAlbumCover(ctx, target.Album.ID, gradient(t, 3, 3))
	if err != nil {
		t.Fatal(err)
	}
	diffAlbum, err := s.SetAlbumCover(ctx, diff.Album.ID, gradient(t, 2, 2))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`UPDATE albums SET cover = ? WHERE id = ?`, targetAlbum.Cover.Name, same.Album.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`UPDATE albums SET cover = NULL WHERE id = ?`, bare.Album.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetAlbumCover(ctx, own.Album.ID, gradient(t, 5, 5)); err != nil {
		t.Fatal(err)
	}
	ownTrack, err := s.SetTrackCover(ctx, own.ID, gradient(t, 4, 4))
	if err != nil {
		t.Fatal(err)
	}

	sources := []int64{diff.Album.ID, same.Album.ID, bare.Album.ID, own.Album.ID}
	album, err := s.MergeAlbums(ctx, target.Album.ID, sources, true)
	if err != nil || album.TrackCount != 5 || album.Cover.Name != targetAlbum.Cover.Name {
		t.Fatalf("merged = %+v, %v", album, err)
	}
	want := map[int64]*Cover{
		target.ID: nil,             // shows the album cover
		diff.ID:   diffAlbum.Cover, // keeps the cover of its old album
		same.ID:   nil,             // its old album had the same cover
		bare.ID:   nil,             // its old album had none
		own.ID:    ownTrack.Cover,  // its own cover stays
	}
	for id, c := range want {
		tr, err := s.Track(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if (tr.Cover == nil) != (c == nil) || c != nil && tr.Cover.Name != c.Name {
			t.Errorf("track %d cover = %+v, want %+v", id, tr.Cover, c)
		}
	}
	if _, err := os.Stat(filepath.Join(s.dataDir, "covers", diffAlbum.Cover.Name)); err != nil {
		t.Errorf("kept cover file: %v", err)
	}
}

func TestMergeAlbumsWithoutCoverTakesTheFirstSourceCover(t *testing.T) {
	for _, keep := range []bool{false, true} {
		s := openTest(t)
		ctx := context.Background()
		bare := newAlbumTrack(t, s, "untagged.mp3", "甲")
		src := newAlbumTrack(t, s, "tagged.mp3", "乙")
		if _, err := s.db.Exec(`UPDATE albums SET cover = NULL WHERE id = ?`, bare.Album.ID); err != nil {
			t.Fatal(err)
		}
		srcAlbum, err := s.SetAlbumCover(ctx, src.Album.ID, gradient(t, 2, 2))
		if err != nil {
			t.Fatal(err)
		}
		album, err := s.MergeAlbums(ctx, bare.Album.ID, []int64{src.Album.ID}, keep)
		if err != nil || album.Cover == nil || album.Cover.Name != srcAlbum.Cover.Name {
			t.Fatalf("keepCovers=%v: merged = %+v, %v", keep, album, err)
		}
		if tr, _ := s.Track(ctx, src.ID); tr.Cover != nil {
			t.Errorf("keepCovers=%v: track cover = %+v, want the album's", keep, tr.Cover)
		}
	}
}

func TestMergeAlbumsUnifyingCoversDropsTheOthers(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	target := newAlbumTrack(t, s, "tagged.mp3", "甲")
	src := newAlbumTrack(t, s, "tagged.flac", "乙")
	if _, err := s.SetAlbumCover(ctx, target.Album.ID, gradient(t, 3, 3)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetAlbumCover(ctx, src.Album.ID, gradient(t, 2, 2)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.MergeAlbums(ctx, target.Album.ID, []int64{src.Album.ID}, false); err != nil {
		t.Fatal(err)
	}
	if tr, _ := s.Track(ctx, src.ID); tr.Cover != nil {
		t.Errorf("track cover = %+v, want the album's", tr.Cover)
	}
	if n := coverFiles(t, s); n != 1 {
		t.Errorf("cover files = %d, want 1", n)
	}
}

func TestSingleAlbums(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	single := newAlbumTrack(t, s, "tagged.mp3", "甲")
	real := newAlbumTrack(t, s, "tagged.flac", "乙")
	if single.Album == nil || real.Album == nil {
		t.Fatal("tracks have no album")
	}
	if a, _ := s.Album(ctx, single.Album.ID); a.Single {
		t.Fatal("a new album starts single")
	}
	yes := true
	a, err := s.UpdateAlbum(ctx, single.Album.ID, AlbumInput{Title: "原声带", Artists: []string{"甲"}, Single: &yes})
	if err != nil || !a.Single {
		t.Fatalf("mark single = %+v, %v", a, err)
	}
	// omitting single keeps the mark
	if a, err = s.UpdateAlbum(ctx, single.Album.ID, AlbumInput{Title: "原声带", Artists: []string{"甲"}, Year: 2020}); err != nil || !a.Single {
		t.Fatalf("keep single = %+v, %v", a, err)
	}
	all, _ := s.Albums(ctx, AlbumQuery{})
	shown, _ := s.Albums(ctx, AlbumQuery{HideSingles: true})
	if all.Total != 2 || shown.Total != 1 || shown.Items[0].ID != real.Album.ID {
		t.Fatalf("all = %d, shown = %+v", all.Total, shown)
	}
	no := false
	if a, err = s.UpdateAlbum(ctx, single.Album.ID, AlbumInput{Title: "原声带", Artists: []string{"甲"}, Single: &no}); err != nil || a.Single {
		t.Fatalf("unmark single = %+v, %v", a, err)
	}
}

func TestAlbumDescription(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	a := newAlbumTrack(t, s, "tagged.mp3", "甲").Album
	b := newAlbumTrack(t, s, "tagged.flac", "乙").Album
	if a == nil || b == nil {
		t.Fatal("tracks have no album")
	}
	desc := "  第一行\n第二行  "
	got, err := s.UpdateAlbum(ctx, b.ID, AlbumInput{Title: b.Title, Artists: []string{"乙"}, Description: &desc})
	if err != nil || got.Description != "第一行\n第二行" {
		t.Fatalf("set = %q, %v", got.Description, err)
	}
	// omitting the description keeps it; an empty one clears it
	if got, _ = s.UpdateAlbum(ctx, b.ID, AlbumInput{Title: b.Title, Artists: []string{"乙"}}); got.Description != "第一行\n第二行" {
		t.Fatalf("keep = %q", got.Description)
	}
	long := strings.Repeat("长", maxAlbumDescription+1)
	if _, err := s.UpdateAlbum(ctx, b.ID, AlbumInput{Title: b.Title, Artists: []string{"乙"}, Description: &long}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("too long: %v", err)
	}
	// a merge target without a description takes the source's
	merged, err := s.MergeAlbums(ctx, a.ID, []int64{b.ID}, false)
	if err != nil || merged.Description != "第一行\n第二行" {
		t.Fatalf("merge = %q, %v", merged.Description, err)
	}
	empty := " "
	if got, _ = s.UpdateAlbum(ctx, a.ID, AlbumInput{Title: merged.Title, Artists: []string{"甲"}, Description: &empty}); got.Description != "" {
		t.Fatalf("clear = %q", got.Description)
	}
}

func TestDeleteRemovesFilesAndOrphans(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	a := ingest(t, s, sample(t, "tagged.mp3"), "a.mp3", nil).Track
	ingest(t, s, sample(t, "tagged.flac"), "b.flac", nil)
	if err := s.DeleteTrack(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if len(mediaFiles(t, s)) != 1 || coverFiles(t, s) != 1 {
		t.Fatalf("after one delete: files = %v, covers = %d", mediaFiles(t, s), coverFiles(t, s))
	}
	if err := s.DeleteAlbum(ctx, a.Album.ID); err != nil {
		t.Fatal(err)
	}
	o, _ := s.Overview(ctx)
	if o != (Overview{}) || len(mediaFiles(t, s)) != 0 || coverFiles(t, s) != 0 {
		t.Fatalf("after album delete: %+v, files = %v, covers = %d", o, mediaFiles(t, s), coverFiles(t, s))
	}
	if err := s.DeleteTrack(ctx, a.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second delete: %v", err)
	}
}

func TestCoversAndLyricsReplaceAndClear(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	tr := ingest(t, s, sample(t, "tagged.m4a"), "a.m4a", nil).Track
	png := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde")

	if _, err := s.SetAlbumCover(ctx, tr.Album.ID, []byte("not an image")); !errors.Is(err, ErrInvalid) {
		t.Fatalf("garbage cover: %v", err)
	}
	album, err := s.SetAlbumCover(ctx, tr.Album.ID, png)
	if err != nil || album.Cover == nil || filepath.Ext(album.Cover.Name) != ".png" {
		t.Fatalf("album = %+v, %v", album, err)
	}
	if n := coverFiles(t, s); n != 1 {
		t.Fatalf("replaced cover kept: %d files", n)
	}
	d, err := s.SetTrackCover(ctx, tr.ID, sample(t, "cover.jpg"))
	if err != nil || d.Cover == nil {
		t.Fatalf("track cover = %+v, %v", d.Cover, err)
	}
	if d, err = s.ClearTrackCover(ctx, tr.ID); err != nil || d.Cover != nil {
		t.Fatalf("cleared track cover = %v, %v", d.Cover, err)
	}
	if _, err := s.ClearAlbumCover(ctx, tr.Album.ID); err != nil || coverFiles(t, s) != 0 {
		t.Fatalf("clear album cover: %v, files = %d", err, coverFiles(t, s))
	}

	if d, err = s.SetTrackLyrics(ctx, tr.ID, []byte("\xEF\xBB\xBF[00:01.00]新歌词")); err != nil || d.Lyrics != "[00:01.00]新歌词" {
		t.Fatalf("lyrics = %q, %v", d.Lyrics, err)
	}
	if d, err = s.SetTrackLyrics(ctx, tr.ID, nil); err != nil || d.HasLyrics {
		t.Fatalf("cleared lyrics = %+v, %v", d, err)
	}
}

func TestLibrarySurvivesReopen(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	tr := ingest(t, s, sample(t, "tagged.opus"), "a.opus", nil).Track
	if _, err := s.CreateFirstAdmin(ctx, "admin", "hash"); err != nil {
		t.Fatal(err)
	}
	s.Store.Close()
	reopened, err := Open(s.dataDir, s.mediaDir)
	if err != nil {
		t.Fatal(err)
	}
	s.Store = reopened
	d, err := reopened.Track(ctx, tr.ID)
	if err != nil || d.Title != "晴天" || d.Format != "opus" {
		t.Fatalf("after reopen: %+v, %v", d, err)
	}
	if has, _ := reopened.HasAdmin(ctx); !has {
		t.Fatal("admin lost on reopen")
	}
	f, err := reopened.OpenAudio(ctx, tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if f.Format.ContentType() != "audio/ogg; codecs=opus" {
		t.Fatalf("content type = %s", f.Format.ContentType())
	}
}

func TestTrackSearchEscapesWildcards(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	ingest(t, s, sample(t, "tagged.mp3"), "a.mp3", nil)
	for q, want := range map[string]int{"晴": 1, "周杰": 1, "%": 0, "_": 0, "nothing": 0} {
		page, err := s.Tracks(ctx, TrackQuery{Q: q})
		if err != nil || page.Total != want {
			t.Errorf("q=%q: total = %d, %v", q, page.Total, err)
		}
	}
}

func TestMultipleArtists(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	a := ingest(t, s, sample(t, "tagged.flac"), "a.flac", nil).Track
	b := ingest(t, s, sample(t, "tagged.mp3"), "b.mp3", nil).Track

	// Blank and repeated names are dropped; album artists default to the
	// track artists.
	d, err := s.UpdateTrack(ctx, a.ID, TrackInput{Title: "千里之外", Artists: []string{"周杰伦", " 费玉清 ", "", "周杰伦"}, Album: "依然范特西"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"周杰伦", "费玉清"}; !slices.Equal(names(d.Artists), want) || !slices.Equal(names(d.Album.Artists), want) {
		t.Fatalf("artists = %v, album artists = %v", names(d.Artists), names(d.Album.Artists))
	}
	// The same artists in another order make the same album.
	d2, err := s.UpdateTrack(ctx, b.ID, TrackInput{Title: "b", Artists: []string{"费玉清", "周杰伦"}, Album: "依然范特西"})
	if err != nil {
		t.Fatal(err)
	}
	if d2.Album.ID != d.Album.ID || !slices.Equal(names(d2.Artists), []string{"费玉清", "周杰伦"}) {
		t.Fatalf("second track = %+v", d2)
	}
	fei := d.Artists[1]
	artist, err := s.Artist(ctx, fei.ID)
	if err != nil || artist.TrackCount != 2 || artist.AlbumCount != 1 {
		t.Fatalf("费玉清 = %+v, %v", artist, err)
	}
	if page, _ := s.Tracks(ctx, TrackQuery{ArtistID: fei.ID}); page.Total != 2 {
		t.Fatalf("tracks of 费玉清 = %d", page.Total)
	}
	if page, _ := s.Tracks(ctx, TrackQuery{Q: "费玉"}); page.Total != 2 {
		t.Fatalf("search 费玉 = %d", page.Total)
	}
	if page, _ := s.Albums(ctx, AlbumQuery{ArtistID: fei.ID}); page.Total != 1 {
		t.Fatalf("albums of 费玉清 = %d", page.Total)
	}

	// Album artists of their own; the old album is pruned.
	album, err := s.UpdateAlbum(ctx, d.Album.ID, AlbumInput{Title: "依然范特西", Artists: []string{"群星"}})
	if err != nil || !slices.Equal(names(album.Artists), []string{"群星"}) {
		t.Fatalf("album = %+v, %v", album, err)
	}
	if artist, _ := s.Artist(ctx, fei.ID); artist.AlbumCount != 0 || artist.TrackCount != 2 {
		t.Fatalf("费玉清 after album edit = %+v", artist)
	}
	// An artist's tracks are those they perform, not those of their albums.
	if page, _ := s.Tracks(ctx, TrackQuery{ArtistID: album.Artists[0].ID}); page.Total != 0 {
		t.Fatalf("tracks of 群星 = %d", page.Total)
	}
	if page, _ := s.Tracks(ctx, TrackQuery{ArtistID: fei.ID}); page.Total != 2 {
		t.Fatalf("tracks of 费玉清 after album edit = %d", page.Total)
	}
	if _, err := s.UpdateTrack(ctx, a.ID, TrackInput{Title: "x", Artists: make([]string, maxArtists+1)}); err != nil {
		t.Fatalf("blank names count toward the limit: %v", err)
	}
	many := make([]string, maxArtists+1)
	for i := range many {
		many[i] = string(rune('一' + i))
	}
	if _, err := s.UpdateTrack(ctx, a.ID, TrackInput{Title: "x", Artists: many}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("too many artists: %v", err)
	}
}

func TestMergeArtistsCollapsesLinksAndAlbums(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	a := ingest(t, s, sample(t, "tagged.flac"), "a.flac", nil).Track
	b := ingest(t, s, sample(t, "tagged.mp3"), "b.mp3", nil).Track
	duet, err := s.UpdateTrack(ctx, a.ID, TrackInput{Title: "a", Artists: []string{"Jay", "周杰伦"}, Album: "Duet"})
	if err != nil {
		t.Fatal(err)
	}
	solo, err := s.UpdateTrack(ctx, b.ID, TrackInput{Title: "b", Artists: []string{"周杰伦"}, Album: "Duet"})
	if err != nil {
		t.Fatal(err)
	}
	if duet.Album.ID == solo.Album.ID {
		t.Fatal("different artist sets must give different albums")
	}
	jay, chou := duet.Artists[0].ID, duet.Artists[1].ID
	merged, err := s.MergeArtists(ctx, chou, []int64{jay})
	if err != nil {
		t.Fatal(err)
	}
	if merged.TrackCount != 2 || merged.AlbumCount != 1 {
		t.Fatalf("merged = %+v", merged)
	}
	d, err := s.Track(ctx, a.ID)
	if err != nil || !slices.Equal(names(d.Artists), []string{"周杰伦"}) || d.Album.ID != solo.Album.ID {
		t.Fatalf("duet after merge = %+v, %v", d, err)
	}
	if albums, _ := s.Albums(ctx, AlbumQuery{}); albums.Total != 1 || albums.Items[0].TrackCount != 2 {
		t.Fatalf("albums = %+v", albums)
	}
}

func TestMigrateKeepsArtistsAndRetiredIDs(t *testing.T) {
	dataDir, mediaDir := t.TempDir(), t.TempDir()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(dataDir, "sonora.db")+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := migrations[0](tx); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`
		PRAGMA user_version = 1;
		INSERT INTO artists (id, name) VALUES (1, '周杰伦');
		INSERT INTO albums (id, title, artist_id) VALUES (1, '叶惠美', 1), (2, '无主', NULL), (3, '已删', NULL);
		INSERT INTO tracks (id, title, artist_id, album_id, duration_ms, format, size, sha256, file, original_name)
		VALUES (1, '晴天', 1, 1, 1000, 'mp3', 1, 'a', 'a.mp3', 'a.mp3'),
		       (2, '无名', NULL, 2, 1000, 'mp3', 1, 'b', 'b.mp3', 'b.mp3'),
		       (9, '已删', NULL, NULL, 1000, 'mp3', 1, 'c', 'c.mp3', 'c.mp3');
		DELETE FROM tracks WHERE id = 9;
		DELETE FROM albums WHERE id = 3;`); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	db.Close()

	s, err := Open(dataDir, mediaDir)
	if err != nil {
		t.Fatal(err)
	}
	ts := &testStore{Store: s, dataDir: dataDir, mediaDir: mediaDir}
	t.Cleanup(func() { s.Close() })
	ctx := context.Background()
	d, err := s.Track(ctx, 1)
	if err != nil || !slices.Equal(names(d.Artists), []string{"周杰伦"}) || !slices.Equal(names(d.Album.Artists), []string{"周杰伦"}) {
		t.Fatalf("migrated track = %+v, %v", d, err)
	}
	if d, err := s.Track(ctx, 2); err != nil || len(d.Artists) != 0 || len(d.Album.Artists) != 0 {
		t.Fatalf("track without artist = %+v, %v", d, err)
	}
	var fk int
	if err := s.db.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil || fk != 1 {
		t.Fatalf("foreign_keys = %d, %v", fk, err)
	}
	// The tagged sample joins the migrated album; deleted IDs stay retired.
	tr := ingest(t, ts, sample(t, "tagged.flac"), "x.flac", nil).Track
	if tr.ID != 10 || tr.Album.ID != 1 {
		t.Fatalf("new track %d in album %d", tr.ID, tr.Album.ID)
	}
	moved, err := s.UpdateTrack(ctx, tr.ID, TrackInput{Title: "x", Album: "新专辑"})
	if err != nil || moved.Album.ID != 4 {
		t.Fatalf("new album = %+v, %v", moved.Album, err)
	}
}

func TestIngestMultiValueArtists(t *testing.T) {
	s := openTest(t)
	// An ID3v2.4 tag whose TPE1 holds two UTF-8 values, before plain audio.
	var body []byte
	for _, f := range [][2]string{{"TIT2", "\x03千里之外"}, {"TPE1", "\x03周杰伦\x00费玉清"}, {"TALB", "\x03依然范特西"}} {
		n := len(f[1])
		body = append(append(body, f[0]...), byte(n>>21&0x7F), byte(n>>14&0x7F), byte(n>>7&0x7F), byte(n&0x7F), 0, 0)
		body = append(body, f[1]...)
	}
	n := len(body)
	data := append([]byte{'I', 'D', '3', 4, 0, 0, byte(n >> 21 & 0x7F), byte(n >> 14 & 0x7F), byte(n >> 7 & 0x7F), byte(n & 0x7F)}, body...)
	data = append(data, sample(t, "untagged.mp3")...)

	tr := ingest(t, s, data, "a.mp3", nil).Track
	if want := []string{"周杰伦", "费玉清"}; !slices.Equal(names(tr.Artists), want) || !slices.Equal(names(tr.Album.Artists), want) {
		t.Fatalf("artists = %v, album artists = %v", names(tr.Artists), names(tr.Album.Artists))
	}
	if artists, _ := s.Artists(context.Background(), "", Window{}); artists.Total != 2 {
		t.Fatalf("artists = %+v", artists)
	}
}

// gradient encodes a w×h PNG.
func gradient(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 200, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestCoverPlaceholdersAndThumbnails(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	tr := ingest(t, s, sample(t, "tagged.flac"), "a.flac", nil).Track
	if tr.Album.Cover == nil || tr.Album.Cover.Thumbhash == "" {
		t.Fatalf("embedded cover = %+v", tr.Album.Cover)
	}
	// The 64×64 sample is smaller than any thumbnail: the original serves.
	orig, _ := s.CoverPath(tr.Album.Cover.Name)
	if p, err := s.CoverThumb(tr.Album.Cover.Name, 96); err != nil || p != orig {
		t.Fatalf("small cover thumb = %s, %v", p, err)
	}

	album, err := s.SetAlbumCover(ctx, tr.Album.ID, gradient(t, 400, 200))
	if err != nil || album.Cover.Thumbhash == "" {
		t.Fatalf("album = %+v, %v", album, err)
	}
	p, err := s.CoverThumb(album.Cover.Name, 96)
	if err != nil || filepath.Ext(p) != ".jpg" {
		t.Fatalf("thumb = %s, %v", p, err)
	}
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	cfg, format, err := image.DecodeConfig(f)
	f.Close()
	if err != nil || format != "jpeg" || cfg.Width != 96 || cfg.Height != 48 {
		t.Fatalf("thumb is %s %dx%d, %v", format, cfg.Width, cfg.Height, err)
	}
	if again, err := s.CoverThumb(album.Cover.Name, 96); err != nil || again != p {
		t.Fatalf("second request = %s, %v", again, err)
	}
	if _, err := s.CoverThumb(album.Cover.Name, 97); !errors.Is(err, ErrInvalid) {
		t.Fatalf("odd size: %v", err)
	}
	if _, err := s.CoverThumb("../sonora.db", 96); !errors.Is(err, ErrNotFound) {
		t.Fatalf("bad name: %v", err)
	}

	// Clearing the cover takes its thumbnail and placeholder along.
	if _, err := s.ClearAlbumCover(ctx, tr.Album.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("thumbnail kept: %v", err)
	}
	var rows int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM covers`).Scan(&rows); err != nil || rows != 0 {
		t.Fatalf("covers rows = %d, %v", rows, err)
	}
}

func TestPlaceholdersBackfilledOnOpen(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	tr := ingest(t, s, sample(t, "tagged.flac"), "a.flac", nil).Track
	want := tr.Album.Cover.Thumbhash
	if _, err := s.db.Exec(`DELETE FROM covers`); err != nil {
		t.Fatal(err)
	}
	s.Store.Close()
	reopened, err := Open(s.dataDir, s.mediaDir)
	if err != nil {
		t.Fatal(err)
	}
	s.Store = reopened
	d, err := reopened.Track(ctx, tr.ID)
	if err != nil || d.Album.Cover.Thumbhash != want {
		t.Fatalf("after reopen: %+v, %v", d.Album.Cover, err)
	}
}
