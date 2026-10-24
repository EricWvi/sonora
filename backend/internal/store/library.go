package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Ref names an artist.
type Ref struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// refs scans the JSON array built by artistsJSON.
type refs []Ref

func (r *refs) Scan(src any) error {
	switch v := src.(type) {
	case string:
		return json.Unmarshal([]byte(v), (*[]Ref)(r))
	case []byte:
		return json.Unmarshal(v, (*[]Ref)(r))
	}
	return fmt.Errorf("scan artists from %T", src)
}

// AlbumRef is the album a track belongs to.
type AlbumRef struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Artists []Ref  `json:"artists"`
	Cover   *Cover `json:"cover"`
}

// Track is a song in the library. Artists is empty and Album nil when no
// tag named them.
type Track struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Artists    []Ref     `json:"artists"`
	Album      *AlbumRef `json:"album"`
	TrackNo    int       `json:"trackNo"`
	DiscNo     int       `json:"discNo"`
	Year       int       `json:"year"`
	DurationMs int64     `json:"durationMs"`
	Format     string    `json:"format"`
	Size       int64     `json:"size"`
	// Cover is the track's own art, set only when it differs from the
	// album cover.
	Cover     *Cover `json:"cover"`
	HasLyrics bool   `json:"hasLyrics"`
	// LyricsFormat is LRC or TTML, empty when the track has no lyrics.
	LyricsFormat string `json:"lyricsFormat"`
	OriginalName string `json:"originalName"`
	CreatedAt    string `json:"createdAt"`
}

// TrackDetail is a track with its lyrics.
type TrackDetail struct {
	Track
	Lyrics string `json:"lyrics"`
}

// Album is a release. Artists are the album artists.
type Album struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Artists    []Ref  `json:"artists"`
	Year       int    `json:"year"`
	Cover      *Cover `json:"cover"`
	TrackCount int    `json:"trackCount"`
	DurationMs int64  `json:"durationMs"`
	CreatedAt  string `json:"createdAt"`
}

// Artist is a performer, of tracks or as an album artist. Image is
// uploaded by the admin and stored like a cover.
type Artist struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Image      *Cover `json:"image"`
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
	maxArtists   = 50
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

// --- artist links ----------------------------------------------------------

// artistsJSON selects, as a JSON array in position order, the artists that
// table (track_artists or album_artists) links to the row whose ID is ref.
func artistsJSON(table, col, ref string) string {
	return `(SELECT json_group_array(json_object('id', a.id, 'name', a.name) ORDER BY x.position)
		FROM ` + table + ` x JOIN artists a ON a.id = x.artist_id WHERE x.` + col + ` = ` + ref + `)`
}

// normalizeNames trims names and drops empty and repeated ones, keeping
// the first spelling.
func normalizeNames(names []string) ([]string, error) {
	out := []string{}
	seen := map[string]bool{}
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" || seen[strings.ToLower(n)] {
			continue
		}
		if utf8.RuneCountInString(n) > maxNameLen {
			return nil, invalid(fmt.Sprintf("名称不能超过 %d 个字符", maxNameLen))
		}
		seen[strings.ToLower(n)] = true
		out = append(out, n)
	}
	if len(out) > maxArtists {
		return nil, invalid(fmt.Sprintf("艺术家不能超过 %d 位", maxArtists))
	}
	return out, nil
}

// artistIDs finds or creates the named artists, in order and without
// repeats.
func artistIDs(ctx context.Context, tx *sql.Tx, names []string) ([]int64, error) {
	ids := []int64{}
	for _, name := range names {
		var id int64
		err := tx.QueryRowContext(ctx, `SELECT id FROM artists WHERE name = ?`, name).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			err = tx.QueryRowContext(ctx, `INSERT INTO artists (name) VALUES (?) RETURNING id`, name).Scan(&id)
		}
		if err != nil {
			return nil, err
		}
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// setArtists replaces the artists table links to the row id.
func setArtists(ctx context.Context, tx *sql.Tx, table, col string, id int64, artists []int64) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM `+table+` WHERE `+col+` = ?`, id); err != nil {
		return err
	}
	for i, a := range artists {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO `+table+` (`+col+`, artist_id, position) VALUES (?, ?, ?)`, id, a, i); err != nil {
			return err
		}
	}
	return nil
}

