package server

import (
	"net/http"

	"sonora/internal/admin"
	"sonora/internal/auth"
	"sonora/internal/store"
)

// PlayerAPIPrefix is where the player API lives. Every endpoint needs a
// listener session.
const PlayerAPIPrefix = "/api/v1/"

// registerPlayer adds the player API to private, which sits behind the
// listener's session.
func (s *server) registerPlayer(private *http.ServeMux) {
	p := PlayerAPIPrefix
	private.HandleFunc("GET "+p+"me", s.getMe)

	private.HandleFunc("GET "+p+"tracks", s.listPlayerTracks)
	private.HandleFunc("GET "+p+"tracks/{id}", s.getPlayerTrack)
	private.HandleFunc("GET "+p+"tracks/{id}/lyrics", s.getPlayerLyrics)
	private.HandleFunc("GET "+p+"tracks/{id}/audio", s.getTrackAudio)

	private.HandleFunc("GET "+p+"albums", s.listPlayerAlbums)
	private.HandleFunc("GET "+p+"albums/{id}", s.getPlayerAlbum)

	private.HandleFunc("GET "+p+"artists", s.listPlayerArtists)
	private.HandleFunc("GET "+p+"artists/{id}", s.getPlayerArtist)

	private.HandleFunc("GET "+p+"covers/{name}", s.getCover)

	private.HandleFunc("GET "+p+"settings", s.getSettings)
	private.HandleFunc("PUT "+p+"settings", s.saveSettings)

	private.HandleFunc("GET "+p+"likes", s.listLikes)
	private.HandleFunc("GET "+p+"likes/ids", s.listLikeIDs)
	private.HandleFunc("PUT "+p+"likes/{trackId}", s.like)
	private.HandleFunc("DELETE "+p+"likes/{trackId}", s.unlike)

	private.HandleFunc("GET "+p+"history", s.listHistory)
	private.HandleFunc("POST "+p+"history", s.recordPlay)
	private.HandleFunc("DELETE "+p+"history", s.clearHistory)
	private.HandleFunc("DELETE "+p+"history/{trackId}", s.removeFromHistory)

	private.HandleFunc("GET "+p+"playlists", s.listPlaylists)
	private.HandleFunc("POST "+p+"playlists", s.createPlaylist)
	private.HandleFunc("GET "+p+"playlists/{id}", s.getPlaylist)
	private.HandleFunc("PUT "+p+"playlists/{id}", s.updatePlaylist)
	private.HandleFunc("DELETE "+p+"playlists/{id}", s.deletePlaylist)
	private.HandleFunc("GET "+p+"playlists/{id}/tracks", s.listPlaylistTracks)
	private.HandleFunc("POST "+p+"playlists/{id}/tracks", s.addToPlaylist)
	private.HandleFunc("DELETE "+p+"playlists/{id}/tracks/{trackId}", s.removeFromPlaylist)
	private.HandleFunc("PUT "+p+"playlists/{id}/order", s.reorderPlaylist)

	private.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		admin.WriteError(w, http.StatusNotFound, "not_found", "接口不存在")
	})
}

// The player sees the library without what only matters to managing it:
// file names, sizes and timestamps.

type playerTrack struct {
	ID         int64           `json:"id"`
	Title      string          `json:"title"`
	Artists    []store.Ref     `json:"artists"`
	Album      *playerAlbumRef `json:"album"`
	TrackNo    int             `json:"trackNo"`
	DiscNo     int             `json:"discNo"`
	Year       int             `json:"year"`
	DurationMs int64           `json:"durationMs"`
	Format     string          `json:"format"`
	Cover      *store.Cover    `json:"cover"`
	HasLyrics  bool            `json:"hasLyrics"`
}

type playerAlbumRef struct {
	ID      int64       `json:"id"`
	Title   string      `json:"title"`
	Artists []store.Ref `json:"artists"`
}

type playerAlbum struct {
	ID         int64        `json:"id"`
	Title      string       `json:"title"`
	Artists    []store.Ref  `json:"artists"`
	Year       int          `json:"year"`
	Cover      *store.Cover `json:"cover"`
	TrackCount int          `json:"trackCount"`
	DurationMs int64        `json:"durationMs"`
}

type playerArtist struct {
	ID         int64        `json:"id"`
	Name       string       `json:"name"`
	Image      *store.Cover `json:"image"`
	TrackCount int          `json:"trackCount"`
	AlbumCount int          `json:"albumCount"`
}

