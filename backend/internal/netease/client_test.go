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
