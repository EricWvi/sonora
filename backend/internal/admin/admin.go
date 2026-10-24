// Package admin authenticates the admin console.
//
// Until an admin exists the service is uninitialised: a one-time setup
// token is generated at startup, printed to the container log and kept
// only in memory. Whoever presents it creates the first admin. After that,
// admins sign in with username and password and get a session cookie
// scoped to the admin API, separate from the player's session.
package admin

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"sonora/internal/store"
)

const (
	// APIPrefix is where the admin API lives; the session cookie is
	// limited to it.
	APIPrefix     = "/admin/api/v1/"
	sessionCookie = "sonora_admin_session"
	// A session ends after IdleTimeout without use or MaxLifetime after
	// sign-in, whichever comes first.
	IdleTimeout = 7 * 24 * time.Hour
	MaxLifetime = 30 * 24 * time.Hour
	// extendAfter limits session writes to one per this interval.
	extendAfter = 5 * time.Minute
)

// DefaultBcryptCost is the password hashing cost.
const DefaultBcryptCost = 12

// Config configures the service.
type Config struct {
	// PublicURL is the trusted origin for state-changing requests. Empty
	// (development mode) trusts any origin.
	PublicURL string
	// Dev reports development mode to the console so it can show the
	// fixed account.
	Dev bool
	// Now returns the current time; nil uses time.Now.
	Now func() time.Time
	// BcryptCost hashes new passwords; 0 uses DefaultBcryptCost.
	BcryptCost int
}

// Service handles setup, sign-in and request authentication.
type Service struct {
	store  *store.Store
	origin string
	secure bool
	dev    bool
	now    func() time.Time

	// mu guards setupToken and serialises setup attempts.
	mu         sync.Mutex
	setupToken string // empty once an admin exists

	setupLimit *limiter
	loginLimit *limiter
	bcryptCost int
	dummyHash  []byte
}

// New returns the service. When no admin exists yet it generates the
// setup token; read it with SetupToken and print it.
func New(ctx context.Context, st *store.Store, cfg Config) (*Service, error) {
	s := &Service{
		store:      st,
		origin:     strings.TrimSuffix(cfg.PublicURL, "/"),
		secure:     strings.HasPrefix(cfg.PublicURL, "https://"),
		dev:        cfg.Dev,
		now:        cfg.Now,
		setupLimit: newLimiter(5, time.Minute),
		loginLimit: newLimiter(10, time.Minute),
		bcryptCost: cfg.BcryptCost,
	}
	if s.now == nil {
		s.now = time.Now
	}
	if s.bcryptCost == 0 {
		s.bcryptCost = DefaultBcryptCost
	}
	// Compared against when the username is unknown, so a failed sign-in
	// takes the same time either way.
	var err error
	if s.dummyHash, err = bcrypt.GenerateFromPassword([]byte("sonora-dummy-password"), s.bcryptCost); err != nil {
		return nil, err
	}
	has, err := st.HasAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if !has {
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			return nil, err
		}
		s.setupToken = base64.RawURLEncoding.EncodeToString(raw)
	}
	return s, nil
}

// SetupToken returns the one-time setup token, or "" once initialised.
func (s *Service) SetupToken() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.setupToken
}

func (s *Service) initialized() bool { return s.SetupToken() == "" }

// HashPassword hashes a password for storage.
func HashPassword(password string) (string, error) {
	return hashPassword(password, DefaultBcryptCost)
}

func hashPassword(password string, cost int) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(h), err
}

// Register adds the unauthenticated endpoints to mux.
func (s *Service) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET "+APIPrefix+"status", s.status)
	mux.HandleFunc("POST "+APIPrefix+"setup", s.setup)
	mux.HandleFunc("POST "+APIPrefix+"login", s.login)
	mux.HandleFunc("POST "+APIPrefix+"logout", s.logout)
}

type adminKey struct{}

// Admin returns the signed-in admin of a request that passed Require.
func Admin(ctx context.Context) store.Admin {
	return ctx.Value(adminKey{}).(store.Admin)
}