// toPlayerTrack resolves the cover the track shows: its own art, or else
// its album's.
func toPlayerTrack(t store.Track) playerTrack {
	out := playerTrack{
		ID: t.ID, Title: t.Title, Artists: t.Artists, TrackNo: t.TrackNo, DiscNo: t.DiscNo, Year: t.Year,
		DurationMs: t.DurationMs, Format: t.Format, Cover: t.Cover, HasLyrics: t.HasLyrics,
	}
	if t.Album != nil {
		out.Album = &playerAlbumRef{ID: t.Album.ID, Title: t.Album.Title, Artists: t.Album.Artists}
		if out.Cover == nil {
			out.Cover = t.Album.Cover
		}
	}
	return out
}

func toPlayerAlbum(a store.Album) playerAlbum {
	return playerAlbum{
		ID: a.ID, Title: a.Title, Artists: a.Artists, Year: a.Year, Cover: a.Cover,
		TrackCount: a.TrackCount, DurationMs: a.DurationMs,
	}
}

func toPlayerArtist(a store.Artist) playerArtist {
	return playerArtist{ID: a.ID, Name: a.Name, Image: a.Image, TrackCount: a.TrackCount, AlbumCount: a.AlbumCount}
}

func mapPage[T, U any](p store.Page[T], f func(T) U) store.Page[U] {
	out := store.Page[U]{Items: make([]U, len(p.Items)), Total: p.Total}
	for i, item := range p.Items {
		out.Items[i] = f(item)
	}
	return out
}

func (s *server) getMe(w http.ResponseWriter, r *http.Request) {
	respond(w, r, auth.User(r.Context()), nil)
}

func (s *server) listPlayerTracks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := s.store.Tracks(r.Context(), store.TrackQuery{
		Q:        q.Get("q"),
		AlbumID:  queryInt(q.Get("albumId")),
		ArtistID: queryInt(q.Get("artistId")),
		Sort:     listSort(r),
		Window:   window(r),
	})
	respond(w, r, mapPage(page, toPlayerTrack), err)
}

func (s *server) getPlayerTrack(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	d, err := s.store.Track(r.Context(), id)
	respond(w, r, toPlayerTrack(d.Track), err)
}

// getPlayerLyrics returns the stored lyrics text as {"format", "lyrics"}:
// format is "lrc" or "ttml", both empty when the track has none. The
// player parses them.
func (s *server) getPlayerLyrics(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	d, err := s.store.Track(r.Context(), id)
	respond(w, r, map[string]string{"format": d.LyricsFormat, "lyrics": d.Lyrics}, err)
}

func (s *server) listPlayerAlbums(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := s.store.Albums(r.Context(), store.AlbumQuery{
		Q:           q.Get("q"),
		ArtistID:    queryInt(q.Get("artistId")),
		HideSingles: q.Get("hideSingles") == "true",
		Sort:        listSort(r),
		Window:      window(r),
	})
	respond(w, r, mapPage(page, toPlayerAlbum), err)
}

func (s *server) getPlayerAlbum(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	a, err := s.store.Album(r.Context(), id)
	respond(w, r, toPlayerAlbum(a), err)
}

func (s *server) listPlayerArtists(w http.ResponseWriter, r *http.Request) {
	page, err := s.store.Artists(r.Context(), r.URL.Query().Get("q"), window(r))
	respond(w, r, mapPage(page, toPlayerArtist), err)
}

func (s *server) getPlayerArtist(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	a, err := s.store.Artist(r.Context(), id)
	respond(w, r, toPlayerArtist(a), err)
}

// --- listener data ---------------------------------------------------------

func userID(r *http.Request) int64 { return auth.User(r.Context()).ID }

func pathTrackID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	return pathInt(w, r, "trackId")
}

// getSettings returns the listener's appearance and lyrics settings as
// saved by the player.
func (s *server) getSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.store.Settings(r.Context(), userID(r))
	respond(w, r, settings, err)
}

// saveSettings takes the whole settings object as the body and replaces
// what was saved.
func (s *server) saveSettings(w http.ResponseWriter, r *http.Request) {
	data, ok := readBody(w, r, store.MaxSettingsBytes, "设置不能超过 256 KB")
	if !ok {
		return
	}
	respondEmpty(w, r, s.store.SaveSettings(r.Context(), userID(r), data))
}

func (s *server) listLikes(w http.ResponseWriter, r *http.Request) {
	page, err := s.store.LikedTracks(r.Context(), userID(r), listSort(r), window(r))
	respond(w, r, mapPage(page, toPlayerTrack), err)
}

