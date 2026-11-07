package server

import (
	"context"
	"net/http"
	"testing"

	"sonora/internal/store"
)

type fillAnswer struct {
	Filled []string `json:"filled"`
	Album  struct {
		Description string `json:"description"`
	} `json:"album"`
	Artist struct {
		Description string `json:"description"`
		Image       *struct {
			Name string `json:"name"`
		} `json:"image"`
	} `json:"artist"`
	Track struct {
		HasLyrics bool `json:"hasLyrics"`
	} `json:"track"`
}

// link edits a library record through the admin API.
func (ts *testServer) link(kind string, id int64, body string) {
	ts.t.Helper()
	if rec := ts.do("PUT", api+kind+"/"+itoa(id), body); rec.Code != http.StatusOK {
		ts.t.Fatalf("edit %s: %d %s", kind, rec.Code, rec.Body)
	}
}

// fillFixture uploads a track and links its song, album and artist.
func fillFixture(t *testing.T, ts *testServer) store.TrackDetail {
	t.Helper()
	tr := decodeBody[store.IngestResult](t, ts.upload("a.mp3", sample(t, "tagged.mp3"), "")).Track
	ctx := context.Background()
	if err := ts.store.SetNeteaseID(ctx, tr.ID, 701); err != nil {
		t.Fatal(err)
	}
	// The sample carries lyrics in its tags.
	if rec := ts.do("DELETE", api+"tracks/"+itoa(tr.ID)+"/lyrics", ""); rec.Code >= 300 {
		t.Fatalf("clear lyrics: %d %s", rec.Code, rec.Body)
	}
	ts.link("albums", tr.Album.ID, `{"title":"`+tr.Album.Title+`","neteaseId":900}`)
	ts.link("artists", tr.Artists[0].ID, `{"name":"`+tr.Artists[0].Name+`","neteaseId":5}`)
	return tr
}

func TestFillOnlyWhatIsEmpty(t *testing.T) {
	ts, f := newImportTest(t)
	f.importAlbums = map[string]string{"900": `{"code":200,"album":{"name":"Al","description":" 专辑简介 "},"songs":[]}`}
	tr := fillFixture(t, ts)
	post := func(path string) fillAnswer {
		t.Helper()
		rec := ts.do("POST", playerAPI+path, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, rec.Code, rec.Body)
		}
		return decodeBody[fillAnswer](t, rec)
	}
	albumPath := "albums/" + itoa(tr.Album.ID) + "/fill"
	got := post(albumPath)
	if got.Album.Description != "专辑简介" || len(got.Filled) != 1 {
		t.Fatalf("album %+v", got)
	}
	if again := post(albumPath); len(again.Filled) != 0 || again.Album.Description != "专辑简介" {
		t.Fatalf("album again %+v", again)
	}

	artistPath := "artists/" + itoa(tr.Artists[0].ID) + "/fill"
	got = post(artistPath)
	if got.Artist.Description != "简介5" || got.Artist.Image == nil || len(got.Filled) != 2 {
		t.Fatalf("artist %+v", got)
	}
	calls := len(f.artistDetails)
	if again := post(artistPath); len(again.Filled) != 0 || len(f.artistDetails) != calls {
		t.Fatalf("artist asked again: %+v, %d calls", again, len(f.artistDetails))
	}

	trackPath := "tracks/" + itoa(tr.ID) + "/fill"
	got = post(trackPath)
	if !got.Track.HasLyrics || len(got.Filled) != 1 {
		t.Fatalf("track %+v", got)
	}
	detail, err := ts.store.Track(context.Background(), tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.LyricsFormat != store.TTML {
		t.Fatalf("lyrics format %q", detail.LyricsFormat)
	}
	if again := post(trackPath); len(again.Filled) != 0 {
		t.Fatalf("track again %+v", again)
	}
}

func TestFillKeepsWhatTheAdminWrote(t *testing.T) {
	ts, _ := newImportTest(t)
	tr := fillFixture(t, ts)
	ts.link("artists", tr.Artists[0].ID, `{"name":"`+tr.Artists[0].Name+`","description":"管理员写的"}`)
	got := decodeBody[fillAnswer](t, ts.do("POST", playerAPI+"artists/"+itoa(tr.Artists[0].ID)+"/fill", ""))
	if got.Artist.Description != "管理员写的" || got.Artist.Image == nil || len(got.Filled) != 1 || got.Filled[0] != store.FilledImage {
		t.Fatalf("artist %+v", got)
	}
}

func TestFillWhenNetEaseHasNothing(t *testing.T) {
	ts, _ := newImportTest(t)
	tr := fillFixture(t, ts)
	// The fake's default /album has no description.
	ts.link("albums", tr.Album.ID, `{"title":"`+tr.Album.Title+`","neteaseId":901}`)
	got := decodeBody[fillAnswer](t, ts.do("POST", playerAPI+"albums/"+itoa(tr.Album.ID)+"/fill", ""))
	if len(got.Filled) != 0 || got.Album.Description != "" {
		t.Fatalf("album %+v", got)
	}
}

func TestFillNeedsTheGroupALinkAndASession(t *testing.T) {
	ts, f := newImportTest(t)
	tr := fillFixture(t, ts)
	path := playerAPI + "albums/" + itoa(tr.Album.ID) + "/fill"
	importer := ts.listener

	ts.signIn("outsider", "Outsider")
	signInNetease(t, ts, f)
	if rec := ts.do("POST", path, ""); rec.Code != http.StatusForbidden || errorCode(t, rec) != "import_forbidden" {
		t.Fatalf("outsider: %d %s", rec.Code, rec.Body)
	}

	ts.listener = importer
	if rec := ts.do("POST", playerAPI+"netease/logout", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("sign out: %d %s", rec.Code, rec.Body)
	}
	if rec := ts.do("POST", path, ""); rec.Code != http.StatusForbidden || errorCode(t, rec) != "netease_signed_out" {
		t.Fatalf("signed out: %d %s", rec.Code, rec.Body)
	}
	signInNetease(t, ts, f)

	ts.link("albums", tr.Album.ID, `{"title":"`+tr.Album.Title+`","neteaseId":null}`)
	if rec := ts.do("POST", path, ""); rec.Code != http.StatusConflict {
		t.Fatalf("not linked: %d %s", rec.Code, rec.Body)
	}
	if rec := ts.do("POST", playerAPI+"albums/999999/fill", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("missing: %d %s", rec.Code, rec.Body)
	}
}