// albumKey identifies a set of album artists regardless of their order.
func albumKey(artists []int64) string {
	sorted := slices.Sorted(slices.Values(artists))
	parts := make([]string, len(sorted))
	for i, id := range sorted {
		parts[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(parts, ",")
}

// placeholderOf selects the ThumbHash of the cover named by col.
func placeholderOf(col string) string {
	return `(SELECT thumbhash FROM covers WHERE name = ` + col + `)`
}

// --- tracks ----------------------------------------------------------------

// TrackQuery filters the track list.
type TrackQuery struct {
	Q string
	// Missing keeps tracks without an artist or album.
	Missing  bool
	AlbumID  int64
	ArtistID int64
	Sort     Sort
	Window
}

var trackColumns = `
	t.id, t.title, t.track_no, t.disc_no, t.year, t.duration_ms, t.format, t.size,
	t.cover, ` + placeholderOf("t.cover") + `, t.lyrics <> '',
	CASE WHEN t.lyrics = '' THEN '' ELSE t.lyrics_format END, t.original_name, t.created_at, ` +
	artistsJSON("track_artists", "track_id", "t.id") + `,
	al.id, al.title, al.cover, ` + placeholderOf("al.cover") + `, ` + artistsJSON("album_artists", "album_id", "al.id")

const trackFrom = `
	FROM tracks t
	LEFT JOIN albums al ON al.id = t.album_id`

type scanner interface{ Scan(...any) error }

func scanTrack(row scanner, extra ...any) (Track, error) {
	var t Track
	var alID sql.NullInt64
	var cover, coverHash, alTitle, alCover, alCoverHash sql.NullString
	var alArtists []Ref
	dest := []any{&t.ID, &t.Title, &t.TrackNo, &t.DiscNo, &t.Year, &t.DurationMs, &t.Format, &t.Size,
		&cover, &coverHash, &t.HasLyrics, &t.LyricsFormat, &t.OriginalName, &t.CreatedAt, (*refs)(&t.Artists),
		&alID, &alTitle, &alCover, &alCoverHash, (*refs)(&alArtists)}
	if err := row.Scan(append(dest, extra...)...); err != nil {
		return t, err
	}
	t.Cover = coverOf(cover, coverHash)
	if alID.Valid {
		t.Album = &AlbumRef{ID: alID.Int64, Title: alTitle.String, Artists: alArtists, Cover: coverOf(alCover, alCoverHash)}
	}
	return t, nil
}

// Tracks lists tracks in q.Sort order; by default an album's in disc and
// track order, otherwise newest first.
func (s *Store) Tracks(ctx context.Context, q TrackQuery) (Page[Track], error) {
	page := Page[Track]{Items: []Track{}}
	natural := `t.id DESC`
	if q.AlbumID > 0 {
		natural = `t.disc_no, t.track_no, t.id`
	}
	order, err := q.Sort.orderBy(trackSortKeys, natural)
	if err != nil {
		return page, err
	}
	var where []string
	var args []any
	if q.Q = strings.TrimSpace(q.Q); q.Q != "" {
		where = append(where, `(t.title LIKE ? ESCAPE '\' OR al.title LIKE ? ESCAPE '\' OR EXISTS (
			SELECT 1 FROM track_artists x JOIN artists a ON a.id = x.artist_id
			WHERE x.track_id = t.id AND a.name LIKE ? ESCAPE '\'))`)
		p := like(q.Q)
		args = append(args, p, p, p)
	}
	if q.Missing {
		where = append(where, `(t.album_id IS NULL OR NOT EXISTS (SELECT 1 FROM track_artists WHERE track_id = t.id))`)
	}
	if q.AlbumID > 0 {
		where = append(where, `t.album_id = ?`)
		args = append(args, q.AlbumID)
	}
	if q.ArtistID > 0 {
		where = append(where, `(EXISTS (SELECT 1 FROM track_artists WHERE track_id = t.id AND artist_id = ?)
			OR EXISTS (SELECT 1 FROM album_artists WHERE album_id = t.album_id AND artist_id = ?))`)
		args = append(args, q.ArtistID, q.ArtistID)
	}
	cond := ""
	if len(where) > 0 {
		cond = " WHERE " + strings.Join(where, " AND ")
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*)`+trackFrom+cond, args...).Scan(&page.Total); err != nil {
		return page, err
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

// TrackInput is the editable metadata of a track. Empty Artists or Album
// leave the track without them; empty AlbumArtists use Artists.
type TrackInput struct {
	Title        string   `json:"title"`
	Artists      []string `json:"artists"`
	Album        string   `json:"album"`
	AlbumArtists []string `json:"albumArtists"`
	TrackNo      int      `json:"trackNo"`
	DiscNo       int      `json:"discNo"`
	Year         int      `json:"year"`
}

func (in *TrackInput) normalize() error {
	in.Title = strings.TrimSpace(in.Title)
	in.Album = strings.TrimSpace(in.Album)
	if in.Title == "" {
		return invalid("标题不能为空")
	}
	if utf8.RuneCountInString(in.Title) > maxNameLen || utf8.RuneCountInString(in.Album) > maxNameLen {
		return invalid(fmt.Sprintf("名称不能超过 %d 个字符", maxNameLen))
	}
	var err error
	if in.Artists, err = normalizeNames(in.Artists); err != nil {
		return err
	}
	if in.AlbumArtists, err = normalizeNames(in.AlbumArtists); err != nil {
		return err
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
		artists, album, err := resolveTrackRefs(ctx, tx, in)
		if err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, `
			UPDATE tracks SET title = ?, album_id = ?, track_no = ?, disc_no = ?, year = ?,
				updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
			WHERE id = ?`, in.Title, album, in.TrackNo, in.DiscNo, in.Year, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		if err := setArtists(ctx, tx, "track_artists", "track_id", id, artists); err != nil {
			return err
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

// resolveTrackRefs finds or creates the artists and album the input names.
func resolveTrackRefs(ctx context.Context, tx *sql.Tx, in TrackInput) (artists []int64, album sql.NullInt64, err error) {
	if artists, err = artistIDs(ctx, tx, in.Artists); err != nil {
		return
	}
	albumArtists := artists
	if len(in.AlbumArtists) > 0 {
		if albumArtists, err = artistIDs(ctx, tx, in.AlbumArtists); err != nil {
			return
		}
	}
	album, err = albumID(ctx, tx, in.Album, albumArtists, in.Year)
	return
}

func albumID(ctx context.Context, tx *sql.Tx, title string, artists []int64, year int) (sql.NullInt64, error) {
	var id sql.NullInt64
	if title == "" {
		return id, nil
	}
	err := tx.QueryRowContext(ctx,
		`SELECT id FROM albums WHERE title = ? COLLATE NOCASE AND artist_key = ?`,
		title, albumKey(artists)).Scan(&id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		err = tx.QueryRowContext(ctx, `INSERT INTO albums (title, artist_key, year) VALUES (?, ?, ?) RETURNING id`,
			title, albumKey(artists), year).Scan(&id)
		if err == nil {
			err = setArtists(ctx, tx, "album_artists", "album_id", id.Int64, artists)
		}
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
		WHERE NOT EXISTS (SELECT 1 FROM track_artists WHERE artist_id = artists.id)
		  AND NOT EXISTS (SELECT 1 FROM album_artists WHERE artist_id = artists.id);`)
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

// AlbumQuery filters the album list. Sort accepts SortCreateTime besides
// the default order by title.
type AlbumQuery struct {
	Q        string
	ArtistID int64
	Sort     Sort
	Window
}

var albumSelect = `
	SELECT al.id, al.title, al.year, al.cover, ` + placeholderOf("al.cover") + `, al.created_at, ` +
	artistsJSON("album_artists", "album_id", "al.id") + `,
		(SELECT COUNT(*) FROM tracks WHERE album_id = al.id),
		(SELECT IFNULL(SUM(duration_ms), 0) FROM tracks WHERE album_id = al.id)
	FROM albums al`

func scanAlbum(row scanner) (Album, error) {
	var a Album
	var cover, coverHash sql.NullString
	err := row.Scan(&a.ID, &a.Title, &a.Year, &cover, &coverHash, &a.CreatedAt, (*refs)(&a.Artists), &a.TrackCount, &a.DurationMs)
	a.Cover = coverOf(cover, coverHash)
	return a, err
}

// Albums lists albums, by title unless q.Sort says otherwise.
func (s *Store) Albums(ctx context.Context, q AlbumQuery) (Page[Album], error) {
	page := Page[Album]{Items: []Album{}}
	order, err := q.Sort.orderBy(albumSortKeys, `al.title COLLATE `+sortCollation+`, al.id`)
	if err != nil {
		return page, err
	}
	var where []string
	var args []any
	if q.Q = strings.TrimSpace(q.Q); q.Q != "" {
		where = append(where, `(al.title LIKE ? ESCAPE '\' OR EXISTS (
			SELECT 1 FROM album_artists x JOIN artists a ON a.id = x.artist_id
			WHERE x.album_id = al.id AND a.name LIKE ? ESCAPE '\'))`)
		p := like(q.Q)
		args = append(args, p, p)
	}
	if q.ArtistID > 0 {
		where = append(where, `EXISTS (SELECT 1 FROM album_artists WHERE album_id = al.id AND artist_id = ?)`)
		args = append(args, q.ArtistID)
	}
	cond := ""
	if len(where) > 0 {
		cond = " WHERE " + strings.Join(where, " AND ")
	}
	err = s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM albums al`+cond, args...).Scan(&page.Total)
	if err != nil {
		return page, err
	}
	w := q.Window.clamp()
	rows, err := s.db.QueryContext(ctx, albumSelect+cond+order+` LIMIT ? OFFSET ?`,
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

// AlbumInput is the editable metadata of an album. Artists are the album
// artists' names; empty leaves the album without any.
type AlbumInput struct {
	Title   string   `json:"title"`
	Artists []string `json:"artists"`
	Year    int      `json:"year"`
}

// UpdateAlbum replaces an album's metadata. Renaming onto another album of
// the same artists is a conflict: merge them instead.
func (s *Store) UpdateAlbum(ctx context.Context, id int64, in AlbumInput) (Album, error) {
	in.Title = strings.TrimSpace(in.Title)
	switch {
	case in.Title == "":
		return Album{}, invalid("专辑名不能为空")
	case utf8.RuneCountInString(in.Title) > maxNameLen:
		return Album{}, invalid(fmt.Sprintf("名称不能超过 %d 个字符", maxNameLen))
	case in.Year < 0 || in.Year > 9999:
		return Album{}, invalid("年份超出范围")
	}
	var err error
	if in.Artists, err = normalizeNames(in.Artists); err != nil {
		return Album{}, err
	}
	s.files.Lock()
	defer s.files.Unlock()
	var a Album
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		artists, err := artistIDs(ctx, tx, in.Artists)
		if err != nil {
			return err
		}
		var other int64
		err = tx.QueryRowContext(ctx,
			`SELECT id FROM albums WHERE title = ? COLLATE NOCASE AND artist_key = ? AND id <> ?`,
			in.Title, albumKey(artists), id).Scan(&other)
		if err == nil {
			return conflict("该艺术家已有同名专辑，请改用合并")
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		res, err := tx.ExecContext(ctx, `UPDATE albums SET title = ?, artist_key = ?, year = ? WHERE id = ?`,
			in.Title, albumKey(artists), in.Year, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrNotFound
		}
		if err := setArtists(ctx, tx, "album_artists", "album_id", id, artists); err != nil {
			return err
		}
		if err := pruneOrphans(ctx, tx); err != nil {
			return err
		}
		a, err = s.album(ctx, tx, id)
		return err
	})
	if err == nil {
		s.collectCovers(ctx) // a pruned artist may leave its image behind
	}
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

var artistSelect = `
	SELECT a.id, a.name, a.image, ` + placeholderOf("a.image") + `, a.created_at,
		(SELECT COUNT(*) FROM track_artists WHERE artist_id = a.id),
		(SELECT COUNT(*) FROM album_artists WHERE artist_id = a.id)
	FROM artists a`

func scanArtist(row scanner) (Artist, error) {
	var a Artist
	var image, imageHash sql.NullString
	err := row.Scan(&a.ID, &a.Name, &image, &imageHash, &a.CreatedAt, &a.TrackCount, &a.AlbumCount)
	a.Image = coverOf(image, imageHash)
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
	rows, err := s.db.QueryContext(ctx, artistSelect+cond+` ORDER BY a.name COLLATE `+sortCollation+`, a.id LIMIT ? OFFSET ?`,
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
// and deletes the sources. The target keeps its own image when set. An album whose artists then match another
// album of the same title is merged into it.
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
			// Where target is already linked, the source link just goes.
			for _, table := range []string{"track_artists", "album_artists"} {
				if _, err := tx.ExecContext(ctx, `UPDATE OR IGNORE `+table+` SET artist_id = ? WHERE artist_id = ?`,
					target, src); err != nil {
					return err
				}
				if _, err := tx.ExecContext(ctx, `DELETE FROM `+table+` WHERE artist_id = ?`, src); err != nil {
					return err
				}
			}
			if _, err := tx.ExecContext(ctx, `UPDATE artists SET image = IFNULL(image, (SELECT image FROM artists WHERE id = ?))
				WHERE id = ?`, src, target); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM artists WHERE id = ?`, src); err != nil {
				return err
			}
		}
		if err := rekeyAlbums(ctx, tx, target); err != nil {
			return err
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

// rekeyAlbums recomputes artist_key for the albums of artist, merging an
// album into another of the same title whose artists it now matches.
func rekeyAlbums(ctx context.Context, tx *sql.Tx, artist int64) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT al.id, al.title, (SELECT group_concat(artist_id) FROM album_artists WHERE album_id = al.id)
		FROM albums al WHERE EXISTS (SELECT 1 FROM album_artists WHERE album_id = al.id AND artist_id = ?)`, artist)
	if err != nil {
		return err
	}
	type album struct {
		id      int64
		title   string
		artists string
	}
	var albums []album
	for rows.Next() {
		var a album
		if err := rows.Scan(&a.id, &a.title, &a.artists); err != nil {
			rows.Close()
			return err
		}
		albums = append(albums, a)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, a := range albums {
		var ids []int64
		for _, part := range strings.Split(a.artists, ",") {
			id, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return err
			}
			ids = append(ids, id)
		}
		key := albumKey(ids)
		var into int64
		err := tx.QueryRowContext(ctx, `SELECT id FROM albums WHERE title = ? COLLATE NOCASE AND artist_key = ? AND id <> ?`,
			a.title, key, a.id).Scan(&into)
		switch {
		case err == nil:
			err = mergeAlbumInto(ctx, tx, into, a.id)
		case errors.Is(err, sql.ErrNoRows):
			_, err = tx.ExecContext(ctx, `UPDATE albums SET artist_key = ? WHERE id = ?`, key, a.id)
		}
		if err != nil {
			return err
		}
	}
	return nil
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
			(SELECT COUNT(*) FROM tracks t
				WHERE t.album_id IS NULL OR NOT EXISTS (SELECT 1 FROM track_artists WHERE track_id = t.id)),
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
