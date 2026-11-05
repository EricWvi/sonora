// Package netease talks to the NetEase Cloud Music API service, which runs as
// a separate container. The service has no authentication of its own, and its
// responses echo cookies back, so nothing here is a generic proxy: every call
// is a named method that returns only the fields the player needs. Cookies
// travel in the POST body, never the URL, so that they stay out of logs.
//
// The service caches responses for two minutes under a key made of the URL and
// the Cookie header, without the POST body. Left alone, one listener would be
// served another's answers, so call opts out of that cache on every request.
package netease

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	requestTimeout = 15 * time.Second
	maxResponse    = 32 << 20
)

// ErrUnavailable means the API service could not be reached or failed; the
// listener's session says nothing about it.
var ErrUnavailable = errors.New("netease: api service unavailable")

// ErrNotFound means NetEase has no such item.
var ErrNotFound = errors.New("netease: not found")

// ErrExpired means NetEase refused the session.
var ErrExpired = errors.New("netease: session expired")

// ErrInvalidCookie means a cookie cannot sign anyone in.
var ErrInvalidCookie = errors.New("netease: cookie has no MUSIC_U")

// QR login states, as /login/qr/check reports them.
const (
	QRExpired   = 800
	QRWaiting   = 801
	QRConfirm   = 802
	QRSucceeded = 803
)

// needLogin is the code NetEase answers with when a call needs a session.
const needLogin = 301

// Client calls the API service.
type Client struct {
	base string
	http *http.Client
	// Downloads fetches files from NetEase's CDN. Unlike API calls it has no
	// overall timeout, since a song can take a while; the request's context
	// bounds it.
	Downloads *http.Client
}

// RejectedError is NetEase refusing a write, such as a playlist name it does
// not accept. Message is NetEase's own explanation and may be empty.
type RejectedError struct {
	Message string
}

func (e *RejectedError) Error() string {
	return "netease rejected the request: " + e.Message
}

// New returns a client for the service at baseURL.
func New(baseURL string) *Client {
	return &Client{
		base:      strings.TrimRight(baseURL, "/"),
		http:      &http.Client{Timeout: requestTimeout},
		Downloads: &http.Client{},
	}
}

// Profile is the signed-in NetEase account.
type Profile struct {
	UserID    int64  `json:"userId"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
	VIPType   int    `json:"vipType"`
}

// QRState is the progress of a QR login. Cookie is set only on QRSucceeded and
// must not reach the browser.
type QRState struct {
	Code      int
	Nickname  string
	AvatarURL string
	Cookie    string
}

// call posts params to path and decodes the JSON answer into out. The service
// answers some business failures with a 4xx and a JSON body, so only a 5xx or
// an unreadable body counts as unavailable. Errors never carry the response
// body, which may hold a cookie.
func (c *Client) call(ctx context.Context, path, cookie string, params url.Values, out any) error {
	form := url.Values{}
	for k, v := range params {
		form[k] = v
	}
	if cookie != "" {
		// SPlayer adds os=pc to every signed-in request as well.
		form.Set("cookie", cookie+"; os=pc")
	}
	// The timestamp gives each request its own URL, as SPlayer does, should a
	// later version of the service stop honoring the bypass header.
	target := c.base + path + "?timestamp=" + strconv.FormatInt(time.Now().UnixNano(), 10)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("%w: %s", ErrUnavailable, path)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("x-apicache-bypass", "true")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %s: request failed", ErrUnavailable, path)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("%w: %s: status %d", ErrUnavailable, path, resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse))
	if err != nil {
		return fmt.Errorf("%w: %s: unreadable response", ErrUnavailable, path)
	}
	var head struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		return fmt.Errorf("%w: %s: unreadable response", ErrUnavailable, path)
	}
	if head.Code == needLogin {
		return fmt.Errorf("%w: %s", ErrExpired, path)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("%w: %s: unreadable response", ErrUnavailable, path)
	}
	return nil
}

// Status returns the account the cookie signs in to, or nil when the cookie
// is no longer valid. An invalid cookie still gets an anonymous account from
// the service, so only the profile tells them apart.
func (c *Client) Status(ctx context.Context, cookie string) (*Profile, error) {
	var out struct {
		Data struct {
			Profile *Profile `json:"profile"`
		} `json:"data"`
	}
	if err := c.call(ctx, "/login/status", cookie, nil, &out); err != nil {
		return nil, err
	}
	if p := out.Data.Profile; p != nil && p.UserID != 0 {
		return p, nil
	}
	return nil, nil
}

// QRKey starts a QR login.
func (c *Client) QRKey(ctx context.Context) (string, error) {
	var out struct {
		Data struct {
			Unikey string `json:"unikey"`
		} `json:"data"`
	}
	if err := c.call(ctx, "/login/qr/key", "", nil, &out); err != nil {
		return "", err
	}
	if out.Data.Unikey == "" {
		return "", fmt.Errorf("%w: /login/qr/key: no key", ErrUnavailable)
	}
	return out.Data.Unikey, nil
}

// QRURL is what the login QR code encodes for key.
func QRURL(key string) string {
	return "https://music.163.com/login?codekey=" + url.QueryEscape(key)
}

// QRCheck reports the progress of the login started with key.
func (c *Client) QRCheck(ctx context.Context, key string) (QRState, error) {
	var out struct {
		Code      int    `json:"code"`
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatarUrl"`
		Cookie    string `json:"cookie"`
	}
	if err := c.call(ctx, "/login/qr/check", "", url.Values{"key": {key}}, &out); err != nil {
		return QRState{}, err
	}
	state := QRState{Code: out.Code, Nickname: out.Nickname, AvatarURL: out.AvatarURL}
	if out.Code == QRSucceeded {
		state.Cookie = out.Cookie
	}
	return state, nil
}

// Refresh renews the session and returns the cookie to keep, which is the
// given one when the service hands out no new cookie.
func (c *Client) Refresh(ctx context.Context, cookie string) (string, error) {
	var out struct {
		Cookie string `json:"cookie"`
	}
	if err := c.call(ctx, "/login/refresh", cookie, nil, &out); err != nil {
		return "", err
	}
	if renewed, err := NormalizeCookie(out.Cookie); err == nil {
		return renewed, nil
	}
	return cookie, nil
}

// cookieAttributes are the Set-Cookie attributes the service leaves in the
// cookie strings it returns.
var cookieAttributes = map[string]bool{
	"max-age": true, "expires": true, "path": true, "domain": true,
	"httponly": true, "secure": true, "samesite": true,
}

// NormalizeCookie reduces what a listener pasted, or what a login returned, to
// a "name=value; name=value" cookie header holding MUSIC_U. Set-Cookie
// attributes are dropped, as is a URL-encoded wrapping; of a repeated cookie
// the last one wins.
func NormalizeCookie(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if !strings.Contains(raw, "MUSIC_U=") && strings.Contains(raw, "%") {
		if decoded, err := url.QueryUnescape(raw); err == nil {
			raw = decoded
		}
	}
	var names []string
	values := map[string]string{}
	for _, part := range strings.Split(raw, ";") {
		name, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if !ok || name == "" || value == "" || cookieAttributes[strings.ToLower(name)] {
			continue
		}
		if _, seen := values[name]; !seen {
			names = append(names, name)
		}
		values[name] = value
	}
	if values["MUSIC_U"] == "" {
		return "", ErrInvalidCookie
	}
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, name+"="+values[name])
	}
	return strings.Join(parts, "; "), nil
}
