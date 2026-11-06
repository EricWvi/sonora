package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"sonora/internal/store"
)

const playerAPI = PlayerAPIPrefix

// signIn starts a player session for a listener, as a completed OIDC
// callback would, and sends it on later requests.
func (ts *testServer) signIn(subject, name string) {
	ts.t.Helper()
	ctx := context.Background()
	u, err := ts.store.UpsertUser(ctx, store.Identity{Issuer: "https://auth.test", Subject: subject, Name: name})
	if err != nil {
		ts.t.Fatal(err)
	}
	if ts.listener, err = ts.store.CreateSession(ctx, u.ID); err != nil {
		ts.t.Fatal(err)
	}
}

func TestPlayerAPIRequiresListenerSession(t *testing.T) {
	ts := newTest(t)
	ts.setup()
	// An admin session does not open the player API.
	if rec := ts.do("GET", playerAPI+"tracks", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("admin session on player API: %d", rec.Code)
	}
	ts.cookie = ""
	ts.signIn("alice", "Alice")
	if rec := ts.do("GET", api+"tracks", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("listener session on admin API: %d", rec.Code)
	}
	rec := ts.do("GET", playerAPI+"me", "")
	if rec.Code != http.StatusOK || decodeBody[map[string]any](t, rec)["name"] != "Alice" {
		t.Fatalf("me: %d %s", rec.Code, rec.Body)
	}
	if rec := ts.do("GET", playerAPI+"nope", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown endpoint: %d", rec.Code)
	}

	// Signing out ends the session.
	if rec := ts.do("POST", "/auth/logout", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("logout: %d", rec.Code)
	}
	if rec := ts.do("GET", playerAPI+"me", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("after logout: %d", rec.Code)
	}
}

