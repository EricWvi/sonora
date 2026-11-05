package store

import (
	"sync"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	sqlite "modernc.org/sqlite"
)

// sortCollation orders names the way SPlayer's song list does with
// Intl.Collator("zh-CN", {numeric: true}): Chinese by pinyin, and digit
// runs by their value, so "2" comes before "10".
const sortCollation = "zh_numeric"

func init() {
	c := collate.New(language.SimplifiedChinese, collate.Numeric)
	// A Collator keeps scratch buffers and is not safe for concurrent use.
	var mu sync.Mutex
	sqlite.MustRegisterCollationUtf8(sortCollation, func(a, b string) int {
		mu.Lock()
		defer mu.Unlock()
		return c.CompareString(a, b)
	})
}

// Sort fields of the track lists, named after SPlayer's sortFieldOptions.
const (
	SortDefault    = "default"
	SortTitle      = "title"
	SortArtist     = "artist"
	SortAlbum      = "album"
	SortDuration   = "duration"
	SortCreateTime = "createTime"
)

// trackSortKeys are the SQL expressions the fields sort by. A track sorts
// by its first artist, as in SPlayer; one without artist or album sorts
// as an empty name.
var trackSortKeys = map[string]string{
	SortTitle: `t.title COLLATE ` + sortCollation,
	SortArtist: `IFNULL((SELECT a.name FROM track_artists x JOIN artists a ON a.id = x.artist_id
		WHERE x.track_id = t.id ORDER BY x.position LIMIT 1), '') COLLATE ` + sortCollation,
	SortAlbum:      `IFNULL(al.title, '') COLLATE ` + sortCollation,
	SortDuration:   `t.duration_ms`,
	SortCreateTime: `t.id`,
}

// albumSortKeys are the fields the album list sorts by.
var albumSortKeys = map[string]string{
	SortCreateTime: `al.id`,
}

// artistSortKeys are the fields the artist list sorts by.
var artistSortKeys = map[string]string{
	SortCreateTime: `a.id`,
}

// Sort orders a list by Field, ascending unless Order is "desc". An empty
// Field or SortDefault keeps the list's own order: newest first, an
// album's by disc and track, a playlist's as arranged.
type Sort struct {
	Field string
	Order string
}

// isDefault reports whether the list keeps its own order.
func (s Sort) isDefault() bool {
	return s.Field == "" || s.Field == SortDefault
}

// orderBy returns the ORDER BY clause for keys, falling back to the list's
// own order natural, which also breaks ties so equal rows keep it.
func (s Sort) orderBy(keys map[string]string, natural string) (string, error) {
	dir := ""
	switch s.Order {
	case "", "asc":
		dir = " ASC"
	case "desc":
		dir = " DESC"
	default:
		return "", invalid("排序方向只支持 asc 或 desc")
	}
	if s.isDefault() {
		return ` ORDER BY ` + natural, nil
	}
	key, ok := keys[s.Field]
	if !ok {
		return "", invalid("不支持按 " + s.Field + " 排序")
	}
	return ` ORDER BY ` + key + dir + `, ` + natural, nil
}
