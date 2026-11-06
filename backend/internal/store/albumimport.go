package store

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// AlbumImport is the NetEase album that an album import fills (Eric,
// 2026-11-06). Unlike a single import, the album artists are NetEase's, so a
// compilation stays one album; each song keeps its own artists.
type AlbumImport struct {
	NeteaseID   int64
	Title       string
	Artists     []string
	Year        int
	Description string
	// Single marks an album NetEase lists with only one song.
	Single bool
}

// albumImportSuffix sets apart an album created by an album import, as
// importedSuffix does for a single; "album:" keeps the NetEase album ID from
// reading as a song ID. rekey keeps it like the other.
const albumImportSuffix = importedSuffix + "album:"

// singleImportGlob matches the artist_key of an album created for a single
// imported song.
const singleImportGlob = "*" + importedSuffix + "[0-9]*"

func (a AlbumImport) normalized() AlbumImport {
	a.Title = truncate(strings.TrimSpace(a.Title))
	var names []string
	for _, n := range truncateAll(a.Artists) {
		if n = strings.TrimSpace(n); n != "" {
			names = append(names, n)
		}
	}
	a.Artists = names
	a.Description = strings.TrimSpace(a.Description)
	if utf8.RuneCountInString(a.Description) > maxAlbumDescription {
		a.Description = string([]rune(a.Description)[:maxAlbumDescription])
	}
	a.Year = min(max(a.Year, 0), 9999)
	return a
}

type rowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// importTargetID finds the album an import of the NetEase album joins: one
// linked to it, which an earlier album import created or the admin linked,
// but not an album created for a single imported song.
func importTargetID(ctx context.Context, q rowQuerier, neteaseID int64) (int64, error) {
	var id int64
	err := q.QueryRowContext(ctx,
		`SELECT id FROM albums WHERE netease_id = ? AND artist_key NOT GLOB ? ORDER BY id LIMIT 1`,
		neteaseID, singleImportGlob).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return id, err
}

// ImportTarget returns the album that importing the NetEase album fills, or
// ErrNotFound when the import would create one.
func (s *Store) ImportTarget(ctx context.Context, neteaseID int64) (Album, error) {
	id, err := importTargetID(ctx, s.db, neteaseID)
	if err != nil {
		return Album{}, err
	}
	return s.Album(ctx, id)
}

// SameTitleAlbums counts the albums titled like the NetEase album that are
// not linked to it, such as an upload of the same release. An album import
// leaves them apart (Eric, 2026-11-06); the admin merges them.
func (s *Store) SameTitleAlbums(ctx context.Context, title string, neteaseID int64) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM albums WHERE title = ? COLLATE NOCASE AND netease_id IS NOT ?`,
		truncate(strings.TrimSpace(title)), neteaseID).Scan(&n)
	return n, err
}

// importAlbum finds or creates the album for a, filling in a year and a
// description it lacks.
func importAlbum(ctx context.Context, tx *sql.Tx, a AlbumImport, artistNetease map[string]int64) (int64, error) {
	id, err := importTargetID(ctx, tx, a.NeteaseID)
	if errors.Is(err, ErrNotFound) {
		var artists []int64
		if artists, err = artistIDs(ctx, tx, a.Artists); err != nil {
			return 0, err
		}
		if err := setNeteaseArtists(ctx, tx, a.Artists, artistNetease); err != nil {
			return 0, err
		}
		key := albumKey(artists) + albumImportSuffix + strconv.FormatInt(a.NeteaseID, 10)
		// The admin may have unlinked an album this import created.
		err = tx.QueryRowContext(ctx, `SELECT id FROM albums WHERE title = ? COLLATE NOCASE AND artist_key = ?`,
			a.Title, key).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			err = tx.QueryRowContext(ctx, `
				INSERT INTO albums (title, artist_key, year, single, description, netease_id)
				VALUES (?, ?, ?, ?, ?, ?) RETURNING id`,
				a.Title, key, a.Year, a.Single, a.Description, a.NeteaseID).Scan(&id)
			if err != nil {
				return 0, err
			}
			return id, setArtists(ctx, tx, "album_artists", "album_id", id, artists)
		}
	}
	if err != nil {
		return 0, err
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE albums SET
			year = CASE WHEN year = 0 THEN ? ELSE year END,
			description = CASE WHEN description = '' THEN ? ELSE description END,
			netease_id = IFNULL(netease_id, ?)
		WHERE id = ?`, a.Year, a.Description, a.NeteaseID, id)
	return id, err
}

// setNeteaseArtists records the NetEase IDs of the named artists. An artist
// found by name keeps the NetEase ID they already have: two NetEase artists
// can share a name, and the first import decides.
func setNeteaseArtists(ctx context.Context, tx *sql.Tx, names []string, ids map[string]int64) error {
	for _, name := range names {
		if id := ids[name]; id != 0 {
			if _, err := tx.ExecContext(ctx, `UPDATE artists SET netease_id = ? WHERE name = ? AND netease_id IS NULL`,
				id, name); err != nil {
				return err
			}
		}
	}
	return nil
}