func TestPlayerCatalog(t *testing.T) {
	ts := newTest(t)
	ts.setup()
	tr := decodeBody[store.IngestResult](t, ts.upload("a.flac", sample(t, "tagged.flac"), "[00:01.00]歌词")).Track
	rec := ts.do("PUT", api+"artists/"+itoa(tr.Artists[0].ID)+"/image", string(sample(t, "cover.jpg")))
	if rec.Code != http.StatusOK || decodeBody[store.Artist](t, rec).Image == nil {
		t.Fatalf("artist image: %d %s", rec.Code, rec.Body)
	}
	artistIn, _ := json.Marshal(map[string]any{"name": tr.Artists[0].Name, "description": "歌手简介"})
	if rec := ts.do("PUT", api+"artists/"+itoa(tr.Artists[0].ID), string(artistIn)); rec.Code != http.StatusOK {
		t.Fatalf("artist description: %d %s", rec.Code, rec.Body)
	}
	albumArtists := []string{}
	for _, a := range tr.Album.Artists {
		albumArtists = append(albumArtists, a.Name)
	}
	albumIn, _ := json.Marshal(map[string]any{"title": tr.Album.Title, "artists": albumArtists, "year": tr.Year, "description": "专辑简介"})
	if rec := ts.do("PUT", api+"albums/"+itoa(tr.Album.ID), string(albumIn)); rec.Code != http.StatusOK {
		t.Fatalf("album description: %d %s", rec.Code, rec.Body)
	}
	ts.cookie = ""
	ts.signIn("alice", "Alice")

	rec = ts.do("GET", playerAPI+"tracks", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("tracks: %d %s", rec.Code, rec.Body)
	}
	for _, field := range []string{"originalName", "size", "createdAt", "lyrics\"", "neteaseId"} {
		if strings.Contains(rec.Body.String(), `"`+field) {
			t.Fatalf("player track exposes %s: %s", field, rec.Body)
		}
	}
	tracks := decodeBody[store.Page[playerTrack]](t, rec)
	if tracks.Total != 1 || tracks.Items[0].ID != tr.ID || !tracks.Items[0].HasLyrics {
		t.Fatalf("tracks = %+v", tracks)
	}
	// The track shows its album's cover when it has none of its own.
	got := tracks.Items[0]
	if tr.Cover != nil || got.Cover == nil || got.Cover.Name != tr.Album.Cover.Name || got.Album.ID != tr.Album.ID {
		t.Fatalf("track cover = %+v, album = %+v", got.Cover, got.Album)
	}

	byAlbum := decodeBody[store.Page[playerTrack]](t, ts.do("GET", playerAPI+"tracks?albumId="+itoa(tr.Album.ID), ""))
	byArtist := decodeBody[store.Page[playerTrack]](t, ts.do("GET", playerAPI+"tracks?artistId="+itoa(tr.Artists[0].ID), ""))
	search := decodeBody[store.Page[playerTrack]](t, ts.do("GET", playerAPI+"tracks?q="+url.QueryEscape(tr.Title), ""))
	none := decodeBody[store.Page[playerTrack]](t, ts.do("GET", playerAPI+"tracks?q=zzzz", ""))
	if byAlbum.Total != 1 || byArtist.Total != 1 || search.Total != 1 || none.Total != 0 {
		t.Fatalf("filters: album %d, artist %d, search %d, none %d", byAlbum.Total, byArtist.Total, search.Total, none.Total)
	}

	album := decodeBody[playerAlbum](t, ts.do("GET", playerAPI+"albums/"+itoa(tr.Album.ID), ""))
	if album.TrackCount != 1 || album.Cover == nil || album.Title != tr.Album.Title || album.Description != "专辑简介" || album.OnlineID != nil {
		t.Fatalf("album = %+v", album)
	}
	albums := decodeBody[store.Page[playerAlbum]](t, ts.do("GET", playerAPI+"albums?artistId="+itoa(tr.Artists[0].ID), ""))
	if albums.Total != 1 {
		t.Fatalf("albums = %+v", albums)
	}
	artist := decodeBody[playerArtist](t, ts.do("GET", playerAPI+"artists/"+itoa(tr.Artists[0].ID), ""))
	artists := decodeBody[store.Page[playerArtist]](t, ts.do("GET", playerAPI+"artists", ""))
	if artist.TrackCount != 1 || artist.AlbumCount != 1 || artist.Image == nil || artist.Description != "歌手简介" || artist.OnlineID != nil || artists.Total != 1 ||
		artists.Items[0].Image == nil {
		t.Fatalf("artist = %+v, artists = %+v", artist, artists)
	}
	if rec := ts.do("GET", playerAPI+"covers/"+artist.Image.Name+"?size=96", ""); rec.Code != http.StatusOK {
		t.Fatalf("artist image: %d", rec.Code)
	}

	// Lists take sort and order; unknown values are rejected.
	for path, want := range map[string]int{
		"tracks?sort=title&order=desc":      http.StatusOK,
		"tracks?sort=size":                  http.StatusBadRequest,
		"tracks?sort=title&order=sideways":  http.StatusBadRequest,
		"albums?sort=createTime&order=desc": http.StatusOK,
		"albums?sort=duration":              http.StatusBadRequest,
		"likes?sort=artist":                 http.StatusOK,
	} {
		if rec := ts.do("GET", playerAPI+path, ""); rec.Code != want {
			t.Errorf("%s: %d %s", path, rec.Code, rec.Body)
		}
	}
	if rec := ts.do("GET", playerAPI+"albums/999", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("missing album: %d", rec.Code)
	}

	lyrics := decodeBody[map[string]string](t, ts.do("GET", playerAPI+"tracks/"+itoa(tr.ID)+"/lyrics", ""))
	if lyrics["lyrics"] != "[00:01.00]歌词" || lyrics["format"] != "lrc" {
		t.Fatalf("lyrics = %q", lyrics)
	}

	rec = ts.do("GET", playerAPI+"covers/"+got.Cover.Name+"?size=300", "")
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatalf("cover: %d %s", rec.Code, rec.Header())
	}

	req, _ := http.NewRequest("GET", playerAPI+"tracks/"+itoa(tr.ID)+"/audio", nil)
	req.Header.Set("Range", "bytes=0-3")
	rec = ts.send(req)
	if rec.Code != http.StatusPartialContent || rec.Body.String() != "fLaC" ||
		rec.Header().Get("Content-Type") != "audio/flac" {
		t.Fatalf("audio range: %d %q %s", rec.Code, rec.Body.String(), rec.Header())
	}
}

