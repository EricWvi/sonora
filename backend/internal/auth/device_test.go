package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"sonora/internal/store"
)

// Adapted from asr (internal/auth/device_test.go), without the Mac client
// and the avatar.

func challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func deviceLoginURL(platform, redirect, verifier string) string {
	return "/auth/device/login?" + url.Values{
		"platform": {platform}, "name": {"Eric 的 Pixel"}, "redirect_uri": {redirect},
		"state": {"st"}, "code_challenge": {challenge(verifier)}, "code_challenge_method": {"S256"},
	}.Encode()
}

func request(h http.Handler, method, path, body string, header http.Header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	for k, v := range header {
		req.Header[k] = v
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func cookieOf(rec *httptest.ResponseRecorder, name string) string {
	for _, c := range rec.Result().Cookies() {
		if c.Name == name && c.MaxAge > 0 {
			return c.Value
		}
	}
	return ""
}

var redirectRe = regexp.MustCompile(`location\.replace\("([^"]+)"\)`)

// approve runs the browser side of a dev-mode device login and returns
// where the browser is sent back to the client.
func approve(t *testing.T, h http.Handler, login, action string) *url.URL {
	t.Helper()
	rec := serve(h, "GET", login)
	id := cookieOf(rec, approvalCookie)
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/auth/device/confirm" || id == "" {
		t.Fatalf("login: status = %d, location = %q", rec.Code, rec.Header().Get("Location"))
	}
	cookie := http.Header{"Cookie": {approvalCookie + "=" + id}}
	rec = request(h, "GET", "/auth/device/confirm", "", cookie)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Eric 的 Pixel") {
		t.Fatalf("confirm: status = %d, body = %s", rec.Code, rec.Body)
	}
	form := cookie.Clone()
	form.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = request(h, "POST", "/auth/device/approve", url.Values{"approval": {id}, "action": {action}}.Encode(), form)
	m := redirectRe.FindStringSubmatch(rec.Body.String())
	if rec.Code != http.StatusOK || m == nil {
		t.Fatalf("approve: status = %d, body = %s", rec.Code, rec.Body)
	}
	// The address sits in a JavaScript string.
	var raw string
	json.Unmarshal([]byte(`"`+m[1]+`"`), &raw)
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	// The approval works only once.
	if rec := request(h, "POST", "/auth/device/approve", url.Values{"approval": {id}, "action": {"allow"}}.Encode(), form); rec.Code != http.StatusBadRequest {
		t.Fatalf("second approve: status = %d", rec.Code)
	}
	return u
}

func exchange(h http.Handler, code, verifier string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]string{"code": code, "code_verifier": verifier})
	return request(h, "POST", "/auth/device/token", string(body), nil)
}

