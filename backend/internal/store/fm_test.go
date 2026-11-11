package store

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
)

func TestFMSegmentSize(t *testing.T) {
	for _, c := range []struct {
		name                           string
		bigPlays, bigs, smallPlays, sm int
		want                           int
	}{
		{"first cycle of the real library", 1, 29, 0, 549, 19},
		{"no small albums", 3, 29, 0, 0, 0},
		{"no big album", 0, 0, 0, 100, fmSoloSegment},
		{"far behind is capped", 290, 29, 0, 549, 24},
		{"far ahead is floored", 0, 29, 5490, 549, 14},
		{"never below one", 0, 100, 500, 5, 1},
	} {
		if got := fmSegmentSize(c.bigPlays, c.bigs, c.smallPlays, c.sm); got != c.want {
			t.Errorf("%s: %d, want %d", c.name, got, c.want)
		}
	}
}

// Simulates listening: counts stay within one of each other across the pools
// and groups, also while the library changes.
func TestFMCyclesEvenOutCounts(t *testing.T) {
	rnd := rand.New(rand.NewPCG(1, 2))
	var albums []fmCandidate
	next := int64(1)
	add := func(n int, group int64, plays int) {
		for range n {
			albums = append(albums, fmCandidate{id: next, group: group, plays: plays})
			next++
		}
	}
	lowest := func(group func(int64) bool) int {
		low := -1
		for _, a := range albums {
			if group(a.group) && (low < 0 || a.plays < low) {
				low = a.plays
			}
		}
		return low
	}
	all := func(int64) bool { return true }
	add(29, 0, 0)
	add(276, 1, 0)
	add(194, 2, 0)
	add(79, 3, 0)
	check := func(segment int) {
		t.Helper()
		low, high := lowest(all), 0
		for _, a := range albums {
			high = max(high, a.plays)
		}
		if high-low > 1 {
			t.Fatalf("segment %d: counts from %d to %d", segment, low, high)
		}
	}
	for segment := range 2000 {
		switch segment {
		case 500:
			add(60, 1, lowest(func(g int64) bool { return g != 0 }))
		case 1000:
			add(8, 0, lowest(func(g int64) bool { return g == 0 }))
		}
		if segment == 500 || segment == 1000 {
			// Give the caps time to make up for the change.
			for range 100 {
				fmListen(albums, rnd)
			}
		}
		fmListen(albums, rnd)
		if (segment+1)%100 == 0 {
			check(segment)
		}
	}
}

func fmListen(albums []fmCandidate, rnd *rand.Rand) {
	for _, id := range fmPlanCycle(albums, rnd) {
		albums[slices.IndexFunc(albums, func(c fmCandidate) bool { return c.id == id })].plays++
	}
}

func TestFMCycleKeepsBigFirstAndGroupsTogether(t *testing.T) {
	rnd := rand.New(rand.NewPCG(3, 4))
	var albums []fmCandidate
	groupOf := map[int64]int64{}
	for i := int64(1); i <= 300; i++ {
		c := fmCandidate{id: i}
		switch {
		case i <= 10:
		case i <= 160:
			c.group = 1
		case i <= 260:
			c.group = 2
		default:
			c.group = 3
		}
		groupOf[i] = c.group
		albums = append(albums, c)
	}
	orders := map[string]bool{}
	for range 200 {
		cycle := fmPlanCycle(albums, rnd)
		if groupOf[cycle[0]] != 0 {
			t.Fatalf("cycle starts with a small album: %v", cycle)
		}
		order := ""
		var last int64
		seen := map[int64]bool{}
		for _, id := range cycle[1:] {
			g := groupOf[id]
			if g == 0 {
				t.Fatalf("second big album in %v", cycle)
			}
			if g != last {
				if seen[g] {
					t.Fatalf("group %d split in %v", g, cycle)
				}
				seen[g] = true
				order += fmt.Sprint(g)
				last = g
			}
		}
		orders[order] = true
		for _, id := range cycle {
			albums[id-1].plays++
		}
	}
	if len(orders) < 6 {
		t.Fatalf("group orders seen: %v", orders)
	}
}

