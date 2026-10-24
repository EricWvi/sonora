package server

import (
	"context"
	"net/http"
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
	if rec.Code != http.StatusOK || decodeBody[map[string]string](t, rec)["name"] != "Alice" {
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
	ts.cookie = ""
	ts.signIn("alice", "Alice")

	rec := ts.do("GET", playerAPI+"tracks", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("tracks: %d %s", rec.Code, rec.Body)
	}
	for _, field := range []string{"originalName", "size", "createdAt", "lyrics\""} {
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
	if album.TrackCount != 1 || album.Cover == nil || album.Title != tr.Album.Title {
		t.Fatalf("album = %+v", album)
	}
	albums := decodeBody[store.Page[playerAlbum]](t, ts.do("GET", playerAPI+"albums?artistId="+itoa(tr.Artists[0].ID), ""))
	if albums.Total != 1 {
		t.Fatalf("albums = %+v", albums)
	}
	artist := decodeBody[playerArtist](t, ts.do("GET", playerAPI+"artists/"+itoa(tr.Artists[0].ID), ""))
	artists := decodeBody[store.Page[playerArtist]](t, ts.do("GET", playerAPI+"artists", ""))
	if artist.TrackCount != 1 || artist.AlbumCount != 1 || artists.Total != 1 {
		t.Fatalf("artist = %+v, artists = %+v", artist, artists)
	}
	if rec := ts.do("GET", playerAPI+"albums/999", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("missing album: %d", rec.Code)
	}

	lyrics := decodeBody[map[string]string](t, ts.do("GET", playerAPI+"tracks/"+itoa(tr.ID)+"/lyrics", ""))
	if lyrics["lyrics"] != "[00:01.00]歌词" {
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
