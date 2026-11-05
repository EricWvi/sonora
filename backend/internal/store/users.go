package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

// User is a listener who signed in to the player through the identity
// provider. Admins are separate accounts, see Admin.
type User struct {
	ID    int64  `json:"-"`
	Name  string `json:"name"`
	Email string `json:"email"`
	// Groups are the identity provider groups as of the last login.
	Groups []string `json:"-"`
}

// Identity is what the identity provider vouches for. Issuer and Subject
// identify the person; Name and Email are display attributes that may
// change between logins, and Groups the memberships at this login.
type Identity struct {
	Issuer  string
	Subject string
	Name    string
	Email   string
	Groups  []string
}

// UpsertUser returns the user bound to the identity, creating it on first
// login and refreshing the display attributes on later ones.
func (s *Store) UpsertUser(ctx context.Context, id Identity) (User, error) {
	if id.Issuer == "" || id.Subject == "" {
		return User{}, invalid("身份缺少 issuer 或 subject")
	}
	u := User{Name: id.Name, Email: id.Email, Groups: id.Groups}
	groups, err := json.Marshal(append([]string{}, id.Groups...))
	if err != nil {
		return User{}, err
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx,
			`SELECT id FROM users WHERE issuer = ? AND subject = ?`, id.Issuer, id.Subject).Scan(&u.ID)
		if err == nil {
			_, err = tx.ExecContext(ctx,
				`UPDATE users SET name = ?, email = ?, idp_groups = ? WHERE id = ?`, id.Name, id.Email, string(groups), u.ID)
			return err
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		res, err := tx.ExecContext(ctx,
			`INSERT INTO users (issuer, subject, name, email, idp_groups) VALUES (?, ?, ?, ?, ?)`,
			id.Issuer, id.Subject, id.Name, id.Email, string(groups))
		if err != nil {
			return err
		}
		u.ID, err = res.LastInsertId()
		return err
	})
	return u, err
}

// CreateSession starts a session for the user and returns the secret the
// browser presents. Only its hash is stored.
func (s *Store) CreateSession(ctx context.Context, userID int64) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	now := time.Now().Unix()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (token_hash, user_id, created_at, last_seen_at) VALUES (?, ?, ?, ?)`,
		hashToken(token), userID, now, now)
	return token, err
}

// SessionUser resolves a session secret to its user and reports when the
// session was last refreshed.
func (s *Store) SessionUser(ctx context.Context, token string) (User, time.Time, error) {
	var u User
	var lastSeen int64
	var groups string
	err := s.db.QueryRowContext(ctx, `
		SELECT u.id, u.name, u.email, u.idp_groups, s.last_seen_at
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = ?`, hashToken(token)).Scan(&u.ID, &u.Name, &u.Email, &groups, &lastSeen)
	if errors.Is(err, sql.ErrNoRows) {
		return u, time.Time{}, ErrNotFound
	}
	if err == nil {
		err = json.Unmarshal([]byte(groups), &u.Groups)
	}
	return u, time.Unix(lastSeen, 0), err
}

// TouchSession records that the session is still in use.
func (s *Store) TouchSession(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE sessions SET last_seen_at = ? WHERE token_hash = ?`, time.Now().Unix(), hashToken(token))
	return err
}

// DeleteSession ends a session. Deleting an unknown session is not an error.
func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = ?`, hashToken(token))
	return err
}
