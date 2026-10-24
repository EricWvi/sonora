package store

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
)

func newUser(t *testing.T, s *testStore, subject string) int64 {
	t.Helper()
	u, err := s.UpsertUser(context.Background(), Identity{Issuer: "https://auth.test", Subject: subject, Name: subject})
	if err != nil {
		t.Fatal(err)
	}
	return u.ID
}

// fakeTracks inserts n tracks without audio files, for tests that only
// need rows.
func fakeTracks(t *testing.T, s *testStore, n int) []int64 {
	t.Helper()
	var ids []int64
	for i := range n {
		var id int64
		err := s.db.QueryRow(`
			INSERT INTO tracks (title, duration_ms, format, size, sha256, file, original_name)
			VALUES (?, 1000, 'mp3', 1, ?, 'x', 'x.mp3') RETURNING id`,
			fmt.Sprintf("fake %d", i), fmt.Sprintf("%d-%d", len(ids), i)+t.Name()).Scan(&id)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	return ids
}

func trackIDs(items []Track) []int64 {
	ids := []int64{}
	for _, t := range items {
		ids = append(ids, t.ID)
	}
	return ids
}

func TestLikesAreIdempotentAndPerUser(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	alice, bob := newUser(t, s, "alice"), newUser(t, s, "bob")
	tracks := fakeTracks(t, s, 3)

	for _, id := range []int64{tracks[0], tracks[2], tracks[0]} {
		if err := s.Like(ctx, alice, id); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Like(ctx, alice, 9999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("like missing track: %v", err)
	}
	ids, _ := s.LikedTrackIDs(ctx, alice)
	if !slices.Equal(ids, []int64{tracks[2], tracks[0]}) {
		t.Fatalf("alice likes = %v", ids)
	}
	page, err := s.LikedTracks(ctx, alice, Sort{}, Window{})
	if err != nil || page.Total != 2 || !slices.Equal(trackIDs(page.Items), ids) {
		t.Fatalf("liked tracks = %+v, %v", page, err)
	}
	if ids, _ := s.LikedTrackIDs(ctx, bob); len(ids) != 0 {
		t.Fatalf("bob sees alice's likes: %v", ids)
	}

	if err := s.Unlike(ctx, alice, tracks[2]); err != nil {
		t.Fatal(err)
	}
	if err := s.Unlike(ctx, alice, tracks[2]); err != nil {
		t.Fatalf("unlike twice: %v", err)
	}
	if ids, _ := s.LikedTrackIDs(ctx, alice); !slices.Equal(ids, []int64{tracks[0]}) {
		t.Fatalf("after unlike = %v", ids)
	}
}

func TestHistoryMovesReplaysToTopAndIsCapped(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	alice := newUser(t, s, "alice")
	tracks := fakeTracks(t, s, MaxHistory+2)

	for _, id := range tracks {
		if err := s.RecordPlay(ctx, alice, id); err != nil {
			t.Fatal(err)
		}
	}
	// Replaying an old track that is still kept moves it to the top.
	if err := s.RecordPlay(ctx, alice, tracks[5]); err != nil {
		t.Fatal(err)
	}
	page, err := s.History(ctx, alice, Window{Limit: 3})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != MaxHistory {
		t.Fatalf("history total = %d, want %d", page.Total, MaxHistory)
	}
	got := []int64{page.Items[0].Track.ID, page.Items[1].Track.ID, page.Items[2].Track.ID}
	if want := []int64{tracks[5], tracks[len(tracks)-1], tracks[len(tracks)-2]}; !slices.Equal(got, want) {
		t.Fatalf("history head = %v, want %v", got, want)
	}
	if page.Items[0].PlayedAt == "" {
		t.Fatal("playedAt missing")
	}
	// The oldest plays fell off.
	var kept int
	s.db.QueryRow(`SELECT COUNT(*) FROM play_history WHERE track_id IN (?, ?)`, tracks[0], tracks[1]).Scan(&kept)
	if kept != 0 {
		t.Fatalf("oldest entries kept: %d", kept)
	}

	if err := s.RemoveFromHistory(ctx, alice, tracks[5]); err != nil {
		t.Fatal(err)
	}
	if p, _ := s.History(ctx, alice, Window{Limit: 1}); p.Items[0].Track.ID == tracks[5] || p.Total != MaxHistory-1 {
		t.Fatalf("after remove: %+v", p)
	}
	if err := s.ClearHistory(ctx, alice); err != nil {
		t.Fatal(err)
	}
	if p, _ := s.History(ctx, alice, Window{}); p.Total != 0 {
		t.Fatalf("after clear: %d", p.Total)
	}
}

func TestPlaylists(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	alice, bob := newUser(t, s, "alice"), newUser(t, s, "bob")
	a := ingest(t, s, sample(t, "tagged.flac"), "a.flac", nil).Track
	tracks := append([]int64{a.ID}, fakeTracks(t, s, 2)...)

	if _, err := s.CreatePlaylist(ctx, alice, PlaylistInput{Name: "  "}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("empty name: %v", err)
	}
	p, err := s.CreatePlaylist(ctx, alice, PlaylistInput{Name: " 跑步 ", Description: "快歌"})
	if err != nil || p.Name != "跑步" || p.TrackCount != 0 || p.Cover != nil {
		t.Fatalf("create = %+v, %v", p, err)
	}

	added, p, err := s.AddToPlaylist(ctx, alice, p.ID, []int64{tracks[1], tracks[0], tracks[1]})
	if err != nil || added != 2 || p.TrackCount != 2 || p.DurationMs != 1000+a.DurationMs {
		t.Fatalf("add = %d, %+v, %v", added, p, err)
	}
	// tracks[1] comes first but has no cover, so the next one's shows.
	if p.Cover == nil || p.Cover.Name != a.Album.Cover.Name {
		t.Fatalf("cover = %+v", p.Cover)
	}
	added, p, _ = s.AddToPlaylist(ctx, alice, p.ID, []int64{tracks[0], tracks[2]})
	if added != 1 || p.TrackCount != 3 {
		t.Fatalf("add again = %d, %+v", added, p)
	}
	if _, _, err := s.AddToPlaylist(ctx, alice, p.ID, []int64{9999}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("add missing track: %v", err)
	}
	page, _ := s.PlaylistTracks(ctx, alice, p.ID, Sort{}, Window{})
	if !slices.Equal(trackIDs(page.Items), []int64{tracks[1], tracks[0], tracks[2]}) {
		t.Fatalf("order = %v", trackIDs(page.Items))
	}

	// Reordering needs exactly the current tracks.
	if _, err := s.ReorderPlaylist(ctx, alice, p.ID, []int64{tracks[0], tracks[1]}); !errors.Is(err, ErrConflict) {
		t.Fatalf("reorder subset: %v", err)
	}
	if _, err := s.ReorderPlaylist(ctx, alice, p.ID, []int64{tracks[0], tracks[1], 9999}); !errors.Is(err, ErrConflict) {
		t.Fatalf("reorder foreign: %v", err)
	}
	p, err = s.ReorderPlaylist(ctx, alice, p.ID, []int64{tracks[0], tracks[2], tracks[1]})
	if err != nil || p.TrackCount != 3 {
		t.Fatalf("reorder = %+v, %v", p, err)
	}
	page, _ = s.PlaylistTracks(ctx, alice, p.ID, Sort{}, Window{})
	if !slices.Equal(trackIDs(page.Items), []int64{tracks[0], tracks[2], tracks[1]}) {
		t.Fatalf("reordered = %v", trackIDs(page.Items))
	}

	p, err = s.RemoveFromPlaylist(ctx, alice, p.ID, tracks[2])
	if err != nil || p.TrackCount != 2 {
		t.Fatalf("remove = %+v, %v", p, err)
	}
	p, err = s.UpdatePlaylist(ctx, alice, p.ID, PlaylistInput{Name: "夜跑"})
	if err != nil || p.Name != "夜跑" || p.Description != "" {
		t.Fatalf("update = %+v, %v", p, err)
	}

	// Bob can neither see nor change Alice's playlist.
	if list, _ := s.Playlists(ctx, bob, "", Window{}); list.Total != 0 {
		t.Fatalf("bob lists %d playlists", list.Total)
	}
	for name, err := range map[string]error{
		"get":    func() error { _, err := s.Playlist(ctx, bob, p.ID); return err }(),
		"tracks": func() error { _, err := s.PlaylistTracks(ctx, bob, p.ID, Sort{}, Window{}); return err }(),
		"add":    func() error { _, _, err := s.AddToPlaylist(ctx, bob, p.ID, tracks[:1]); return err }(),
		"remove": func() error { _, err := s.RemoveFromPlaylist(ctx, bob, p.ID, tracks[0]); return err }(),
		"order":  func() error { _, err := s.ReorderPlaylist(ctx, bob, p.ID, tracks[:2]); return err }(),
		"update": func() error { _, err := s.UpdatePlaylist(ctx, bob, p.ID, PlaylistInput{Name: "x"}); return err }(),
		"delete": s.DeletePlaylist(ctx, bob, p.ID),
	} {
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("bob %s: %v", name, err)
		}
	}
	if list, _ := s.Playlists(ctx, alice, "夜", Window{}); list.Total != 1 {
		t.Fatalf("search = %+v", list)
	}
	if err := s.DeletePlaylist(ctx, alice, p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Playlist(ctx, alice, p.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("after delete: %v", err)
	}
}

func TestDeletingTrackRemovesItFromListenerData(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	alice := newUser(t, s, "alice")
	tr := ingest(t, s, sample(t, "tagged.mp3"), "a.mp3", nil).Track
	keep := fakeTracks(t, s, 1)[0]

	s.Like(ctx, alice, tr.ID)
	s.RecordPlay(ctx, alice, tr.ID)
	p, _ := s.CreatePlaylist(ctx, alice, PlaylistInput{Name: "p"})
	s.AddToPlaylist(ctx, alice, p.ID, []int64{tr.ID, keep})

	if err := s.DeleteTrack(ctx, tr.ID); err != nil {
		t.Fatal(err)
	}
	ids, _ := s.LikedTrackIDs(ctx, alice)
	history, _ := s.History(ctx, alice, Window{})
	p, _ = s.Playlist(ctx, alice, p.ID)
	if len(ids) != 0 || history.Total != 0 || p.TrackCount != 1 {
		t.Fatalf("likes %v, history %d, playlist %d tracks", ids, history.Total, p.TrackCount)
	}
}
