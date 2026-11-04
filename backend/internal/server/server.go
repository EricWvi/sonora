// Package server exposes the library over the admin and player APIs and
// serves the embedded player and admin console.
package server

import (
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"sonora/internal/admin"
	"sonora/internal/auth"
	"sonora/internal/media"
	"sonora/internal/store"
)

const (
	// MaxAudioBytes caps one uploaded audio file.
	MaxAudioBytes = 1 << 30
	// uploadOverhead covers the multipart framing and a lyrics part.
	uploadOverhead = store.MaxLyricsBytes + 1<<20
)

// Options are server-wide settings.
type Options struct {
	Version string
}

type server struct {
	store *store.Store
	opts  Options
}

// New returns the HTTP handler for the admin API and the admin console
// under /admin/, the player's sign-in under /auth/, the player API under
// /api/v1/ and the player at the root. authn guards the admin API, listeners
// the player API.
func New(st *store.Store, authn *admin.Service, listeners *auth.Service, player, console fs.FS, opts Options) (http.Handler, error) {
	s := &server{store: st, opts: opts}
	playerStatic, err := newStatic(player, "player not built; run scripts/build.sh")
	if err != nil {
		return nil, err
	}
	consoleStatic, err := newStatic(console, "admin console not built; run scripts/build.sh")
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/version", s.getVersion)
	authn.Register(mux)
	listeners.Register(mux)

	playerAPI := http.NewServeMux()
	s.registerPlayer(playerAPI)
	mux.Handle(PlayerAPIPrefix, listeners.Require(playerAPI))

	// Everything registered on private needs an admin session.
	private := http.NewServeMux()
	mux.Handle(admin.APIPrefix, authn.Require(private))
	mux.Handle("/admin/", http.StripPrefix("/admin", consoleStatic))
	mux.Handle("GET /admin", http.RedirectHandler("/admin/", http.StatusMovedPermanently))
	mux.Handle("/", playerStatic)

	p := admin.APIPrefix
	private.HandleFunc("GET "+p+"overview", s.getOverview)

	private.HandleFunc("GET "+p+"tracks", s.listTracks)
	private.HandleFunc("POST "+p+"tracks", s.uploadTrack)
	private.HandleFunc("GET "+p+"tracks/{id}", s.getTrack)
	private.HandleFunc("PUT "+p+"tracks/{id}", s.updateTrack)
	private.HandleFunc("DELETE "+p+"tracks/{id}", s.deleteTrack)
	private.HandleFunc("PUT "+p+"tracks/{id}/cover", s.setTrackCover)
	private.HandleFunc("DELETE "+p+"tracks/{id}/cover", s.clearTrackCover)
	private.HandleFunc("PUT "+p+"tracks/{id}/lyrics", s.setTrackLyrics)
	private.HandleFunc("DELETE "+p+"tracks/{id}/lyrics", s.clearTrackLyrics)
	private.HandleFunc("GET "+p+"tracks/{id}/audio", s.getTrackAudio)

	private.HandleFunc("GET "+p+"albums", s.listAlbums)
	private.HandleFunc("GET "+p+"albums/{id}", s.getAlbum)
	private.HandleFunc("PUT "+p+"albums/{id}", s.updateAlbum)
	private.HandleFunc("DELETE "+p+"albums/{id}", s.deleteAlbum)
	private.HandleFunc("POST "+p+"albums/{id}/merge", s.mergeAlbums)
	private.HandleFunc("PUT "+p+"albums/{id}/cover", s.setAlbumCover)
	private.HandleFunc("DELETE "+p+"albums/{id}/cover", s.clearAlbumCover)

	private.HandleFunc("GET "+p+"artists", s.listArtists)
	private.HandleFunc("GET "+p+"artists/{id}", s.getArtist)
	private.HandleFunc("PUT "+p+"artists/{id}", s.updateArtist)
	private.HandleFunc("POST "+p+"artists/{id}/merge", s.mergeArtists)
	private.HandleFunc("PUT "+p+"artists/{id}/image", s.setArtistImage)
	private.HandleFunc("DELETE "+p+"artists/{id}/image", s.clearArtistImage)

	private.HandleFunc("GET "+p+"covers/{name}", s.getCover)
	private.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		admin.WriteError(w, http.StatusNotFound, "not_found", "接口不存在")
	})
	return mux, nil
}

