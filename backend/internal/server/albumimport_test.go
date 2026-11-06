package server

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"sonora/internal/media"
	"sonora/internal/netease"
	"sonora/internal/store"
)

// importAlbumSong is a song of a fake /album answer.
func importAlbumSong(id, no, cd, artist, st string) string {
	return `{"id":` + id + `,"name":"song` + id + `","ar":[{"id":5,"name":"` + artist + `"}],"dt":1000,"no":` + no + `,"cd":"` + cd + `",
		"al":{"id":77,"name":"Comp","picUrl":"CDN/a77.jpg"},"privilege":{"st":` + st + `}}`
}

func newAlbumImportTest(t *testing.T) (*testServer, *fakeNetease) {
	t.Helper()
	ts, f := newImportTest(t)
	f.importAlbums = map[string]string{
		// 701 was imported on its own before; 772 is grey, 790 a FLAC.
		"77": `{"code":200,"album":{"name":"Comp","picUrl":"CDN/a77.jpg","description":"notes","publishTime":1500000000000,
			"artists":[{"id":9,"name":"群星"}]},"songs":[` + strings.Join([]string{
			importAlbumSong("701", "1", "01", "A", "0"),
			importAlbumSong("771", "2", "01", "B", "0"),
			importAlbumSong("772", "3", "01", "B", "-200"),
			importAlbumSong("790", "4", "01", "B", "0"),
			importAlbumSong("773", "1", "02 [Bonus]", "C", "0"),
		}, ",") + `]}`,
		// NetEase limits the account at 779.
		"79": `{"code":200,"album":{"name":"Busy","artists":[{"id":5,"name":"A"}]},"songs":[` + strings.Join([]string{
			importAlbumSong("791", "1", "1", "A", "0"),
			importAlbumSong("779", "2", "1", "A", "0"),
			importAlbumSong("792", "3", "1", "A", "0"),
		}, ",") + `]}`,
		// The session expires at 778.
		"78": `{"code":200,"album":{"name":"Stop","artists":[{"id":5,"name":"A"}]},"songs":[` + strings.Join([]string{
			importAlbumSong("781", "1", "1", "A", "0"),
			importAlbumSong("778", "2", "1", "A", "0"),
			importAlbumSong("782", "3", "1", "A", "0"),
		}, ",") + `]}`,
	}
	return ts, f
}

