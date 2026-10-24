package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"golang.org/x/crypto/bcrypt"

	"sonora/internal/admin"
	"sonora/internal/store"
)

const (
	origin = "https://music.test"
	api    = admin.APIPrefix
)

type testServer struct {
	http.Handler
	t        *testing.T
	store    *store.Store
	authn    *admin.Service
	mediaDir string
	now      time.Time
	// cookie is sent on every request made with do.
	cookie string
}

func newTest(t *testing.T) *testServer {
	t.Helper()
	mediaDir := t.TempDir()
	st, err := store.Open(t.TempDir(), mediaDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	ts := &testServer{t: t, store: st, mediaDir: mediaDir, now: time.Unix(1_800_000_000, 0)}
	ts.restart()
	return ts
}

// restart builds a fresh service on the same store, as a process restart
// would.
func (ts *testServer) restart() {
	ts.t.Helper()
	authn, err := admin.New(context.Background(), ts.store, admin.Config{
		PublicURL:  origin,
		Now:        func() time.Time { return ts.now },
		BcryptCost: bcrypt.MinCost,
	})
	if err != nil {
		ts.t.Fatal(err)
	}
	player := fstest.MapFS{"index.html": {Data: []byte("player")}}
	console := fstest.MapFS{
		"index.html":      {Data: []byte("console")},
		"assets/app-1.js": {Data: []byte("console.log(1)")},
		"favicon.svg":     {Data: []byte("<svg/>")},
		".gitkeep":        {Data: nil},
	}
	h, err := New(ts.store, authn, player, console, Options{Version: "test"})
	if err != nil {
		ts.t.Fatal(err)
	}
	ts.Handler, ts.authn = h, authn
}

func (ts *testServer) send(req *http.Request) *httptest.ResponseRecorder {
	if ts.cookie != "" {
		req.AddCookie(&http.Cookie{Name: "sonora_admin_session", Value: ts.cookie})
	}
	rec := httptest.NewRecorder()
	ts.ServeHTTP(rec, req)
	return rec
}

func (ts *testServer) do(method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if method != http.MethodGet {
		req.Header.Set("Origin", origin)
	}
	return ts.send(req)
}

// setup initialises the service and keeps the session cookie.
func (ts *testServer) setup() {
	ts.t.Helper()
	body := `{"token":"` + ts.authn.SetupToken() + `","username":"eric","password":"correct horse"}`
	rec := ts.do("POST", api+"setup", body)
	if rec.Code != http.StatusOK {
		ts.t.Fatalf("setup: %d %s", rec.Code, rec.Body)
	}
	ts.cookie = sessionCookie(ts.t, rec)
}

func sessionCookie(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	for _, c := range rec.Result().Cookies() {
		if c.Name == "sonora_admin_session" {
			if c.Path != api || !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteStrictMode {
				t.Fatalf("cookie attributes = %+v", c)
			}
			return c.Value
		}
	}
	t.Fatal("no session cookie")
	return ""
}

func decodeBody[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %s: %v", rec.Body, err)
	}
	return v
}

func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	e := decodeBody[admin.Error](t, rec)
	if e.Message == "" {
		t.Fatalf("error without message: %s", rec.Body)
	}
	return e.Code
}

func TestUninitialisedServiceOnlyAllowsSetup(t *testing.T) {
	ts := newTest(t)
	if ts.authn.SetupToken() == "" {
		t.Fatal("no setup token before initialisation")
	}
	st := decodeBody[admin.Status](t, ts.do("GET", api+"status", ""))
	if st.Initialized || st.Admin != nil {
		t.Fatalf("status = %+v", st)
	}
	for _, path := range []string{"overview", "tracks", "albums"} {
		rec := ts.do("GET", api+path, "")
		if rec.Code != http.StatusForbidden || errorCode(t, rec) != "not_initialized" {
			t.Errorf("%s: %d %s", path, rec.Code, rec.Body)
		}
	}
	rec := ts.do("POST", api+"login", `{"username":"eric","password":"correct horse"}`)
	if rec.Code != http.StatusForbidden {
		t.Errorf("login before setup: %d", rec.Code)
	}
}

