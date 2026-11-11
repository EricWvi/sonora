package server

import (
	"net/http"
	"strconv"

	"sonora/internal/admin"
	"sonora/internal/store"
)

func (s *server) registerFM(private *http.ServeMux) {
	p := PlayerAPIPrefix + "fm"
	private.HandleFunc("GET "+p, s.getFM)
	private.HandleFunc("POST "+p+"/albums/{id}/finish", s.finishFMAlbum)
	private.HandleFunc("POST "+p+"/albums/{id}/skip", s.skipFMAlbum)
	private.HandleFunc("PUT "+p+"/position", s.setFMPosition)
	private.HandleFunc("PUT "+p+"/groups", s.setFMGroups)
}

type playerFMAlbum struct {
	ID     int64         `json:"id"`
	Tracks []playerTrack `json:"tracks"`
}

type playerFM struct {
	Albums   []playerFMAlbum `json:"albums"`
	TrackID  *int64          `json:"trackId"`
	GroupIDs []int64         `json:"groupIds"`
}

func respondFM(w http.ResponseWriter, r *http.Request, state store.FMState, err error) {
	out := playerFM{Albums: make([]playerFMAlbum, len(state.Albums)), TrackID: state.TrackID, GroupIDs: state.GroupIDs}
	for i, album := range state.Albums {
		tracks := make([]playerTrack, len(album.Tracks))
		for j, t := range album.Tracks {
			tracks[j] = toPlayerTrack(t)
		}
		out.Albums[i] = playerFMAlbum{ID: album.ID, Tracks: tracks}
	}
	respond(w, r, out, err)
}

// getFM returns the radio. count (1–10, default 2) is how many albums to
// return, the current one included; the Android home screen asks for five.
func (s *server) getFM(w http.ResponseWriter, r *http.Request) {
	count := 2
	if v := r.URL.Query().Get("count"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > store.FMMaxAhead {
			admin.WriteError(w, http.StatusBadRequest, "invalid", "count 必须是 1 到 10 的整数")
			return
		}
		count = n
	}
	state, err := s.store.FMAhead(r.Context(), userID(r), count)
	respondFM(w, r, state, err)
}

func (s *server) finishFMAlbum(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	state, err := s.store.FMFinish(r.Context(), userID(r), id)
	respondFM(w, r, state, err)
}

func (s *server) skipFMAlbum(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	state, err := s.store.FMSkip(r.Context(), userID(r), id)
	respondFM(w, r, state, err)
}

func (s *server) setFMPosition(w http.ResponseWriter, r *http.Request) {
	var in struct {
		TrackID int64 `json:"trackId"`
	}
	if !admin.Decode(w, r, &in) {
		return
	}
	respondEmpty(w, r, s.store.FMSetPosition(r.Context(), userID(r), in.TrackID))
}

func (s *server) setFMGroups(w http.ResponseWriter, r *http.Request) {
	var in struct {
		PlaylistIDs []int64 `json:"playlistIds"`
	}
	if !admin.Decode(w, r, &in) {
		return
	}
	state, err := s.store.FMSetGroups(r.Context(), userID(r), in.PlaylistIDs)
	respondFM(w, r, state, err)
}
