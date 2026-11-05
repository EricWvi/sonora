package store

import (
	"context"
	"slices"
	"testing"
)

func TestSearchOrdersByRelevance(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	ids := fakeTracks(t, s, 6)
	meta := []struct{ title, artist, album string }{
		{"爱", "阿桑", "单曲"},
		{"我爱你", "卢广仲", "我爱你"},
		{"晴天", "爱乐团", "叶惠美"},
		{"x", "某人", "爱"},
		{"爱情转移", "陈奕迅", "认了吧"},
		{"无关", "阿爱", "无关"},
	}
	for i, m := range meta {
		if _, err := s.UpdateTrack(ctx, ids[i], TrackInput{Title: m.title, Artists: []string{m.artist}, Album: m.album}); err != nil {
			t.Fatal(err)
		}
	}
	exact, contains, artistPrefix, albumExact, prefix, artistContains := ids[0], ids[1], ids[2], ids[3], ids[4], ids[5]

	// Titles first: exact, then prefix, then contains; then artist and
	// album names the same way. Ties keep the default order.
	page, err := s.Tracks(ctx, TrackQuery{Q: "爱"})
	if want := []int64{exact, prefix, contains, albumExact, artistPrefix, artistContains}; err != nil || !slices.Equal(trackIDs(page.Items), want) {
		t.Fatalf("tracks: %v, %v; want %v", trackIDs(page.Items), err, want)
	}
	if page.Total != 6 {
		t.Fatalf("total = %d", page.Total)
	}
	// Paging follows the same order.
	page, _ = s.Tracks(ctx, TrackQuery{Q: "爱", Window: Window{Offset: 2, Limit: 2}})
	if want := []int64{contains, albumExact}; !slices.Equal(trackIDs(page.Items), want) {
		t.Fatalf("page 2: %v; want %v", trackIDs(page.Items), want)
	}
	// An explicit sort still wins.
	page, _ = s.Tracks(ctx, TrackQuery{Q: "爱", Sort: Sort{Field: SortCreateTime}})
	if !slices.Equal(trackIDs(page.Items), ids) {
		t.Fatalf("by create time: %v", trackIDs(page.Items))
	}
	// LIKE wildcards in the query are literal here too.
	if page, _ := s.Tracks(ctx, TrackQuery{Q: "%"}); page.Total != 0 {
		t.Fatalf("q=%%: %d", page.Total)
	}

	// Albums: by title, then by artist name. By title alone 无关 (wu)
	// would come before 叶惠美 (ye).
	albums, err := s.Albums(ctx, AlbumQuery{Q: "爱"})
	var titles []string
	for _, a := range albums.Items {
		titles = append(titles, a.Title)
	}
	if want := []string{"爱", "我爱你", "叶惠美", "无关"}; err != nil || !slices.Equal(titles, want) {
		t.Fatalf("albums: %v, %v; want %v", titles, err, want)
	}

	// Artists: by name alone 阿爱 (a ai) sorts first.
	artists, err := s.Artists(ctx, ArtistQuery{Q: "爱"})
	var names []string
	for _, a := range artists.Items {
		names = append(names, a.Name)
	}
	if want := []string{"爱乐团", "阿爱"}; err != nil || !slices.Equal(names, want) {
		t.Fatalf("artists: %v, %v; want %v", names, err, want)
	}

	// Playlists: newest first unless the name matches better.
	alice := newUser(t, s, "alice")
	for _, name := range []string{"爱", "我爱", "爱的歌"} {
		if _, err := s.CreatePlaylist(ctx, alice, PlaylistInput{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	lists, err := s.Playlists(ctx, alice, "爱", Window{})
	names = nil
	for _, p := range lists.Items {
		names = append(names, p.Name)
	}
	if want := []string{"爱", "爱的歌", "我爱"}; err != nil || !slices.Equal(names, want) {
		t.Fatalf("playlists: %v, %v; want %v", names, err, want)
	}
}
