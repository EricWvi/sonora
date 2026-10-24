package store

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
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

func coverFiles(t *testing.T, s *testStore) int {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(s.dataDir, "covers"))
	if err != nil {
		t.Fatal(err)
	}
	return len(entries)
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
	if tr.Title != "晴天" || tr.Artist == nil || tr.Artist.Name != "周杰伦" || tr.Album == nil || tr.Album.Title != "叶惠美" {
		t.Fatalf("track = %+v", tr)
	}
	if tr.Album.Artist == nil || tr.Album.Artist.ID != tr.Artist.ID {
		t.Fatalf("album artist = %+v", tr.Album.Artist)
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
	if res.Track.Title != "无名之歌" || res.Track.Artist != nil || res.Track.Album != nil {
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
	if a.Album.ID != b.Album.ID || a.Artist.ID != b.Artist.ID {
		t.Fatalf("albums %d/%d, artists %d/%d", a.Album.ID, b.Album.ID, a.Artist.ID, b.Artist.ID)
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
	d, err := s.UpdateTrack(ctx, tr.ID, TrackInput{Title: "晴天 (Live)", Artist: "Jay Chou", Album: "Live"})
	if err != nil {
		t.Fatal(err)
	}
	if d.Artist.Name != "Jay Chou" || d.Album.Title != "Live" || d.Album.Artist.Name != "Jay Chou" {
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
	b2, err := s.UpdateTrack(ctx, b.ID, TrackInput{Title: b.Title, Artist: "Jay Chou", Album: "叶惠美"})
	if err != nil {
		t.Fatal(err)
	}
	if b2.Album.ID == a.Album.ID {
		t.Fatal("different album artists must give different albums")
	}
	merged, err := s.MergeArtists(ctx, a.Artist.ID, []int64{b2.Artist.ID})
	if err != nil {
		t.Fatal(err)
	}
	if merged.TrackCount != 2 || merged.AlbumCount != 1 {
		t.Fatalf("merged = %+v", merged)
	}
	if _, err := s.Artist(ctx, b2.Artist.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("source artist remains: %v", err)
	}
	album, err := s.Album(ctx, a.Album.ID)
	if err != nil || album.TrackCount != 2 {
		t.Fatalf("album = %+v, %v", album, err)
	}
	if _, err := s.MergeArtists(ctx, a.Artist.ID, []int64{a.Artist.ID}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("self merge: %v", err)
	}
}

func TestMergeAlbumsAndRenameConflicts(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	a := ingest(t, s, sample(t, "tagged.mp3"), "a.mp3", nil).Track
	b := ingest(t, s, sample(t, "untagged.mp3"), "b.mp3", nil).Track
	b2, err := s.UpdateTrack(ctx, b.ID, TrackInput{Title: "b", Artist: "周杰伦", Album: "叶惠美 (Deluxe)"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateAlbum(ctx, b2.Album.ID, AlbumInput{Title: "叶惠美", Artist: "周杰伦"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("rename onto existing album: %v", err)
	}
	album, err := s.MergeAlbums(ctx, a.Album.ID, []int64{b2.Album.ID})
	if err != nil || album.TrackCount != 2 || album.Cover == nil {
		t.Fatalf("merged = %+v, %v", album, err)
	}
	page, _ := s.Tracks(ctx, TrackQuery{AlbumID: a.Album.ID})
	if page.Total != 2 {
		t.Fatalf("album tracks = %d", page.Total)
	}
	other, err := s.UpdateTrack(ctx, b.ID, TrackInput{Title: "b", Artist: "Someone"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.RenameArtist(ctx, other.Artist.ID, "周杰伦"); !errors.Is(err, ErrConflict) {
		t.Fatalf("rename onto existing artist: %v", err)
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
	if err != nil || album.Cover == nil || filepath.Ext(*album.Cover) != ".png" {
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
