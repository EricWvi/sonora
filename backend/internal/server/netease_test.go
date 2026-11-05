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
	expired  bool
	likes    []string
	liked    []map[string]int64 // trackIds of the liked playlist
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
		case "/user/playlist":
			if f.expired {
				w.Write([]byte(`{"code":301,"msg":"需要登录"}`))
				return
			}
			w.Write([]byte(`{"code":200,"playlist":[
				{"id":100,"name":"liked","specialType":5,"subscribed":false,"trackCount":2},
				{"id":101,"name":"mine","specialType":0,"subscribed":false,"coverImgUrl":"http://p1/c.jpg","trackCount":1,"description":null},
				{"id":102,"name":"theirs","specialType":0,"subscribed":true,"trackCount":1,"description":"d"}]}`))
		case "/playlist/detail":
			if r.PostForm.Get("id") == "404" {
				w.Write([]byte(`{"code":404}`))
				return
			}
			body, _ := json.Marshal(map[string]any{"code": 200, "playlist": map[string]any{
				"id": 100, "name": "liked", "trackCount": len(f.liked), "trackIds": f.liked,
			}})
			w.Write(body)
		case "/song/detail":
			var songs, privileges []string
			for _, id := range strings.Split(r.PostForm.Get("ids"), ",") {
				if id == "999" {
					continue // gone from NetEase
				}
				songs = append(songs, `{"id":`+id+`,"name":"song`+id+`","ar":[{"id":5,"name":"A"}],"dt":1000,
					"publishTime":1500000000000,"al":{"id":6,"name":"Al","picUrl":"http://p1/`+id+`.jpg"}}`)
				st := "0"
				if id == "888" {
					st = "-200"
				}
				privileges = append(privileges, `{"id":`+id+`,"st":`+st+`}`)
			}
			w.Write([]byte(`{"code":200,"songs":[` + strings.Join(songs, ",") + `],"privileges":[` + strings.Join(privileges, ",") + `]}`))
		case "/song/url/v1":
			switch r.PostForm.Get("id") {
			case "1":
				w.Write([]byte(`{"code":200,"data":[{"url":"http://m804.music.126.net/a.mp3","freeTrialInfo":null}]}`))
			case "2":
				w.Write([]byte(`{"code":200,"data":[{"url":"https://m7.music.126.net/t.mp3","freeTrialInfo":{"start":0}}]}`))
			default:
				w.Write([]byte(`{"code":200,"data":[{"url":null}]}`))
			}
		case "/lyric/new":
			w.Write([]byte(`{"code":200,"lrc":{"lyric":"[00:01.00]line"}}`))
		case "/likelist":
			w.Write([]byte(`{"code":200,"ids":[1,2]}`))
		case "/like":
			f.likes = append(f.likes, r.PostForm.Get("id")+"="+r.PostForm.Get("like"))
			w.Write([]byte(`{"code":200}`))
		case "/album/sublist":
			w.Write([]byte(`{"code":200,"hasMore":false,"data":[{"id":7,"name":"Al","picUrl":"http://p1/a.jpg","size":3,"artists":[{"id":5,"name":"A"}]}]}`))
		case "/artist/sublist":
			w.Write([]byte(`{"code":200,"hasMore":false,"data":[{"id":5,"name":"A","picUrl":"http://p1/r.jpg","albumSize":9}]}`))
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

// likedAt is a time in the fake's "at" form.
func likedAt(t *testing.T, s string) int64 {
	t.Helper()
	at, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return at.UnixMilli()
}

func signInNetease(t *testing.T, ts *testServer, f *fakeNetease) {
	t.Helper()
	f.accounts["good"] = 7
	if rec := ts.do("POST", playerAPI+"netease/login/cookie", `{"cookie":"MUSIC_U=good"}`); rec.Code != http.StatusOK {
		t.Fatalf("login: %d %s", rec.Code, rec.Body)
	}
}

type mixedRow struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Source   string `json:"source"`
	Playable bool   `json:"playable"`
}

type mixedResult struct {
	Items        []mixedRow `json:"items"`
	Total        int        `json:"total"`
	OnlineFailed bool       `json:"onlineFailed"`
}

// mixedFixture has a listener signed in to Sonora, and one local track.
func mixedFixture(t *testing.T) (*testServer, *fakeNetease, store.TrackDetail) {
	t.Helper()
	f := newFakeNetease(t)
	ts := newNeteaseTest(t, f)
	ts.setup()
	tr := decodeBody[store.IngestResult](t, ts.upload("a.flac", sample(t, "tagged.flac"), "")).Track
	ts.cookie = ""
	ts.signIn("alice", "Alice")
	return ts, f, tr
}

func mixedIDs(rows []mixedRow) string {
	var out []string
	for _, r := range rows {
		out = append(out, r.Source+":"+itoa(r.ID))
	}
	return strings.Join(out, " ")
}

