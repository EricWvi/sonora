package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// NeteaseSession binds a listener to their NetEase account. The cookie stays
// on the server: it has no JSON form so it cannot end up in a response.
type NeteaseSession struct {
	Cookie      string    `json:"-"`
	AccountID   int64     `json:"-"`
	RefreshedAt time.Time `json:"-"`
}

// NeteaseSession returns the listener's session, or ErrNotFound without one.
func (s *Store) NeteaseSession(ctx context.Context, userID int64) (NeteaseSession, error) {
	var out NeteaseSession
	var refreshed string
	err := s.db.QueryRowContext(ctx,
		`SELECT cookie, account_id, refreshed_at FROM netease_sessions WHERE user_id = ?`, userID).
		Scan(&out.Cookie, &out.AccountID, &refreshed)
	if errors.Is(err, sql.ErrNoRows) {
		return out, ErrNotFound
	}
	if err != nil {
		return out, err
	}
	out.RefreshedAt, err = time.Parse(time.RFC3339, refreshed)
	return out, err
}

// SaveNeteaseSession stores a session just signed in or refreshed, replacing
// the listener's previous one.
func (s *Store) SaveNeteaseSession(ctx context.Context, userID int64, session NeteaseSession) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO netease_sessions (user_id, cookie, account_id, refreshed_at) VALUES (?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET cookie = excluded.cookie, account_id = excluded.account_id,
			refreshed_at = excluded.refreshed_at`,
		userID, session.Cookie, session.AccountID, session.RefreshedAt.UTC().Format(time.RFC3339))
	return err
}

// DeleteNeteaseSession forgets the listener's session. It is idempotent.
func (s *Store) DeleteNeteaseSession(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM netease_sessions WHERE user_id = ?`, userID)
	return err
}

// OnlinePlay is a NetEase song in a listener's history. Song is the snapshot
// the player sent; the store does not interpret it.
type OnlinePlay struct {
	NeteaseID int64
	Song      string
	PlayedAt  string
}

// LikedEntry is a liked track and when it was liked.
type LikedEntry struct {
	Track   Track
	LikedAt string
}

// LikedEntries returns every track the user liked, most recent first.
func (s *Store) LikedEntries(ctx context.Context, userID int64) ([]LikedEntry, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+trackColumns+`, l.created_at FROM likes l
		JOIN tracks t ON t.id = l.track_id LEFT JOIN albums al ON al.id = t.album_id
		WHERE l.user_id = ? ORDER BY l.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []LikedEntry{}
	for rows.Next() {
		var e LikedEntry
		if e.Track, err = scanTrack(rows, &e.LikedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ImportedTracks maps the NetEase ID of every imported track to its track ID.
func (s *Store) ImportedTracks(ctx context.Context) (map[int64]int64, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT netease_id, id FROM tracks WHERE netease_id IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]int64{}
	for rows.Next() {
		var neteaseID, id int64
		if err := rows.Scan(&neteaseID, &id); err != nil {
			return nil, err
		}
		out[neteaseID] = id
	}
	return out, rows.Err()
}

// SetNeteaseID records which NetEase song a track was imported from.
func (s *Store) SetNeteaseID(ctx context.Context, trackID, neteaseID int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE tracks SET netease_id = ? WHERE id = ?`, neteaseID, trackID)
	return err
}

// OnlineHistory returns the NetEase songs the user played, most recent first.
func (s *Store) OnlineHistory(ctx context.Context, userID int64) ([]OnlinePlay, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT netease_id, song, played_at FROM online_history WHERE user_id = ? ORDER BY id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []OnlinePlay{}
	for rows.Next() {
		var p OnlinePlay
		if err := rows.Scan(&p.NeteaseID, &p.Song, &p.PlayedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// RecordOnlinePlay moves the NetEase song to the top of the user's history.
func (s *Store) RecordOnlinePlay(ctx context.Context, userID, neteaseID int64, song string) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM online_history WHERE user_id = ? AND netease_id = ?`, userID, neteaseID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO online_history (user_id, netease_id, song) VALUES (?, ?, ?)`, userID, neteaseID, song); err != nil {
			return err
		}
		return trimHistory(ctx, tx, userID)
	})
}

// RemoveFromOnlineHistory drops one NetEase song from the user's history.
func (s *Store) RemoveFromOnlineHistory(ctx context.Context, userID, neteaseID int64) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM online_history WHERE user_id = ? AND netease_id = ?`, userID, neteaseID)
	return err
}

// trimHistory keeps the newest MaxHistory entries across local and NetEase
// songs. Within one second, local entries count as the newer.
func trimHistory(ctx context.Context, tx *sql.Tx, userID int64) error {
	const keep = `WITH keep AS (
		SELECT 'local' AS kind, id, played_at FROM play_history WHERE user_id = ?1
		UNION ALL
		SELECT 'online', id, played_at FROM online_history WHERE user_id = ?1
		ORDER BY played_at DESC, kind ASC, id DESC LIMIT ?2)`
	if _, err := tx.ExecContext(ctx, keep+`
		DELETE FROM play_history WHERE user_id = ?1
			AND id NOT IN (SELECT id FROM keep WHERE kind = 'local')`, userID, MaxHistory); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, keep+`
		DELETE FROM online_history WHERE user_id = ?1
			AND id NOT IN (SELECT id FROM keep WHERE kind = 'online')`, userID, MaxHistory)
	return err
}
