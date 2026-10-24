package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// Ref names an artist.
type Ref struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// AlbumRef is the album a track belongs to.
type AlbumRef struct {
	ID     int64   `json:"id"`
	Title  string  `json:"title"`
	Artist *Ref    `json:"artist"`
	Cover  *string `json:"cover"`
}

// Track is a song in the library. Artist and Album are nil when no tag
// named them.
type Track struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Artist     *Ref      `json:"artist"`
	Album      *AlbumRef `json:"album"`
	TrackNo    int       `json:"trackNo"`
	DiscNo     int       `json:"discNo"`
	Year       int       `json:"year"`
	DurationMs int64     `json:"durationMs"`
	Format     string    `json:"format"`
	Size       int64     `json:"size"`
	// Cover is the track's own art, set only when it differs from the
	// album cover.
	Cover        *string `json:"cover"`
	HasLyrics    bool    `json:"hasLyrics"`
	OriginalName string  `json:"originalName"`
	CreatedAt    string  `json:"createdAt"`
}

// TrackDetail is a track with its lyrics.
type TrackDetail struct {
	Track
	Lyrics string `json:"lyrics"`
}

// Album is a release. Artist is the album artist.
type Album struct {
	ID         int64   `json:"id"`
	Title      string  `json:"title"`
	Artist     *Ref    `json:"artist"`
	Year       int     `json:"year"`
	Cover      *string `json:"cover"`
	TrackCount int     `json:"trackCount"`
	DurationMs int64   `json:"durationMs"`
	CreatedAt  string  `json:"createdAt"`
}

// Artist is a performer, either of tracks or as an album artist.
type Artist struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	TrackCount int    `json:"trackCount"`
	AlbumCount int    `json:"albumCount"`
	CreatedAt  string `json:"createdAt"`
}