// Require rejects requests without a valid admin session and makes the
// admin available through Admin. Before setup every request is refused.
func (s *Service) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.sameOrigin(r) {
			WriteError(w, http.StatusForbidden, "bad_origin", "请求来源不受信任")
			return
		}
		if !s.initialized() {
			WriteError(w, http.StatusForbidden, "not_initialized", "服务尚未初始化")
			return
		}
		a, ok := s.authenticate(w, r)
		if !ok {
			WriteError(w, http.StatusUnauthorized, "unauthenticated", "未登录或登录已过期")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), adminKey{}, a)))
	})
}

func (s *Service) authenticate(w http.ResponseWriter, r *http.Request) (store.Admin, bool) {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return store.Admin{}, false
	}
	sess, err := s.store.LookupAdminSession(r.Context(), c.Value)
	now := s.now()
	if err == nil && !now.Before(sess.ExpiresAt) {
		s.store.DeleteAdminSession(r.Context(), c.Value)
		err = store.ErrNotFound
	}
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			slog.Error("look up admin session", "err", err)
		}
		s.clearCookie(w)
		return store.Admin{}, false
	}
	if now.Sub(sess.LastSeen) > extendAfter {
		if err := s.store.ExtendAdminSession(r.Context(), c.Value, now, s.expiry(sess.CreatedAt, now)); err != nil {
			slog.Error("extend admin session", "err", err)
		}
	}
	return sess.Admin, true
}

func (s *Service) expiry(created, now time.Time) time.Time {
	idle, hard := now.Add(IdleTimeout), created.Add(MaxLifetime)
	if idle.Before(hard) {
		return idle
	}
	return hard
}

// Status is what the console needs before it decides which page to show.
type Status struct {
	Initialized bool         `json:"initialized"`
	DevMode     bool         `json:"devMode"`
	Admin       *store.Admin `json:"admin"`
}

func (s *Service) status(w http.ResponseWriter, r *http.Request) {
	st := Status{Initialized: s.initialized(), DevMode: s.dev}
	if st.Initialized {
		if a, ok := s.authenticate(w, r); ok {
			st.Admin = &a
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	WriteJSON(w, http.StatusOK, st)
}

type credentials struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Service) setup(w http.ResponseWriter, r *http.Request) {
	if !s.sameOrigin(r) {
		WriteError(w, http.StatusForbidden, "bad_origin", "请求来源不受信任")
		return
	}
	var in credentials
	if !Decode(w, r, &in) {
		return
	}
	if wait, ok := s.setupLimit.allow("setup", s.now()); !ok {
		tooMany(w, wait)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.setupToken == "" {
		WriteError(w, http.StatusConflict, "already_initialized", "服务已经初始化，请直接登录")
		return
	}
	// Hashing first makes the comparison independent of the input length.
	got, want := sha256.Sum256([]byte(in.Token)), sha256.Sum256([]byte(s.setupToken))
	if subtle.ConstantTimeCompare(got[:], want[:]) != 1 {
		s.setupLimit.fail("setup", s.now())
		WriteError(w, http.StatusForbidden, "invalid_token", "初始化 token 不正确")
		return
	}
	username := strings.TrimSpace(in.Username)
	if msg := checkCredentials(username, in.Password); msg != "" {
		WriteError(w, http.StatusBadRequest, "invalid", msg)
		return
	}
	hash, err := hashPassword(in.Password, s.bcryptCost)
	if err != nil {
		s.internal(w, r, err)
		return
	}
	a, err := s.store.CreateFirstAdmin(r.Context(), username, hash)
	if errors.Is(err, store.ErrConflict) {
		s.setupToken = ""
		WriteError(w, http.StatusConflict, "already_initialized", "服务已经初始化，请直接登录")
		return
	}
	if err != nil {
		s.internal(w, r, err)
		return
	}
	s.setupToken = ""
	slog.Info("admin created; setup token revoked", "admin", a.Username)
	s.startSession(w, r, a)
}

func (s *Service) login(w http.ResponseWriter, r *http.Request) {
	if !s.sameOrigin(r) {
		WriteError(w, http.StatusForbidden, "bad_origin", "请求来源不受信任")
		return
	}
	if !s.initialized() {
		WriteError(w, http.StatusForbidden, "not_initialized", "服务尚未初始化")
		return
	}
	var in credentials
	if !Decode(w, r, &in) {
		return
	}
	if wait, ok := s.loginLimit.allow("login", s.now()); !ok {
		tooMany(w, wait)
		return
	}
	a, hash, err := s.store.AdminByUsername(r.Context(), strings.TrimSpace(in.Username))
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		s.internal(w, r, err)
		return
	}
	if err != nil {
		bcrypt.CompareHashAndPassword(s.dummyHash, []byte(in.Password))
	}
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) != nil {
		s.loginLimit.fail("login", s.now())
		WriteError(w, http.StatusUnauthorized, "invalid_credentials", "用户名或密码不正确")
		return
	}
	s.startSession(w, r, a)
}