func (s *server) listLikeIDs(w http.ResponseWriter, r *http.Request) {
	ids, err := s.store.LikedTrackIDs(r.Context(), userID(r))
	respond(w, r, map[string][]int64{"ids": ids}, err)
}

func (s *server) like(w http.ResponseWriter, r *http.Request) {
	id, ok := pathTrackID(w, r)
	if !ok {
		return
	}
	respondEmpty(w, r, s.store.Like(r.Context(), userID(r), id))
}

func (s *server) unlike(w http.ResponseWriter, r *http.Request) {
	id, ok := pathTrackID(w, r)
	if !ok {
		return
	}
	respondEmpty(w, r, s.store.Unlike(r.Context(), userID(r), id))
}

type playerHistoryEntry struct {
	Track    playerTrack `json:"track"`
	PlayedAt string      `json:"playedAt"`
}

func (s *server) listHistory(w http.ResponseWriter, r *http.Request) {
	page, err := s.store.History(r.Context(), userID(r), window(r))
	respond(w, r, mapPage(page, func(e store.HistoryEntry) playerHistoryEntry {
		return playerHistoryEntry{Track: toPlayerTrack(e.Track), PlayedAt: e.PlayedAt}
	}), err)
}

// recordPlay takes {"trackId": n}. The player decides when a play counts.
func (s *server) recordPlay(w http.ResponseWriter, r *http.Request) {
	var in struct {
		TrackID int64 `json:"trackId"`
	}
	if !admin.Decode(w, r, &in) {
		return
	}
	if in.TrackID <= 0 {
		admin.WriteError(w, http.StatusBadRequest, "bad_request", "无效的 trackId")
		return
	}
	respondEmpty(w, r, s.store.RecordPlay(r.Context(), userID(r), in.TrackID))
}

func (s *server) clearHistory(w http.ResponseWriter, r *http.Request) {
	respondEmpty(w, r, s.store.ClearHistory(r.Context(), userID(r)))
}

func (s *server) removeFromHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathTrackID(w, r)
	if !ok {
		return
	}
	respondEmpty(w, r, s.store.RemoveFromHistory(r.Context(), userID(r), id))
}

func (s *server) listPlaylists(w http.ResponseWriter, r *http.Request) {
	page, err := s.store.Playlists(r.Context(), userID(r), r.URL.Query().Get("q"), window(r))
	respond(w, r, page, err)
}

func (s *server) createPlaylist(w http.ResponseWriter, r *http.Request) {
	var in store.PlaylistInput
	if !admin.Decode(w, r, &in) {
		return
	}
	p, err := s.store.CreatePlaylist(r.Context(), userID(r), in)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	admin.WriteJSON(w, http.StatusCreated, p)
}

func (s *server) getPlaylist(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	p, err := s.store.Playlist(r.Context(), userID(r), id)
	respond(w, r, p, err)
}

func (s *server) updatePlaylist(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in store.PlaylistInput
	if !admin.Decode(w, r, &in) {
		return
	}
	p, err := s.store.UpdatePlaylist(r.Context(), userID(r), id, in)
	respond(w, r, p, err)
}

func (s *server) deletePlaylist(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	respondEmpty(w, r, s.store.DeletePlaylist(r.Context(), userID(r), id))
}

func (s *server) listPlaylistTracks(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	page, err := s.store.PlaylistTracks(r.Context(), userID(r), id, listSort(r), window(r))
	respond(w, r, mapPage(page, toPlayerTrack), err)
}

type trackIDsInput struct {
	TrackIDs []int64 `json:"trackIds"`
}

// addToPlaylist appends {"trackIds": [...]} and answers
// {"added": n, "playlist": {...}}; tracks already in it are skipped.
func (s *server) addToPlaylist(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in trackIDsInput
	if !admin.Decode(w, r, &in) {
		return
	}
	added, p, err := s.store.AddToPlaylist(r.Context(), userID(r), id, in.TrackIDs)
	respond(w, r, map[string]any{"added": added, "playlist": p}, err)
}

func (s *server) removeFromPlaylist(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	trackID, ok := pathTrackID(w, r)
	if !ok {
		return
	}
	p, err := s.store.RemoveFromPlaylist(r.Context(), userID(r), id, trackID)
	respond(w, r, p, err)
}

// reorderPlaylist takes the playlist's complete track list in the new
// order as {"trackIds": [...]}.
func (s *server) reorderPlaylist(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in trackIDsInput
	if !admin.Decode(w, r, &in) {
		return
	}
	p, err := s.store.ReorderPlaylist(r.Context(), userID(r), id, in.TrackIDs)
	respond(w, r, p, err)
}
