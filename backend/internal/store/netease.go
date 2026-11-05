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
