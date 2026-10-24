package store

import (
	"context"
	"errors"
	"testing"
)

func TestArtistImages(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	tr := ingest(t, s, sample(t, "tagged.m4a"), "a.m4a", nil).Track
	if _, err := s.ClearAlbumCover(ctx, tr.Album.ID); err != nil {
		t.Fatal(err)
	}
	artist := tr.Artists[0].ID
	if _, err := s.SetArtistImage(ctx, artist, []byte("not an image")); !errors.Is(err, ErrInvalid) {
		t.Fatalf("garbage image: %v", err)
	}
	a, err := s.SetArtistImage(ctx, artist, sample(t, "cover.jpg"))
	if err != nil || a.Image == nil || a.Image.Thumbhash == "" {
		t.Fatalf("artist = %+v, %v", a, err)
	}
	if _, err := s.CoverThumb(a.Image.Name, 96); err != nil {
		t.Fatalf("thumbnail: %v", err)
	}
	if _, err := s.SetArtistImage(ctx, 9999, sample(t, "cover.jpg")); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing artist: %v", err)
	}

	// A merged artist hands its image to a target without one.
	other, err := s.UpdateTrack(ctx, ingest(t, s, sample(t, "tagged.opus"), "b.opus", nil).Track.ID,
		TrackInput{Title: "b", Artists: []string{"Someone"}})
	if err != nil {
		t.Fatal(err)
	}
	merged, err := s.MergeArtists(ctx, other.Artists[0].ID, []int64{artist})
	if err != nil || merged.Image == nil || merged.Image.Name != a.Image.Name {
		t.Fatalf("merged = %+v, %v", merged, err)
	}
	if n := coverFiles(t, s); n != 1 {
		t.Fatalf("image files after merge = %d", n)
	}

	// An artist pruned with its last track takes its image along.
	page, _ := s.Tracks(ctx, TrackQuery{ArtistID: merged.ID})
	for _, it := range page.Items {
		if err := s.DeleteTrack(ctx, it.ID); err != nil {
			t.Fatal(err)
		}
	}
	if n := coverFiles(t, s); n != 0 {
		t.Fatalf("image files after artist is gone = %d", n)
	}
}