// Page is one slice of a list and the size of the whole list.
type Page[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

// Window is the requested slice of a list.
type Window struct {
	Offset int
	Limit  int
}

const (
	defaultLimit = 50
	maxLimit     = 500
	maxNameLen   = 300
)

func (w Window) clamp() Window {
	if w.Offset < 0 {
		w.Offset = 0
	}
	if w.Limit <= 0 {
		w.Limit = defaultLimit
	}
	w.Limit = min(w.Limit, maxLimit)
	return w
}

// like turns a search phrase into a LIKE pattern with ESCAPE '\'.
func like(q string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(q) + "%"
}

// --- tracks ----------------------------------------------------------------

// TrackQuery filters the track list.
type TrackQuery struct {
	Q string
	// Missing keeps tracks without an artist or album.
	Missing  bool
	AlbumID  int64
	ArtistID int64
	Window
}

const trackColumns = `
	t.id, t.title, t.track_no, t.disc_no, t.year, t.duration_ms, t.format, t.size,
	t.cover, t.lyrics <> '', t.original_name, t.created_at,
	ar.id, ar.name, al.id, al.title, al.cover, aa.id, aa.name`

const trackFrom = `
	FROM tracks t
	LEFT JOIN artists ar ON ar.id = t.artist_id
	LEFT JOIN albums al ON al.id = t.album_id
	LEFT JOIN artists aa ON aa.id = al.artist_id`

type scanner interface{ Scan(...any) error }

func scanTrack(row scanner, extra ...any) (Track, error) {
	var t Track
	var arID, alID, aaID sql.NullInt64
	var arName, alTitle, aaName, alCover sql.NullString
	dest := []any{&t.ID, &t.Title, &t.TrackNo, &t.DiscNo, &t.Year, &t.DurationMs, &t.Format, &t.Size,
		&t.Cover, &t.HasLyrics, &t.OriginalName, &t.CreatedAt,
		&arID, &arName, &alID, &alTitle, &alCover, &aaID, &aaName}
	if err := row.Scan(append(dest, extra...)...); err != nil {
		return t, err
	}
	if arID.Valid {
		t.Artist = &Ref{ID: arID.Int64, Name: arName.String}
	}
	if alID.Valid {
		t.Album = &AlbumRef{ID: alID.Int64, Title: alTitle.String}
		if alCover.Valid {
			t.Album.Cover = &alCover.String
		}
		if aaID.Valid {
			t.Album.Artist = &Ref{ID: aaID.Int64, Name: aaName.String}
		}
	}
	return t, nil
}

// Tracks lists tracks: an album's in disc and track order, otherwise
// newest first.
func (s *Store) Tracks(ctx context.Context, q TrackQuery) (Page[Track], error) {
	page := Page[Track]{Items: []Track{}}
	var where []string
	var args []any
	if q.Q = strings.TrimSpace(q.Q); q.Q != "" {
		where = append(where, `(t.title LIKE ? ESCAPE '\' OR ar.name LIKE ? ESCAPE '\' OR al.title LIKE ? ESCAPE '\')`)
		p := like(q.Q)
		args = append(args, p, p, p)
	}
	if q.Missing {
		where = append(where, `(t.artist_id IS NULL OR t.album_id IS NULL)`)
	}
	if q.AlbumID > 0 {
		where = append(where, `t.album_id = ?`)
		args = append(args, q.AlbumID)
	}
	if q.ArtistID > 0 {
		where = append(where, `(t.artist_id = ? OR al.artist_id = ?)`)
		args = append(args, q.ArtistID, q.ArtistID)
	}
	cond := ""
	if len(where) > 0 {
		cond = " WHERE " + strings.Join(where, " AND ")
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*)`+trackFrom+cond, args...).Scan(&page.Total); err != nil {
		return page, err
	}
	order := ` ORDER BY t.id DESC`
	if q.AlbumID > 0 {
		order = ` ORDER BY t.disc_no, t.track_no, t.id`
	}
	w := q.Window.clamp()
	rows, err := s.db.QueryContext(ctx, `SELECT `+trackColumns+trackFrom+cond+order+` LIMIT ? OFFSET ?`,
		append(args, w.Limit, w.Offset)...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		t, err := scanTrack(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, t)
	}
	return page, rows.Err()
}

// Track returns one track with its lyrics.
func (s *Store) Track(ctx context.Context, id int64) (TrackDetail, error) {
	return s.track(ctx, s.db, id)
}

type querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (s *Store) track(ctx context.Context, q querier, id int64) (TrackDetail, error) {
	var d TrackDetail
	t, err := scanTrack(q.QueryRowContext(ctx, `SELECT `+trackColumns+`, t.lyrics`+trackFrom+` WHERE t.id = ?`, id), &d.Lyrics)
	if errors.Is(err, sql.ErrNoRows) {
		return d, ErrNotFound
	}
	d.Track = t
	return d, err
}

// TrackInput is the editable metadata of a track. An empty Artist or
// Album leaves the track without one; an empty AlbumArtist uses Artist.
type TrackInput struct {
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	AlbumArtist string `json:"albumArtist"`
	TrackNo     int    `json:"trackNo"`
	DiscNo      int    `json:"discNo"`
	Year        int    `json:"year"`
}

func (in *TrackInput) normalize() error {
	in.Title = strings.TrimSpace(in.Title)
	in.Artist = strings.TrimSpace(in.Artist)
	in.Album = strings.TrimSpace(in.Album)
	in.AlbumArtist = strings.TrimSpace(in.AlbumArtist)
	if in.Title == "" {
		return invalid("标题不能为空")
	}
	for _, v := range []string{in.Title, in.Artist, in.Album, in.AlbumArtist} {
		if utf8.RuneCountInString(v) > maxNameLen {
			return invalid(fmt.Sprintf("名称不能超过 %d 个字符", maxNameLen))
		}
	}
	if in.TrackNo < 0 || in.TrackNo > 9999 || in.DiscNo < 0 || in.DiscNo > 999 {
		return invalid("音轨号或碟号超出范围")
	}
	if in.Year < 0 || in.Year > 9999 {
		return invalid("年份超出范围")
	}
	return nil
}

// UpdateTrack replaces a track's metadata, creating the named artists and
// album when they do not exist yet.
func (s *Store) UpdateTrack(ctx context.Context, id int64, in TrackInput) (TrackDetail, error) {
	if err := in.normalize(); err != nil {
		return TrackDetail{}, err
	}
	s.files.Lock()
	defer s.files.Unlock()
	var d TrackDetail
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		artist, album, err := resolveTrackRefs(ctx, tx, in)
		if err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, `
			UPDATE tracks SET title = ?, artist_id = ?, album_id = ?, track_no = ?, disc_no = ?, year = ?,
				updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
			WHERE id = ?`, in.Title, artist, album, in.TrackNo, in.DiscNo, in.Year, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		if err := pruneOrphans(ctx, tx); err != nil {
			return err
		}
		d, err = s.track(ctx, tx, id)
		return err
	})
	if err == nil {
		s.collectCovers(ctx)
	}
	return d, err
}

// resolveTrackRefs finds or creates the artist and album the input names.
func resolveTrackRefs(ctx context.Context, tx *sql.Tx, in TrackInput) (artist, album sql.NullInt64, err error) {
	if artist, err = artistID(ctx, tx, in.Artist); err != nil {
		return
	}
	albumArtistName := in.AlbumArtist
	if albumArtistName == "" {
		albumArtistName = in.Artist
	}
	albumArtist, err := artistID(ctx, tx, albumArtistName)
	if err != nil {
		return
	}
	album, err = albumID(ctx, tx, in.Album, albumArtist, in.Year)
	return
}

func artistID(ctx context.Context, tx *sql.Tx, name string) (sql.NullInt64, error) {
	var id sql.NullInt64
	if name == "" {
		return id, nil
	}
	err := tx.QueryRowContext(ctx, `SELECT id FROM artists WHERE name = ?`, name).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `INSERT INTO artists (name) VALUES (?) RETURNING id`, name).Scan(&id)
	}
	return id, err
}

func albumID(ctx context.Context, tx *sql.Tx, title string, artist sql.NullInt64, year int) (sql.NullInt64, error) {
	var id sql.NullInt64
	if title == "" {
		return id, nil
	}
	err := tx.QueryRowContext(ctx,
		`SELECT id FROM albums WHERE title = ? COLLATE NOCASE AND IFNULL(artist_id, 0) = IFNULL(?, 0)`,
		title, artist).Scan(&id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		err = tx.QueryRowContext(ctx,
			`INSERT INTO albums (title, artist_id, year) VALUES (?, ?, ?) RETURNING id`, title, artist, year).Scan(&id)
	case err == nil && year > 0:
		_, err = tx.ExecContext(ctx, `UPDATE albums SET year = ? WHERE id = ? AND year = 0`, year, id)
	}
	return id, err
}

// pruneOrphans removes albums without tracks and artists that neither
// perform a track nor own an album. The library is driven by its tracks.
func pruneOrphans(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM albums WHERE NOT EXISTS (SELECT 1 FROM tracks WHERE album_id = albums.id);
		DELETE FROM artists
		WHERE NOT EXISTS (SELECT 1 FROM tracks WHERE artist_id = artists.id)
		  AND NOT EXISTS (SELECT 1 FROM albums WHERE artist_id = artists.id);`)
	return err
}

