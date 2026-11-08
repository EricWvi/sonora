package server

import (
	"net/http"
	"strconv"

	"sonora/internal/admin"
	"sonora/internal/store"
)

// The native clients keep a copy of the library and update it from the
// change log. Items have the player's shapes, plus what the client needs
// to filter and sort its copy as the lists do: the album's single mark,
// and sort keys for the names.

type syncTrack struct {
	playerTrack
	// SortKey orders titles as the lists' zh_numeric collation does.
	SortKey string `json:"sortKey"`
}

type syncAlbum struct {
	playerAlbum
	Single  bool   `json:"single"`
	SortKey string `json:"sortKey"`
}

type syncArtist struct {
	playerArtist
	SortKey string `json:"sortKey"`
}

type libraryChanges struct {
	Cursor  int64              `json:"cursor"`
	More    bool               `json:"more"`
	Reset   bool               `json:"reset"`
	Tracks  []syncTrack        `json:"tracks"`
	Albums  []syncAlbum        `json:"albums"`
	Artists []syncArtist       `json:"artists"`
	Deleted store.DeletedItems `json:"deleted"`
}

func (s *server) listLibraryChanges(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var since int64
	if v := q.Get("since"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < 0 {
			admin.WriteError(w, http.StatusBadRequest, "invalid", "since 必须是非负整数")
			return
		}
		since = n
	}
	c, err := s.store.Changes(r.Context(), since, int(queryInt(q.Get("limit"))))
	if err != nil {
		respond(w, r, nil, err)
		return
	}
	keys := store.NewSortKeys()
	out := libraryChanges{
		Cursor: c.Cursor, More: c.More, Reset: c.Reset, Deleted: c.Deleted,
		Tracks:  make([]syncTrack, len(c.Tracks)),
		Albums:  make([]syncAlbum, len(c.Albums)),
		Artists: make([]syncArtist, len(c.Artists)),
	}
	for i, t := range c.Tracks {
		out.Tracks[i] = syncTrack{toPlayerTrack(t), keys.Key(t.Title)}
	}
	for i, a := range c.Albums {
		out.Albums[i] = syncAlbum{toPlayerAlbum(a), a.Single, keys.Key(a.Title)}
	}
	for i, a := range c.Artists {
		out.Artists[i] = syncArtist{toPlayerArtist(a), keys.Key(a.Name)}
	}
	respond(w, r, out, nil)
}
