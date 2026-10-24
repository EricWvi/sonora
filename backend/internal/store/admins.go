package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"
)

// Admin is an account of the admin console.
type Admin struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

// HasAdmin reports whether the service has been initialised.
func (s *Store) HasAdmin(ctx context.Context) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admins`).Scan(&n)
	return n > 0, err
}

// CreateFirstAdmin creates the initial admin. It fails with ErrConflict
// once any admin exists, so concurrent setups cannot both succeed.
func (s *Store) CreateFirstAdmin(ctx context.Context, username, passwordHash string) (Admin, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO admins (username, password_hash)
		 SELECT ?, ? WHERE NOT EXISTS (SELECT 1 FROM admins)`, username, passwordHash)
	if err != nil {
		return Admin{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return Admin{}, conflict("管理员已创建")
	}
	id, err := res.LastInsertId()
	return Admin{ID: id, Username: username}, err
}

// AdminByUsername returns the admin and its password hash.
func (s *Store) AdminByUsername(ctx context.Context, username string) (Admin, string, error) {
	var a Admin
	var hash string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash FROM admins WHERE username = ?`, username).Scan(&a.ID, &a.Username, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		return a, "", ErrNotFound
	}
	return a, hash, err
}

// AdminSession is a signed-in admin console browser.
type AdminSession struct {
	Admin     Admin
	CreatedAt time.Time
	LastSeen  time.Time
	ExpiresAt time.Time
}

// CreateAdminSession starts a session and returns the secret the browser
// presents. Only its hash is stored.
func (s *Store) CreateAdminSession(ctx context.Context, adminID int64, now, expires time.Time) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO admin_sessions (token_hash, admin_id, created_at, last_seen_at, expires_at) VALUES (?, ?, ?, ?, ?)`,
		hashToken(token), adminID, now.Unix(), now.Unix(), expires.Unix())
	return token, err
}

// LookupAdminSession resolves a session secret. Expiry is the caller's
// decision.
func (s *Store) LookupAdminSession(ctx context.Context, token string) (AdminSession, error) {
	var as AdminSession
	var created, seen, expires int64
	err := s.db.QueryRowContext(ctx, `
		SELECT a.id, a.username, s.created_at, s.last_seen_at, s.expires_at
		FROM admin_sessions s JOIN admins a ON a.id = s.admin_id
		WHERE s.token_hash = ?`, hashToken(token)).Scan(&as.Admin.ID, &as.Admin.Username, &created, &seen, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return as, ErrNotFound
	}
	as.CreatedAt, as.LastSeen, as.ExpiresAt = time.Unix(created, 0), time.Unix(seen, 0), time.Unix(expires, 0)
	return as, err
}

// ExtendAdminSession records use of the session and moves its expiry.
func (s *Store) ExtendAdminSession(ctx context.Context, token string, now, expires time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE admin_sessions SET last_seen_at = ?, expires_at = ? WHERE token_hash = ?`,
		now.Unix(), expires.Unix(), hashToken(token))
	return err
}

// DeleteAdminSession ends a session. Deleting an unknown session is not an
// error.
func (s *Store) DeleteAdminSession(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM admin_sessions WHERE token_hash = ?`, hashToken(token))
	return err
}

// DeleteExpiredAdminSessions removes sessions that expired before now.
func (s *Store) DeleteExpiredAdminSessions(ctx context.Context, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM admin_sessions WHERE expires_at <= ?`, now.Unix())
	return err
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
