package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sonora/internal/netease"
	"sonora/internal/store"
)

// fakeNetease stands in for the API service: a cookie is valid when it is in
// accounts, and every answer echoes cookies as the real service does.
type fakeNetease struct {
	srv      *httptest.Server
	accounts map[string]int64 // MUSIC_U value to account ID
	qrCode   int
	qrCookie string
	down     bool
	calls    map[string]int
}

func newFakeNetease(t *testing.T) *fakeNetease {
	t.Helper()
	f := &fakeNetease{accounts: map[string]int64{}, qrCode: netease.QRWaiting, calls: map[string]int{}}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f.down {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		f.calls[r.URL.Path]++
		r.ParseForm()
		cookie := r.PostForm.Get("cookie")
		switch r.URL.Path {
		case "/login/status":
			var id int64
			for value, account := range f.accounts {
				if strings.Contains(cookie, "MUSIC_U="+value) {
					id = account
				}
			}
			profile := "null"
			if id != 0 {
				profile = `{"userId":` + itoa(id) + `,"nickname":"N` + itoa(id) + `","avatarUrl":"https://a/b.jpg","vipType":11}`
			}
			w.Write([]byte(`{"data":{"code":200,"cookie":"MUSIC_U=echoed","profile":` + profile + `}}`))
		case "/login/qr/key":
			w.Write([]byte(`{"data":{"code":200,"unikey":"key1"},"cookie":"NMTID=echoed"}`))
		case "/login/qr/check":
			body, _ := json.Marshal(map[string]any{
				"code": f.qrCode, "nickname": "Scanner", "cookie": f.qrCookie,
			})
			w.Write(body)
		case "/login/refresh":
			w.Write([]byte(`{"code":200,"cookie":"MUSIC_U=renewed; Path=/"}`))
		default:
			t.Errorf("unexpected call to %s", r.URL.Path)
		}
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func newNeteaseTest(t *testing.T, f *fakeNetease) *testServer {
	t.Helper()
	ts := newTest(t)
	ts.opts.Netease = netease.New(f.srv.URL)
	ts.restart()
	return ts
}

func neteaseState(t *testing.T, ts *testServer) neteaseStatus {
	t.Helper()
	rec := ts.do("GET", playerAPI+"netease/status", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d %s", rec.Code, rec.Body)
	}
	assertNoCookie(t, rec)
	return decodeBody[neteaseStatus](t, rec)
}

func assertNoCookie(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	body := rec.Body.String()
	for _, secret := range []string{"MUSIC_U", "NMTID", "cookie"} {
		if strings.Contains(body, secret) {
			t.Fatalf("response leaks %s: %s", secret, body)
		}
	}
}

func TestNeteaseDisabled(t *testing.T) {
	ts := newTest(t)
	ts.signIn("alice", "Alice")
	if got := neteaseState(t, ts); got.State != neteaseDisabled {
		t.Fatalf("state = %q", got.State)
	}
	for _, path := range []string{"netease/login/qr", "netease/login/cookie"} {
		rec := ts.do("POST", playerAPI+path, `{"cookie":"MUSIC_U=a"}`)
		if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "netease_disabled") {
			t.Fatalf("%s: %d %s", path, rec.Code, rec.Body)
		}
	}
}

func TestNeteaseRequiresListenerSession(t *testing.T) {
	ts := newNeteaseTest(t, newFakeNetease(t))
	if rec := ts.do("GET", playerAPI+"netease/status", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status: %d", rec.Code)
	}
}

func TestNeteaseCookieLogin(t *testing.T) {
	f := newFakeNetease(t)
	f.accounts["good"] = 7
	ts := newNeteaseTest(t, f)
	ts.signIn("alice", "Alice")
	if got := neteaseState(t, ts); got.State != neteaseSignedOut || got.Profile != nil {
		t.Fatalf("before login: %+v", got)
	}

	for _, body := range []string{`{"cookie":"nonsense"}`, `{"cookie":"MUSIC_U=stale;"}`} {
		rec := ts.do("POST", playerAPI+"netease/login/cookie", body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: %d %s", body, rec.Code, rec.Body)
		}
	}
	if got := neteaseState(t, ts); got.State != neteaseSignedOut {
		t.Fatalf("a rejected cookie was kept: %+v", got)
	}

	rec := ts.do("POST", playerAPI+"netease/login/cookie", `{"cookie":"NMTID=n; MUSIC_U=good; Path=/;"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("login: %d %s", rec.Code, rec.Body)
	}
	assertNoCookie(t, rec)
	got := neteaseState(t, ts)
	if got.State != neteaseSignedIn || got.Profile.UserID != 7 || got.Profile.VIPType != 11 {
		t.Fatalf("after login: %+v", got)
	}

	if rec := ts.do("POST", playerAPI+"netease/logout", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("logout: %d %s", rec.Code, rec.Body)
	}
	if got := neteaseState(t, ts); got.State != neteaseSignedOut {
		t.Fatalf("after logout: %+v", got)
	}
}

func TestNeteaseQRLogin(t *testing.T) {
	f := newFakeNetease(t)
	f.accounts["scanned"] = 9
	ts := newNeteaseTest(t, f)
	ts.signIn("alice", "Alice")

	rec := ts.do("POST", playerAPI+"netease/login/qr", "")
	qr := decodeBody[map[string]string](t, rec)
	if rec.Code != http.StatusOK || qr["key"] != "key1" || qr["url"] != netease.QRURL("key1") {
		t.Fatalf("qr key: %d %s", rec.Code, rec.Body)
	}
	assertNoCookie(t, rec)

	f.qrCode, f.qrCookie = netease.QRConfirm, "NMTID=anonymous"
	rec = ts.do("GET", playerAPI+"netease/login/qr/key1", "")
	if rec.Code != http.StatusOK || decodeBody[map[string]any](t, rec)["nickname"] != "Scanner" {
		t.Fatalf("confirming: %d %s", rec.Code, rec.Body)
	}
	assertNoCookie(t, rec)
	if got := neteaseState(t, ts); got.State != neteaseSignedOut {
		t.Fatalf("signed in before confirming: %+v", got)
	}

	f.qrCode, f.qrCookie = netease.QRSucceeded, "MUSIC_U=scanned; Max-Age=100; Path=/;__csrf=c; Path=/"
	rec = ts.do("GET", playerAPI+"netease/login/qr/key1", "")
	if rec.Code != http.StatusOK || decodeBody[map[string]any](t, rec)["code"] != float64(netease.QRSucceeded) {
		t.Fatalf("succeeded: %d %s", rec.Code, rec.Body)
	}
	assertNoCookie(t, rec)
	if got := neteaseState(t, ts); got.State != neteaseSignedIn || got.Profile.UserID != 9 {
		t.Fatalf("after qr: %+v", got)
	}
}

func TestNeteaseSessionsBelongToOneListener(t *testing.T) {
	f := newFakeNetease(t)
	f.accounts["alice"], f.accounts["bob"] = 1, 2
	ts := newNeteaseTest(t, f)

	ts.signIn("alice", "Alice")
	alice := ts.listener
	if rec := ts.do("POST", playerAPI+"netease/login/cookie", `{"cookie":"MUSIC_U=alice"}`); rec.Code != http.StatusOK {
		t.Fatalf("alice: %d %s", rec.Code, rec.Body)
	}
	ts.signIn("bob", "Bob")
	if got := neteaseState(t, ts); got.State != neteaseSignedOut {
		t.Fatalf("bob sees alice's session: %+v", got)
	}
	if rec := ts.do("POST", playerAPI+"netease/login/cookie", `{"cookie":"MUSIC_U=bob"}`); rec.Code != http.StatusOK {
		t.Fatalf("bob: %d %s", rec.Code, rec.Body)
	}
	ts.listener = alice
	if got := neteaseState(t, ts); got.Profile == nil || got.Profile.UserID != 1 {
		t.Fatalf("alice after bob: %+v", got)
	}
	// Signing out of the player leaves the NetEase session alone.
	if rec := ts.do("POST", "/auth/logout", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("logout: %d", rec.Code)
	}
	ts.signIn("alice", "Alice")
	if got := neteaseState(t, ts); got.Profile == nil || got.Profile.UserID != 1 {
		t.Fatalf("alice after player logout: %+v", got)
	}
}

func TestNeteaseExpiredAndUnavailable(t *testing.T) {
	f := newFakeNetease(t)
	f.accounts["good"] = 7
	ts := newNeteaseTest(t, f)
	ts.signIn("alice", "Alice")
	if rec := ts.do("POST", playerAPI+"netease/login/cookie", `{"cookie":"MUSIC_U=good"}`); rec.Code != http.StatusOK {
		t.Fatalf("login: %d %s", rec.Code, rec.Body)
	}

	f.down = true
	if got := neteaseState(t, ts); got.State != neteaseUnavailable {
		t.Fatalf("service down: %+v", got)
	}
	rec := ts.do("POST", playerAPI+"netease/login/qr", "")
	if rec.Code != http.StatusBadGateway || !strings.Contains(rec.Body.String(), "netease_unavailable") {
		t.Fatalf("qr while down: %d %s", rec.Code, rec.Body)
	}
	f.down = false
	if got := neteaseState(t, ts); got.State != neteaseSignedIn {
		t.Fatalf("service back: %+v", got)
	}

	delete(f.accounts, "good")
	if got := neteaseState(t, ts); got.State != neteaseExpired {
		t.Fatalf("dead session: %+v", got)
	}
}

func TestNeteaseRefreshesAgingSession(t *testing.T) {
	f := newFakeNetease(t)
	f.accounts["old"], f.accounts["renewed"] = 7, 7
	ts := newNeteaseTest(t, f)
	ts.signIn("alice", "Alice")
	ctx := context.Background()
	u, err := ts.store.UpsertUser(ctx, store.Identity{Issuer: "https://auth.test", Subject: "alice", Name: "Alice"})
	if err != nil {
		t.Fatal(err)
	}

	fresh := time.Now().Add(-neteaseRefreshAge + time.Hour)
	if err := ts.store.SaveNeteaseSession(ctx, u.ID, store.NeteaseSession{Cookie: "MUSIC_U=old", AccountID: 7, RefreshedAt: fresh}); err != nil {
		t.Fatal(err)
	}
	neteaseState(t, ts)
	if f.calls["/login/refresh"] != 0 {
		t.Fatal("refreshed a fresh session")
	}

	stale := time.Now().Add(-neteaseRefreshAge - time.Hour)
	if err := ts.store.SaveNeteaseSession(ctx, u.ID, store.NeteaseSession{Cookie: "MUSIC_U=old", AccountID: 7, RefreshedAt: stale}); err != nil {
		t.Fatal(err)
	}
	neteaseState(t, ts)
	got, err := ts.store.NeteaseSession(ctx, u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if f.calls["/login/refresh"] != 1 || got.Cookie != "MUSIC_U=renewed" || time.Since(got.RefreshedAt) > time.Minute {
		t.Fatalf("refresh: %d calls, session %+v", f.calls["/login/refresh"], got.RefreshedAt)
	}
}
