package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sonora/internal/media"
)

// Upload is an audio file received into the incoming area and identified,
// but not yet part of the library. Call Ingest or Discard.
type Upload struct {
	path   string
	sha256 string
	size   int64
	name   string
	info   media.Info
}

// Receive copies an uploaded file into the incoming area and identifies
// it. name is the browser's file name, used as the title when the file has
// none. Errors wrapping media.ErrUnsupported or media.ErrCorrupt describe
// the file; nothing is kept.
func (s *Store) Receive(body io.Reader, name string) (*Upload, error) {
	tmp, err := os.CreateTemp(s.incoming, "upload-*")
	if err != nil {
		return nil, err
	}
	u := &Upload{path: tmp.Name(), name: baseName(name)}
	hash := sha256.New()
	u.size, err = io.Copy(io.MultiWriter(tmp, hash), body)
	if err == nil {
		u.sha256 = hex.EncodeToString(hash.Sum(nil))
		u.info, err = media.Probe(tmp, u.size)
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		u.Discard()
		return nil, err
	}
	return u, nil
}

// Discard drops an upload that was not ingested. It is safe to call after
// Ingest.
func (u *Upload) Discard() {
	if err := os.Remove(u.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Error("remove upload", "err", err)
	}
}

// IngestStatus tells a new track from a file the library already had.
type IngestStatus string

const (
	Created   IngestStatus = "created"
	Duplicate IngestStatus = "duplicate"
)

// IngestResult is the outcome of adding one file.
type IngestResult struct {
	Status IngestStatus `json:"status"`
	Track  TrackDetail  `json:"track"`
}

// Ingest adds the upload to the library. A file whose content is already
// in the library is not stored again; the existing track is returned.
// lyrics is an optional sidecar LRC, TTML or YRC file that takes precedence
// over lyrics embedded in the tags.
func (s *Store) Ingest(ctx context.Context, u *Upload, lyrics []byte) (IngestResult, error) {
	return s.ingest(ctx, u, lyrics, NeteaseIDs{})
}

// IngestImported adds a song imported from NetEase. Imported on its own, unlike
// an upload, it always gets a new album marked single, and keeps its cover as
// well as giving it to the album, so that the song keeps it when albums are
// merged later; with ids.FromAlbum it joins that album instead. ids.Song must
// be set.
func (s *Store) IngestImported(ctx context.Context, u *Upload, lyrics []byte, ids NeteaseIDs) (IngestResult, error) {
	if ids.Song == 0 {
		return IngestResult{}, invalid("导入的歌曲缺少网易云 ID")
	}
	return s.ingest(ctx, u, lyrics, ids)
}

// NeteaseIDs are the NetEase IDs of an imported song, its album and its
// artists. Artists maps an artist's name, as written to the tags, to their
// ID; a missing name or a zero ID leaves that one unset.
type NeteaseIDs struct {
	Song    int64
	Album   int64
	Artists map[string]int64
	// FromAlbum is set when the song comes with an album import: it joins
	// that album instead of getting one of its own, and its cover is kept
	// only when it differs from the album's, as on upload.
	FromAlbum *AlbumImport
}

// IncomingFile creates a temporary file in the incoming area, on the same
// volume as the library. The caller removes it.
func (s *Store) IncomingFile() (*os.File, error) {
	return os.CreateTemp(s.incoming, "import-*")
}

// ImportedTrack returns the track imported from the NetEase song.
func (s *Store) ImportedTrack(ctx context.Context, neteaseID int64) (TrackDetail, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, `SELECT id FROM tracks WHERE netease_id = ?`, neteaseID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return TrackDetail{}, ErrNotFound
	}
	if err != nil {
		return TrackDetail{}, err
	}
	return s.Track(ctx, id)
}

