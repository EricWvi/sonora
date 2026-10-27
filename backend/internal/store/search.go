package store

import (
	"fmt"
	"strings"
)

// escapeLike escapes LIKE wildcards for ESCAPE '\'.
func escapeLike(q string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(q)
}

// like turns a search phrase into a LIKE pattern with ESCAPE '\'.
func like(q string) string {
	return "%" + escapeLike(q) + "%"
}

// searchField is text a search matches: col, or, when wrap is set, col
// inside an EXISTS subquery with one %s for the condition, for names
// reached through a link table.
type searchField struct {
	col, wrap string
}

func (f searchField) like() string {
	cond := f.col + ` LIKE ? ESCAPE '\'`
	if f.wrap == "" {
		return cond
	}
	return fmt.Sprintf(f.wrap, cond)
}

var (
	trackArtistField = searchField{`a.name`, `EXISTS (SELECT 1 FROM track_artists x JOIN artists a ON a.id = x.artist_id
		WHERE x.track_id = t.id AND %s)`}
	albumArtistField = searchField{`a.name`, `EXISTS (SELECT 1 FROM album_artists x JOIN artists a ON a.id = x.artist_id
		WHERE x.album_id = al.id AND %s)`}
)

// searchCond matches rows where any field contains q.
func searchCond(q string, tiers ...[]searchField) (string, []any) {
	var conds []string
	var args []any
	for _, tier := range tiers {
		for _, f := range tier {
			conds = append(conds, f.like())
			args = append(args, like(q))
		}
	}
	return "(" + strings.Join(conds, " OR ") + ")", args
}

// searchOrder orders search results by relevance, then by the list's own
// order natural.
func searchOrder(q, natural string, tiers ...[]searchField) (string, []any) {
	rank, args := relevance(q, tiers...)
	return ` ORDER BY ` + rank + `, ` + natural, args
}

// relevance ranks rows that match q, for ORDER BY: tiers are groups of
// fields from most to least important, and within a tier the text equal
// to q comes first, then text starting with q, then text containing it.
// LIKE makes all three case-insensitive for ASCII, as the search is.
// Rows matching nothing rank last.
func relevance(q string, tiers ...[]searchField) (string, []any) {
	var b strings.Builder
	var args []any
	b.WriteString("CASE")
	e := escapeLike(q)
	rank := 0
	for _, tier := range tiers {
		for _, pattern := range []string{e, e + "%", "%" + e + "%"} {
			var conds []string
			for _, f := range tier {
				conds = append(conds, f.like())
				args = append(args, pattern)
			}
			fmt.Fprintf(&b, " WHEN %s THEN %d", strings.Join(conds, " OR "), rank)
			rank++
		}
	}
	fmt.Fprintf(&b, " ELSE %d END", rank)
	return b.String(), args
}