// DeleteTrack removes a track and its audio file.
func (s *Store) DeleteTrack(ctx context.Context, id int64) error {
	return s.deleteTracks(ctx, `id = ?`, id)
}

// DeleteAlbum removes an album together with its tracks and their files.
func (s *Store) DeleteAlbum(ctx context.Context, id int64) error {
	return s.deleteTracks(ctx, `album_id = ?`, id)
}

func (s *Store) deleteTracks(ctx context.Context, cond string, id int64) error {
	s.files.Lock()
	defer s.files.Unlock()
	var files []string
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `DELETE FROM tracks WHERE `+cond+` RETURNING file`, id)
		if err != nil {
			return err
		}
		for rows.Next() {
			var f string
			if err := rows.Scan(&f); err != nil {
				rows.Close()
				return err
			}
			files = append(files, f)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		if len(files) == 0 {
			return ErrNotFound
		}
		return pruneOrphans(ctx, tx)
	})
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Remove(filepath.Join(s.mediaDir, f)); err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.Error("remove media file", "err", err)
		}
	}
	s.collectCovers(ctx)
	return nil
}

// --- albums ----------------------------------------------------------------

// AlbumQuery filters the album list.
type AlbumQuery struct {
	Q        string
	ArtistID int64
	Window
}

const albumSelect = `
	SELECT al.id, al.title, al.year, al.cover, al.created_at, aa.id, aa.name,
		(SELECT COUNT(*) FROM tracks WHERE album_id = al.id),
		(SELECT IFNULL(SUM(duration_ms), 0) FROM tracks WHERE album_id = al.id)
	FROM albums al LEFT JOIN artists aa ON aa.id = al.artist_id`