func (s *Store) ingest(ctx context.Context, u *Upload, lyrics []byte, ids NeteaseIDs) (IngestResult, error) {
	defer u.Discard()
	// Embedded lyrics are taken as LRC.
	text, format := u.info.Tags.Lyrics, LRC
	if len(lyrics) > 0 {
		var err error
		if text, format, err = readLyrics(lyrics); err != nil {
			return IngestResult{}, err
		}
	}

	s.files.Lock()
	defer s.files.Unlock()

	var existing int64
	err := s.db.QueryRowContext(ctx, `SELECT id FROM tracks WHERE sha256 = ?`, u.sha256).Scan(&existing)
	if err == nil {
		d, err := s.Track(ctx, existing)
		return IngestResult{Status: Duplicate, Track: d}, err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return IngestResult{}, err
	}

	rel := filepath.Join(u.sha256[:2], u.sha256+"."+string(u.info.Format))
	dst := filepath.Join(s.mediaDir, rel)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return IngestResult{}, err
	}
	if err := os.Rename(u.path, dst); err != nil {
		return IngestResult{}, err
	}

	tags := u.info.Tags
	var cover sql.NullString
	if p := tags.Picture; p != nil {
		// A broken embedded picture is not worth rejecting the song for.
		if name, err := s.saveCover(ctx, p.Data); err == nil {
			cover = sql.NullString{String: name, Valid: true}
		} else {
			slog.Warn("ignore embedded cover", "file", u.name, "err", err)
		}
	}
	in := TrackInput{
		Title:        tags.Title,
		Artists:      truncateAll(tags.Artists),
		Album:        tags.Album,
		AlbumArtists: truncateAll(tags.AlbumArtists),
		TrackNo:      min(max(tags.Track, 0), 9999),
		DiscNo:       min(max(tags.Disc, 0), 999),
		Year:         tags.Year,
	}
	if in.Title == "" {
		in.Title = strings.TrimSuffix(u.name, filepath.Ext(u.name))
	}
	in.Title = truncate(in.Title)
	in.Album = truncate(in.Album)
	if err := in.normalize(); err != nil {
		in.Title = "未命名"
		if err := in.normalize(); err != nil {
			os.Remove(dst)
			return IngestResult{}, err
		}
	}

	var d TrackDetail
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		var artists []int64
		var album sql.NullInt64
		var err error
		switch {
		case ids.FromAlbum != nil:
			artists, album, err = albumImportRefs(ctx, tx, in, ids)
		case ids.Song != 0:
			artists, album, err = newImportedAlbum(ctx, tx, in, ids, cover)
		default:
			artists, album, err = resolveTrackRefs(ctx, tx, in)
		}
		if err != nil {
			return err
		}
		trackCover := cover
		if (ids.Song == 0 || ids.FromAlbum != nil) && album.Valid && cover.Valid {
			// The first embedded picture of an album becomes its cover; a
			// track keeps its own only when the picture differs.
			var albumCover sql.NullString
			if err := tx.QueryRowContext(ctx, `SELECT cover FROM albums WHERE id = ?`, album).Scan(&albumCover); err != nil {
				return err
			}
			switch {
			case !albumCover.Valid:
				if _, err := tx.ExecContext(ctx, `UPDATE albums SET cover = ? WHERE id = ?`, cover, album); err != nil {
					return err
				}
				trackCover = sql.NullString{}
			case albumCover.String == cover.String:
				trackCover = sql.NullString{}
			}
		}
		var id int64
		err = tx.QueryRowContext(ctx, `
			INSERT INTO tracks (title, album_id, track_no, disc_no, year, duration_ms, format, size,
				sha256, file, original_name, cover, lyrics, lyrics_format, netease_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, 0)) RETURNING id`,
			in.Title, album, in.TrackNo, in.DiscNo, in.Year, u.info.Duration.Milliseconds(),
			string(u.info.Format), u.size, u.sha256, rel, u.name, trackCover, text, format, ids.Song).Scan(&id)
		if err != nil {
			return err
		}
		if err := setArtists(ctx, tx, "track_artists", "track_id", id, artists); err != nil {
			return err
		}
		d, err = s.track(ctx, tx, id)
		return err
	})
	if err != nil {
		os.Remove(dst)
		s.collectCovers(ctx)
		return IngestResult{}, err
	}
	return IngestResult{Status: Created, Track: d}, nil
}

// newImportedAlbum creates the album of an imported song, see importedSuffix.
func newImportedAlbum(ctx context.Context, tx *sql.Tx, in TrackInput, ids NeteaseIDs, cover sql.NullString) ([]int64, sql.NullInt64, error) {
	artists, err := artistIDs(ctx, tx, in.Artists)
	if err != nil {
		return nil, sql.NullInt64{}, err
	}
	if err := setNeteaseArtists(ctx, tx, in.Artists, ids.Artists); err != nil {
		return nil, sql.NullInt64{}, err
	}
	var album sql.NullInt64
	if in.Album == "" {
		return artists, album, nil
	}
	err = tx.QueryRowContext(ctx,
		`INSERT INTO albums (title, artist_key, year, cover, single, netease_id) VALUES (?, ?, ?, ?, 1, NULLIF(?, 0)) RETURNING id`,
		in.Album, importedKey(artists, ids.Song), in.Year, cover, ids.Album).Scan(&album)
	if err != nil {
		return nil, album, err
	}
	return artists, album, setArtists(ctx, tx, "album_artists", "album_id", album.Int64, artists)
}

// albumImportRefs resolves the artists of a song that comes with an album
// import, and the album it joins.
func albumImportRefs(ctx context.Context, tx *sql.Tx, in TrackInput, ids NeteaseIDs) ([]int64, sql.NullInt64, error) {
	artists, err := artistIDs(ctx, tx, in.Artists)
	if err != nil {
		return nil, sql.NullInt64{}, err
	}
	if err := setNeteaseArtists(ctx, tx, in.Artists, ids.Artists); err != nil {
		return nil, sql.NullInt64{}, err
	}
	id, err := importAlbum(ctx, tx, ids.FromAlbum.normalized(), ids.Artists)
	return artists, sql.NullInt64{Int64: id, Valid: err == nil}, err
}

// AudioFile is an opened track file for streaming.
type AudioFile struct {
	*os.File
	Format  media.Format
	ModTime time.Time
}

// OpenAudio opens the audio file of a track. The caller closes it.
func (s *Store) OpenAudio(ctx context.Context, id int64) (AudioFile, error) {
	var rel, format string
	err := s.db.QueryRowContext(ctx, `SELECT file, format FROM tracks WHERE id = ?`, id).Scan(&rel, &format)
	if errors.Is(err, sql.ErrNoRows) {
		return AudioFile{}, ErrNotFound
	}
	if err != nil {
		return AudioFile{}, err
	}
	f, err := os.Open(filepath.Join(s.mediaDir, rel))
	if err != nil {
		return AudioFile{}, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return AudioFile{}, err
	}
	return AudioFile{File: f, Format: media.Format(format), ModTime: st.ModTime()}, nil
}

// baseName keeps only the last element of a browser-supplied file name.
func baseName(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, `\`, "/"))
	if name == "." || name == "/" {
		return ""
	}
	return truncate(name)
}

// truncateAll fits tag names within the limits normalize enforces, so odd
// tags never reject an upload.
func truncateAll(names []string) []string {
	out := make([]string, 0, min(len(names), maxArtists))
	for _, n := range names[:min(len(names), maxArtists)] {
		out = append(out, truncate(n))
	}
	return out
}

func truncate(s string) string {
	r := []rune(s)
	if len(r) > maxNameLen {
		return string(r[:maxNameLen])
	}
	return s
}