func TestMixedLikes(t *testing.T) {
	ts, f, tr := mixedFixture(t)
	if rec := ts.do("PUT", playerAPI+"likes/"+itoa(tr.ID), ""); rec.Code != http.StatusNoContent {
		t.Fatalf("like: %d", rec.Code)
	}
	got := decodeBody[mixedResult](t, ts.do("GET", playerAPI+"likes/mixed", ""))
	if mixedIDs(got.Items) != ":"+itoa(tr.ID) || got.OnlineFailed {
		t.Fatalf("signed out of NetEase: %+v", got)
	}

	signInNetease(t, ts, f)
	future := likedAt(t, "2999-01-01T00:00:00Z")
	past := likedAt(t, "2001-01-01T00:00:00Z")
	f.liked = []map[string]int64{{"id": 1, "at": future}, {"id": 888, "at": future - 1000}, {"id": 999, "at": future - 2000}, {"id": 2, "at": past}}
	got = decodeBody[mixedResult](t, ts.do("GET", playerAPI+"likes/mixed", ""))
	if want := "online:1 online:888 :" + itoa(tr.ID) + " online:2"; mixedIDs(got.Items) != want {
		t.Fatalf("order: got %s, want %s", mixedIDs(got.Items), want)
	}
	if !got.Items[0].Playable || got.Items[1].Playable {
		t.Fatalf("playable: %+v", got.Items)
	}

	// The window applies to the merged list.
	page := decodeBody[mixedResult](t, ts.do("GET", playerAPI+"likes/mixed?offset=1&limit=1", ""))
	if mixedIDs(page.Items) != "online:888" {
		t.Fatalf("page: %+v", page)
	}

	// The same song liked on both sides shows once, as the local track, at
	// the later of the two times.
	if err := ts.store.SetNeteaseID(context.Background(), tr.ID, 1); err != nil {
		t.Fatal(err)
	}
	got = decodeBody[mixedResult](t, ts.do("GET", playerAPI+"likes/mixed", ""))
	if want := ":" + itoa(tr.ID) + " online:888 online:2"; mixedIDs(got.Items) != want {
		t.Fatalf("dedup: got %s, want %s", mixedIDs(got.Items), want)
	}

	// NetEase failing leaves the local half, and says so.
	f.down = true
	got = decodeBody[mixedResult](t, ts.do("GET", playerAPI+"likes/mixed", ""))
	if mixedIDs(got.Items) != ":"+itoa(tr.ID) || !got.OnlineFailed {
		t.Fatalf("service down: %+v", got)
	}
	f.down, f.expired = false, true
	got = decodeBody[mixedResult](t, ts.do("GET", playerAPI+"likes/mixed", ""))
	if mixedIDs(got.Items) != ":"+itoa(tr.ID) || !got.OnlineFailed {
		t.Fatalf("expired: %+v", got)
	}
}

func onlineSong(id string) string {
	return `{"song":{"id":` + id + `,"title":"song` + id + `","artists":[{"id":5,"name":"A"}],"durationMs":1000,
		"source":"anything","cover":{"name":"","thumbhash":"","url":"https://p1/c.jpg"}}}`
}

func historyRows(t *testing.T, ts *testServer) []mixedRow {
	t.Helper()
	page := decodeBody[struct {
		Items []struct {
			Track    mixedRow `json:"track"`
			PlayedAt string   `json:"playedAt"`
		} `json:"items"`
	}](t, ts.do("GET", playerAPI+"history/mixed", ""))
	rows := make([]mixedRow, len(page.Items))
	for i, e := range page.Items {
		if e.PlayedAt == "" {
			t.Fatalf("no playedAt: %+v", e)
		}
		rows[i] = e.Track
	}
	return rows
}

func TestMixedHistory(t *testing.T) {
	ts, f, tr := mixedFixture(t)
	if rec := ts.do("POST", playerAPI+"history", `{"trackId":`+itoa(tr.ID)+`}`); rec.Code != http.StatusNoContent {
		t.Fatalf("play: %d", rec.Code)
	}
	for _, body := range []string{`{"song":{"id":0,"title":"x"}}`, `{"song":{"id":1,"title":""}}`,
		`{"song":{"id":1,"title":"` + strings.Repeat("x", maxSnapshotBytes) + `"}}`} {
		if rec := ts.do("POST", playerAPI+"history/online", body); rec.Code != http.StatusBadRequest {
			t.Fatalf("%.40s: %d", body, rec.Code)
		}
	}
	if rec := ts.do("POST", playerAPI+"history/online", onlineSong("1")); rec.Code != http.StatusNoContent {
		t.Fatalf("online play: %d %s", rec.Code, rec.Body)
	}
	// Without a NetEase login the list is local only.
	if got := mixedIDs(historyRows(t, ts)); got != ":"+itoa(tr.ID) {
		t.Fatalf("signed out: %s", got)
	}
	signInNetease(t, ts, f)
	rows := historyRows(t, ts)
	if len(rows) != 2 || rows[0].Source == rows[1].Source {
		t.Fatalf("merged: %+v", rows)
	}
	for _, row := range rows {
		if row.Source == "online" && row.Title != "song1" {
			t.Fatalf("snapshot: %+v", row)
		}
	}
	if rec := ts.do("DELETE", playerAPI+"history/online/1", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("remove: %d", rec.Code)
	}
	if got := mixedIDs(historyRows(t, ts)); got != ":"+itoa(tr.ID) {
		t.Fatalf("after remove: %s", got)
	}
	if rec := ts.do("POST", playerAPI+"history/online", onlineSong("2")); rec.Code != http.StatusNoContent {
		t.Fatalf("online play: %d", rec.Code)
	}
	if rec := ts.do("DELETE", playerAPI+"history", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("clear: %d", rec.Code)
	}
	if rows := historyRows(t, ts); len(rows) != 0 {
		t.Fatalf("after clear: %+v", rows)
	}
}

