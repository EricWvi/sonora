package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"sonora/internal/admin"
	"sonora/internal/auth"
	"sonora/internal/netease"
	"sonora/internal/store"
)

// neteaseRefreshAge is how old a session gets before it is renewed. SPlayer
// renews after three days too; it does so from the browser, Sonora here.
const neteaseRefreshAge = 3 * 24 * time.Hour

// maxCookieBytes caps a pasted cookie.
const maxCookieBytes = 16 << 10

// The states of a listener's NetEase account, see neteaseStatus.
const (
	neteaseDisabled    = "disabled"    // the server has no API service configured
	neteaseSignedOut   = "signedOut"   // no session
	neteaseSignedIn    = "signedIn"    // session works
	neteaseExpired     = "expired"     // session no longer works; sign in again
	neteaseUnavailable = "unavailable" // session present, API service unreachable
)

type neteaseStatus struct {
	State   string           `json:"state"`
	Profile *netease.Profile `json:"profile"`
}

func (s *server) registerNetease(private *http.ServeMux) {
	p := PlayerAPIPrefix + "netease/"
	private.HandleFunc("GET "+p+"status", s.getNeteaseStatus)
	private.HandleFunc("POST "+p+"login/qr", s.startNeteaseQR)
	private.HandleFunc("GET "+p+"login/qr/{key}", s.checkNeteaseQR)
	private.HandleFunc("POST "+p+"login/cookie", s.neteaseCookieLogin)
	private.HandleFunc("POST "+p+"logout", s.neteaseLogout)
	private.HandleFunc("GET "+p+"playlists", s.listNeteasePlaylists)
	private.HandleFunc("GET "+p+"playlists/{id}", s.getNeteasePlaylist)
	private.HandleFunc("GET "+p+"albums", s.listNeteaseAlbums)
	private.HandleFunc("GET "+p+"albums/{id}", s.getNeteaseAlbum)
	private.HandleFunc("GET "+p+"artists", s.listNeteaseArtists)
	private.HandleFunc("GET "+p+"artists/{id}", s.getNeteaseArtist)
	private.HandleFunc("GET "+p+"songs/{id}/url", s.getNeteaseSongURL)
	private.HandleFunc("GET "+p+"songs/{id}/lyrics", s.getNeteaseLyrics)
	private.HandleFunc("GET "+p+"likes/ids", s.listNeteaseLikeIDs)
	private.HandleFunc("PUT "+p+"likes/{id}", s.setNeteaseLike)
	private.HandleFunc("DELETE "+p+"likes/{id}", s.setNeteaseLike)

	private.HandleFunc("GET "+PlayerAPIPrefix+"likes/mixed", s.listMixedLikes)
	private.HandleFunc("GET "+PlayerAPIPrefix+"history/mixed", s.listMixedHistory)
	private.HandleFunc("POST "+PlayerAPIPrefix+"history/online", s.recordOnlinePlay)
	private.HandleFunc("DELETE "+PlayerAPIPrefix+"history/online/{id}", s.removeOnlinePlay)
}

// neteaseReady answers for the endpoints that need the API service.
func (s *server) neteaseReady(w http.ResponseWriter) bool {
	if s.opts.Netease == nil {
		admin.WriteError(w, http.StatusServiceUnavailable, "netease_disabled", "服务器没有配置网易云接口服务")
		return false
	}
	return true
}

func writeNeteaseError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, netease.ErrUnavailable):
		slog.Warn("netease api", "path", r.URL.Path, "err", err)
		admin.WriteError(w, http.StatusBadGateway, "netease_unavailable", "网易云接口服务不可用，请稍后重试")
		return
	case errors.Is(err, netease.ErrExpired):
		admin.WriteError(w, http.StatusForbidden, "netease_expired", "网易云登录已失效，请重新登录")
		return
	case errors.Is(err, netease.ErrNotFound):
		admin.WriteError(w, http.StatusNotFound, "not_found", "内容不存在或已被删除")
		return
	}
	writeStoreError(w, r, err)
}

// neteaseSession returns the listener's session for an endpoint that needs
// one, or answers the request. The status is 403, not 401: only NetEase is
// signed out, and the player treats 401 as its own session ending.
func (s *server) neteaseSession(w http.ResponseWriter, r *http.Request) (store.NeteaseSession, bool) {
	if !s.neteaseReady(w) {
		return store.NeteaseSession{}, false
	}
	session, err := s.store.NeteaseSession(r.Context(), userID(r))
	if errors.Is(err, store.ErrNotFound) {
		admin.WriteError(w, http.StatusForbidden, "netease_signed_out", "没有登录网易云")
		return session, false
	}
	if err != nil {
		writeStoreError(w, r, err)
		return session, false
	}
	return session, true
}