func TestFMDrawSkipsBlocked(t *testing.T) {
	rnd := rand.New(rand.NewPCG(5, 6))
	cycle := fmPlanCycle([]fmCandidate{
		{id: 1, plays: 0, blocked: true},
		{id: 2, plays: 3},
		{id: 3, plays: 0, group: 7, blocked: true},
		{id: 4, plays: 5, group: 7},
	}, rnd)
	if !slices.Equal(cycle, []int64{2, 4}) {
		t.Fatalf("cycle %v", cycle)
	}
}

// fmAlbum inserts an album with n tracks without audio files.
func fmAlbum(t *testing.T, s *testStore, n int) (int64, []int64) {
	t.Helper()
	var album int64
	key := fmt.Sprintf("%s-%d", t.Name(), n)
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM albums`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if err := s.db.QueryRow(`INSERT INTO albums (title, artist_key) VALUES (?, ?) RETURNING id`,
		fmt.Sprintf("album %d", count), key).Scan(&album); err != nil {
		t.Fatal(err)
	}
	tracks := fakeTracksIn(t, s, album, n)
	return album, tracks
}

func fakeTracksIn(t *testing.T, s *testStore, album int64, n int) []int64 {
	t.Helper()
	var ids []int64
	for i := range n {
		var id int64
		if err := s.db.QueryRow(`
			INSERT INTO tracks (title, album_id, track_no, duration_ms, format, size, sha256, file, original_name)
			VALUES (?, ?, ?, 1000, 'mp3', 1, ?, 'x', 'x.mp3') RETURNING id`,
			fmt.Sprintf("track %d", i), album, n-i, fmt.Sprintf("%d-%d-%s", album, i, t.Name())).Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	return ids
}

func fmPlays(t *testing.T, s *testStore, user, album int64) (plays int, skipped bool) {
	t.Helper()
	if err := s.db.QueryRow(`SELECT plays, skipped FROM fm_albums WHERE user_id = ? AND album_id = ?`, user, album).
		Scan(&plays, &skipped); err != nil {
		t.Fatal(err)
	}
	return plays, skipped
}

func TestFMPlansAlternateAndCountOnFinish(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	user := newUser(t, s, "alice")
	big, _ := fmAlbum(t, s, 10)
	small, smallTracks := fmAlbum(t, s, 9)
	ungrouped, _ := fmAlbum(t, s, 1)

	// Without groups only the big album plays.
	state, err := s.FM(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Albums) != 1 || state.Albums[0].ID != big {
		t.Fatalf("albums %+v", state.Albums)
	}

	pl, err := s.CreatePlaylist(ctx, user, PlaylistInput{Name: "CN"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.AddToPlaylist(ctx, user, pl.ID, smallTracks[:1]); err != nil {
		t.Fatal(err)
	}
	state, err = s.FMSetGroups(ctx, user, []int64{pl.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Albums) != 2 || state.Albums[0].ID != big || state.Albums[1].ID != small {
		t.Fatalf("albums %+v", state.Albums)
	}
	if !slices.Equal(state.GroupIDs, []int64{pl.ID}) {
		t.Fatalf("groups %v", state.GroupIDs)
	}
	// Playing order inside the album, not insertion order.
	if tr := state.Albums[1].Tracks; tr[0].TrackNo != 1 || tr[len(tr)-1].TrackNo != 9 {
		t.Fatalf("order %d..%d", tr[0].TrackNo, tr[len(tr)-1].TrackNo)
	}

	// Being planned is not being heard; finishing is, once.
	if plays, _ := fmPlays(t, s, user, big); plays != 0 {
		t.Fatalf("plays %d", plays)
	}
	if state, err = s.FMFinish(ctx, user, big); err != nil {
		t.Fatal(err)
	}
	if state.Albums[0].ID != small {
		t.Fatalf("current %d", state.Albums[0].ID)
	}
	if _, err = s.FMFinish(ctx, user, big); err != nil {
		t.Fatal(err)
	}
	if plays, _ := fmPlays(t, s, user, big); plays != 1 {
		t.Fatalf("plays after a repeated finish %d", plays)
	}

	// The small album outside the groups never joined.
	var rows int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM fm_albums WHERE album_id = ?`, ungrouped).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 0 {
		t.Fatal("ungrouped album was drawn")
	}
}