func TestNeteaseCollections(t *testing.T) {
	ts, f, _ := mixedFixture(t)
	for _, path := range []string{"playlists", "playlists/1", "albums", "artists"} {
		rec := ts.do("GET", playerAPI+"netease/"+path, "")
		if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "netease_signed_out") {
			t.Fatalf("%s signed out: %d %s", path, rec.Code, rec.Body)
		}
	}
	signInNetease(t, ts, f)

	rec := ts.do("GET", playerAPI+"netease/playlists", "")
	lists := decodeBody[struct {
		Created, Collected []netease.Playlist
	}](t, rec)
	if len(lists.Created) != 1 || lists.Created[0].Name != "mine" || len(lists.Collected) != 1 || lists.Collected[0].Name != "theirs" {
		t.Fatalf("playlists: %s", rec.Body)
	}
	if c := lists.Created[0].Cover; c == nil || c.URL != "https://p1/c.jpg" {
		t.Fatalf("cover: %+v", c)
	}

	f.liked = []map[string]int64{{"id": 3, "at": 1}}
	rec = ts.do("GET", playerAPI+"netease/playlists/101", "")
	detail := decodeBody[struct {
		Tracks []netease.Song `json:"tracks"`
	}](t, rec)
	if len(detail.Tracks) != 1 || detail.Tracks[0].Year != 2017 || detail.Tracks[0].Cover.URL != "https://p1/3.jpg" {
		t.Fatalf("detail: %s", rec.Body)
	}
	if rec := ts.do("GET", playerAPI+"netease/playlists/404", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("missing: %d %s", rec.Code, rec.Body)
	}

	if rec := ts.do("GET", playerAPI+"netease/albums", ""); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"trackCount":3`) {
		t.Fatalf("albums: %d %s", rec.Code, rec.Body)
	}
	if rec := ts.do("GET", playerAPI+"netease/artists", ""); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"albumCount":9`) {
		t.Fatalf("artists: %d %s", rec.Code, rec.Body)
	}

	f.expired = true
	rec = ts.do("GET", playerAPI+"netease/playlists", "")
	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "netease_expired") {
		t.Fatalf("expired: %d %s", rec.Code, rec.Body)
	}
}

func TestNeteasePlayback(t *testing.T) {
	ts, f, _ := mixedFixture(t)
	signInNetease(t, ts, f)

	rec := ts.do("GET", playerAPI+"netease/songs/1/url", "")
	if got := decodeBody[map[string]string](t, rec)["url"]; got != "https://m801.music.126.net/a.mp3" {
		t.Fatalf("url: %d %s", rec.Code, rec.Body)
	}
	for _, id := range []string{"2", "3"} { // a trial, and no address at all
		rec := ts.do("GET", playerAPI+"netease/songs/"+id+"/url", "")
		if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "netease_no_url") {
			t.Fatalf("song %s: %d %s", id, rec.Code, rec.Body)
		}
	}
	rec = ts.do("GET", playerAPI+"netease/songs/1/lyrics", "")
	if got := decodeBody[map[string]string](t, rec); got["format"] != "lrc" || got["lyrics"] == "" {
		t.Fatalf("lyrics: %s", rec.Body)
	}
	if rec := ts.do("GET", playerAPI+"netease/likes/ids", ""); !strings.Contains(rec.Body.String(), "[1,2]") {
		t.Fatalf("ids: %s", rec.Body)
	}
	if rec := ts.do("PUT", playerAPI+"netease/likes/5", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("like: %d %s", rec.Code, rec.Body)
	}
	if rec := ts.do("DELETE", playerAPI+"netease/likes/5", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("unlike: %d %s", rec.Code, rec.Body)
	}
	if got := strings.Join(f.likes, " "); got != "5=true 5=false" {
		t.Fatalf("likes sent: %s", got)
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