func (s *server) getNeteaseStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.neteaseStatus(r.Context(), auth.User(r.Context()).ID)
	respond(w, r, status, err)
}

func (s *server) neteaseStatus(ctx context.Context, userID int64) (neteaseStatus, error) {
	if s.opts.Netease == nil {
		return neteaseStatus{State: neteaseDisabled}, nil
	}
	session, err := s.store.NeteaseSession(ctx, userID)
	if errors.Is(err, store.ErrNotFound) {
		return neteaseStatus{State: neteaseSignedOut}, nil
	}
	if err != nil {
		return neteaseStatus{}, err
	}
	profile, err := s.opts.Netease.Status(ctx, session.Cookie)
	switch {
	case errors.Is(err, netease.ErrUnavailable):
		slog.Warn("netease api", "err", err)
		return neteaseStatus{State: neteaseUnavailable}, nil
	case err != nil:
		return neteaseStatus{}, err
	case profile == nil:
		// Kept, not deleted: the listener should learn the session ended
		// rather than find themselves signed out without a word.
		return neteaseStatus{State: neteaseExpired}, nil
	}
	s.refreshNetease(ctx, userID, session)
	return neteaseStatus{State: neteaseSignedIn, Profile: profile}, nil
}

// refreshNetease renews an aging session. It is best effort: the session in
// hand still works, and the next status check tries again.
func (s *server) refreshNetease(ctx context.Context, userID int64, session store.NeteaseSession) {
	if time.Since(session.RefreshedAt) < neteaseRefreshAge {
		return
	}
	cookie, err := s.opts.Netease.Refresh(ctx, session.Cookie)
	if err != nil {
		slog.Warn("netease refresh", "err", err)
		return
	}
	session.Cookie, session.RefreshedAt = cookie, time.Now()
	if err := s.store.SaveNeteaseSession(ctx, userID, session); err != nil {
		slog.Error("netease refresh", "err", err)
	}
}

func (s *server) startNeteaseQR(w http.ResponseWriter, r *http.Request) {
	if !s.neteaseReady(w) {
		return
	}
	key, err := s.opts.Netease.QRKey(r.Context())
	if err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	// The URL comes from here so that the player never names a NetEase host.
	respond(w, r, map[string]string{"key": key, "url": netease.QRURL(key)}, nil)
}

func (s *server) checkNeteaseQR(w http.ResponseWriter, r *http.Request) {
	if !s.neteaseReady(w) {
		return
	}
	state, err := s.opts.Netease.QRCheck(r.Context(), r.PathValue("key"))
	if err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	if state.Code == netease.QRSucceeded {
		if _, err := s.signInNetease(r.Context(), auth.User(r.Context()).ID, state.Cookie); err != nil {
			writeNeteaseError(w, r, err)
			return
		}
	}
	respond(w, r, map[string]any{
		"code": state.Code, "nickname": state.Nickname, "avatarUrl": state.AvatarURL,
	}, nil)
}

func (s *server) neteaseCookieLogin(w http.ResponseWriter, r *http.Request) {
	if !s.neteaseReady(w) {
		return
	}
	var body struct {
		Cookie string `json:"cookie"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxCookieBytes)
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		admin.WriteError(w, http.StatusBadRequest, "bad_request", "请求格式错误")
		return
	}
	profile, err := s.signInNetease(r.Context(), auth.User(r.Context()).ID, body.Cookie)
	if err != nil {
		writeNeteaseError(w, r, err)
		return
	}
	respond(w, r, neteaseStatus{State: neteaseSignedIn, Profile: profile}, nil)
}

// signInNetease checks the cookie against NetEase before keeping it: a
// pasted cookie may be stale, and SPlayer only finds out later.
func (s *server) signInNetease(ctx context.Context, userID int64, raw string) (*netease.Profile, error) {
	cookie, err := netease.NormalizeCookie(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: Cookie 无效，必须包含 MUSIC_U", store.ErrInvalid)
	}
	profile, err := s.opts.Netease.Status(ctx, cookie)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, fmt.Errorf("%w: Cookie 无效或已过期", store.ErrInvalid)
	}
	err = s.store.SaveNeteaseSession(ctx, userID, store.NeteaseSession{
		Cookie: cookie, AccountID: profile.UserID, RefreshedAt: time.Now(),
	})
	return profile, err
}

// neteaseLogout only forgets the session here. SPlayer also calls /logout, but
// that ends the session on NetEase's side, and a pasted cookie may be the one
// another client is using.
func (s *server) neteaseLogout(w http.ResponseWriter, r *http.Request) {
	respondEmpty(w, r, s.store.DeleteNeteaseSession(r.Context(), auth.User(r.Context()).ID))
}