func TestSetupRejectsMissingWrongAndStaleTokens(t *testing.T) {
	ts := newTest(t)
	creds := `","username":"eric","password":"correct horse"}`
	for name, body := range map[string]string{
		"missing": `{"username":"eric","password":"correct horse"}`,
		"wrong":   `{"token":"guess` + creds,
	} {
		rec := ts.do("POST", api+"setup", body)
		if rec.Code != http.StatusForbidden || errorCode(t, rec) != "invalid_token" {
			t.Errorf("%s token: %d %s", name, rec.Code, rec.Body)
		}
	}
	stale := ts.authn.SetupToken()
	ts.restart()
	if ts.authn.SetupToken() == stale {
		t.Fatal("restart kept the token")
	}
	rec := ts.do("POST", api+"setup", `{"token":"`+stale+creds)
	if rec.Code != http.StatusForbidden {
		t.Errorf("token from before restart: %d %s", rec.Code, rec.Body)
	}
	rec = ts.do("POST", api+"setup", `{"token":"`+ts.authn.SetupToken()+`","username":"eric","password":"short"}`)
	if rec.Code != http.StatusBadRequest || errorCode(t, rec) != "invalid" {
		t.Errorf("weak password: %d %s", rec.Code, rec.Body)
	}
}

func TestSetupRateLimitsWrongTokens(t *testing.T) {
	ts := newTest(t)
	for range 5 {
		ts.do("POST", api+"setup", `{"token":"guess","username":"eric","password":"correct horse"}`)
	}
	rec := ts.do("POST", api+"setup", `{"token":"`+ts.authn.SetupToken()+`","username":"eric","password":"correct horse"}`)
	if rec.Code != http.StatusTooManyRequests || rec.Header().Get("Retry-After") == "" {
		t.Fatalf("after 5 failures: %d %s", rec.Code, rec.Body)
	}
	ts.now = ts.now.Add(time.Minute)
	ts.setup()
}

func TestSetupSucceedsOnceAndRevokesToken(t *testing.T) {
	ts := newTest(t)
	token := ts.authn.SetupToken()
	ts.setup()
	if ts.authn.SetupToken() != "" {
		t.Fatal("token still valid after setup")
	}
	rec := ts.do("POST", api+"setup", `{"token":"`+token+`","username":"mallory","password":"correct horse"}`)
	if rec.Code != http.StatusConflict || errorCode(t, rec) != "already_initialized" {
		t.Fatalf("second setup: %d %s", rec.Code, rec.Body)
	}
	st := decodeBody[admin.Status](t, ts.do("GET", api+"status", ""))
	if !st.Initialized || st.Admin == nil || st.Admin.Username != "eric" {
		t.Fatalf("status = %+v", st)
	}
	// A restart after setup does not generate a token again.
	ts.restart()
	if ts.authn.SetupToken() != "" {
		t.Fatal("token generated after initialisation")
	}
}

