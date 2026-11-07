package server

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"

	"sonora/internal/admin"
	"sonora/internal/auth"
	"sonora/internal/store"
)

// Filling in a local album, artist or song from the NetEase one it is linked
// to (Eric, 2026-11-07). Like importing, this changes the shared library, so
// it takes the same group and the listener's own NetEase session.

func (s *server) registerFill(private *http.ServeMux) {
	p := PlayerAPIPrefix
	private.HandleFunc("POST "+p+"albums/{id}/fill", s.fillAlbum)
	private.HandleFunc("POST "+p+"artists/{id}/fill", s.fillArtist)
	private.HandleFunc("POST "+p+"tracks/{id}/fill", s.fillTrack)
}

// fillSession checks the listener may fill in anything and returns their
// NetEase session, or answers the request.
func (s *server) fillSession(w http.ResponseWriter, r *http.Request) (string, bool) {
	if !s.canImport(auth.User(r.Context())) {
		admin.WriteError(w, http.StatusForbidden, "import_forbidden", "没有补全资料的权限")
		return "", false
	}
	session, ok := s.neteaseSession(w, r)
	return session.Cookie, ok
}

func writeNotLinked(w http.ResponseWriter) {
	admin.WriteError(w, http.StatusConflict, "conflict", "还没有关联网易云")
}

func filledFields(filled bool, field string) []string {
	if filled {
		return []string{field}
	}
	return []string{}
}

func (s *server) fillAlbum(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	cookie, ok := s.fillSession(w, r)
	if !ok {
		return
	}
	album, err := s.store.Album(r.Context(), id)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}
	if album.NeteaseID == nil {
		writeNotLinked(w)
		return
	}
	filled := false
	if album.Description == "" {
		online, _, err := s.opts.Netease.AlbumSongs(r.Context(), cookie, *album.NeteaseID)
		if err != nil {
			writeNeteaseError(w, r, err)
			return
		}
		album, filled, err = s.store.FillAlbumDescription(r.Context(), id, online.Description)
		if err != nil {
			writeStoreError(w, r, err)
			return
		}
	}
	respond(w, r, map[string]any{"album": toPlayerAlbum(album), "filled": filledFields(filled, store.FilledDescription)}, nil)
}

func (s *server) fillArtist(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	cookie, ok := s.fillSession(w, r)
	if !ok {
		return
	}
	artist, err := s.store.Artist(r.Context(), id)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}
	if artist.NeteaseID == nil {
		writeNotLinked(w)
		return
	}
	filled := []string{}
	if artist.Image == nil || artist.Description == "" {
		profile, err := s.opts.Netease.ArtistProfile(r.Context(), cookie, *artist.NeteaseID)
		if err != nil {
			writeNeteaseError(w, r, err)
			return
		}
		var image []byte
		if artist.Image == nil && profile.ImageURL != "" {
			image = s.downloadImage(r.Context(), id, profile.ImageURL)
		}
		artist, filled, err = s.store.FillArtistProfile(r.Context(), id, image, profile.Description)
		if err != nil {
			writeStoreError(w, r, err)
			return
		}
	}
	respond(w, r, map[string]any{"artist": toPlayerArtist(artist), "filled": filled}, nil)
}

// downloadImage fetches an artist's picture. Failing leaves it out: the
// description can still be filled in, and the answer says what was.
func (s *server) downloadImage(ctx context.Context, artist int64, url string) []byte {
	var buf bytes.Buffer
	if err := s.opts.Netease.Download(ctx, url, &buf, maxImportCover); err != nil {
		slog.Warn("fill artist image", "artist", artist, "err", err)
		return nil
	}
	return buf.Bytes()
}

func (s *server) fillTrack(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	cookie, ok := s.fillSession(w, r)
	if !ok {
		return
	}
	track, err := s.store.Track(r.Context(), id)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}
	if track.NeteaseID == nil {
		writeNotLinked(w)
		return
	}
	filled := false
	if !track.HasLyrics {
		lyrics, err := s.opts.Netease.Lyrics(r.Context(), cookie, *track.NeteaseID)
		if err != nil {
			writeNeteaseError(w, r, err)
			return
		}
		text, _ := lyrics.Compose()
		track, filled, err = s.store.FillTrackLyrics(r.Context(), id, []byte(text))
		if err != nil {
			writeStoreError(w, r, err)
			return
		}
	}
	respond(w, r, map[string]any{"track": toPlayerTrack(track.Track), "filled": filledFields(filled, "lyrics")}, nil)
}
