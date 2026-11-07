package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"sonora/internal/store"
)

/*
 * Native clients sign in with the system browser, after RFC 8252. Adapted
 * from asr (internal/auth/device.go); Sonora only has the Android client,
 * so there is no loopback redirect and no avatar on the confirm page.
 *
 *   1. The client opens /auth/device/login with its platform, name,
 *      redirect_uri, state and a PKCE code_challenge (S256).
 *   2. The browser goes through the usual OIDC login; the device request
 *      rides along in the login cookie and comes back to /auth/callback,
 *      so the identity provider needs no extra redirect URI.
 *   3. /auth/device/confirm asks the signed-in user to allow the device.
 *   4. Allowing sends the browser to redirect_uri?code=…&state=…; the
 *      client checks state and posts code and code_verifier to
 *      /auth/device/token, which returns the device token once.
 *
 * Codes live for a minute and work once. Without the verifier a stolen
 * code is useless, which matters because another Android app can claim
 * the same URL scheme.
 */

const (
	approvalCookie = "sonora_device"
	// AndroidRedirect is where the Android client receives the code.
	AndroidRedirect = "sonora://auth/callback"
	codeLifetime    = time.Minute
	// A device token is touched at most this often to record use.
	deviceTouchAfter = 5 * time.Minute
)

// deviceRequest is what a native client asks for in step 1.
type deviceRequest struct {
	Platform    string `json:"p"`
	Name        string `json:"m"`
	RedirectURI string `json:"r"`
	State       string `json:"s"`
	Challenge   string `json:"c"`
}

func parseDeviceRequest(q url.Values) (deviceRequest, bool) {
	req := deviceRequest{
		Platform:    q.Get("platform"),
		RedirectURI: q.Get("redirect_uri"),
		State:       q.Get("state"),
		Challenge:   q.Get("code_challenge"),
	}
	name, err := store.CleanDeviceName(q.Get("name"))
	req.Name = name
	ok := err == nil &&
		req.Platform == store.PlatformAndroid && req.RedirectURI == AndroidRedirect &&
		req.State != "" && len(req.State) <= 256 &&
		q.Get("code_challenge_method") == "S256" && validChallenge(req.Challenge)
	return req, ok
}

// validChallenge accepts the base64url SHA-256 of a verifier.
func validChallenge(c string) bool {
	b, err := base64.RawURLEncoding.DecodeString(c)
	return err == nil && len(b) == sha256.Size
}

// approval waits for the user on the confirm page; issued waits for the
// client to exchange its code.
type approval struct {
	user    store.User
	req     deviceRequest
	expires time.Time
}

type issued struct {
	userID  int64
	req     deviceRequest
	expires time.Time
}

// deviceFlows keeps both in memory: a restart only fails sign-ins that
// are under way.
type deviceFlows struct {
	mu        sync.Mutex
	approvals map[string]approval
	codes     map[string]issued
}

func newDeviceFlows() *deviceFlows {
	return &deviceFlows{approvals: map[string]approval{}, codes: map[string]issued{}}
}

func (f *deviceFlows) addApproval(a approval) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sweep()
	id := random()
	f.approvals[id] = a
	return id
}

func (f *deviceFlows) approval(id string) (approval, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	a, ok := f.approvals[id]
	return a, ok && time.Now().Before(a.expires)
}

func (f *deviceFlows) takeApproval(id string) (approval, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	a, ok := f.approvals[id]
	delete(f.approvals, id)
	return a, ok && time.Now().Before(a.expires)
}

func (f *deviceFlows) addCode(i issued) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sweep()
	code := random()
	f.codes[code] = i
	return code
}

// takeCode consumes the code whatever happens next.
func (f *deviceFlows) takeCode(code string) (issued, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	i, ok := f.codes[code]
	delete(f.codes, code)
	return i, ok && time.Now().Before(i.expires)
}

func (f *deviceFlows) sweep() {
	now := time.Now()
	for id, a := range f.approvals {
		if now.After(a.expires) {
			delete(f.approvals, id)
		}
	}
	for c, i := range f.codes {
		if now.After(i.expires) {
			delete(f.codes, c)
		}
	}
}