func (s *server) getVersion(w http.ResponseWriter, r *http.Request) {
	admin.WriteJSON(w, http.StatusOK, map[string]string{"version": s.opts.Version})
}

func (s *server) getOverview(w http.ResponseWriter, r *http.Request) {
	o, err := s.store.Overview(r.Context())
	respond(w, r, o, err)
}

// --- tracks ----------------------------------------------------------------

func (s *server) listTracks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := s.store.Tracks(r.Context(), store.TrackQuery{
		Q:        q.Get("q"),
		Missing:  q.Get("missing") == "true",
		AlbumID:  queryInt(q.Get("albumId")),
		ArtistID: queryInt(q.Get("artistId")),
		Sort:     listSort(r),
		Window:   window(r),
	})
	respond(w, r, page, err)
}

// uploadTrack takes a multipart form with one "file" part and an optional
// "lyrics" part holding a sidecar lyrics file. It answers 201 with the new
// track, or 200 with the existing one when the file is a duplicate.
func (s *server) uploadTrack(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxAudioBytes+uploadOverhead)
	mr, err := r.MultipartReader()
	if err != nil {
		admin.WriteError(w, http.StatusBadRequest, "bad_request", "请以 multipart/form-data 上传")
		return
	}
	var upload *store.Upload
	defer func() {
		if upload != nil {
			upload.Discard()
		}
	}()
	var lyrics []byte
	for {
		var part *multipart.Part
		part, err = mr.NextPart()
		if errors.Is(err, io.EOF) {
			err = nil
			break
		}
		if err != nil {
			break
		}
		switch part.FormName() {
		case "file":
			if upload != nil {
				admin.WriteError(w, http.StatusBadRequest, "bad_request", "每次请求只能上传一个音频文件")
				return
			}
			upload, err = s.store.Receive(part, part.FileName())
		case "lyrics":
			lyrics, err = io.ReadAll(io.LimitReader(part, store.MaxLyricsBytes+1))
		default:
			_, err = io.Copy(io.Discard, part)
		}
		part.Close()
		if err != nil {
			break
		}
	}
	if err != nil {
		writeUploadError(w, r, err)
		return
	}
	if upload == nil {
		admin.WriteError(w, http.StatusBadRequest, "bad_request", "缺少音频文件")
		return
	}
	res, err := s.store.Ingest(r.Context(), upload, lyrics)
	upload = nil
	if err != nil {
		writeStoreError(w, r, err)
		return
	}
	status := http.StatusCreated
	if res.Status == store.Duplicate {
		status = http.StatusOK
	}
	admin.WriteJSON(w, status, res)
}

func writeUploadError(w http.ResponseWriter, r *http.Request, err error) {
	var tooBig *http.MaxBytesError
	switch {
	case errors.As(err, &tooBig):
		admin.WriteError(w, http.StatusRequestEntityTooLarge, "too_large", "单个文件不能超过 1 GB")
	case errors.Is(err, media.ErrUnsupported):
		admin.WriteError(w, http.StatusUnsupportedMediaType, "unsupported_format",
			"不支持的文件格式，仅支持 MP3、FLAC、AAC（M4A）、Ogg Vorbis 与 Opus")
	case errors.Is(err, media.ErrCorrupt):
		admin.WriteError(w, http.StatusUnprocessableEntity, "corrupt_file",
			"文件已损坏或不完整："+detail(err, media.ErrCorrupt))
	case errors.Is(err, io.ErrUnexpectedEOF), errors.Is(err, multipart.ErrMessageTooLarge):
		admin.WriteError(w, http.StatusBadRequest, "bad_request", "上传中断")
	default:
		writeStoreError(w, r, err)
	}
}