func scanAlbum(row scanner) (Album, error) {
	var a Album
	var aaID sql.NullInt64
	var aaName sql.NullString
	err := row.Scan(&a.ID, &a.Title, &a.Year, &a.Cover, &a.CreatedAt, &aaID, &aaName, &a.TrackCount, &a.DurationMs)
	if aaID.Valid {
		a.Artist = &Ref{ID: aaID.Int64, Name: aaName.String}
	}
	return a, err
}

// Albums lists albums by title.
func (s *Store) Albums(ctx context.Context, q AlbumQuery) (Page[Album], error) {
	page := Page[Album]{Items: []Album{}}
	var where []string
	var args []any
	if q.Q = strings.TrimSpace(q.Q); q.Q != "" {
		where = append(where, `(al.title LIKE ? ESCAPE '\' OR aa.name LIKE ? ESCAPE '\')`)
		p := like(q.Q)
		args = append(args, p, p)
	}
	if q.ArtistID > 0 {
		where = append(where, `al.artist_id = ?`)
		args = append(args, q.ArtistID)
	}
	cond := ""
	if len(where) > 0 {
		cond = " WHERE " + strings.Join(where, " AND ")
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM albums al LEFT JOIN artists aa ON aa.id = al.artist_id`+cond, args...).Scan(&page.Total)
	if err != nil {
		return page, err
	}
	w := q.Window.clamp()
	rows, err := s.db.QueryContext(ctx, albumSelect+cond+` ORDER BY al.title COLLATE NOCASE, al.id LIMIT ? OFFSET ?`,
		append(args, w.Limit, w.Offset)...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		a, err := scanAlbum(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, a)
	}
	return page, rows.Err()
}

// Album returns one album.
func (s *Store) Album(ctx context.Context, id int64) (Album, error) {
	return s.album(ctx, s.db, id)
}

func (s *Store) album(ctx context.Context, q querier, id int64) (Album, error) {
	a, err := scanAlbum(q.QueryRowContext(ctx, albumSelect+` WHERE al.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return a, ErrNotFound
	}
	return a, err
}

// AlbumInput is the editable metadata of an album. Artist is the album
// artist's name; empty leaves the album without one.
type AlbumInput struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Year   int    `json:"year"`
}

// UpdateAlbum replaces an album's metadata. Renaming onto another album of
// the same artist is a conflict: merge them instead.
func (s *Store) UpdateAlbum(ctx context.Context, id int64, in AlbumInput) (Album, error) {
	in.Title, in.Artist = strings.TrimSpace(in.Title), strings.TrimSpace(in.Artist)
	switch {
	case in.Title == "":
		return Album{}, invalid("专辑名不能为空")
	case utf8.RuneCountInString(in.Title) > maxNameLen || utf8.RuneCountInString(in.Artist) > maxNameLen:
		return Album{}, invalid(fmt.Sprintf("名称不能超过 %d 个字符", maxNameLen))
	case in.Year < 0 || in.Year > 9999:
		return Album{}, invalid("年份超出范围")
	}
	s.files.Lock()
	defer s.files.Unlock()
	var a Album
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		artist, err := artistID(ctx, tx, in.Artist)
		if err != nil {
			return err
		}
		var other int64
		err = tx.QueryRowContext(ctx,
			`SELECT id FROM albums WHERE title = ? COLLATE NOCASE AND IFNULL(artist_id, 0) = IFNULL(?, 0) AND id <> ?`,
			in.Title, artist, id).Scan(&other)
		if err == nil {
			return conflict("该艺术家已有同名专辑，请改用合并")
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		res, err := tx.ExecContext(ctx, `UPDATE albums SET title = ?, artist_id = ?, year = ? WHERE id = ?`,
			in.Title, artist, in.Year, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		if err := pruneOrphans(ctx, tx); err != nil {
			return err
		}
		a, err = s.album(ctx, tx, id)
		return err
	})
	return a, err
}