func (s *Service) deviceLogin(w http.ResponseWriter, r *http.Request) {
	req, ok := parseDeviceRequest(r.URL.Query())
	if !ok {
		s.devicePage(w, http.StatusBadRequest, devicePageData{Kind: "error", Title: "无法登录", Message: "登录请求无效，请从应用重新登录"})
		return
	}
	if s.devUser != "" {
		user, ok := s.authenticate(w, r)
		if !ok {
			s.devicePage(w, http.StatusInternalServerError, devicePageData{Kind: "error", Title: "无法登录", Message: "请稍后重试"})
			return
		}
		s.startApproval(w, r, user, req)
		return
	}
	s.startLogin(w, r, &req)
}

// startApproval remembers the signed-in user's pending device and shows
// the confirm page.
func (s *Service) startApproval(w http.ResponseWriter, r *http.Request, user store.User, req deviceRequest) {
	id := s.flows.addApproval(approval{user: user, req: req, expires: time.Now().Add(loginLifetime)})
	http.SetCookie(w, &http.Cookie{
		Name:     approvalCookie,
		Value:    id,
		Path:     "/auth/device/",
		MaxAge:   int(loginLifetime.Seconds()),
		HttpOnly: true,
		Secure:   s.secure,
		// Lax, not Strict: the cookie must come along on the redirect that
		// started at the identity provider.
		SameSite: http.SameSiteLaxMode,
	})
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, "/auth/device/confirm", http.StatusFound)
}

func (s *Service) deviceConfirm(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(approvalCookie)
	if err != nil {
		s.deviceExpired(w)
		return
	}
	a, ok := s.flows.approval(c.Value)
	if !ok {
		s.deviceExpired(w)
		return
	}
	s.devicePage(w, http.StatusOK, devicePageData{
		Kind:     "ask",
		Approval: c.Value,
		Device:   a.req.Name,
		User:     a.user.Name,
		Email:    a.user.Email,
	})
}

func (s *Service) deviceApprove(w http.ResponseWriter, r *http.Request) {
	if !s.sameOrigin(r) {
		s.devicePage(w, http.StatusForbidden, devicePageData{Kind: "error", Title: "无法登录", Message: "请求来源不受信任"})
		return
	}
	c, err := r.Cookie(approvalCookie)
	s.clearCookie(w, approvalCookie, "/auth/device/")
	if err != nil || subtle.ConstantTimeCompare([]byte(r.PostFormValue("approval")), []byte(c.Value)) != 1 {
		s.deviceExpired(w)
		return
	}
	a, ok := s.flows.takeApproval(c.Value)
	if !ok {
		s.deviceExpired(w)
		return
	}
	back := url.Values{"state": {a.req.State}}
	data := devicePageData{Kind: "denied"}
	if r.PostFormValue("action") == "allow" {
		back.Set("code", s.flows.addCode(issued{userID: a.user.ID, req: a.req, expires: time.Now().Add(codeLifetime)}))
		data.Kind = "allowed"
	} else {
		back.Set("error", "access_denied")
	}
	data.Redirect = a.req.RedirectURI + "?" + back.Encode()
	s.devicePage(w, http.StatusOK, data)
}

