package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"sonora/internal/media"
)

// importMP3 is an MP3 tagged as an album import writes it.
func importMP3(t *testing.T, st *Store, m media.Meta) *Upload {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "media", "testdata", "untagged.mp3"))
	if err != nil {
		t.Fatal(err)
	}
	// Every song is its own file, as on NetEase.
	data = append(data, []byte(m.Title)...)
	path := filepath.Join(t.TempDir(), "song.mp3")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if m.Picture == nil {
		if m.Picture, err = os.ReadFile(filepath.Join("..", "media", "testdata", "cover.jpg")); err != nil {
			t.Fatal(err)
		}
	}
	if err := media.WriteTags(path, m); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	u, err := st.Receive(f, m.Title+".mp3")
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func TestAlbumImport(t *testing.T) {
	st := openTest(t).Store
	ctx := context.Background()
	single := func(title string, song, album int64) IngestResult {
		t.Helper()
		res, err := st.IngestImported(ctx, importMP3(t, st, media.Meta{Title: title, Artists: []string{"A"}, AlbumArtists: []string{"A"}, Album: "Comp"}),
			nil, NeteaseIDs{Song: song, Album: album})
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	alone := single("alone", 11, 21)
	merged := single("merged", 12, 21)
	upload, err := st.Ingest(ctx, importMP3(t, st, media.Meta{Title: "up", Artists: []string{"U"}, AlbumArtists: []string{"U"}, Album: "Comp"}), nil)
	if err != nil {
		t.Fatal(err)
	}
	// The admin merges one single with another single of the album: neither
	// is alone any more.
	other := single("other", 13, 21)
	if _, err := st.MergeAlbums(ctx, other.Track.Album.ID, []int64{merged.Track.Album.ID}, false); err != nil {
		t.Fatal(err)
	}

	places, err := st.ImportedPlacements(ctx, []int64{11, 12, 13, 14})
	if err != nil {
		t.Fatal(err)
	}
	if len(places) != 3 || !places[11].Alone || places[12].Alone || places[13].Alone || places[11].TrackID != alone.Track.ID {
		t.Fatalf("placements: %+v", places)
	}
	// Albums made for singles are no target.
	if _, err := st.ImportTarget(ctx, 21); err != ErrNotFound {
		t.Fatalf("target before: %v", err)
	}
	if n, err := st.SameTitleAlbums(ctx, "comp", 21); err != nil || n != 1 {
		t.Fatalf("same title: %d %v", n, err)
	}

	a := AlbumImport{NeteaseID: 21, Title: "Comp", Artists: []string{"群星"}, Year: 2017, Description: " notes "}
	ids := map[string]int64{"群星": 99}
	moved, err := st.MoveImported(ctx, alone.Track.ID, a, ids, 3, 2)
	if err != nil || !moved {
		t.Fatalf("move: %v %v", moved, err)
	}
	if moved, err := st.MoveImported(ctx, merged.Track.ID, a, ids, 4, 1); err != nil || moved {
		t.Fatalf("merged song moved: %v %v", moved, err)
	}
	target, err := st.ImportTarget(ctx, 21)
	if err != nil {
		t.Fatal(err)
	}
	if target.Title != "Comp" || len(target.Artists) != 1 || target.Artists[0].Name != "群星" || target.Single ||
		target.Year != 2017 || target.Description != "notes" || target.Cover == nil || target.TrackCount != 1 {
		t.Fatalf("target: %+v", target)
	}
	if _, err := st.Album(ctx, alone.Track.Album.ID); err != ErrNotFound {
		t.Fatalf("emptied single album: %v", err)
	}
	d, err := st.Track(ctx, alone.Track.ID)
	if err != nil || d.Album.ID != target.ID || d.TrackNo != 3 || d.DiscNo != 2 || d.Cover != nil || d.Artists[0].Name != "A" {
		t.Fatalf("moved track: %v %+v", err, d.Track)
	}
	if va, err := st.Artist(ctx, target.Artists[0].ID); err != nil || va.NeteaseID == nil || *va.NeteaseID != 99 {
		t.Fatalf("album artist: %v %+v", err, va)
	}

	// A new song joins the same album at its place; with the album's cover it
	// keeps none of its own, with another it does.
	join := func(title string, song int64, picture []byte) IngestResult {
		t.Helper()
		res, err := st.IngestImported(ctx, importMP3(t, st, media.Meta{Title: title, Artists: []string{"B"}, AlbumArtists: []string{"群星"},
			Album: "Comp", Year: 2017, TrackNo: 5, DiscNo: 1, Picture: picture}), nil, NeteaseIDs{Song: song, Album: 21, FromAlbum: &a})
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	same := join("new", 14, nil)
	if same.Track.Album.ID != target.ID || same.Track.TrackNo != 5 || same.Track.DiscNo != 1 || same.Track.Cover != nil {
		t.Fatalf("joined: %+v", same.Track)
	}
	png, err := os.ReadFile(filepath.Join("..", "media", "testdata", "cover.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	png = append(png, 0) // another file, still a JPEG
	if differs := join("own cover", 15, png); differs.Track.Cover == nil {
		t.Fatalf("own cover dropped: %+v", differs.Track)
	}
	// The upload of the same title stays apart.
	if u, err := st.Track(ctx, upload.Track.ID); err != nil || u.Album.ID == target.ID {
		t.Fatalf("upload: %v %+v", err, u.Album)
	}

	// An album the admin linked to a NetEase album is where its import goes.
	if _, err := st.UpdateAlbum(ctx, upload.Track.Album.ID, AlbumInput{Title: "Comp", Artists: []string{"U"}, NeteaseID: OptionalID{Set: true, Value: ptr(int64(31))}}); err != nil {
		t.Fatal(err)
	}
	b := AlbumImport{NeteaseID: 31, Title: "Comp", Artists: []string{"U"}, Year: 2020, Description: "linked"}
	res, err := st.IngestImported(ctx, importMP3(t, st, media.Meta{Title: "linked", Artists: []string{"U"}, Album: "Comp"}), nil, NeteaseIDs{Song: 41, Album: 31, FromAlbum: &b})
	if err != nil || res.Track.Album.ID != upload.Track.Album.ID {
		t.Fatalf("linked: %v %+v", err, res.Track.Album)
	}
	if al, err := st.Album(ctx, upload.Track.Album.ID); err != nil || al.Year != 2020 || al.Description != "linked" {
		t.Fatalf("linked album filled: %v %+v", err, al)
	}

	// Editing the imported album keeps it apart from a same-titled one.
	if _, err := st.UpdateAlbum(ctx, target.ID, AlbumInput{Title: "Comp", Artists: []string{"U"}}); err != nil {
		t.Fatalf("edit: %v", err)
	}
}

func ptr[T any](v T) *T { return &v }

// Editing an imported track without touching its album keeps it there; the
// album's mark used to make the lookup miss and move the track out.
func TestEditKeepsTrackInItsAlbum(t *testing.T) {
	st := openTest(t).Store
	ctx := context.Background()
	single, err := st.IngestImported(ctx, importMP3(t, st, media.Meta{Title: "single", Artists: []string{"A"}, AlbumArtists: []string{"A"}, Album: "Comp"}),
		nil, NeteaseIDs{Song: 11})
	if err != nil {
		t.Fatal(err)
	}
	a := AlbumImport{NeteaseID: 21, Title: "Comp", Artists: []string{"群星"}}
	whole, err := st.IngestImported(ctx, importMP3(t, st, media.Meta{Title: "whole", Artists: []string{"B"}, AlbumArtists: []string{"群星"}, Album: "Comp", TrackNo: 2}),
		nil, NeteaseIDs{Song: 12, FromAlbum: &a})
	if err != nil {
		t.Fatal(err)
	}
	upload, err := st.Ingest(ctx, importMP3(t, st, media.Meta{Title: "up", Artists: []string{"A"}, AlbumArtists: []string{"A"}, Album: "Comp"}), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		res          IngestResult
		albumArtists []string
	}{{single, []string{"A"}}, {whole, []string{"群星"}}, {upload, []string{"A"}}} {
		tr := tc.res.Track
		// The album title differs only in ASCII case, which is the same album.
		got, err := st.UpdateTrack(ctx, tr.ID, TrackInput{Title: tr.Title + " (edited)", Artists: []string{tr.Artists[0].Name},
			Album: "comp", AlbumArtists: tc.albumArtists, TrackNo: tr.TrackNo})
		if err != nil || got.Album == nil || got.Album.ID != tr.Album.ID {
			t.Fatalf("%s: %v moved from %d to %+v", tr.Title, err, tr.Album.ID, got.Album)
		}
	}
	// The album artists of the whole album's track left out means the track's
	// own, which is another album: it moves.
	moved, err := st.UpdateTrack(ctx, whole.Track.ID, TrackInput{Title: "whole", Artists: []string{"B"}, Album: "Comp"})
	if err != nil || moved.Album.ID == whole.Track.Album.ID {
		t.Fatalf("other album artists: %v %+v", err, moved.Album)
	}
	renamed, err := st.UpdateTrack(ctx, single.Track.ID, TrackInput{Title: "single", Artists: []string{"A"}, Album: "Other", AlbumArtists: []string{"A"}})
	if err != nil || renamed.Album.ID == single.Track.Album.ID || renamed.Album.Title != "Other" {
		t.Fatalf("renamed: %v %+v", err, renamed.Album)
	}
}
