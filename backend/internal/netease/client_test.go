package netease

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalizeCookie(t *testing.T) {
	for _, tt := range []struct{ name, in, want string }{
		{"plain", "MUSIC_U=abc;", "MUSIC_U=abc"},
		{"browser copy", " NMTID=n; MUSIC_U=abc; __csrf=c ", "NMTID=n; MUSIC_U=abc; __csrf=c"},
		{
			"set-cookie attributes",
			"MUSIC_U=abc; Max-Age=15552000; Expires=Tue, 05 May 2027 01:02:03 GMT; Path=/; HTTPOnly;__csrf=c; Path=/; Secure",
			"MUSIC_U=abc; __csrf=c",
		},
		{"url encoded", "MUSIC_U%3Dabc%3B%20__csrf%3Dc", "MUSIC_U=abc; __csrf=c"},
		{"repeated", "MUSIC_U=old; MUSIC_U=new", "MUSIC_U=new"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeCookie(tt.in)
			if err != nil || got != tt.want {
				t.Fatalf("got %q, %v; want %q", got, err, tt.want)
			}
		})
	}
	for _, in := range []string{"", "__csrf=c", "MUSIC_U=;", "abc"} {
		if _, err := NormalizeCookie(in); !errors.Is(err, ErrInvalidCookie) {
			t.Errorf("%q: %v", in, err)
		}
	}
}

func TestCookieStaysInTheBody(t *testing.T) {
	var query, form, bypass string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		bypass = r.Header.Get("x-apicache-bypass")
		r.ParseForm()
		form = r.PostForm.Get("cookie")
		w.Write([]byte(`{"data":{"profile":{"userId":7,"nickname":"N","vipType":11}}}`))
	}))
	defer srv.Close()
	p, err := New(srv.URL+"/").Status(context.Background(), "MUSIC_U=secret")
	if err != nil || p == nil || p.UserID != 7 || p.VIPType != 11 {
		t.Fatalf("status: %+v, %v", p, err)
	}
	if strings.Contains(query, "MUSIC_U") || form != "MUSIC_U=secret; os=pc" {
		t.Fatalf("query %q, form %q", query, form)
	}
	if bypass != "true" {
		t.Fatalf("cache bypass header = %q", bypass)
	}
}

// The service keys its cache on the URL, not the body holding the cookie, so
// two calls must never share one.
func TestEveryCallHasItsOwnURL(t *testing.T) {
	seen := map[string]bool{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen[r.URL.String()] = true
		w.Write([]byte(`{"data":{}}`))
	}))
	defer srv.Close()
	c := New(srv.URL)
	for range 20 {
		if _, err := c.Status(context.Background(), "MUSIC_U=a"); err != nil {
			t.Fatal(err)
		}
	}
	if len(seen) != 20 {
		t.Fatalf("%d distinct URLs for 20 calls", len(seen))
	}
}

func TestStatusWithoutProfileMeansInvalid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// What the service answers for a dead cookie: an anonymous account.
		w.Write([]byte(`{"data":{"code":200,"account":{"id":1},"profile":null}}`))
	}))
	defer srv.Close()
	if p, err := New(srv.URL).Status(context.Background(), "MUSIC_U=dead"); p != nil || err != nil {
		t.Fatalf("got %+v, %v", p, err)
	}
}

func TestQRCheckDropsCookieUntilSucceeded(t *testing.T) {
	code := QRWaiting
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if code == QRWaiting {
			w.Write([]byte(`{"code":801,"cookie":"NMTID=anonymous"}`))
			return
		}
		w.Write([]byte(`{"code":803,"cookie":"MUSIC_U=abc"}`))
	}))
	defer srv.Close()
	c := New(srv.URL)
	if s, err := c.QRCheck(context.Background(), "k"); err != nil || s.Cookie != "" {
		t.Fatalf("waiting: %+v, %v", s, err)
	}
	code = QRSucceeded
	if s, err := c.QRCheck(context.Background(), "k"); err != nil || s.Cookie != "MUSIC_U=abc" {
		t.Fatalf("succeeded: %+v, %v", s, err)
	}
}

