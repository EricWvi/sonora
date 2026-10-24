package store

import (
	"context"
	"errors"
	"slices"
	"testing"
)

func TestSortedTrackLists(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	alice := newUser(t, s, "alice")
	ids := fakeTracks(t, s, 4)
	meta := []struct {
		title, artist, album string
		ms                   int
	}{
		{"第10首", "周杰伦", "叶惠美", 3000},
		{"第2首", "Beyond", "", 1000},
		{"爱", "阿桑", "叶惠美", 2000},
		{"Zebra", "", "七里香", 1000},
	}
	for i, m := range meta {
		in := TrackInput{Title: m.title, Album: m.album}
		if m.artist != "" {
			in.Artists = []string{m.artist}
		}
		if _, err := s.UpdateTrack(ctx, ids[i], in); err != nil {
			t.Fatal(err)
		}
		if _, err := s.db.Exec(`UPDATE tracks SET duration_ms = ? WHERE id = ?`, m.ms, ids[i]); err != nil {
			t.Fatal(err)
		}
	}
	a, b, c, d := ids[0], ids[1], ids[2], ids[3]
	for _, tc := range []struct {
		sort Sort
		want []int64
	}{
		{Sort{}, []int64{d, c, b, a}}, // newest first
		{Sort{Field: SortDefault, Order: "asc"}, []int64{d, c, b, a}},
		// Pinyin, digits by value: ai < di2 < di10, Latin first.
		{Sort{Field: SortTitle}, []int64{d, c, b, a}},
		{Sort{Field: SortTitle, Order: "desc"}, []int64{a, b, c, d}},
		// No artist first, then Beyond, a-sang, zhou.
		{Sort{Field: SortArtist}, []int64{d, b, c, a}},
		// Ties keep the default order, newest first.
		{Sort{Field: SortAlbum}, []int64{b, d, c, a}},
		{Sort{Field: SortDuration, Order: "desc"}, []int64{a, c, d, b}},
		{Sort{Field: SortCreateTime}, []int64{a, b, c, d}},
	} {
		page, err := s.Tracks(ctx, TrackQuery{Sort: tc.sort})
		if err != nil || !slices.Equal(trackIDs(page.Items), tc.want) {
			t.Errorf("%+v: got %v, %v; want %v", tc.sort, trackIDs(page.Items), err, tc.want)
		}
	}
	for _, bad := range []Sort{{Field: "size"}, {Field: SortTitle, Order: "up"}} {
		if _, err := s.Tracks(ctx, TrackQuery{Sort: bad}); !errors.Is(err, ErrInvalid) {
			t.Errorf("%+v: %v, want ErrInvalid", bad, err)
		}
	}

	// Likes and playlists sort the same way, defaulting to their own order.
	for _, id := range []int64{b, a, c} {
		if err := s.Like(ctx, alice, id); err != nil {
			t.Fatal(err)
		}
	}
	liked, _ := s.LikedTracks(ctx, alice, Sort{}, Window{})
	byTitle, _ := s.LikedTracks(ctx, alice, Sort{Field: SortTitle}, Window{})
	if !slices.Equal(trackIDs(liked.Items), []int64{c, a, b}) || !slices.Equal(trackIDs(byTitle.Items), []int64{c, b, a}) {
		t.Fatalf("likes: %v, by title %v", trackIDs(liked.Items), trackIDs(byTitle.Items))
	}
	p, err := s.CreatePlaylist(ctx, alice, PlaylistInput{Name: "p"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.AddToPlaylist(ctx, alice, p.ID, []int64{a, d, b}); err != nil {
		t.Fatal(err)
	}
	listed, _ := s.PlaylistTracks(ctx, alice, p.ID, Sort{}, Window{})
	byLength, _ := s.PlaylistTracks(ctx, alice, p.ID, Sort{Field: SortDuration}, Window{})
	if !slices.Equal(trackIDs(listed.Items), []int64{a, d, b}) || !slices.Equal(trackIDs(byLength.Items), []int64{d, b, a}) {
		t.Fatalf("playlist: %v, by duration %v", trackIDs(listed.Items), trackIDs(byLength.Items))
	}

	// Albums: by title unless asked for the newest.
	albums, _ := s.Albums(ctx, AlbumQuery{})
	newest, _ := s.Albums(ctx, AlbumQuery{Sort: Sort{Field: SortCreateTime, Order: "desc"}})
	albumIDs := func(p Page[Album]) []int64 {
		ids := []int64{}
		for _, a := range p.Items {
			ids = append(ids, a.ID)
		}
		return ids
	}
	// 叶惠美 twice, by different artists; 七里香 (qi) sorts before 叶 (ye).
	qili := trackAlbum(t, s, d)
	if !slices.Equal(albumIDs(albums), []int64{qili, trackAlbum(t, s, a), trackAlbum(t, s, c)}) ||
		!slices.Equal(albumIDs(newest), []int64{qili, trackAlbum(t, s, c), trackAlbum(t, s, a)}) {
		t.Fatalf("albums %v, newest %v", albumIDs(albums), albumIDs(newest))
	}
	if _, err := s.Albums(ctx, AlbumQuery{Sort: Sort{Field: SortTitle}}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("album sort by title: %v", err)
	}
}

func trackAlbum(t *testing.T, s *testStore, id int64) int64 {
	t.Helper()
	d, err := s.Track(context.Background(), id)
	if err != nil || d.Album == nil {
		t.Fatalf("track %d: %+v, %v", id, d.Album, err)
	}
	return d.Album.ID
}