func waitImport(t *testing.T, ts *testServer, id int64) importJob {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		jobs := decodeBody[struct{ Items []importJob }](t, ts.do("GET", playerAPI+"netease/imports", "")).Items
		for _, j := range jobs {
			if j.ID == id && (j.State == importDone || j.State == importFailed) {
				return j
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("import %d did not finish", id)
	return importJob{}
}

func TestNeteaseAlbumImport(t *testing.T) {
	ts, f := newAlbumImportTest(t)
	ctx := context.Background()
	single := decodeBody[importAnswer](t, ts.do("POST", playerAPI+"netease/songs/701/import", ""))

	preview := decodeBody[albumImportPreview](t, ts.do("GET", playerAPI+"netease/albums/77/import", ""))
	if preview != (albumImportPreview{Title: "Comp", Total: 5, ToImport: 3, InLibrary: 1, Moving: 1, Unplayable: 1}) {
		t.Fatalf("preview: %+v", preview)
	}

	f.downloads = nil
	rec := ts.do("POST", playerAPI+"netease/albums/77/import", "")
	if rec.Code != http.StatusAccepted {
		t.Fatalf("start: %d %s", rec.Code, rec.Body)
	}
	job := waitImport(t, ts, decodeBody[importJob](t, rec).ID)
	if job.State != importDone || job.Total != 5 || job.Done != 5 || job.Created != 2 || job.Moved != 1 || job.Unplayable != 1 ||
		job.Existing != 0 || len(job.Failed) != 1 || job.Failed[0] != (importFailure{Title: "song790", Reason: "不是 MP3"}) {
		t.Fatalf("job: %+v", job)
	}
	// The songs share one cover download.
	covers := 0
	for _, d := range f.downloads {
		if strings.HasSuffix(d, ".jpg") {
			covers++
		}
	}
	if covers != 1 {
		t.Fatalf("downloads: %v", f.downloads)
	}

	album, err := ts.store.ImportTarget(ctx, 77)
	if err != nil {
		t.Fatal(err)
	}
	if album.Title != "Comp" || len(album.Artists) != 1 || album.Artists[0].Name != "群星" || album.Single ||
		album.Year != 2017 || album.Description != "notes" || album.TrackCount != 3 || album.Cover == nil {
		t.Fatalf("album: %+v", album)
	}
	tracks, err := ts.store.Tracks(ctx, store.TrackQuery{AlbumID: album.ID})
	if err != nil {
		t.Fatal(err)
	}
	var places []string
	for _, tr := range tracks.Items {
		places = append(places, tr.Title+":"+itoa(int64(tr.DiscNo))+"/"+itoa(int64(tr.TrackNo))+":"+tr.Artists[0].Name)
		if tr.Cover != nil {
			t.Errorf("%s keeps a cover of its own", tr.Title)
		}
	}
	if !slices.Equal(places, []string{"song701:1/1:A", "song771:1/2:B", "song773:2/1:C"}) {
		t.Fatalf("tracks: %v", places)
	}
	// The single's own album is gone; the imported files carry their place.
	if _, err := ts.store.Album(ctx, single.Track.Album.ID); err == nil {
		t.Fatal("single album kept")
	}
	for _, tr := range tracks.Items {
		if tr.Title != "song773" {
			continue
		}
		audio, err := ts.store.OpenAudio(ctx, tr.ID)
		if err != nil {
			t.Fatal(err)
		}
		stat, _ := audio.Stat()
		info, err := media.Probe(audio, stat.Size())
		audio.Close()
		if err != nil || info.Tags.Track != 1 || info.Tags.Disc != 2 || info.Tags.Album != "Comp" || !slices.Equal(info.Tags.AlbumArtists, []string{"群星"}) {
			t.Fatalf("tags: %v %+v", err, info.Tags)
		}
	}

	// Again: the songs in stay, only the FLAC is left, and nothing moves.
	again := decodeBody[albumImportPreview](t, ts.do("GET", playerAPI+"netease/albums/77/import", ""))
	if again.ToImport != 1 || again.InLibrary != 3 || again.Moving != 0 || again.Target == nil || again.Target.ID != album.ID {
		t.Fatalf("preview again: %+v", again)
	}
	second := waitImport(t, ts, decodeBody[importJob](t, ts.do("POST", playerAPI+"netease/albums/77/import", "")).ID)
	if second.Existing != 3 || second.Created != 0 || len(second.Failed) != 1 {
		t.Fatalf("second: %+v", second)
	}
}

func TestNeteaseAlbumImportStopsWhenSessionExpires(t *testing.T) {
	ts, _ := newAlbumImportTest(t)
	job := waitImport(t, ts, decodeBody[importJob](t, ts.do("POST", playerAPI+"netease/albums/78/import", "")).ID)
	if job.State != importFailed || job.Done != 1 || job.Created != 1 || !strings.Contains(job.Error, "失效") {
		t.Fatalf("job: %+v", job)
	}
	jobs := decodeBody[struct{ Items []importJob }](t, ts.do("GET", playerAPI+"netease/imports", "")).Items
	if len(jobs) != 1 {
		t.Fatalf("jobs: %+v", jobs)
	}
}

func TestNeteaseAlbumImportNeedsTheGroup(t *testing.T) {
	ts, f := newAlbumImportTest(t)
	ts.signIn("outsider", "outsider")
	for _, req := range [][2]string{
		{"GET", "netease/albums/77/import"}, {"POST", "netease/albums/77/import"}, {"GET", "netease/imports"},
	} {
		if rec := ts.do(req[0], playerAPI+req[1], ""); rec.Code != http.StatusForbidden {
			t.Errorf("%s %s: %d", req[0], req[1], rec.Code)
		}
	}
	if len(f.downloads) != 0 {
		t.Fatalf("downloaded: %v", f.downloads)
	}
}

func TestImportQueueKeepsOneJobPerAlbum(t *testing.T) {
	var q importQueue
	album := netease.ImportAlbum{ID: 7, Title: "A"}
	first, start := q.add(1, album)
	again, startAgain := q.add(1, album)
	other, _ := q.add(2, album)
	if !start || startAgain || again.ID != first.ID || other.ID == first.ID {
		t.Fatalf("first %d %v, again %d %v, other %d", first.ID, start, again.ID, startAgain, other.ID)
	}
	if got := q.of(1); len(got) != 1 || got[0].ID != first.ID {
		t.Fatalf("of: %+v", got)
	}
}

func TestNeteaseAlbumImportStopsWhenThrottled(t *testing.T) {
	ts, f := newAlbumImportTest(t)
	job := waitImport(t, ts, decodeBody[importJob](t, ts.do("POST", playerAPI+"netease/albums/79/import", "")).ID)
	if job.State != importFailed || job.Done != 1 || job.Created != 1 || len(job.Failed) != 0 || !strings.Contains(job.Error, "频繁") {
		t.Fatalf("job: %+v", job)
	}
	// Nothing after the refusal was asked for.
	for _, d := range f.downloads {
		if strings.Contains(d, "792") {
			t.Fatalf("downloads: %v", f.downloads)
		}
	}
}

// Songs fetched from NetEase are albumImportGap apart, counted from the last
// fetch, so songs that fetch nothing add no time of their own (Eric,
// 2026-11-07).
func TestAlbumImportWaitsBetweenSongs(t *testing.T) {
	ts, f := newAlbumImportTest(t)
	albumImportGap = 500 * time.Millisecond
	decodeBody[importAnswer](t, ts.do("POST", playerAPI+"netease/songs/701/import", ""))
	f.downloads, f.downloadedAt = nil, nil
	start := time.Now()
	waitImport(t, ts, decodeBody[importJob](t, ts.do("POST", playerAPI+"netease/albums/77/import", "")).ID)
	elapsed := time.Since(start)
	// 701 moves and 772 is grey; 771, 790 (refused as FLAC) and 773 are
	// fetched, so two waits.
	var audio []time.Time
	for i, d := range f.downloads {
		if strings.HasSuffix(d, ".mp3") {
			audio = append(audio, f.downloadedAt[i])
		}
	}
	if len(audio) != 2 || audio[1].Sub(audio[0]) < 2*albumImportGap {
		t.Fatalf("audio downloads %v at %v", f.downloads, audio)
	}
	// Waiting a full gap before every song would take five gaps.
	if elapsed > 3*albumImportGap+albumImportGap/2 {
		t.Fatalf("took %v", elapsed)
	}
}