func TestFMResumesAtPositionUntilAlbumChanges(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	user := newUser(t, s, "alice")
	first, firstTracks := fmAlbum(t, s, 10)
	second, secondTracks := fmAlbum(t, s, 11)

	state, err := s.FM(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if state.TrackID != nil {
		t.Fatalf("position %v", *state.TrackID)
	}
	current, later := first, second
	currentTracks, laterTracks := firstTracks, secondTracks
	if state.Albums[0].ID == second {
		current, later = second, first
		currentTracks, laterTracks = secondTracks, firstTracks
	}
	if err := s.FMSetPosition(ctx, user, currentTracks[3]); err != nil {
		t.Fatal(err)
	}
	state, _ = s.FM(ctx, user)
	if state.TrackID == nil || *state.TrackID != currentTracks[3] {
		t.Fatalf("position %v", state.TrackID)
	}
	// Jumping into the next album passes over the current one like a skip.
	if err := s.FMSetPosition(ctx, user, laterTracks[0]); err != nil {
		t.Fatal(err)
	}
	state, _ = s.FM(ctx, user)
	if state.Albums[0].ID != later || *state.TrackID != laterTracks[0] {
		t.Fatalf("after jump %+v", state)
	}
	if plays, _ := fmPlays(t, s, user, current); plays != 0 {
		t.Fatalf("passed-over album counted: %d", plays)
	}
	// A track of no planned album changes nothing.
	other := fakeTracks(t, s, 1)[0]
	if err := s.FMSetPosition(ctx, user, other); err != nil {
		t.Fatal(err)
	}
	state, _ = s.FM(ctx, user)
	if state.Albums[0].ID != later {
		t.Fatalf("current %d", state.Albums[0].ID)
	}
}

func TestFMSkipSitsOutOneDraw(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	user := newUser(t, s, "alice")
	a, _ := fmAlbum(t, s, 10)
	b, _ := fmAlbum(t, s, 11)
	c, _ := fmAlbum(t, s, 12)

	state, err := s.FM(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	skipped := state.Albums[0].ID
	if state, err = s.FMSkip(ctx, user, skipped); err != nil {
		t.Fatal(err)
	}
	for _, album := range state.Albums {
		if album.ID == skipped {
			t.Fatalf("skipped album back at once: %d", album.ID)
		}
	}
	if plays, _ := fmPlays(t, s, user, skipped); plays != 0 {
		t.Fatalf("skip counted: %d", plays)
	}
	// Finishing the others brings it back without the count having moved.
	seen := map[int64]bool{}
	for range 6 {
		seen[state.Albums[0].ID] = true
		if state, err = s.FMFinish(ctx, user, state.Albums[0].ID); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []int64{a, b, c} {
		if !seen[id] {
			t.Fatalf("album %d never came up: %v", id, seen)
		}
	}
}

func TestFMNewAlbumStartsAtLowestCount(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	user := newUser(t, s, "alice")
	a, _ := fmAlbum(t, s, 10)
	b, _ := fmAlbum(t, s, 11)
	if _, err := s.FM(ctx, user); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`UPDATE fm_albums SET plays = 7 WHERE album_id = ?`, a); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`UPDATE fm_albums SET plays = 5 WHERE album_id = ?`, b); err != nil {
		t.Fatal(err)
	}
	joined, _ := fmAlbum(t, s, 12)
	if _, err := s.db.Exec(`DELETE FROM fm_plan WHERE user_id = ?`, user); err != nil {
		t.Fatal(err)
	}
	if _, err := s.FM(ctx, user); err != nil {
		t.Fatal(err)
	}
	if plays, _ := fmPlays(t, s, user, joined); plays != 5 {
		t.Fatalf("new album at %d", plays)
	}
}

func TestFMGroupsAlbumByMostTracksThenPriority(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	user := newUser(t, s, "alice")
	_, tracks := fmAlbum(t, s, 4)
	cn, _ := s.CreatePlaylist(ctx, user, PlaylistInput{Name: "CN"})
	en, _ := s.CreatePlaylist(ctx, user, PlaylistInput{Name: "EN"})
	if _, _, err := s.AddToPlaylist(ctx, user, cn.ID, tracks[:1]); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.AddToPlaylist(ctx, user, en.ID, tracks[1:2]); err != nil {
		t.Fatal(err)
	}
	var album int64
	if err := s.db.QueryRow(`SELECT album_id FROM tracks WHERE id = ?`, tracks[0]).Scan(&album); err != nil {
		t.Fatal(err)
	}
	groupOf := func(order ...int64) int64 {
		t.Helper()
		if _, err := s.FMSetGroups(ctx, user, order); err != nil {
			t.Fatal(err)
		}
		var group int64
		err := s.inTx(ctx, func(tx *sql.Tx) error {
			assigned, err := fmAlbumGroups(ctx, tx, user, order)
			group = assigned[album]
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		return group
	}
	if g := groupOf(cn.ID, en.ID); g != cn.ID {
		t.Fatalf("tie went to %d", g)
	}
	if g := groupOf(en.ID, cn.ID); g != en.ID {
		t.Fatalf("tie went to %d", g)
	}
	if _, _, err := s.AddToPlaylist(ctx, user, cn.ID, tracks[2:3]); err != nil {
		t.Fatal(err)
	}
	if g := groupOf(en.ID, cn.ID); g != cn.ID {
		t.Fatalf("most tracks lost to %d", g)
	}
}

func TestFMForgetsDeletedAlbumsAndForeignPlaylists(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	alice, bob := newUser(t, s, "alice"), newUser(t, s, "bob")
	album, _ := fmAlbum(t, s, 10)
	if _, err := s.FM(ctx, alice); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`DELETE FROM tracks WHERE album_id = ?`, album); err != nil {
		t.Fatal(err)
	}
	state, err := s.FM(ctx, alice)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Albums) != 0 {
		t.Fatalf("albums %+v", state.Albums)
	}
	theirs, _ := s.CreatePlaylist(ctx, alice, PlaylistInput{Name: "CN"})
	if _, err := s.FMSetGroups(ctx, bob, []int64{theirs.ID}); err == nil {
		t.Fatal("another listener's playlist was accepted")
	}
}

func TestFMAheadLinesUpMoreAlbums(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	user := newUser(t, s, "alice")
	for i := range 8 {
		fmAlbum(t, s, 10+i)
	}

	// Only big albums: one per draw, so six need more draws than two.
	ahead, err := s.FMAhead(ctx, user, 6)
	if err != nil {
		t.Fatal(err)
	}
	if len(ahead.Albums) != 6 {
		t.Fatalf("albums %d", len(ahead.Albums))
	}
	seen := map[int64]bool{}
	for _, a := range ahead.Albums {
		if seen[a.ID] {
			t.Fatalf("album %d twice", a.ID)
		}
		seen[a.ID] = true
	}
	// They are the radio's real plan: the player's two are the first two.
	state, err := s.FM(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Albums) != 2 || state.Albums[0].ID != ahead.Albums[0].ID || state.Albums[1].ID != ahead.Albums[1].ID {
		t.Fatalf("player %+v, ahead %+v", state.Albums, ahead.Albums[:2])
	}

	// Clamped to 1..FMMaxAhead; the library only has eight albums.
	if state, _ = s.FMAhead(ctx, user, 0); len(state.Albums) != 1 {
		t.Fatalf("clamped low %d", len(state.Albums))
	}
	if state, _ = s.FMAhead(ctx, user, 99); len(state.Albums) != 8 {
		t.Fatalf("clamped high %d", len(state.Albums))
	}
}
