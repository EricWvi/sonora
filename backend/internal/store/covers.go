package store

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // register decoders
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"sonora/internal/thumbhash"
)

// MaxCoverBytes caps an uploaded or embedded cover image, and an artist
// image, which is stored the same way.
const MaxCoverBytes = 10 << 20

// maxCoverPixels bounds the images decoded for placeholders and
// thumbnails; a larger one is still stored and served as is.
const maxCoverPixels = 25_000_000

// CoverSizes are the thumbnail edge lengths served besides the original.
var CoverSizes = []int{96, 300}

var coverName = regexp.MustCompile(`^[0-9a-f]{32}\.(jpg|png|webp)$`)

// usedCovers selects the names of the stored images something refers to.
const usedCovers = `
	SELECT cover FROM albums WHERE cover IS NOT NULL
	UNION SELECT cover FROM tracks WHERE cover IS NOT NULL
	UNION SELECT image FROM artists WHERE image IS NOT NULL`

// Cover is a stored cover image. Thumbhash is its blurred placeholder,
// empty when the image could not be decoded.
type Cover struct {
	Name      string `json:"name"`
	Thumbhash string `json:"thumbhash"`
}

func coverOf(name, hash sql.NullString) *Cover {
	if !name.Valid {
		return nil
	}
	return &Cover{Name: name.String, Thumbhash: hash.String}
}

