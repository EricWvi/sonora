package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	_ "embed"
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

//go:embed assets/gflex.woff2
var deviceFont []byte

// deviceFontPath serves the font of the confirm page.
const deviceFontPath = "/auth/device/font.woff2"

func (s *Service) deviceFontFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "font/woff2")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Write(deviceFont)
}

// The confirm page follows the visual language of PixelPlayerOSS, as the
// Android client does; PixelPlayerOSS has no such page, so the layout is
// modelled on its NavidromeLoginScreen (top app bar, round badge,
// headlineSmall title, a surfaceContainerLow card with 28dp corners, 52dp
// buttons with 18dp corners). Differences from the app, which a web page
// cannot follow:
//   - Colours are the static LightColorScheme and DarkColorScheme of
//     ui/theme/Theme.kt; roles the app leaves to Material 3 use the Material
//     3 baseline. The app itself uses the wallpaper colours on Android 12+.
//   - AbsoluteSmoothCornerShape (squircle) is drawn as plain rounded corners.
//   - The font is Google Sans Flex pinned at ROND 100 like RoundedSans,
//     subset to Latin (assets/gflex.woff2); Chinese falls back to the system
//     font, as it does in the app.
//
// The card is centred by its own wrapper, not by body: browser extensions
// append elements to body.
var devicePageTmpl = template.Must(template.New("device").Parse(`<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="color-scheme" content="light dark"><title>Sonora 设备登录</title>
<style>
@font-face{font-family:"Rounded Sans";src:url(/auth/device/font.woff2) format("woff2");font-weight:300 700;font-display:swap}
:root{--bg:#f7f2ff;--on-surface:#1e1237;--on-surface-variant:#4d4165;--primary:#6c4ff5;--on-primary:#fff;
--container-lowest:#fff;--container-low:#f7f2fa;--container-high:#ece6f0;--secondary-container:rgb(240 98 146/.15);--on-secondary-container:rgb(240 98 146/.85)}
@media (prefers-color-scheme:dark){:root{--bg:#1e1234;--on-surface:#e1bee7;--on-surface-variant:#cac4d0;--primary:#ab47bc;--on-primary:#fff;
--container-lowest:#0f0d13;--container-low:#1d1b20;--container-high:#2b2930;--secondary-container:#4a4458;--on-secondary-container:#e8def8}}
*{box-sizing:border-box}html,body{margin:0;min-height:100%}
body{background:var(--bg);color:var(--on-surface);font:400 14px/20px "Rounded Sans",system-ui,sans-serif;letter-spacing:.25px;-webkit-font-smoothing:antialiased}
.bar{height:64px;display:flex;align-items:center;padding:0 16px;background:var(--container-high);font-size:22px;line-height:28px;font-weight:600;letter-spacing:0}
.page{max-width:480px;margin:0 auto;padding:24px 20px 32px;display:flex;flex-direction:column;align-items:center;text-align:center}
.badge{width:64px;height:64px;border-radius:50%;background:#fff;display:grid;place-items:center;color:var(--primary);
box-shadow:0 1px 2px rgb(0 0 0/.3),0 2px 6px 2px rgb(0 0 0/.15)}
svg{display:block;fill:currentColor}.badge svg{width:42px;height:42px}
h1{margin:16px 0 0;font-size:24px;line-height:32px;font-weight:700;letter-spacing:0}
.sub{margin:0;color:var(--on-surface-variant)}
.card{width:100%;margin-top:24px;padding:18px;border-radius:28px;background:var(--container-low);text-align:left}
.card h2{margin:0;font-size:18px;line-height:24px;font-weight:700;letter-spacing:.15px}
.card .hint{margin:4px 0 18px;font-size:12px;line-height:16px;letter-spacing:.4px;color:var(--on-surface-variant)}
.field{display:flex;align-items:center;gap:12px;min-height:56px;padding:8px 16px 8px 12px;border-radius:18px;background:var(--container-high)}
.field+.field{margin-top:14px}.field svg{width:24px;height:24px;flex:none;color:var(--primary)}
.field .label{font-size:12px;line-height:16px;letter-spacing:.4px;color:var(--on-surface-variant)}
.field .value{font-size:16px;line-height:24px;letter-spacing:.5px;overflow-wrap:anywhere}
form{width:100%;margin:24px 0 0}
button{display:block;width:100%;height:52px;border:0;border-radius:18px;font:600 14px/20px "Rounded Sans",system-ui,sans-serif;letter-spacing:.1px;cursor:pointer}
.primary{background:var(--primary);color:var(--on-primary)}
.tonal{margin-top:12px;background:var(--secondary-container);color:var(--on-secondary-container)}
.foot{margin-top:16px;font-size:12px;line-height:16px;letter-spacing:.4px;color:var(--on-surface-variant)}
.foot a{color:var(--primary)}
</style></head><body>
<div class="bar">设备登录</div>
<main class="page">
<div class="badge">{{if eq .Kind "ask"}}<svg viewBox="0 0 24 24"><path d="M16 1H8C6.34 1 5 2.34 5 4v16c0 1.66 1.34 3 3 3h8c1.66 0 3-1.34 3-3V4c0-1.66-1.34-3-3-3zm-2.5 20h-3c-.28 0-.5-.22-.5-.5s.22-.5.5-.5h3c.28 0 .5.22.5.5s-.22.5-.5.5zm3.5-3H7V4h10v14z"/></svg>{{else if eq .Kind "allowed"}}<svg viewBox="0 0 24 24"><path d="M8.8 15.9l-3.5-3.5a.984.984 0 0 0-1.4 0c-.39.39-.39 1.01 0 1.4l4.19 4.19c.39.39 1.02.39 1.41 0L20.1 7.39c.39-.39.39-1.01 0-1.4a.984.984 0 0 0-1.4 0L8.8 15.9z"/></svg>{{else}}<svg viewBox="0 0 24 24"><path d="M18.3 5.71a.996.996 0 0 0-1.41 0L12 10.59 7.11 5.7A.996.996 0 1 0 5.7 7.11L10.59 12 5.7 16.89a.996.996 0 1 0 1.41 1.41L12 13.41l4.89 4.89a.996.996 0 1 0 1.41-1.41L13.41 12l4.89-4.89c.38-.38.38-1.02 0-1.4z"/></svg>{{end}}</div>
{{if eq .Kind "ask"}}
<h1>允许这台设备登录？</h1>
<p class="sub">允许后，这台设备上的 Sonora 会以你的账号使用曲库</p>
<section class="card">
<h2>登录信息</h2>
<p class="hint">确认是你自己的设备再允许</p>
<div class="field"><svg viewBox="0 0 24 24"><path d="M16 1H8C6.34 1 5 2.34 5 4v16c0 1.66 1.34 3 3 3h8c1.66 0 3-1.34 3-3V4c0-1.66-1.34-3-3-3zm-2.5 20h-3c-.28 0-.5-.22-.5-.5s.22-.5.5-.5h3c.28 0 .5.22.5.5s-.22.5-.5.5zm3.5-3H7V4h10v14z"/></svg><div><div class="label">安卓设备</div><div class="value">{{.Device}}</div></div></div>
<div class="field"><svg viewBox="0 0 24 24"><path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v1c0 .55.45 1 1 1h14c.55 0 1-.45 1-1v-1c0-2.66-5.33-4-8-4z"/></svg><div><div class="label">账号</div><div class="value">{{.User}}</div>{{if .Email}}<div class="label">{{.Email}}</div>{{end}}</div></div>
</section>
<form method="post" action="/auth/device/approve">
<input type="hidden" name="approval" value="{{.Approval}}">
<button class="primary" name="action" value="allow">允许</button>
<button class="tonal" name="action" value="deny">取消</button>
</form>
{{else if eq .Kind "allowed"}}
<h1>已允许登录</h1>
<p class="sub">正在返回应用</p>
<p class="foot">没有跳转的话<a href="{{.RedirectURL}}">点这里</a></p>
<script>location.replace({{.Redirect}})</script>
{{else if eq .Kind "denied"}}
<h1>已取消登录</h1>
<script>location.replace({{.Redirect}})</script>
{{else}}
<h1>{{.Title}}</h1>
<p class="sub">{{.Message}}</p>
{{end}}
</main></body></html>`))
