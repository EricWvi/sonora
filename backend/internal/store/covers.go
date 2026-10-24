package store

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"image"
	_ "image/jpeg" // register decoders for DecodeConfig
	_ "image/png"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	_ "golang.org/x/image/webp"
)

// MaxCoverBytes caps an uploaded or embedded cover image.
const MaxCoverBytes = 10 << 20

var coverName = regexp.MustCompile(`^[0-9a-f]{32}\.(jpg|png|webp)$`)

// saveCover stores an image under its content hash and returns the file
// name. The caller holds s.files.
func (s *Store) saveCover(data []byte) (string, error) {
	if len(data) > MaxCoverBytes {
		return "", invalid("封面不能超过 10 MB")
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", invalid("封面只支持 JPEG、PNG 或 WebP 图片")
	}
	ext := map[string]string{"jpeg": ".jpg", "png": ".png", "webp": ".webp"}[format]
	if ext == "" {
		return "", invalid("封面只支持 JPEG、PNG 或 WebP 图片")
	}
	sum := sha256.Sum256(data)
	name := hex.EncodeToString(sum[:16]) + ext
	path := filepath.Join(s.coverDir, name)
	if _, err := os.Stat(path); err == nil {
		return name, nil
	}
	tmp, err := os.CreateTemp(s.coverDir, ".cover-*")
	if err != nil {
		return "", err
	}
	_, err = tmp.Write(data)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp.Name(), path)
	}
	if err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	return name, nil
}

// CoverPath returns the file holding the named cover.
func (s *Store) CoverPath(name string) (string, error) {
	if !coverName.MatchString(name) {
		return "", ErrNotFound
	}
	return filepath.Join(s.coverDir, name), nil
}

// SetAlbumCover replaces an album's cover.
func (s *Store) SetAlbumCover(ctx context.Context, id int64, data []byte) (Album, error) {
	return setCover(s, ctx, `UPDATE albums SET cover = ? WHERE id = ?`, id, data, s.album)
}

// ClearAlbumCover removes an album's cover.
func (s *Store) ClearAlbumCover(ctx context.Context, id int64) (Album, error) {
	return setCover(s, ctx, `UPDATE albums SET cover = ? WHERE id = ?`, id, nil, s.album)
}

// SetTrackCover gives a track its own cover, shown instead of the album's.
func (s *Store) SetTrackCover(ctx context.Context, id int64, data []byte) (TrackDetail, error) {
	return setCover(s, ctx, `UPDATE tracks SET cover = ? WHERE id = ?`, id, data, s.track)
}

// ClearTrackCover removes a track's own cover so the album's shows.
func (s *Store) ClearTrackCover(ctx context.Context, id int64) (TrackDetail, error) {
	return setCover(s, ctx, `UPDATE tracks SET cover = ? WHERE id = ?`, id, nil, s.track)
}

// setCover stores data (nil clears) and points the row at it.
func setCover[T any](s *Store, ctx context.Context, query string, id int64, data []byte,
	get func(context.Context, querier, int64) (T, error)) (T, error) {
	var zero T
	s.files.Lock()
	defer s.files.Unlock()
	var name sql.NullString
	if data != nil {
		n, err := s.saveCover(data)
		if err != nil {
			return zero, err
		}
		name = sql.NullString{String: n, Valid: true}
	}
	var out T
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, query, name, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		out, err = get(ctx, tx, id)
		return err
	})
	s.collectCovers(ctx)
	if err != nil {
		return zero, err
	}
	return out, nil
}

// collectCovers removes cover files no album or track refers to. The
// caller holds s.files. Failures are logged; a leftover file is harmless.
func (s *Store) collectCovers(ctx context.Context) {
	used := map[string]bool{}
	rows, err := s.db.QueryContext(ctx,
		`SELECT cover FROM albums WHERE cover IS NOT NULL UNION SELECT cover FROM tracks WHERE cover IS NOT NULL`)
	if err != nil {
		slog.Error("collect covers", "err", err)
		return
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			slog.Error("collect covers", "err", err)
			return
		}
		used[name] = true
	}
	rows.Close()
	if rows.Err() != nil {
		return
	}
	entries, err := os.ReadDir(s.coverDir)
	if err != nil {
		slog.Error("collect covers", "err", err)
		return
	}
	for _, e := range entries {
		if !used[e.Name()] && coverName.MatchString(e.Name()) {
			if err := os.Remove(filepath.Join(s.coverDir, e.Name())); err != nil && !errors.Is(err, os.ErrNotExist) {
				slog.Error("remove cover", "err", err)
			}
		}
	}
}