// MergeAlbums moves the tracks of the source albums into target and
// deletes the sources. The target keeps its own cover and year when set.
func (s *Store) MergeAlbums(ctx context.Context, target int64, sources []int64) (Album, error) {
	if err := checkMerge(target, sources); err != nil {
		return Album{}, err
	}
	s.files.Lock()
	defer s.files.Unlock()
	var a Album
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := mustExist(ctx, tx, "albums", append([]int64{target}, sources...)); err != nil {
			return err
		}
		for _, src := range sources {
			if err := mergeAlbumInto(ctx, tx, target, src); err != nil {
				return err
			}
		}
		if err := pruneOrphans(ctx, tx); err != nil {
			return err
		}
		var err error
		a, err = s.album(ctx, tx, target)
		return err
	})
	if err == nil {
		s.collectCovers(ctx)
	}
	return a, err
}

func mergeAlbumInto(ctx context.Context, tx *sql.Tx, target, src int64) error {
	if _, err := tx.ExecContext(ctx, `UPDATE tracks SET album_id = ? WHERE album_id = ?`, target, src); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE albums SET
			cover = IFNULL(cover, (SELECT cover FROM albums WHERE id = ?)),
			year = CASE WHEN year = 0 THEN (SELECT year FROM albums WHERE id = ?) ELSE year END
		WHERE id = ?`, src, src, target); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `DELETE FROM albums WHERE id = ?`, src)
	return err
}

// --- artists ---------------------------------------------------------------

const artistSelect = `
	SELECT a.id, a.name, a.created_at,
		(SELECT COUNT(*) FROM tracks WHERE artist_id = a.id),
		(SELECT COUNT(*) FROM albums WHERE artist_id = a.id)
	FROM artists a`

func scanArtist(row scanner) (Artist, error) {
	var a Artist
	err := row.Scan(&a.ID, &a.Name, &a.CreatedAt, &a.TrackCount, &a.AlbumCount)
	return a, err
}

// Artists lists artists by name.
func (s *Store) Artists(ctx context.Context, q string, w Window) (Page[Artist], error) {
	page := Page[Artist]{Items: []Artist{}}
	cond, args := "", []any{}
	if q = strings.TrimSpace(q); q != "" {
		cond, args = ` WHERE a.name LIKE ? ESCAPE '\'`, []any{like(q)}
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM artists a`+cond, args...).Scan(&page.Total); err != nil {
		return page, err
	}
	w = w.clamp()
	rows, err := s.db.QueryContext(ctx, artistSelect+cond+` ORDER BY a.name COLLATE NOCASE, a.id LIMIT ? OFFSET ?`,
		append(args, w.Limit, w.Offset)...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	for rows.Next() {
		a, err := scanArtist(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, a)
	}
	return page, rows.Err()
}

// Artist returns one artist.
func (s *Store) Artist(ctx context.Context, id int64) (Artist, error) {
	return s.artist(ctx, s.db, id)
}

func (s *Store) artist(ctx context.Context, q querier, id int64) (Artist, error) {
	a, err := scanArtist(q.QueryRowContext(ctx, artistSelect+` WHERE a.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return a, ErrNotFound
	}
	return a, err
}

// RenameArtist renames an artist. Taking another artist's name is a
// conflict: merge them instead.
func (s *Store) RenameArtist(ctx context.Context, id int64, name string) (Artist, error) {
	name = strings.TrimSpace(name)
	switch {
	case name == "":
		return Artist{}, invalid("艺术家名不能为空")
	case utf8.RuneCountInString(name) > maxNameLen:
		return Artist{}, invalid(fmt.Sprintf("名称不能超过 %d 个字符", maxNameLen))
	}
	var a Artist
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var other int64
		err := tx.QueryRowContext(ctx, `SELECT id FROM artists WHERE name = ? AND id <> ?`, name, id).Scan(&other)
		if err == nil {
			return conflict("已有同名艺术家，请改用合并")
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		res, err := tx.ExecContext(ctx, `UPDATE artists SET name = ? WHERE id = ?`, name, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		a, err = s.artist(ctx, tx, id)
		return err
	})
	return a, err
}

