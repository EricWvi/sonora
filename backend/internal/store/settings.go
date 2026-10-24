package store

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

// MaxSettingsBytes caps a listener's saved settings.
const MaxSettingsBytes = 256 << 10

// Settings returns the user's saved appearance and lyrics settings, a JSON
// object the player owns; "{}" when none are saved yet.
func (s *Store) Settings(ctx context.Context, userID int64) (json.RawMessage, error) {
	var out string
	err := s.db.QueryRowContext(ctx, `SELECT settings FROM user_settings WHERE user_id = ?`, userID).Scan(&out)
	if errors.Is(err, sql.ErrNoRows) {
		return json.RawMessage(`{}`), nil
	}
	return json.RawMessage(out), err
}

// SaveSettings replaces the user's settings with data, which must be a
// JSON object. Its contents are not interpreted.
func (s *Store) SaveSettings(ctx context.Context, userID int64, data []byte) error {
	if len(data) > MaxSettingsBytes {
		return invalid("设置不能超过 256 KB")
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, data); err != nil {
		return invalid("设置必须是 JSON 对象")
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("{")) {
		return invalid("设置必须是 JSON 对象")
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_settings (user_id, settings) VALUES (?, ?)
		ON CONFLICT (user_id) DO UPDATE SET settings = excluded.settings,
			updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')`, userID, buf.String())
	return err
}
