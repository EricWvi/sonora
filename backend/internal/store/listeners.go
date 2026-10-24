package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// What a listener keeps: liked tracks, playlists and play history. Every
// method takes the user it acts for; another user's playlist is reported
// as missing, not forbidden.

const (
	// MaxHistory is how many distinct tracks a listener's history keeps.
	MaxHistory = 500
	// MaxPlaylistTracks caps one playlist.
	MaxPlaylistTracks = 10000
	maxPlaylistName   = 100
	maxDescription    = 1000
)

// trackPage runs count and then items, both taking args, and hands each
// row of the window of items to scan. It returns the count.
func (s *Store) trackPage(ctx context.Context, count, items string, args []any, w Window,
	scan func(rows *sql.Rows) error) (int, error) {
	var total int
	if err := s.db.QueryRowContext(ctx, count, args...).Scan(&total); err != nil {
		return 0, err
	}
	w = w.clamp()
	rows, err := s.db.QueryContext(ctx, items+` LIMIT ? OFFSET ?`, append(args, w.Limit, w.Offset)...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return 0, err
		}
	}
	return total, rows.Err()
}

func trackExists(ctx context.Context, q querier, id int64) error {
	var one int
	err := q.QueryRowContext(ctx, `SELECT 1 FROM tracks WHERE id = ?`, id).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// --- likes -----------------------------------------------------------------

// LikedTracks lists the tracks the user liked, most recent first.
func (s *Store) LikedTracks(ctx context.Context, userID int64, w Window) (Page[Track], error) {
	page := Page[Track]{Items: []Track{}}
	var err error
	page.Total, err = s.trackPage(ctx,
		`SELECT COUNT(*) FROM likes WHERE user_id = ?`,
		`SELECT `+trackColumns+` FROM likes l JOIN tracks t ON t.id = l.track_id
			LEFT JOIN albums al ON al.id = t.album_id
		WHERE l.user_id = ? ORDER BY l.id DESC`,
		[]any{userID}, w, func(rows *sql.Rows) error {
			t, err := scanTrack(rows)
			page.Items = append(page.Items, t)
			return err
		})
	return page, err
}

// LikedTrackIDs returns the IDs of every track the user liked, most
// recent first, so the player can mark them in any list.
func (s *Store) LikedTrackIDs(ctx context.Context, userID int64) ([]int64, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT track_id FROM likes WHERE user_id = ? ORDER BY id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// Like adds a track to the user's likes. Liking it again changes nothing.
func (s *Store) Like(ctx context.Context, userID, trackID int64) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := trackExists(ctx, tx, trackID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO likes (user_id, track_id) VALUES (?, ?) ON CONFLICT DO NOTHING`, userID, trackID)
		return err
	})
}

// Unlike removes a track from the user's likes, if it is there.
func (s *Store) Unlike(ctx context.Context, userID, trackID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM likes WHERE user_id = ? AND track_id = ?`, userID, trackID)
	return err
}

// --- history ---------------------------------------------------------------

// HistoryEntry is a track the user played and when they last did.
type HistoryEntry struct {
	Track    Track  `json:"track"`
	PlayedAt string `json:"playedAt"`
}

// History lists the tracks the user played, most recent first.
func (s *Store) History(ctx context.Context, userID int64, w Window) (Page[HistoryEntry], error) {
	page := Page[HistoryEntry]{Items: []HistoryEntry{}}
	var err error
	page.Total, err = s.trackPage(ctx,
		`SELECT COUNT(*) FROM play_history WHERE user_id = ?`,
		`SELECT `+trackColumns+`, h.played_at FROM play_history h JOIN tracks t ON t.id = h.track_id
			LEFT JOIN albums al ON al.id = t.album_id
		WHERE h.user_id = ? ORDER BY h.id DESC`,
		[]any{userID}, w, func(rows *sql.Rows) error {
			var e HistoryEntry
			var err error
			e.Track, err = scanTrack(rows, &e.PlayedAt)
			page.Items = append(page.Items, e)
			return err
		})
	return page, err
}

// RecordPlay moves the track to the top of the user's history, dropping
// the oldest entries beyond MaxHistory.
func (s *Store) RecordPlay(ctx context.Context, userID, trackID int64) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := trackExists(ctx, tx, trackID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM play_history WHERE user_id = ? AND track_id = ?`, userID, trackID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO play_history (user_id, track_id) VALUES (?, ?)`, userID, trackID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `
			DELETE FROM play_history WHERE user_id = ? AND id NOT IN (
				SELECT id FROM play_history WHERE user_id = ? ORDER BY id DESC LIMIT ?)`,
			userID, userID, MaxHistory)
		return err
	})
}

// RemoveFromHistory drops one track from the user's history.
func (s *Store) RemoveFromHistory(ctx context.Context, userID, trackID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM play_history WHERE user_id = ? AND track_id = ?`, userID, trackID)
	return err
}

// ClearHistory empties the user's history.
func (s *Store) ClearHistory(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM play_history WHERE user_id = ?`, userID)
	return err
}