// deviceToken exchanges a code for a device token. It carries no cookie,
// so it needs no origin check.
func (s *Service) deviceToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code     string `json:"code"`
		Verifier string `json:"code_verifier"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式不正确")
		return
	}
	i, ok := s.flows.takeCode(body.Code)
	sum := sha256.Sum256([]byte(body.Verifier))
	if !ok || subtle.ConstantTimeCompare([]byte(base64.RawURLEncoding.EncodeToString(sum[:])), []byte(i.req.Challenge)) != 1 {
		writeError(w, http.StatusBadRequest, "登录已过期或无效，请重新登录")
		return
	}
	d, token, err := s.store.CreateDevice(r.Context(), i.userID, i.req.Name, i.req.Platform)
	if err != nil {
		slog.Error("create device", "err", err)
		writeError(w, http.StatusInternalServerError, "服务器内部错误")
		return
	}
	slog.Info("device signed in", "user", i.userID, "device", d.ID, "platform", d.Platform)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]any{"token": token, "device": d})
}

// bearer authenticates a native client by its device token.
func (s *Service) bearer(r *http.Request, header string) (store.User, store.Device, bool) {
	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || token == "" {
		return store.User{}, store.Device{}, false
	}
	d, u, err := s.store.DeviceUser(r.Context(), token)
	if err != nil {
		if err != store.ErrNotFound {
			slog.Error("look up device", "err", err)
		}
		return u, d, false
	}
	if time.Since(d.LastUsed) > deviceTouchAfter {
		if err := s.store.TouchDevice(r.Context(), d.ID); err != nil {
			slog.Error("touch device", "err", err)
		}
	}
	return u, d, true
}

func (s *Service) deviceExpired(w http.ResponseWriter) {
	s.devicePage(w, http.StatusBadRequest, devicePageData{Kind: "expired", Title: "登录已过期", Message: "请回到应用重新登录"})
}

type devicePageData struct {
	Kind     string // ask, allowed, denied, expired or error
	Approval string
	Device   string
	User     string
	Email    string
	Redirect string
	Title    string
	Message  string
}

// RedirectURL marks the client's address as safe for href; it was checked
// against the allow list in parseDeviceRequest.
func (d devicePageData) RedirectURL() template.URL { return template.URL(d.Redirect) }

func (s *Service) devicePage(w http.ResponseWriter, status int, data devicePageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	// Not no-referrer: under it the form posts "Origin: null" and fails sameOrigin.
	w.Header().Set("Referrer-Policy", "same-origin")
	w.WriteHeader(status)
	if err := devicePageTmpl.Execute(w, data); err != nil {
		slog.Error("render device page", "err", err)
	}
}

// The page reuses the colours of the plain login error page (pageTmpl);
// it is a placeholder look until Eric approves one. The card is centred by
// its own wrapper, not by body: browser extensions append elements to body.
var devicePageTmpl = template.Must(template.New("device").Parse(`<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="color-scheme" content="light dark"><title>Sonora 设备登录</title>
<style>
:root{--bg:#efede8;--surface:#fff;--ink:#3a3833;--ink-2:#7a776f;--line:#e2dfd8;--accent:#3a3833;--on-accent:#fff}
@media (prefers-color-scheme:dark){:root{--bg:#1f2226;--surface:#2a2d32;--ink:#d4d6da;--ink-2:#9a9da3;--line:#3a3e44;--accent:#d4d6da;--on-accent:#1f2226}}
*{box-sizing:border-box}html,body{height:100%;margin:0}
body{background:var(--bg);color:var(--ink);font:15px/1.5 system-ui,sans-serif}
.page{min-height:100%;display:grid;place-items:center;padding:24px}
.card{width:100%;max-width:360px;padding:24px;border-radius:16px;background:var(--surface);border:1px solid var(--line)}
.brand{font-weight:600;font-size:17px;margin-bottom:16px}
h1{margin:0 0 12px;font-size:18px;font-weight:600}
.rows{border:1px solid var(--line);border-radius:12px}
.row{padding:10px 14px}.row+.row{border-top:1px solid var(--line)}
.sub{font-size:13px;color:var(--ink-2)}
.actions{display:flex;gap:10px;margin-top:20px}
button{flex:1;height:40px;border-radius:10px;font:inherit;cursor:pointer;border:1px solid var(--line);background:none;color:inherit}
button.primary{background:var(--accent);border-color:var(--accent);color:var(--on-accent)}
p{margin:0;color:var(--ink-2)}a{color:inherit}
</style></head><body>
<div class="page"><main class="card">
<div class="brand">Sonora</div>
{{if eq .Kind "ask"}}
<h1>允许这台设备登录？</h1>
<div class="rows">
<div class="row"><div>{{.Device}}</div><div class="sub">安卓客户端</div></div>
<div class="row"><div>{{.User}}</div>{{if .Email}}<div class="sub">{{.Email}}</div>{{end}}</div>
</div>
<form method="post" action="/auth/device/approve" class="actions">
<input type="hidden" name="approval" value="{{.Approval}}">
<button name="action" value="deny">取消</button>
<button class="primary" name="action" value="allow">允许</button>
</form>
{{else if eq .Kind "allowed"}}
<h1>已允许登录</h1><p>正在返回应用，没有跳转的话<a href="{{.RedirectURL}}">点这里</a></p>
<script>location.replace({{.Redirect}})</script>
{{else if eq .Kind "denied"}}
<h1>已取消登录</h1>
<script>location.replace({{.Redirect}})</script>
{{else}}
<h1>{{.Title}}</h1><p>{{.Message}}</p>
{{end}}
</main></div></body></html>`))