// decodeCover decodes an image within the pixel budget.
func decodeCover(data []byte) (image.Image, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if cfg.Width*cfg.Height > maxCoverPixels {
		return nil, fmt.Errorf("cover is %dx%d", cfg.Width, cfg.Height)
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

// placeholder returns the ThumbHash of an image, or "" when it cannot be
// decoded.
func placeholder(data []byte) string {
	img, err := decodeCover(data)
	if err != nil {
		slog.Warn("cover placeholder", "err", err)
		return ""
	}
	return thumbhash.Encode(img)
}

// saveCover stores an image under its content hash, records its
// placeholder and returns the file name. The caller holds s.files.
func (s *Store) saveCover(ctx context.Context, data []byte) (string, error) {
	if len(data) > MaxCoverBytes {
		return "", invalid("图片不能超过 10 MB")
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", invalid("图片只支持 JPEG、PNG 或 WebP 格式")
	}
	ext := map[string]string{"jpeg": ".jpg", "png": ".png", "webp": ".webp"}[format]
	if ext == "" {
		return "", invalid("图片只支持 JPEG、PNG 或 WebP 格式")
	}
	sum := sha256.Sum256(data)
	name := hex.EncodeToString(sum[:16]) + ext
	var known int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM covers WHERE name = ?`, name).Scan(&known); err != nil {
		return "", err
	}
	if known == 0 {
		if _, err := s.db.ExecContext(ctx, `INSERT INTO covers (name, thumbhash) VALUES (?, ?)`,
			name, placeholder(data)); err != nil {
			return "", err
		}
	}
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

func (s *Store) thumbPath(name string, size int) string {
	return filepath.Join(s.coverDir, "thumbs", fmt.Sprintf("%s-%d.jpg", strings.TrimSuffix(name, filepath.Ext(name)), size))
}

// CoverThumb returns a JPEG of the named cover scaled to fit size×size,
// one of CoverSizes, creating it on first use. A cover already that small,
// or one that cannot be decoded, is returned as is.
func (s *Store) CoverThumb(name string, size int) (string, error) {
	orig, err := s.CoverPath(name)
	if err != nil {
		return "", err
	}
	if !slices.Contains(CoverSizes, size) {
		return "", invalid(fmt.Sprintf("封面尺寸只支持 %v", CoverSizes))
	}
	path := s.thumbPath(name, size)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	data, err := os.ReadFile(orig)
	if errors.Is(err, os.ErrNotExist) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	s.decoding <- struct{}{} // bound the memory of concurrent decodes
	defer func() { <-s.decoding }()
	img, err := decodeCover(data)
	if err != nil {
		slog.Warn("cover thumbnail", "cover", name, "err", err)
		return orig, nil
	}
	b := img.Bounds()
	if max(b.Dx(), b.Dy()) <= size {
		return orig, nil
	}
	w, h := size, size*b.Dy()/b.Dx()
	if b.Dy() > b.Dx() {
		w, h = size*b.Dx()/b.Dy(), size
	}
	dst := image.NewRGBA(image.Rect(0, 0, max(w, 1), max(h, 1)))
	draw.Draw(dst, dst.Bounds(), image.White, image.Point{}, draw.Src) // JPEG has no alpha
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 82}); err != nil {
		return "", err
	}
	// Concurrent requests may race to write the same bytes; rename keeps
	// readers from seeing a partial file.
	if err := writeFileAtomic(filepath.Dir(path), path, buf.Bytes()); err != nil {
		return "", err
	}
	return path, nil
}

func writeFileAtomic(dir, path string, data []byte) error {
	tmp, err := os.CreateTemp(dir, ".thumb-*")
	if err != nil {
		return err
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
	}
	return err
}

// backfillPlaceholders records placeholders for covers stored before they
// existed.
func (s *Store) backfillPlaceholders(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, usedCovers+` EXCEPT SELECT name FROM covers`)
	if err != nil {
		return err
	}
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			rows.Close()
			return err
		}
		names = append(names, n)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, n := range names {
		hash := ""
		if data, err := os.ReadFile(filepath.Join(s.coverDir, n)); err == nil {
			hash = placeholder(data)
		} else {
			slog.Warn("cover placeholder", "cover", n, "err", err)
		}
		if _, err := s.db.ExecContext(ctx, `INSERT INTO covers (name, thumbhash) VALUES (?, ?)`, n, hash); err != nil {
			return err
		}
	}
	if len(names) > 0 {
		slog.Info("cover placeholders created", "count", len(names))
	}
	return nil
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

// SetArtistImage replaces an artist's image.
func (s *Store) SetArtistImage(ctx context.Context, id int64, data []byte) (Artist, error) {
	return setCover(s, ctx, `UPDATE artists SET image = ? WHERE id = ?`, id, data, s.artist)
}

// ClearArtistImage removes an artist's image.
func (s *Store) ClearArtistImage(ctx context.Context, id int64) (Artist, error) {
	return setCover(s, ctx, `UPDATE artists SET image = ? WHERE id = ?`, id, nil, s.artist)
}

// setCover stores data (nil clears) and points the row at it.
func setCover[T any](s *Store, ctx context.Context, query string, id int64, data []byte,
	get func(context.Context, querier, int64) (T, error)) (T, error) {
	var zero T
	s.files.Lock()
	defer s.files.Unlock()
	var name sql.NullString
	if data != nil {
		n, err := s.saveCover(ctx, data)
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

// collectCovers removes image files no album, track or artist refers to. The
// caller holds s.files. Failures are logged; a leftover file is harmless.
func (s *Store) collectCovers(ctx context.Context) {
	used := map[string]bool{}
	rows, err := s.db.QueryContext(ctx, usedCovers)
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
	if _, err := s.db.ExecContext(ctx, `
		DELETE FROM covers WHERE name NOT IN (`+usedCovers+`)`); err != nil {
		slog.Error("collect covers", "err", err)
	}
	entries, err := os.ReadDir(s.coverDir)
	if err != nil {
		slog.Error("collect covers", "err", err)
		return
	}
	usedBase := map[string]bool{}
	for _, e := range entries {
		if !coverName.MatchString(e.Name()) {
			continue
		}
		if used[e.Name()] {
			usedBase[strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))] = true
		} else {
			removeFile(filepath.Join(s.coverDir, e.Name()))
		}
	}
	thumbs, err := os.ReadDir(filepath.Join(s.coverDir, "thumbs"))
	if err != nil {
		slog.Error("collect covers", "err", err)
		return
	}
	for _, e := range thumbs {
		// Dot files are thumbnails being written.
		if base, _, ok := strings.Cut(e.Name(), "-"); ok && !strings.HasPrefix(base, ".") && !usedBase[base] {
			removeFile(filepath.Join(s.coverDir, "thumbs", e.Name()))
		}
	}
}

func removeFile(path string) {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Error("remove cover", "err", err)
	}
}