func TestDeviceLoginIssuesTokenOnce(t *testing.T) {
	svc, h := newService(t, Config{DevUser: "eric", DevGroups: []string{"sonora-import"}})
	back := approve(t, h, deviceLoginURL("android", AndroidRedirect, "verifier-1"), "allow")
	code := back.Query().Get("code")
	if !strings.HasPrefix(back.String(), AndroidRedirect+"?") || back.Query().Get("state") != "st" || code == "" {
		t.Fatalf("redirect = %s", back)
	}

	// A wrong verifier burns the code.
	if rec := exchange(h, code, "verifier-2"); rec.Code != http.StatusBadRequest {
		t.Fatalf("wrong verifier: status = %d", rec.Code)
	}
	if rec := exchange(h, code, "verifier-1"); rec.Code != http.StatusBadRequest {
		t.Fatalf("burnt code: status = %d", rec.Code)
	}

	back = approve(t, h, deviceLoginURL("android", AndroidRedirect, "verifier-3"), "allow")
	rec := exchange(h, back.Query().Get("code"), "verifier-3")
	var out struct {
		Token  string       `json:"token"`
		Device store.Device `json:"device"`
	}
	json.Unmarshal(rec.Body.Bytes(), &out)
	if rec.Code != http.StatusOK || !strings.HasPrefix(out.Token, store.DeviceTokenPrefix) || out.Device.Platform != "android" || out.Device.Name != "Eric 的 Pixel" {
		t.Fatalf("token: status = %d, body = %s", rec.Code, rec.Body)
	}
	if rec := exchange(h, back.Query().Get("code"), "verifier-3"); rec.Code != http.StatusBadRequest {
		t.Fatalf("reused code: status = %d", rec.Code)
	}

	// The token signs requests in as the device, from any origin, and the
	// user keeps its groups.
	rec = request(h, "POST", "/whoami", "", http.Header{"Authorization": {"Bearer " + out.Token}, "Origin": {"https://evil.test"}})
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"eric"`) {
		t.Fatalf("whoami: status = %d, body = %s", rec.Code, rec.Body)
	}
	_, u, err := svc.store.DeviceUser(t.Context(), out.Token)
	if err != nil || !slices.Equal(u.Groups, []string{"sonora-import"}) {
		t.Fatalf("device user = %+v, %v", u, err)
	}
	if rec := request(h, "GET", "/whoami", "", http.Header{"Authorization": {"Bearer nope"}}); rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad token: status = %d", rec.Code)
	}
}

func TestDeviceLoginCanBeDenied(t *testing.T) {
	_, h := newService(t, Config{DevUser: "eric"})
	back := approve(t, h, deviceLoginURL("android", AndroidRedirect, "v"), "deny")
	q := back.Query()
	if q.Get("error") != "access_denied" || q.Get("state") != "st" || q.Has("code") {
		t.Fatalf("redirect = %s", back)
	}
}

func TestDeviceLoginRejectsBadRequests(t *testing.T) {
	_, h := newService(t, Config{DevUser: "eric"})
	good := deviceLoginURL("android", AndroidRedirect, "v")
	if rec := serve(h, "GET", good); rec.Code != http.StatusFound {
		t.Fatalf("good request: status = %d", rec.Code)
	}
	for _, path := range []string{
		deviceLoginURL("android", "craft://auth/callback", "v"),
		deviceLoginURL("android", "https://evil.test/callback", "v"),
		deviceLoginURL("mac", "http://127.0.0.1:53682/callback", "v"),
		deviceLoginURL("mac", AndroidRedirect, "v"),
		strings.Replace(good, "S256", "plain", 1),
		strings.Replace(good, "state=st", "state=", 1),
		strings.Replace(good, "code_challenge="+challenge("v"), "code_challenge=short", 1),
	} {
		if rec := serve(h, "GET", path); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d", path, rec.Code)
		}
	}
}

func TestDeviceApprovalChecks(t *testing.T) {
	svc, h := newService(t, Config{PublicURL: "https://sonora.test", Issuer: "https://idp.invalid", ClientID: "x", ClientSecret: "y"})
	req := deviceRequest{Platform: "android", Name: "Pixel", RedirectURI: AndroidRedirect, State: "st", Challenge: challenge("v")}
	id := svc.flows.addApproval(approval{user: store.User{ID: 1, Name: "eric"}, req: req, expires: time.Now().Add(time.Minute)})
	form := http.Header{"Cookie": {approvalCookie + "=" + id}, "Content-Type": {"application/x-www-form-urlencoded"}}
	body := url.Values{"approval": {id}, "action": {"allow"}}.Encode()

	// Another site cannot submit the form.
	cross := form.Clone()
	cross.Set("Origin", "https://evil.test")
	if rec := request(h, "POST", "/auth/device/approve", body, cross); rec.Code != http.StatusForbidden {
		t.Fatalf("cross-origin approve: status = %d", rec.Code)
	}
	// The form must name the approval in the cookie.
	other := url.Values{"approval": {"other"}, "action": {"allow"}}.Encode()
	if rec := request(h, "POST", "/auth/device/approve", other, form); rec.Code != http.StatusBadRequest {
		t.Fatalf("mismatched approve: status = %d", rec.Code)
	}

	expired := svc.flows.addApproval(approval{req: req, expires: time.Now().Add(-time.Second)})
	if rec := request(h, "GET", "/auth/device/confirm", "", http.Header{"Cookie": {approvalCookie + "=" + expired}}); rec.Code != http.StatusBadRequest {
		t.Fatalf("expired confirm: status = %d", rec.Code)
	}
	code := svc.flows.addCode(issued{userID: 1, req: req, expires: time.Now().Add(-time.Second)})
	if rec := exchange(h, code, "v"); rec.Code != http.StatusBadRequest {
		t.Fatalf("expired code: status = %d", rec.Code)
	}
}

// Under Referrer-Policy: no-referrer a browser posts the form with
// "Origin: null", which the origin check refuses.
func TestDeviceConfirmLetsTheFormNameItsOrigin(t *testing.T) {
	svc, h := newService(t, Config{PublicURL: "https://sonora.test", Issuer: "https://idp.invalid", ClientID: "x", ClientSecret: "y"})
	req := deviceRequest{Platform: "android", Name: "Pixel", RedirectURI: AndroidRedirect, State: "st", Challenge: challenge("v")}
	id := svc.flows.addApproval(approval{user: store.User{ID: 1, Name: "eric"}, req: req, expires: time.Now().Add(time.Minute)})
	rec := request(h, "GET", "/auth/device/confirm", "", http.Header{"Cookie": {approvalCookie + "=" + id}})
	if got := rec.Header().Get("Referrer-Policy"); got != "same-origin" {
		t.Fatalf("Referrer-Policy = %q, want same-origin", got)
	}
}

func TestDeviceConfirmUsesClientColors(t *testing.T) {
	_, h := newService(t, Config{DevUser: "eric"})
	colors := "d0bcff,381e72,141218,e6e0e9,cac4d0,0f0d13,1d1b20,2b2930,4a4458,e8def8"
	rec := serve(h, "GET", deviceLoginURL("android", AndroidRedirect, "v")+"&colors="+url.QueryEscape(colors))
	cookie := http.Header{"Cookie": {approvalCookie + "=" + cookieOf(rec, approvalCookie)}}
	body := request(h, "GET", "/auth/device/confirm", "", cookie).Body.String()
	if !strings.Contains(body, "--primary:#d0bcff;") || !strings.Contains(body, "color-scheme:dark}") {
		t.Fatalf("confirm page without the client colours: %s", body)
	}
	// Without colors the static scheme stays.
	rec = serve(h, "GET", deviceLoginURL("android", AndroidRedirect, "v"))
	cookie = http.Header{"Cookie": {approvalCookie + "=" + cookieOf(rec, approvalCookie)}}
	if body := request(h, "GET", "/auth/device/confirm", "", cookie).Body.String(); strings.Contains(body, "color-scheme:dark}") {
		t.Fatalf("static page got a client theme: %s", body)
	}
	for _, bad := range []string{"d0bcff", strings.Replace(colors, "d0bcff", "red;}x", 1), strings.Replace(colors, "d0bcff", "d0bcfg", 1)} {
		if rec := serve(h, "GET", deviceLoginURL("android", AndroidRedirect, "v")+"&colors="+url.QueryEscape(bad)); rec.Code != http.StatusBadRequest {
			t.Errorf("colors %q: status = %d", bad, rec.Code)
		}
	}
}