func (s *server) getTrack(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	d, err := s.store.Track(r.Context(), id)
	respond(w, r, d, err)
}

func (s *server) updateTrack(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in store.TrackInput
	if !admin.Decode(w, r, &in) {
		return
	}
	d, err := s.store.UpdateTrack(r.Context(), id, in)
	respond(w, r, d, err)
}

func (s *server) deleteTrack(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	respondEmpty(w, r, s.store.DeleteTrack(r.Context(), id))
}

// setTrackCover takes the raw image bytes as the request body.
func (s *server) setTrackCover(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	data, ok := readBody(w, r, store.MaxCoverBytes, "图片不能超过 10 MB")
	if !ok {
		return
	}
	d, err := s.store.SetTrackCover(r.Context(), id, data)
	respond(w, r, d, err)
}

func (s *server) clearTrackCover(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	d, err := s.store.ClearTrackCover(r.Context(), id)
	respond(w, r, d, err)
}

// setTrackLyrics takes the lyrics file or edited text as the raw body.
func (s *server) setTrackLyrics(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	data, ok := readBody(w, r, store.MaxLyricsBytes, "歌词不能超过 512 KB")
	if !ok {
		return
	}
	d, err := s.store.SetTrackLyrics(r.Context(), id, data)
	respond(w, r, d, err)
}

func (s *server) clearTrackLyrics(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	d, err := s.store.SetTrackLyrics(r.Context(), id, nil)
	respond(w, r, d, err)
}

// getTrackAudio streams a track with Range support so the console can
// preview it.
func (s *server) getTrackAudio(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	f, err := s.store.OpenAudio(r.Context(), id)
	if err != nil {
		writeStoreError(w, r, err)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", f.Format.ContentType())
	w.Header().Set("Cache-Control", "private, no-cache")
	http.ServeContent(w, r, "", f.ModTime, f)
}

// --- albums ----------------------------------------------------------------

func (s *server) listAlbums(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := s.store.Albums(r.Context(), store.AlbumQuery{
		Q:        q.Get("q"),
		ArtistID: queryInt(q.Get("artistId")),
		Sort:     listSort(r),
		Window:   window(r),
	})
	respond(w, r, page, err)
}

func (s *server) getAlbum(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	a, err := s.store.Album(r.Context(), id)
	respond(w, r, a, err)
}

func (s *server) updateAlbum(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in store.AlbumInput
	if !admin.Decode(w, r, &in) {
		return
	}
	a, err := s.store.UpdateAlbum(r.Context(), id, in)
	respond(w, r, a, err)
}

func (s *server) deleteAlbum(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	respondEmpty(w, r, s.store.DeleteAlbum(r.Context(), id))
}

type mergeInput struct {
	SourceIDs []int64 `json:"sourceIds"`
	// KeepCovers applies to albums only, see store.MergeAlbums.
	KeepCovers bool `json:"keepCovers"`
}

// mergeAlbums merges the albums in the body into the one in the path.
func (s *server) mergeAlbums(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in mergeInput
	if !admin.Decode(w, r, &in) {
		return
	}
	a, err := s.store.MergeAlbums(r.Context(), id, in.SourceIDs, in.KeepCovers)
	respond(w, r, a, err)
}

func (s *server) setAlbumCover(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	data, ok := readBody(w, r, store.MaxCoverBytes, "图片不能超过 10 MB")
	if !ok {
		return
	}
	a, err := s.store.SetAlbumCover(r.Context(), id, data)
	respond(w, r, a, err)
}

func (s *server) clearAlbumCover(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	a, err := s.store.ClearAlbumCover(r.Context(), id)
	respond(w, r, a, err)
}

// --- artists ---------------------------------------------------------------

func (s *server) listArtists(w http.ResponseWriter, r *http.Request) {
	page, err := s.store.Artists(r.Context(), r.URL.Query().Get("q"), window(r))
	respond(w, r, page, err)
}

func (s *server) getArtist(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	a, err := s.store.Artist(r.Context(), id)
	respond(w, r, a, err)
}

func (s *server) updateArtist(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in store.ArtistInput
	if !admin.Decode(w, r, &in) {
		return
	}
	a, err := s.store.UpdateArtist(r.Context(), id, in)
	respond(w, r, a, err)
}

func (s *server) mergeArtists(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var in mergeInput
	if !admin.Decode(w, r, &in) {
		return
	}
	a, err := s.store.MergeArtists(r.Context(), id, in.SourceIDs)
	respond(w, r, a, err)
}

// setArtistImage takes the raw image bytes as the request body.
func (s *server) setArtistImage(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	data, ok := readBody(w, r, store.MaxCoverBytes, "图片不能超过 10 MB")
	if !ok {
		return
	}
	a, err := s.store.SetArtistImage(r.Context(), id, data)
	respond(w, r, a, err)
}

func (s *server) clearArtistImage(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	a, err := s.store.ClearArtistImage(r.Context(), id)
	respond(w, r, a, err)
}

// getCover serves a cover image or artist image. Names are content hashes, so the response
// never changes and may be cached for a year.
//
// ?size= picks a thumbnail edge length from store.CoverSizes.
func (s *server) getCover(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var path string
	var err error
	if size := r.URL.Query().Get("size"); size != "" {
		n, perr := strconv.Atoi(size)
		if perr != nil {
			admin.WriteError(w, http.StatusBadRequest, "invalid", "封面尺寸无效")
			return
		}
		path, err = s.store.CoverThumb(name, n)
	} else {
		path, err = s.store.CoverPath(name)
	}
	if err != nil {
		writeStoreError(w, r, err)
		return
	}
	h := w.Header()
	h.Set("Cache-Control", "private, max-age=31536000, immutable")
	h.Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, path)
}