func TestLoginLogoutAndExpiry(t *testing.T) {
	ts := newTest(t)
	ts.setup()
	ts.cookie = ""
	if rec := ts.do("GET", api+"overview", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("without session: %d", rec.Code)
	}
	if rec := ts.do("POST", api+"login", `{"username":"eric","password":"wrong password"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password: %d", rec.Code)
	}
	rec := ts.do("POST", api+"login", `{"username":"ERIC","password":"correct horse"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("login: %d %s", rec.Code, rec.Body)
	}
	ts.cookie = sessionCookie(t, rec)
	if rec := ts.do("GET", api+"overview", ""); rec.Code != http.StatusOK {
		t.Fatalf("with session: %d", rec.Code)
	}

	// Use keeps the session alive past the idle timeout...
	ts.now = ts.now.Add(admin.IdleTimeout - time.Hour)
	if rec := ts.do("GET", api+"overview", ""); rec.Code != http.StatusOK {
		t.Fatalf("before idle timeout: %d", rec.Code)
	}
	// ...but idleness ends it.
	ts.now = ts.now.Add(admin.IdleTimeout)
	if rec := ts.do("GET", api+"overview", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("after idle timeout: %d", rec.Code)
	}

	rec = ts.do("POST", api+"login", `{"username":"eric","password":"correct horse"}`)
	ts.cookie = sessionCookie(t, rec)
	if rec := ts.do("POST", api+"logout", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("logout: %d", rec.Code)
	}
	if rec := ts.do("GET", api+"overview", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("after logout: %d", rec.Code)
	}
}

func TestSessionHasHardLifetime(t *testing.T) {
	ts := newTest(t)
	ts.setup()
	for elapsed := time.Duration(0); elapsed < admin.MaxLifetime-24*time.Hour; elapsed += 24 * time.Hour {
		ts.now = ts.now.Add(24 * time.Hour)
		if rec := ts.do("GET", api+"overview", ""); rec.Code != http.StatusOK {
			t.Fatalf("day %v: %d", elapsed, rec.Code)
		}
	}
	ts.now = ts.now.Add(48 * time.Hour)
	if rec := ts.do("GET", api+"overview", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("past max lifetime: %d", rec.Code)
	}
}

func TestCrossOriginWritesRejected(t *testing.T) {
	ts := newTest(t)
	ts.setup()
	req := httptest.NewRequest("POST", api+"logout", nil)
	req.Header.Set("Origin", "https://evil.test")
	if rec := ts.send(req); rec.Code != http.StatusForbidden {
		t.Fatalf("cross-origin logout: %d", rec.Code)
	}
	req = httptest.NewRequest("DELETE", api+"tracks/1", nil)
	req.Header.Set("Origin", "https://evil.test")
	if rec := ts.send(req); rec.Code != http.StatusForbidden || errorCode(t, rec) != "bad_origin" {
		t.Fatalf("cross-origin delete: %d", rec.Code)
	}
}

func sample(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "media", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func (ts *testServer) upload(name string, data []byte, lyrics string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", name)
	fw.Write(data)
	if lyrics != "" {
		lw, _ := mw.CreateFormFile("lyrics", "lyrics.lrc")
		lw.Write([]byte(lyrics))
	}
	mw.Close()
	req := httptest.NewRequest("POST", api+"tracks", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Origin", origin)
	return ts.send(req)
}

func TestUploadOutcomes(t *testing.T) {
	ts := newTest(t)
	ts.setup()

	rec := ts.upload("晴天.mp3", sample(t, "tagged.mp3"), "[00:00.00]外置歌词")
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload: %d %s", rec.Code, rec.Body)
	}
	res := decodeBody[store.IngestResult](t, rec)
	if res.Status != store.Created || res.Track.Lyrics != "[00:00.00]外置歌词" {
		t.Fatalf("result = %+v", res)
	}
	// File paths stay inside the service.
	if strings.Contains(rec.Body.String(), ts.mediaDir) || strings.Contains(rec.Body.String(), `"file"`) ||
		strings.Contains(rec.Body.String(), "sha256") {
		t.Fatalf("response leaks storage details: %s", rec.Body)
	}

	rec = ts.upload("again.mp3", sample(t, "tagged.mp3"), "")
	if rec.Code != http.StatusOK || decodeBody[store.IngestResult](t, rec).Status != store.Duplicate {
		t.Fatalf("duplicate: %d %s", rec.Code, rec.Body)
	}
	rec = ts.upload("bad.flac", sample(t, "truncated.flac"), "")
	if rec.Code != http.StatusUnprocessableEntity || errorCode(t, rec) != "corrupt_file" {
		t.Fatalf("corrupt: %d %s", rec.Code, rec.Body)
	}
	rec = ts.upload("cover.mp3", sample(t, "cover.jpg"), "")
	if rec.Code != http.StatusUnsupportedMediaType || errorCode(t, rec) != "unsupported_format" {
		t.Fatalf("unsupported: %d %s", rec.Code, rec.Body)
	}
	rec = ts.do("POST", api+"tracks", "not multipart")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("not multipart: %d", rec.Code)
	}

	o := decodeBody[store.Overview](t, ts.do("GET", api+"overview", ""))
	if o.Tracks != 1 || o.Albums != 1 || o.Artists != 1 {
		t.Fatalf("overview = %+v", o)
	}
}

func TestLibraryEndpoints(t *testing.T) {
	ts := newTest(t)
	ts.setup()
	tr := decodeBody[store.IngestResult](t, ts.upload("a.flac", sample(t, "tagged.flac"), "")).Track
	other := decodeBody[store.IngestResult](t, ts.upload("b.mp3", sample(t, "untagged.mp3"), "")).Track

	missing := decodeBody[store.Page[store.Track]](t, ts.do("GET", api+"tracks?missing=true", ""))
	if missing.Total != 1 || missing.Items[0].ID != other.ID {
		t.Fatalf("missing = %+v", missing)
	}
	rec := ts.do("PUT", api+"tracks/"+itoa(other.ID), `{"title":"B","artists":["Jay Chou"],"album":"叶惠美 (Live)"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("update: %d %s", rec.Code, rec.Body)
	}
	b := decodeBody[store.TrackDetail](t, rec)

	rec = ts.do("POST", api+"artists/"+itoa(tr.Artists[0].ID)+"/merge", `{"sourceIds":[`+itoa(b.Artists[0].ID)+`]}`)
	if rec.Code != http.StatusOK || decodeBody[store.Artist](t, rec).TrackCount != 2 {
		t.Fatalf("merge artists: %d %s", rec.Code, rec.Body)
	}
	rec = ts.do("POST", api+"albums/"+itoa(tr.Album.ID)+"/merge", `{"sourceIds":[`+itoa(b.Album.ID)+`]}`)
	if rec.Code != http.StatusOK || decodeBody[store.Album](t, rec).TrackCount != 2 {
		t.Fatalf("merge albums: %d %s", rec.Code, rec.Body)
	}
	albums := decodeBody[store.Page[store.Album]](t, ts.do("GET", api+"albums", ""))
	if albums.Total != 1 || albums.Items[0].Cover == nil {
		t.Fatalf("albums = %+v", albums)
	}

	// The cover is served by its hash name.
	cover := albums.Items[0].Cover
	if cover.Thumbhash == "" {
		t.Fatalf("cover without placeholder: %+v", cover)
	}
	rec = ts.do("GET", api+"covers/"+cover.Name, "")
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatalf("cover: %d %s", rec.Code, rec.Header())
	}
	rec = ts.do("GET", api+"covers/"+cover.Name+"?size=96", "")
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/jpeg" ||
		!strings.Contains(rec.Header().Get("Cache-Control"), "immutable") {
		t.Fatalf("cover thumb: %d %s", rec.Code, rec.Header())
	}
	for _, size := range []string{"97", "big"} {
		if rec := ts.do("GET", api+"covers/"+cover.Name+"?size="+size, ""); rec.Code != http.StatusBadRequest {
			t.Fatalf("size %s: %d", size, rec.Code)
		}
	}
	if rec := ts.do("GET", api+"covers/..%2Fsonora.db", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("cover traversal: %d", rec.Code)
	}

	rec = ts.do("PUT", api+"tracks/"+itoa(tr.ID)+"/lyrics", "[00:01.00]改过的歌词")
	if rec.Code != http.StatusOK || decodeBody[store.TrackDetail](t, rec).Lyrics != "[00:01.00]改过的歌词" {
		t.Fatalf("lyrics: %d %s", rec.Code, rec.Body)
	}
	rec = ts.do("PUT", api+"albums/"+itoa(tr.Album.ID)+"/cover", "garbage")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("garbage cover: %d", rec.Code)
	}

	// Range requests let the console seek in a preview.
	req := httptest.NewRequest("GET", api+"tracks/"+itoa(tr.ID)+"/audio", nil)
	req.Header.Set("Range", "bytes=0-3")
	rec = ts.send(req)
	if rec.Code != http.StatusPartialContent || rec.Body.String() != "fLaC" || rec.Header().Get("Content-Type") != "audio/flac" {
		t.Fatalf("audio range: %d %q %s", rec.Code, rec.Body, rec.Header())
	}

	if rec := ts.do("DELETE", api+"albums/"+itoa(tr.Album.ID), ""); rec.Code != http.StatusNoContent {
		t.Fatalf("delete album: %d %s", rec.Code, rec.Body)
	}
	if rec := ts.do("GET", api+"tracks/"+itoa(tr.ID), ""); rec.Code != http.StatusNotFound || errorCode(t, rec) != "not_found" {
		t.Fatalf("deleted track: %d", rec.Code)
	}
	if rec := ts.do("GET", api+"nope", ""); rec.Code != http.StatusNotFound || errorCode(t, rec) != "not_found" {
		t.Fatalf("unknown endpoint: %d %s", rec.Code, rec.Body)
	}
}

func TestStaticApps(t *testing.T) {
	ts := newTest(t)
	for path, want := range map[string]string{
		"/":                      "player",
		"/library/albums":        "player",
		"/admin/":                "console",
		"/admin/tracks":          "console",
		"/admin/assets/app-1.js": "console.log(1)",
	} {
		rec := ts.do("GET", path, "")
		if rec.Code != http.StatusOK || rec.Body.String() != want {
			t.Errorf("%s: %d %q", path, rec.Code, rec.Body)
		}
	}
	if rec := ts.do("GET", "/admin", ""); rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != "/admin/" {
		t.Errorf("/admin: %d %s", rec.Code, rec.Header())
	}
	if rec := ts.do("GET", "/admin/assets/missing.js", ""); rec.Code != http.StatusNotFound {
		t.Errorf("missing asset: %d", rec.Code)
	}
	if rec := ts.do("GET", "/admin/.gitkeep", ""); rec.Code != http.StatusNotFound {
		t.Errorf("dotfile: %d", rec.Code)
	}
	rec := ts.do("GET", "/api/version", "")
	if body, _ := io.ReadAll(rec.Body); !strings.Contains(string(body), `"test"`) {
		t.Errorf("version: %s", body)
	}
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

func TestConcurrentSetupOnlyOneWins(t *testing.T) {
	ts := newTest(t)
	token := ts.authn.SetupToken()
	codes := make(chan int, 6)
	var wg sync.WaitGroup
	for i := range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body := `{"token":"` + token + `","username":"admin` + strconv.Itoa(i) + `","password":"correct horse"}`
			codes <- ts.do("POST", api+"setup", body).Code
		}()
	}
	wg.Wait()
	close(codes)
	won := 0
	for c := range codes {
		switch c {
		case http.StatusOK:
			won++
		case http.StatusConflict:
		default:
			t.Errorf("unexpected status %d", c)
		}
	}
	if won != 1 {
		t.Fatalf("%d setups succeeded", won)
	}
}