// --- playlists -------------------------------------------------------------

// Playlist is a listener's own list of tracks. Cover is the cover of its
// first track that shows one.
type Playlist struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cover       *Cover `json:"cover"`
	TrackCount  int    `json:"trackCount"`
	DurationMs  int64  `json:"durationMs"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// PlaylistInput is the editable part of a playlist.
type PlaylistInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (in *PlaylistInput) normalize() error {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	switch {
	case in.Name == "":
		return invalid("歌单名不能为空")
	case utf8.RuneCountInString(in.Name) > maxPlaylistName:
		return invalid(fmt.Sprintf("歌单名不能超过 %d 个字符", maxPlaylistName))
	case utf8.RuneCountInString(in.Description) > maxDescription:
		return invalid(fmt.Sprintf("简介不能超过 %d 个字符", maxDescription))
	}
	return nil
}

// firstCover selects column col (name or thumbhash) of the cover of the
// playlist's first track that shows one.
func firstCover(col string) string {
	return `(SELECT ` + col + ` FROM playlist_tracks pt JOIN tracks t ON t.id = pt.track_id
		LEFT JOIN albums al ON al.id = t.album_id
		JOIN covers c ON c.name = IFNULL(t.cover, al.cover)
		WHERE pt.playlist_id = p.id ORDER BY pt.position LIMIT 1)`
}

var playlistSelect = `
	SELECT p.id, p.name, p.description, ` + firstCover("c.name") + `, ` + firstCover("c.thumbhash") + `,
		(SELECT COUNT(*) FROM playlist_tracks WHERE playlist_id = p.id),
		(SELECT IFNULL(SUM(t.duration_ms), 0) FROM playlist_tracks pt JOIN tracks t ON t.id = pt.track_id
			WHERE pt.playlist_id = p.id),
		p.created_at, p.updated_at
	FROM playlists p`

func scanPlaylist(row scanner) (Playlist, error) {
	var p Playlist
	var cover, coverHash sql.NullString
	err := row.Scan(&p.ID, &p.Name, &p.Description, &cover, &coverHash, &p.TrackCount, &p.DurationMs,
		&p.CreatedAt, &p.UpdatedAt)
	p.Cover = coverOf(cover, coverHash)
	return p, err
}

// Playlists lists the user's playlists, newest first. q matches names.
func (s *Store) Playlists(ctx context.Context, userID int64, q string, w Window) (Page[Playlist], error) {
	page := Page[Playlist]{Items: []Playlist{}}
	cond, args := ` WHERE p.user_id = ?`, []any{userID}
	if q = strings.TrimSpace(q); q != "" {
		cond += ` AND p.name LIKE ? ESCAPE '\'`
		args = append(args, like(q))
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM playlists p`+cond, args...).Scan(&page.Total); err != nil {
		return page, err
	}
	w = w.clamp()
	rows, err := s.db.QueryContext(ctx, playlistSelect+cond+` ORDER BY p.id DESC LIMIT ? OFFSET ?`,
		append(args, w.Limit, w.Offset)...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		p, err := scanPlaylist(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, p)
	}
	return page, rows.Err()
}

// Playlist returns one of the user's playlists.
func (s *Store) Playlist(ctx context.Context, userID, id int64) (Playlist, error) {
	return s.playlist(ctx, s.db, userID, id)
}

func (s *Store) playlist(ctx context.Context, q querier, userID, id int64) (Playlist, error) {
	p, err := scanPlaylist(q.QueryRowContext(ctx, playlistSelect+` WHERE p.id = ? AND p.user_id = ?`, id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, err
}

// CreatePlaylist creates an empty playlist for the user.
func (s *Store) CreatePlaylist(ctx context.Context, userID int64, in PlaylistInput) (Playlist, error) {
	if err := in.normalize(); err != nil {
		return Playlist{}, err
	}
	var p Playlist
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var id int64
		err := tx.QueryRowContext(ctx,
			`INSERT INTO playlists (user_id, name, description) VALUES (?, ?, ?) RETURNING id`,
			userID, in.Name, in.Description).Scan(&id)
		if err != nil {
			return err
		}
		p, err = s.playlist(ctx, tx, userID, id)
		return err
	})
	return p, err
}

// UpdatePlaylist renames a playlist and replaces its description.
func (s *Store) UpdatePlaylist(ctx context.Context, userID, id int64, in PlaylistInput) (Playlist, error) {
	if err := in.normalize(); err != nil {
		return Playlist{}, err
	}
	var p Playlist
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `
			UPDATE playlists SET name = ?, description = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
			WHERE id = ? AND user_id = ?`, in.Name, in.Description, id, userID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		p, err = s.playlist(ctx, tx, userID, id)
		return err
	})
	return p, err
}

// DeletePlaylist deletes one of the user's playlists.
func (s *Store) DeletePlaylist(ctx context.Context, userID, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM playlists WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// PlaylistTracks lists a playlist's tracks in order.
func (s *Store) PlaylistTracks(ctx context.Context, userID, id int64, w Window) (Page[Track], error) {
	page := Page[Track]{Items: []Track{}}
	if err := ownPlaylist(ctx, s.db, userID, id); err != nil {
		return page, err
	}
	var err error
	page.Total, err = s.trackPage(ctx,
		`SELECT COUNT(*) FROM playlist_tracks WHERE playlist_id = ?`,
		`SELECT `+trackColumns+` FROM playlist_tracks pt JOIN tracks t ON t.id = pt.track_id
			LEFT JOIN albums al ON al.id = t.album_id
		WHERE pt.playlist_id = ? ORDER BY pt.position`,
		[]any{id}, w, func(rows *sql.Rows) error {
			t, err := scanTrack(rows)
			page.Items = append(page.Items, t)
			return err
		})
	return page, err
}

func ownPlaylist(ctx context.Context, q querier, userID, id int64) error {
	var one int
	err := q.QueryRowContext(ctx, `SELECT 1 FROM playlists WHERE id = ? AND user_id = ?`, id, userID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func touchPlaylist(ctx context.Context, tx *sql.Tx, id int64) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE playlists SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = ?`, id)
	return err
}