func TestListenerDataIsPerUser(t *testing.T) {
	ts := newTest(t)
	ts.setup()
	a := decodeBody[store.IngestResult](t, ts.upload("a.flac", sample(t, "tagged.flac"), "")).Track
	b := decodeBody[store.IngestResult](t, ts.upload("b.mp3", sample(t, "tagged.mp3"), "")).Track
	ts.cookie = ""
	ts.signIn("alice", "Alice")
	alice := ts.listener

	if rec := ts.do("PUT", playerAPI+"likes/"+itoa(a.ID), ""); rec.Code != http.StatusNoContent {
		t.Fatalf("like: %d %s", rec.Code, rec.Body)
	}
	if rec := ts.do("PUT", playerAPI+"likes/999", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("like missing: %d", rec.Code)
	}
	if rec := ts.do("POST", playerAPI+"history", `{"trackId":`+itoa(b.ID)+`}`); rec.Code != http.StatusNoContent {
		t.Fatalf("record play: %d %s", rec.Code, rec.Body)
	}
	if rec := ts.do("POST", playerAPI+"history", `{}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("record play without track: %d", rec.Code)
	}
	rec := ts.do("POST", playerAPI+"playlists", `{"name":"跑步"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create playlist: %d %s", rec.Code, rec.Body)
	}
	pl := decodeBody[store.Playlist](t, rec)
	rec = ts.do("POST", playerAPI+"playlists/"+itoa(pl.ID)+"/tracks", `{"trackIds":[`+itoa(a.ID)+`,`+itoa(b.ID)+`]}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"added":2`) {
		t.Fatalf("add to playlist: %d %s", rec.Code, rec.Body)
	}
	rec = ts.do("PUT", playerAPI+"playlists/"+itoa(pl.ID)+"/order", `{"trackIds":[`+itoa(b.ID)+`]}`)
	if rec.Code != http.StatusConflict || errorCode(t, rec) != "conflict" {
		t.Fatalf("stale reorder: %d %s", rec.Code, rec.Body)
	}
	rec = ts.do("PUT", playerAPI+"playlists/"+itoa(pl.ID)+"/order", `{"trackIds":[`+itoa(b.ID)+`,`+itoa(a.ID)+`]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("reorder: %d %s", rec.Code, rec.Body)
	}
	tracks := decodeBody[store.Page[playerTrack]](t, ts.do("GET", playerAPI+"playlists/"+itoa(pl.ID)+"/tracks", ""))
	if tracks.Total != 2 || tracks.Items[0].ID != b.ID {
		t.Fatalf("playlist tracks = %+v", tracks)
	}
	req := httptest.NewRequest("POST", playerAPI+"playlists", strings.NewReader(`{"name":"x"}`))
	req.Header.Set("Origin", "https://evil.test")
	if rec := ts.send(req); rec.Code != http.StatusForbidden {
		t.Fatalf("cross-origin create: %d", rec.Code)
	}

	// Bob sees none of it and cannot touch Alice's playlist.
	ts.signIn("bob", "Bob")
	for path, want := range map[string]string{
		"likes":     `"total":0`,
		"likes/ids": `"ids":[]`,
		"history":   `"total":0`,
		"playlists": `"total":0`,
	} {
		if rec := ts.do("GET", playerAPI+path, ""); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), want) {
			t.Errorf("bob %s: %d %s", path, rec.Code, rec.Body)
		}
	}
	for _, c := range []struct{ method, path, body string }{
		{"GET", "playlists/" + itoa(pl.ID), ""},
		{"GET", "playlists/" + itoa(pl.ID) + "/tracks", ""},
		{"POST", "playlists/" + itoa(pl.ID) + "/tracks", `{"trackIds":[` + itoa(a.ID) + `]}`},
		{"DELETE", "playlists/" + itoa(pl.ID), ""},
	} {
		if rec := ts.do(c.method, playerAPI+c.path, c.body); rec.Code != http.StatusNotFound {
			t.Errorf("bob %s %s: %d", c.method, c.path, rec.Code)
		}
	}

	ts.listener = alice
	ids := decodeBody[map[string][]int64](t, ts.do("GET", playerAPI+"likes/ids", ""))["ids"]
	history := decodeBody[store.Page[playerHistoryEntry]](t, ts.do("GET", playerAPI+"history", ""))
	if len(ids) != 1 || ids[0] != a.ID || history.Total != 1 || history.Items[0].Track.ID != b.ID {
		t.Fatalf("alice likes %v, history %+v", ids, history)
	}
	for _, path := range []string{"likes/" + itoa(a.ID), "history/" + itoa(b.ID), "playlists/" + itoa(pl.ID) + "/tracks/" + itoa(a.ID)} {
		if rec := ts.do("DELETE", playerAPI+path, ""); rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
			t.Fatalf("delete %s: %d %s", path, rec.Code, rec.Body)
		}
	}
	if rec := ts.do("DELETE", playerAPI+"playlists/"+itoa(pl.ID), ""); rec.Code != http.StatusNoContent {
		t.Fatalf("delete playlist: %d", rec.Code)
	}
}

func TestSettingsSync(t *testing.T) {
	ts := newTest(t)
	ts.signIn("alice", "Alice")
	alice := ts.listener
	if rec := ts.do("GET", playerAPI+"settings", ""); rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "{}" {
		t.Fatalf("unsaved: %d %s", rec.Code, rec.Body)
	}
	if rec := ts.do("PUT", playerAPI+"settings", `{"themeMode":"dark","lyricFontSize":46}`); rec.Code != http.StatusNoContent {
		t.Fatalf("save: %d %s", rec.Code, rec.Body)
	}
	if rec := ts.do("PUT", playerAPI+"settings", `["not","an","object"]`); rec.Code != http.StatusBadRequest {
		t.Fatalf("save array: %d", rec.Code)
	}
	big := `{"x":"` + strings.Repeat("a", store.MaxSettingsBytes) + `"}`
	if rec := ts.do("PUT", playerAPI+"settings", big); rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("save too large: %d", rec.Code)
	}
	ts.signIn("bob", "Bob")
	if rec := ts.do("GET", playerAPI+"settings", ""); strings.TrimSpace(rec.Body.String()) != "{}" {
		t.Fatalf("bob sees %s", rec.Body)
	}
	ts.listener = alice
	got := decodeBody[map[string]any](t, ts.do("GET", playerAPI+"settings", ""))
	if got["themeMode"] != "dark" || got["lyricFontSize"] != float64(46) {
		t.Fatalf("alice settings = %v", got)
	}
}
