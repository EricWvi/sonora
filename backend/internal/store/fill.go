package store

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"
	"unicode/utf8"
)

// Filling in what a linked NetEase album, artist or song has (Eric,
// 2026-11-07). Each only writes what is empty, so nothing the admin wrote is
// replaced, also when the request races with an edit.

func trimDescription(description string) string {
	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(description) > maxAlbumDescription {
		description = string([]rune(description)[:maxAlbumDescription])
	}
	return description
}

// FillAlbumDescription sets an album's description if it has none. It
// reports whether it did.
func (s *Store) FillAlbumDescription(ctx context.Context, id int64, description string) (Album, bool, error) {
	description = trimDescription(description)
	filled := false
	if description != "" {
		res, err := s.db.ExecContext(ctx, `UPDATE albums SET description = ? WHERE id = ? AND description = ''`, description, id)
		if err != nil {
			return Album{}, false, err
		}
		n, _ := res.RowsAffected()
		filled = n > 0
	}
	album, err := s.Album(ctx, id)
	return album, filled, err
}

// Fields FillArtistProfile can fill, as the API names them.
const (
	FilledImage       = "image"
	FilledDescription = "description"
)

// FillArtistProfile sets an artist's image and description where they are
// empty and returns which it set. image may be nil; one the library cannot
// read is left out.
func (s *Store) FillArtistProfile(ctx context.Context, id int64, image []byte, description string) (Artist, []string, error) {
	description = trimDescription(description)
	s.files.Lock()
	defer s.files.Unlock()
	before, err := s.Artist(ctx, id)
	if err != nil {
		return before, nil, err
	}
	var cover sql.NullString
	if image != nil && before.Image == nil {
		if name, err := s.saveCover(ctx, image); err == nil {
			cover = sql.NullString{String: name, Valid: true}
		} else {
			slog.Warn("ignore artist image", "artist", id, "err", err)
		}
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE artists SET image = IFNULL(image, ?),
			description = CASE WHEN description = '' THEN ? ELSE description END
		WHERE id = ?`, cover, description, id)
	if cover.Valid {
		s.collectCovers(ctx)
	}
	if err != nil {
		return before, nil, err
	}
	after, err := s.Artist(ctx, id)
	if err != nil {
		return after, nil, err
	}
	filled := []string{}
	if before.Image == nil && after.Image != nil {
		filled = append(filled, FilledImage)
	}
	if before.Description == "" && after.Description != "" {
		filled = append(filled, FilledDescription)
	}
	return after, filled, nil
}

// FillTrackLyrics sets a track's lyrics if it has none. It reports whether
// it did.
func (s *Store) FillTrackLyrics(ctx context.Context, id int64, data []byte) (TrackDetail, bool, error) {
	filled := false
	if len(data) > 0 {
		text, format, err := readLyrics(data)
		if err != nil {
			return TrackDetail{}, false, err
		}
		res, err := s.db.ExecContext(ctx, `
			UPDATE tracks SET lyrics = ?, lyrics_format = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
			WHERE id = ? AND lyrics = ''`, text, format, id)
		if err != nil {
			return TrackDetail{}, false, err
		}
		n, _ := res.RowsAffected()
		filled = n > 0
	}
	track, err := s.Track(ctx, id)
	return track, filled, err
}
