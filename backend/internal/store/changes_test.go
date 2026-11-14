package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"slices"
	"sort"
	"testing"
)

// changeSet is what a page or pages of changes named, by kind.
type changeSet struct {
	tracks, albums, artists                      []int64
	deletedTracks, deletedAlbums, deletedArtists []int64
}

func sorted(ids []int64) []int64 {
	out := slices.Clone(ids)
	slices.Sort(out)
	return out
}

// changesSince reads every change after cursor, a page of limit at a
// time, and returns them with the new cursor.
func changesSince(t *testing.T, s *testStore, cursor int64, limit int) (changeSet, int64) {
	t.Helper()
	var cs changeSet
	for {
		c, err := s.Changes(context.Background(), cursor, limit)
		if err != nil {
			t.Fatal(err)
		}
		if c.Reset {
			t.Fatal("unexpected reset")
		}
		for _, x := range c.Tracks {
			cs.tracks = append(cs.tracks, x.ID)
		}
		for _, x := range c.Albums {
			cs.albums = append(cs.albums, x.ID)
		}
		for _, x := range c.Artists {
			cs.artists = append(cs.artists, x.ID)
		}
		cs.deletedTracks = append(cs.deletedTracks, c.Deleted.Tracks...)
		cs.deletedAlbums = append(cs.deletedAlbums, c.Deleted.Albums...)
		cs.deletedArtists = append(cs.deletedArtists, c.Deleted.Artists...)
		cursor = c.Cursor
		if !c.More {
			break
		}
	}
	for _, ids := range []*[]int64{&cs.tracks, &cs.albums, &cs.artists, &cs.deletedTracks, &cs.deletedAlbums, &cs.deletedArtists} {
		*ids = sorted(*ids)
	}
	return cs, cursor
}

func ids(xs ...int64) []int64 { return sorted(xs) }

