package store

import (
	"context"
	"database/sql"
	"encoding/hex"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// LibraryChanges is one page of the library change log after a cursor:
// the tracks, albums and artists that changed, and the IDs of those
// deleted. Each item appears once, in its latest state.
type LibraryChanges struct {
	// Cursor is the seq to ask from next; it stays at the asked one when
	// nothing changed.
	Cursor int64
	// More is set when changes after Cursor remain.
	More bool
	// Reset is set when the asked cursor is ahead of the log, as after the
	// database was restored from a backup: the client drops its copy and
	// syncs again from 0. Nothing else is set then.
	Reset   bool
	Tracks  []Track
	Albums  []Album
	Artists []Artist
	Deleted DeletedItems
}

// DeletedItems are the IDs of deleted library items.
type DeletedItems struct {
	Tracks  []int64 `json:"tracks"`
	Albums  []int64 `json:"albums"`
	Artists []int64 `json:"artists"`
}

// Changes returns up to limit changes after seq since, oldest first; see
// migration 16 for what counts as a change.
func (s *Store) Changes(ctx context.Context, since int64, limit int) (LibraryChanges, error) {
	out := LibraryChanges{Cursor: since, Tracks: []Track{}, Albums: []Album{}, Artists: []Artist{},
		Deleted: DeletedItems{Tracks: []int64{}, Albums: []int64{}, Artists: []int64{}}}
	limit = Window{Limit: limit}.clamp().Limit
	// One read transaction, so the log and the items agree.
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return out, err
	}
	defer tx.Rollback()

	var last int64
	if err := tx.QueryRowContext(ctx, `SELECT IFNULL((SELECT seq FROM sqlite_sequence WHERE name = 'library_changes'), 0)`).
		Scan(&last); err != nil {
		return out, err
	}
	if since > last {
		out.Reset = true
		return out, nil
	}
	rows, err := tx.QueryContext(ctx, `SELECT seq, kind, item_id FROM library_changes WHERE seq > ? ORDER BY seq LIMIT ?`,
		since, limit+1)
	if err != nil {
		return out, err
	}
	ids := map[string][]int64{}
	n := 0
	for rows.Next() {
		var seq, id int64
		var kind string
		if err := rows.Scan(&seq, &kind, &id); err != nil {
			rows.Close()
			return out, err
		}
		if n++; n > limit {
			out.More = true
			break
		}
		out.Cursor = seq
		ids[kind] = append(ids[kind], id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return out, err
	}

	if err := changedItems(ctx, tx, `SELECT `+trackColumns+trackFrom+` WHERE t.id IN `, ids["track"],
		func(r scanner) (int64, error) {
			t, err := scanTrack(r)
			out.Tracks = append(out.Tracks, t)
			return t.ID, err
		}, &out.Deleted.Tracks); err != nil {
		return out, err
	}
	if err := changedItems(ctx, tx, albumSelect+` WHERE al.id IN `, ids["album"],
		func(r scanner) (int64, error) {
			a, err := scanAlbum(r)
			out.Albums = append(out.Albums, a)
			return a.ID, err
		}, &out.Deleted.Albums); err != nil {
		return out, err
	}
	err = changedItems(ctx, tx, artistSelect+` WHERE a.id IN `, ids["artist"],
		func(r scanner) (int64, error) {
			a, err := scanArtist(r)
			out.Artists = append(out.Artists, a)
			return a.ID, err
		}, &out.Deleted.Artists)
	return out, err
}

// changedItems scans the items of ids that query finds and adds the rest
// to deleted, keeping the order of ids.
func changedItems(ctx context.Context, tx *sql.Tx, query string, ids []int64,
	scan func(scanner) (int64, error), deleted *[]int64) error {
	if len(ids) == 0 {
		return nil
	}
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	rows, err := tx.QueryContext(ctx, query+`(?`+strings.Repeat(`, ?`, len(ids)-1)+`)`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	found := map[int64]bool{}
	for rows.Next() {
		id, err := scan(rows)
		if err != nil {
			return err
		}
		found[id] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, id := range ids {
		if !found[id] {
			*deleted = append(*deleted, id)
		}
	}
	return nil
}

// SortKeys turns names into keys that order, compared as strings, the way
// the lists do with the zh_numeric collation. A SortKeys is not safe for
// concurrent use.
type SortKeys struct {
	c   *collate.Collator
	buf collate.Buffer
}

// NewSortKeys returns the keys of the zh_numeric collation.
func NewSortKeys() *SortKeys {
	return &SortKeys{c: collate.New(language.SimplifiedChinese, collate.Numeric)}
}

// Key is name's sort key, in lowercase hex so that the byte order of the
// collation key survives as string order.
func (k *SortKeys) Key(name string) string {
	key := hex.EncodeToString(k.c.KeyFromString(&k.buf, name))
	k.buf.Reset()
	return key
}