// AddToPlaylist appends tracks to the end of a playlist in the given
// order, skipping those already in it. It reports how many were added.
// A missing track fails the whole request.
func (s *Store) AddToPlaylist(ctx context.Context, userID, id int64, trackIDs []int64) (int, Playlist, error) {
	if len(trackIDs) == 0 {
		return 0, Playlist{}, invalid("请选择要添加的歌曲")
	}
	var added int
	var p Playlist
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := ownPlaylist(ctx, tx, userID, id); err != nil {
			return err
		}
		var count, next int
		if err := tx.QueryRowContext(ctx,
			`SELECT COUNT(*), IFNULL(MAX(position) + 1, 0) FROM playlist_tracks WHERE playlist_id = ?`, id).
			Scan(&count, &next); err != nil {
			return err
		}
		for _, t := range trackIDs {
			if err := trackExists(ctx, tx, t); err != nil {
				return err
			}
			res, err := tx.ExecContext(ctx,
				`INSERT INTO playlist_tracks (playlist_id, track_id, position) VALUES (?, ?, ?) ON CONFLICT DO NOTHING`,
				id, t, next)
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n > 0 {
				added++
				next++
			}
		}
		if count+added > MaxPlaylistTracks {
			return invalid(fmt.Sprintf("一个歌单最多 %d 首歌", MaxPlaylistTracks))
		}
		if added > 0 {
			if err := touchPlaylist(ctx, tx, id); err != nil {
				return err
			}
		}
		var err error
		p, err = s.playlist(ctx, tx, userID, id)
		return err
	})
	return added, p, err
}

// RemoveFromPlaylist removes a track from a playlist, if it is there.
func (s *Store) RemoveFromPlaylist(ctx context.Context, userID, id, trackID int64) (Playlist, error) {
	var p Playlist
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := ownPlaylist(ctx, tx, userID, id); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx,
			`DELETE FROM playlist_tracks WHERE playlist_id = ? AND track_id = ?`, id, trackID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			if err := touchPlaylist(ctx, tx, id); err != nil {
				return err
			}
		}
		p, err = s.playlist(ctx, tx, userID, id)
		return err
	})
	return p, err
}

// ReorderPlaylist puts a playlist's tracks in the given order. trackIDs
// must be exactly the tracks the playlist holds; otherwise the client
// worked from a stale copy and the request is a conflict.
func (s *Store) ReorderPlaylist(ctx context.Context, userID, id int64, trackIDs []int64) (Playlist, error) {
	var p Playlist
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := ownPlaylist(ctx, tx, userID, id); err != nil {
			return err
		}
		var count int
		if err := tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM playlist_tracks WHERE playlist_id = ?`, id).Scan(&count); err != nil {
			return err
		}
		stale := conflict("歌单已在别处修改，请刷新后重试")
		if len(trackIDs) != count {
			return stale
		}
		seen := make(map[int64]bool, len(trackIDs))
		for i, t := range trackIDs {
			if seen[t] {
				return invalid("歌曲重复")
			}
			seen[t] = true
			res, err := tx.ExecContext(ctx,
				`UPDATE playlist_tracks SET position = ? WHERE playlist_id = ? AND track_id = ?`, i, id, t)
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n == 0 {
				return stale
			}
		}
		if err := touchPlaylist(ctx, tx, id); err != nil {
			return err
		}
		var err error
		p, err = s.playlist(ctx, tx, userID, id)
		return err
	})
	return p, err
}
