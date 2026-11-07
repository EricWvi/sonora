package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

// Adapted from asr (internal/store/devices.go). Unlike asr, Sonora only has
// an Android client, tokens start with sonora_d_, times are reported as
// RFC 3339 strings like the other createdAt fields, and the user comes with
// its identity provider groups so a device can import like the web.

// PlatformAndroid is the only native client.
const PlatformAndroid = "android"

// DeviceTokenPrefix starts every device token so a leaked one is easy to
// recognise.
const DeviceTokenPrefix = "sonora_d_"

// MaxDeviceName caps the name a client reports for itself, in characters.
const MaxDeviceName = 64

// Device is a native client signed in as a user.
type Device struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"-"`
	Name       string `json:"name"`
	Platform   string `json:"platform"`
	CreatedAt  string `json:"createdAt"`
	LastUsedAt string `json:"lastUsedAt"`
	// LastUsed is LastUsedAt as a time, for deciding when to touch it.
	LastUsed time.Time `json:"-"`
}

// ValidPlatform reports whether p names a supported native client.
func ValidPlatform(p string) bool { return p == PlatformAndroid }

// CleanDeviceName trims a reported device name and checks its length.
func CleanDeviceName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", invalid("设备名不能为空")
	}
	if utf8.RuneCountInString(name) > MaxDeviceName {
		return "", invalid("设备名太长")
	}
	return name, nil
}

func (d *Device) setTimes(created, used int64) {
	d.CreatedAt = formatMillis(created)
	d.LastUsed = time.UnixMilli(used)
	d.LastUsedAt = formatMillis(used)
}

func formatMillis(ms int64) string {
	return time.UnixMilli(ms).UTC().Format("2006-01-02T15:04:05Z")
}

// CreateDevice records a newly signed-in device and returns it with the
// token the client presents from now on. Only the token's hash is stored.
func (s *Store) CreateDevice(ctx context.Context, userID int64, name, platform string) (Device, string, error) {
	name, err := CleanDeviceName(name)
	if err != nil {
		return Device{}, "", err
	}
	if !ValidPlatform(platform) {
		return Device{}, "", invalid("不支持的平台")
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return Device{}, "", err
	}
	token := DeviceTokenPrefix + base64.RawURLEncoding.EncodeToString(raw)
	now := time.Now().UnixMilli()
	d := Device{UserID: userID, Name: name, Platform: platform}
	d.setTimes(now, now)
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO devices (user_id, name, platform, token_hash, created_at, last_used_at)
		VALUES (?, ?, ?, ?, ?, ?)`, userID, name, platform, hashToken(token), now, now)
	if err != nil {
		return Device{}, "", err
	}
	d.ID, err = res.LastInsertId()
	return d, token, err
}

// DeviceUser resolves a device token to its device and user.
func (s *Store) DeviceUser(ctx context.Context, token string) (Device, User, error) {
	var d Device
	var u User
	var groups string
	var created, used int64
	err := s.db.QueryRowContext(ctx, `
		SELECT d.id, d.name, d.platform, d.created_at, d.last_used_at, u.id, u.name, u.email, u.idp_groups
		FROM devices d JOIN users u ON u.id = d.user_id
		WHERE d.token_hash = ?`, hashToken(token)).
		Scan(&d.ID, &d.Name, &d.Platform, &created, &used, &u.ID, &u.Name, &u.Email, &groups)
	if errors.Is(err, sql.ErrNoRows) {
		return d, u, ErrNotFound
	}
	if err != nil {
		return d, u, err
	}
	d.UserID = u.ID
	d.setTimes(created, used)
	return d, u, json.Unmarshal([]byte(groups), &u.Groups)
}

// TouchDevice records that the device is still in use.
func (s *Store) TouchDevice(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE devices SET last_used_at = ? WHERE id = ?`, time.Now().UnixMilli(), id)
	return err
}

// ListDevices returns the user's devices, most recently used first.
func (s *Store) ListDevices(ctx context.Context, userID int64) ([]Device, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, platform, created_at, last_used_at FROM devices
		WHERE user_id = ? ORDER BY last_used_at DESC, id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Device{}
	for rows.Next() {
		d := Device{UserID: userID}
		var created, used int64
		if err := rows.Scan(&d.ID, &d.Name, &d.Platform, &created, &used); err != nil {
			return nil, err
		}
		d.setTimes(created, used)
		out = append(out, d)
	}
	return out, rows.Err()
}

// DeleteDevice revokes a device; its token stops working at once.
func (s *Store) DeleteDevice(ctx context.Context, userID, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM devices WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