func TestChangesFollowEveryEdit(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	a := ingest(t, s, sample(t, "tagged.mp3"), "a.mp3", nil).Track
	b := ingest(t, s, sample(t, "tagged.flac"), "b.flac", nil).Track
	artist, album := a.Artists[0].ID, a.Album.ID
	if b.Album.ID != album || b.Artists[0].ID != artist {
		t.Fatalf("samples should share album and artist: %+v, %+v", a, b)
	}

	cs, cursor := changesSince(t, s, 0, 0)
	if !slices.Equal(cs.tracks, ids(a.ID, b.ID)) || !slices.Equal(cs.albums, ids(album)) ||
		!slices.Equal(cs.artists, ids(artist)) || cs.deletedTracks != nil {
		t.Fatalf("first sync = %+v", cs)
	}
	if again, c := changesSince(t, s, cursor, 0); c != cursor || again.tracks != nil || again.albums != nil {
		t.Fatalf("nothing changed, got %+v at %d (was %d)", again, c, cursor)
	}

	// Renaming the artist changes what their tracks and albums show.
	if _, err := s.UpdateArtist(ctx, artist, ArtistInput{Name: "Jay"}); err != nil {
		t.Fatal(err)
	}
	cs, cursor = changesSince(t, s, cursor, 0)
	if !slices.Equal(cs.tracks, ids(a.ID, b.ID)) || !slices.Equal(cs.albums, ids(album)) || !slices.Equal(cs.artists, ids(artist)) {
		t.Fatalf("after rename = %+v", cs)
	}

	// A new album cover is what its tracks show.
	if _, err := s.SetAlbumCover(ctx, album, sample(t, "cover.jpg")); err != nil {
		t.Fatal(err)
	}
	cs, cursor = changesSince(t, s, cursor, 0)
	if !slices.Equal(cs.tracks, ids(a.ID, b.ID)) || !slices.Equal(cs.albums, ids(album)) || cs.artists != nil {
		t.Fatalf("after album cover = %+v", cs)
	}

	// Lyrics only change the track.
	if _, err := s.SetTrackLyrics(ctx, a.ID, []byte("[00:01.00]词")); err != nil {
		t.Fatal(err)
	}
	cs, cursor = changesSince(t, s, cursor, 0)
	if !slices.Equal(cs.tracks, ids(a.ID)) || cs.albums != nil || cs.artists != nil {
		t.Fatalf("after lyrics = %+v", cs)
	}

	// Moving a track to a new album and artist: the old album and artist
	// lose a track, the new ones appear.
	moved, err := s.UpdateTrack(ctx, b.ID, TrackInput{Title: "b", Artists: []string{"Other"}, Album: "Live"})
	if err != nil {
		t.Fatal(err)
	}
	cs, cursor = changesSince(t, s, cursor, 0)
	if !slices.Equal(cs.tracks, ids(b.ID)) || !slices.Equal(cs.albums, ids(album, moved.Album.ID)) ||
		!slices.Equal(cs.artists, ids(artist, moved.Artists[0].ID)) {
		t.Fatalf("after move = %+v", cs)
	}

	// Merging goes through UPDATE OR IGNORE, whose conflict policy the
	// triggers inherit.
	if _, err := s.MergeArtists(ctx, artist, []int64{moved.Artists[0].ID}); err != nil {
		t.Fatal(err)
	}
	cs, cursor = changesSince(t, s, cursor, 0)
	if !slices.Contains(cs.tracks, b.ID) || !slices.Contains(cs.artists, artist) ||
		!slices.Equal(cs.deletedArtists, ids(moved.Artists[0].ID)) {
		t.Fatalf("after merge = %+v", cs)
	}

	// Deleting the last tracks prunes their albums and artist.
	if err := s.DeleteTrack(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteTrack(ctx, b.ID); err != nil {
		t.Fatal(err)
	}
	cs, _ = changesSince(t, s, cursor, 0)
	if cs.tracks != nil || cs.albums != nil || cs.artists != nil ||
		!slices.Equal(cs.deletedTracks, ids(a.ID, b.ID)) || !slices.Equal(cs.deletedAlbums, ids(album, moved.Album.ID)) ||
		!slices.Equal(cs.deletedArtists, ids(artist)) {
		t.Fatalf("after delete = %+v", cs)
	}
}

func TestChangesPages(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	for _, f := range []string{"tagged.mp3", "tagged.flac", "tagged.ogg", "untagged.mp3"} {
		ingest(t, s, sample(t, f), f, nil)
	}
	all, cursor := changesSince(t, s, 0, 0)
	paged, pagedCursor := changesSince(t, s, 0, 1)
	if len(all.tracks) != 4 || !slices.Equal(paged.tracks, all.tracks) || !slices.Equal(paged.albums, all.albums) ||
		!slices.Equal(paged.artists, all.artists) || pagedCursor != cursor {
		t.Fatalf("paged = %+v at %d, all = %+v at %d", paged, pagedCursor, all, cursor)
	}
	c, err := s.Changes(ctx, 0, 2)
	if err != nil || !c.More || len(c.Tracks)+len(c.Albums)+len(c.Artists) != 2 {
		t.Fatalf("first page = %+v, %v", c, err)
	}

	// A cursor ahead of the log means the database went back in time.
	c, err = s.Changes(ctx, cursor+1, 0)
	if err != nil || !c.Reset || len(c.Tracks) != 0 {
		t.Fatalf("ahead = %+v, %v", c, err)
	}
}

func TestMigrationLogsTheExistingLibrary(t *testing.T) {
	// Fill a library at version 15, then let migration 16 run on it.
	s := openTest(t)
	tr := ingest(t, s, sample(t, "tagged.mp3"), "a.mp3", nil).Track
	s.Store.Close()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(s.dataDir, "sonora.db"))
	if err != nil {
		t.Fatal(err)
	}
	// Undo migration 17 first.
	for _, col := range []string{"loudness", "true_peak", "loudness_version"} {
		if _, err := db.Exec(`ALTER TABLE tracks DROP COLUMN ` + col); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type = 'trigger' AND name LIKE 'library_%'`)
	if err != nil {
		t.Fatal(err)
	}
	undo := []string{`DROP TABLE library_changes`, `PRAGMA user_version = 15`}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		undo = append(undo, `DROP TRIGGER `+name)
	}
	rows.Close()
	if len(undo) != 2+18 {
		t.Fatalf("triggers = %d", len(undo)-2)
	}
	for _, q := range undo {
		if _, err := db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	db.Close()
	reopened, err := Open(s.dataDir, s.mediaDir)
	if err != nil {
		t.Fatal(err)
	}
	s.Store = reopened
	cs, _ := changesSince(t, s, 0, 0)
	if !slices.Equal(cs.tracks, ids(tr.ID)) || !slices.Equal(cs.albums, ids(tr.Album.ID)) ||
		!slices.Equal(cs.artists, ids(tr.Artists[0].ID)) {
		t.Fatalf("after migration = %+v", cs)
	}
}

func TestSortKeysMatchCollation(t *testing.T) {
	s := openTest(t)
	names := []string{"周杰伦", "阿杜", "Beyond", "abc", "Abd", "10 年", "2 年", "陈奕迅", "张学友", "ZARD", "王菲", "林俊杰", "a10", "a9"}
	rows, err := s.db.Query(`WITH n(name) AS (VALUES `+valuesOf(len(names))+`)
		SELECT name FROM n ORDER BY name COLLATE `+sortCollation, anyOf(names)...)
	if err != nil {
		t.Fatal(err)
	}
	var bySQL []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatal(err)
		}
		bySQL = append(bySQL, n)
	}
	rows.Close()
	keys := NewSortKeys()
	byKey := slices.Clone(names)
	sort.SliceStable(byKey, func(i, j int) bool { return keys.Key(byKey[i]) < keys.Key(byKey[j]) })
	if !slices.Equal(byKey, bySQL) {
		t.Fatalf("by key = %v\nby SQL = %v", byKey, bySQL)
	}
}

func valuesOf(n int) string {
	out := "(?)"
	for range n - 1 {
		out += ", (?)"
	}
	return out
}

func anyOf(xs []string) []any {
	out := make([]any, len(xs))
	for i, x := range xs {
		out[i] = x
	}
	return out
}