// --- helpers ---------------------------------------------------------------

func respond(w http.ResponseWriter, r *http.Request, v any, err error) {
	if err != nil {
		writeStoreError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	admin.WriteJSON(w, http.StatusOK, v)
}

func respondEmpty(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		writeStoreError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeStoreError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, store.ErrInvalid):
		admin.WriteError(w, http.StatusBadRequest, "invalid", detail(err, store.ErrInvalid))
	case errors.Is(err, store.ErrConflict):
		admin.WriteError(w, http.StatusConflict, "conflict", detail(err, store.ErrConflict))
	case errors.Is(err, store.ErrNotFound):
		admin.WriteError(w, http.StatusNotFound, "not_found", "内容不存在或已被删除")
	default:
		slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
		admin.WriteError(w, http.StatusInternalServerError, "internal", "服务器内部错误")
	}
}

// detail strips the sentinel prefix from a wrapped error.
func detail(err, sentinel error) string {
	return strings.TrimPrefix(err.Error(), sentinel.Error()+": ")
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	return pathInt(w, r, "id")
}

// pathInt reads a positive ID from the named path segment.
func pathInt(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		admin.WriteError(w, http.StatusBadRequest, "bad_request", "无效的 id")
		return 0, false
	}
	return id, true
}

func queryInt(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// listSort reads the sort and order query parameters; the store rejects
// values it does not know.
func listSort(r *http.Request) store.Sort {
	q := r.URL.Query()
	return store.Sort{Field: q.Get("sort"), Order: q.Get("order")}
}

func window(r *http.Request) store.Window {
	q := r.URL.Query()
	return store.Window{Offset: int(queryInt(q.Get("offset"))), Limit: int(queryInt(q.Get("limit")))}
}

// readBody reads a raw upload body of at most limit bytes.
func readBody(w http.ResponseWriter, r *http.Request, limit int64, tooBigMsg string) ([]byte, bool) {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, limit))
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			admin.WriteError(w, http.StatusRequestEntityTooLarge, "too_large", tooBigMsg)
			return nil, false
		}
		admin.WriteError(w, http.StatusBadRequest, "bad_request", "读取上传内容失败")
		return nil, false
	}
	return data, true
}