func (s *Service) startSession(w http.ResponseWriter, r *http.Request, a store.Admin) {
	now := s.now()
	if err := s.store.DeleteExpiredAdminSessions(r.Context(), now); err != nil {
		slog.Error("delete expired admin sessions", "err", err)
	}
	token, err := s.store.CreateAdminSession(r.Context(), a.ID, now, s.expiry(now, now))
	if err != nil {
		s.internal(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     APIPrefix,
		MaxAge:   int(MaxLifetime.Seconds()),
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteStrictMode,
	})
	w.Header().Set("Cache-Control", "no-store")
	WriteJSON(w, http.StatusOK, Status{Initialized: true, DevMode: s.dev, Admin: &a})
}

func (s *Service) logout(w http.ResponseWriter, r *http.Request) {
	if !s.sameOrigin(r) {
		WriteError(w, http.StatusForbidden, "bad_origin", "请求来源不受信任")
		return
	}
	if c, err := r.Cookie(sessionCookie); err == nil {
		if err := s.store.DeleteAdminSession(r.Context(), c.Value); err != nil {
			s.internal(w, r, err)
			return
		}
	}
	s.clearCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// sameOrigin rejects state-changing requests sent from other sites,
// including sibling subdomains that SameSite treats as same-site.
func (s *Service) sameOrigin(r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	origin := r.Header.Get("Origin")
	return s.origin == "" || origin == "" || origin == s.origin
}

func (s *Service) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Path: APIPrefix, MaxAge: -1, HttpOnly: true, Secure: s.secure, SameSite: http.SameSiteStrictMode,
	})
}

func (s *Service) internal(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	WriteError(w, http.StatusInternalServerError, "internal", "服务器内部错误")
}

// checkCredentials returns a user-facing problem with a new account, or "".
func checkCredentials(username, password string) string {
	n := utf8.RuneCountInString(username)
	if n == 0 || n > 32 {
		return "用户名需要 1–32 个字符"
	}
	for _, r := range username {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return "用户名不能包含空白或控制字符"
		}
	}
	if len(password) < 8 {
		return "密码至少需要 8 个字符"
	}
	// bcrypt ignores everything past 72 bytes.
	if len(password) > 72 {
		return "密码不能超过 72 字节"
	}
	return ""
}

func tooMany(w http.ResponseWriter, wait time.Duration) {
	w.Header().Set("Retry-After", strconv.Itoa(int(wait.Seconds())+1))
	WriteError(w, http.StatusTooManyRequests, "rate_limited", "尝试次数过多，请稍后再试")
}

// Error is the body of every error response of the Sonora API. Clients
// branch on the HTTP status first; Code distinguishes cases that share a
// status, and Message is shown to people.
type Error struct {
	Message string `json:"error"`
	Code    string `json:"code"`
}

// WriteError writes an error response.
func WriteError(w http.ResponseWriter, status int, code, msg string) {
	WriteJSON(w, status, Error{Message: msg, Code: code})
}

// WriteJSON writes v as a JSON response.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// Decode reads a JSON request body into v, answering 400 on failure.
func Decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256<<10)).Decode(v); err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", "请求格式不正确")
		return false
	}
	return true
}