func TestUnavailableErrorsHideTheBody(t *testing.T) {
	for name, handler := range map[string]http.HandlerFunc{
		"server error": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`MUSIC_U=leak`))
		},
		"not json": func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`MUSIC_U=leak`)) },
	} {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(handler)
			defer srv.Close()
			_, err := New(srv.URL).Status(context.Background(), "MUSIC_U=secret")
			if !errors.Is(err, ErrUnavailable) || strings.Contains(err.Error(), "MUSIC_U") {
				t.Fatalf("err = %v", err)
			}
		})
	}
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()
	if _, err := New(srv.URL).QRKey(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("closed server: %v", err)
	}
}

// The disc numbers are the forms seen on real albums (2026-11-06).
func TestAlbumForImport(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"code":200,"album":{"name":"Al","picUrl":"http://p/a.jpg","description":"d",
			"publishTime":1500000000000,"artists":[{"id":9,"name":"群星"}]},"songs":[
			{"id":1,"name":"a","ar":[{"id":5,"name":"A"}],"no":1,"cd":"01","al":{"id":3,"name":"Al","picUrl":"http://p/1.jpg"},"privilege":{"st":0}},
			{"id":2,"name":"b","ar":[{"id":6,"name":"B"}],"no":"2","cd":"02 [Target Deluxe Edition]","al":{"id":3,"name":"Al"},"privilege":{"st":-200}},
			{"id":3,"name":"c","ar":[],"no":3,"cd":null,"al":{"id":3,"name":"Al"}}]}`))
	}))
	defer srv.Close()
	a, err := New(srv.URL).AlbumForImport(context.Background(), "MUSIC_U=x", 3)
	if err != nil {
		t.Fatal(err)
	}
	if a.Title != "Al" || a.Year != 2017 || a.Description != "d" || len(a.Artists) != 1 || a.Artists[0].Name != "群星" || len(a.Tracks) != 3 {
		t.Fatalf("album: %+v", a)
	}
	for i, want := range []struct {
		no, cd   int
		playable bool
		cover    string
	}{{1, 1, true, "https://p/1.jpg"}, {2, 2, false, "https://p/a.jpg"}, {3, 0, true, "https://p/a.jpg"}} {
		tr := a.Tracks[i]
		if tr.TrackNo != want.no || tr.DiscNo != want.cd || tr.Song.Playable != want.playable ||
			tr.Song.Cover == nil || tr.Song.Cover.URL != want.cover || tr.Song.Year != 2017 || tr.Song.TrackNo != 0 {
			t.Errorf("track %d: %+v", i, tr)
		}
	}
}

// The service passes NetEase's code on as the HTTP status, and -460, which is
// no HTTP status, as 400.
func TestThrottledCodes(t *testing.T) {
	for _, tt := range []struct {
		status int
		body   string
	}{
		{405, `{"code":405,"message":"操作频繁，请稍候再试"}`},
		{400, `{"code":-460,"message":"Cheating"}`},
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tt.status)
			w.Write([]byte(tt.body))
		}))
		_, err := New(srv.URL).SongFile(context.Background(), "MUSIC_U=x", 1)
		srv.Close()
		if !errors.Is(err, ErrThrottled) {
			t.Errorf("%s: %v", tt.body, err)
		}
	}
}

func TestArtistProfile(t *testing.T) {
	for _, tt := range []struct{ body, image string }{
		{`{"code":200,"data":{"artist":{"cover":"http://p/c.jpg","avatar":"http://p/a.jpg","briefDesc":"简介"}}}`, "https://p/a.jpg"},
		{`{"code":200,"data":{"artist":{"cover":"http://p/c.jpg","avatar":"","briefDesc":"简介"}}}`, "https://p/c.jpg"},
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(tt.body)) }))
		p, err := New(srv.URL).ArtistProfile(context.Background(), "MUSIC_U=x", 5)
		srv.Close()
		if err != nil || p.ImageURL != tt.image || p.Description != "简介" {
			t.Errorf("%+v %v", p, err)
		}
	}
}