// ImportedPlacement is where an imported NetEase song is in the library.
type ImportedPlacement struct {
	TrackID int64
	// Alone means the song is still the only track of the album created when
	// it was imported on its own, so an album import may move it.
	Alone bool
}

// ImportedPlacements returns the placement of each of the NetEase songs that
// is in the library.
func (s *Store) ImportedPlacements(ctx context.Context, neteaseIDs []int64) (map[int64]ImportedPlacement, error) {
	out := map[int64]ImportedPlacement{}
	if len(neteaseIDs) == 0 {
		return out, nil
	}
	args := make([]any, len(neteaseIDs))
	for i, id := range neteaseIDs {
		args[i] = id
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.netease_id, t.id, `+aloneSQL+`
		FROM tracks t LEFT JOIN albums al ON al.id = t.album_id
		WHERE t.netease_id IN (?`+strings.Repeat(",?", len(args)-1)+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var neteaseID int64
		var p ImportedPlacement
		if err := rows.Scan(&neteaseID, &p.TrackID, &p.Alone); err != nil {
			return nil, err
		}
		out[neteaseID] = p
	}
	return out, rows.Err()
}

// aloneSQL tells, for track t of album al, whether al is the album created
// when t was imported on its own and holds nothing else.
const aloneSQL = `IFNULL(
	al.artist_key LIKE '%` + importedSuffix + `' || t.netease_id
	AND (SELECT COUNT(*) FROM tracks WHERE album_id = al.id) = 1, 0)`

// MoveImported moves a song imported on its own into the album that an album
// import fills, at its place on it, and drops the album made for it. Only a
// song still alone in that album moves; one the admin merged elsewhere stays
// where it is. The file's tags are left as they are. It reports whether the
// song moved.
func (s *Store) MoveImported(ctx context.Context, trackID int64, a AlbumImport, artistNetease map[string]int64, trackNo, discNo int) (bool, error) {
	a = a.normalized()
	s.files.Lock()
	defer s.files.Unlock()
	moved := false
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var alone bool
		var cover sql.NullString
		err := tx.QueryRowContext(ctx, `
			SELECT `+aloneSQL+`, t.cover FROM tracks t LEFT JOIN albums al ON al.id = t.album_id
			WHERE t.id = ?`, trackID).Scan(&alone, &cover)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil || !alone {
			return err
		}
		album, err := importAlbum(ctx, tx, a, artistNetease)
		if err != nil {
			return err
		}
		// The song's cover becomes the album's when it has none, and the song
		// keeps its own only when it differs, as on upload.
		if cover.Valid {
			if _, err := tx.ExecContext(ctx, `UPDATE albums SET cover = ? WHERE id = ? AND cover IS NULL`, cover, album); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE tracks SET album_id = ?1, track_no = ?2, disc_no = ?3,
				cover = CASE WHEN cover = (SELECT cover FROM albums WHERE id = ?1) THEN NULL ELSE cover END
			WHERE id = ?4`,
			album, min(max(trackNo, 0), 9999), min(max(discNo, 0), 999), trackID); err != nil {
			return err
		}
		moved = true
		return pruneOrphans(ctx, tx)
	})
	if err == nil && moved {
		s.collectCovers(ctx)
	}
	return moved, err
}

// MissingArtists returns those of the names that no artist in the library has
// yet, as the library would store them: an import that names them creates
// them.
func (s *Store) MissingArtists(ctx context.Context, names []string) ([]string, error) {
	var missing []string
	for _, name := range names {
		name = truncate(strings.TrimSpace(name))
		if name == "" || slices.Contains(missing, name) {
			continue
		}
		var id int64
		err := s.db.QueryRowContext(ctx, `SELECT id FROM artists WHERE name = ?`, name).Scan(&id)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			missing = append(missing, name)
		case err != nil:
			return nil, err
		}
	}
	return missing, nil
}

// FillImportedArtist gives the artist an import created, found by name and
// NetEase ID, the image and description NetEase has (Eric, 2026-11-07). An
// image or description the artist already has is kept, so nothing the admin
// wrote is replaced. image may be nil; one the library cannot read is left
// out. It reports whether the artist was found.
func (s *Store) FillImportedArtist(ctx context.Context, name string, neteaseID int64, image []byte, description string) (bool, error) {
	name = truncate(strings.TrimSpace(name))
	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(description) > maxAlbumDescription {
		description = string([]rune(description)[:maxAlbumDescription])
	}
	s.files.Lock()
	defer s.files.Unlock()
	var cover sql.NullString
	if image != nil {
		if n, err := s.saveCover(ctx, image); err == nil {
			cover = sql.NullString{String: n, Valid: true}
		} else {
			slog.Warn("ignore artist image", "artist", name, "err", err)
		}
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE artists SET image = IFNULL(image, ?),
			description = CASE WHEN description = '' THEN ? ELSE description END
		WHERE name = ? AND netease_id = ?`, cover, description, name, neteaseID)
	if cover.Valid {
		s.collectCovers(ctx)
	}
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