// MergeArtists moves the tracks and albums of the source artists to target
// and deletes the sources. Albums that then share a title with one of the
// target's albums are merged into it.
func (s *Store) MergeArtists(ctx context.Context, target int64, sources []int64) (Artist, error) {
	if err := checkMerge(target, sources); err != nil {
		return Artist{}, err
	}
	s.files.Lock()
	defer s.files.Unlock()
	var a Artist
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := mustExist(ctx, tx, "artists", append([]int64{target}, sources...)); err != nil {
			return err
		}
		for _, src := range sources {
			if _, err := tx.ExecContext(ctx, `UPDATE tracks SET artist_id = ? WHERE artist_id = ?`, target, src); err != nil {
				return err
			}
			rows, err := tx.QueryContext(ctx, `
				SELECT a.id, b.id FROM albums a
				LEFT JOIN albums b ON b.artist_id = ? AND b.title = a.title COLLATE NOCASE
				WHERE a.artist_id = ?`, target, src)
			if err != nil {
				return err
			}
			type move struct {
				album int64
				into  sql.NullInt64
			}
			var moves []move
			for rows.Next() {
				var m move
				if err := rows.Scan(&m.album, &m.into); err != nil {
					rows.Close()
					return err
				}
				moves = append(moves, m)
			}
			rows.Close()
			for _, m := range moves {
				if m.into.Valid {
					err = mergeAlbumInto(ctx, tx, m.into.Int64, m.album)
				} else {
					_, err = tx.ExecContext(ctx, `UPDATE albums SET artist_id = ? WHERE id = ?`, target, m.album)
				}
				if err != nil {
					return err
				}
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM artists WHERE id = ?`, src); err != nil {
				return err
			}
		}
		if err := pruneOrphans(ctx, tx); err != nil {
			return err
		}
		var err error
		a, err = s.artist(ctx, tx, target)
		return err
	})
	if err == nil {
		s.collectCovers(ctx)
	}
	return a, err
}

func checkMerge(target int64, sources []int64) error {
	if target <= 0 || len(sources) == 0 {
		return invalid("请选择要合并的项目")
	}
	seen := map[int64]bool{target: true}
	for _, id := range sources {
		if seen[id] {
			return invalid("合并目标与来源重复")
		}
		seen[id] = true
	}
	return nil
}

func mustExist(ctx context.Context, tx *sql.Tx, table string, ids []int64) error {
	for _, id := range ids {
		var one int
		err := tx.QueryRowContext(ctx, `SELECT 1 FROM `+table+` WHERE id = ?`, id).Scan(&one)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// --- overview --------------------------------------------------------------

// Overview summarises the library.
type Overview struct {
	Tracks      int   `json:"tracks"`
	Albums      int   `json:"albums"`
	Artists     int   `json:"artists"`
	MissingTags int   `json:"missingTags"`
	DurationMs  int64 `json:"durationMs"`
	MediaBytes  int64 `json:"mediaBytes"`
	CoverBytes  int64 `json:"coverBytes"`
}

// Overview counts the library and the storage it uses.
func (s *Store) Overview(ctx context.Context) (Overview, error) {
	var o Overview
	err := s.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM tracks),
			(SELECT COUNT(*) FROM albums),
			(SELECT COUNT(*) FROM artists),
			(SELECT COUNT(*) FROM tracks WHERE artist_id IS NULL OR album_id IS NULL),
			(SELECT IFNULL(SUM(duration_ms), 0) FROM tracks),
			(SELECT IFNULL(SUM(size), 0) FROM tracks)`).
		Scan(&o.Tracks, &o.Albums, &o.Artists, &o.MissingTags, &o.DurationMs, &o.MediaBytes)
	if err != nil {
		return o, err
	}
	entries, err := os.ReadDir(s.coverDir)
	if err != nil {
		return o, err
	}
	for _, e := range entries {
		if info, err := e.Info(); err == nil && info.Mode().IsRegular() {
			o.CoverBytes += info.Size()
		}
	}
	return o, nil
}
