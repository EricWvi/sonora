package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"sonora/internal/store"
)

func TestDevicesListAndRevoke(t *testing.T) {
	ts := newTest(t)
	ts.signIn("alice", "Alice")
	ctx := context.Background()
	u, err := ts.store.UpsertUser(ctx, store.Identity{Issuer: "https://auth.test", Subject: "alice", Name: "Alice"})
	if err != nil {
		t.Fatal(err)
	}
	phone, phoneToken, err := ts.store.CreateDevice(ctx, u.ID, "Phone", store.PlatformAndroid)
	if err != nil {
		t.Fatal(err)
	}
	tablet, tabletToken, err := ts.store.CreateDevice(ctx, u.ID, "Tablet", store.PlatformAndroid)
	if err != nil {
		t.Fatal(err)
	}
	bearer := func(method, path, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		// A device needs no origin; a foreign one is ignored.
		req.Header.Set("Origin", "https://evil.test")
		rec := httptest.NewRecorder()
		ts.ServeHTTP(rec, req)
		return rec
	}

	// The device token opens the player API without the session.
	listener := ts.listener
	ts.listener = ""
	if rec := bearer("GET", playerAPI+"me", phoneToken); rec.Code != http.StatusOK || decodeBody[map[string]any](t, rec)["name"] != "Alice" {
		t.Fatalf("me with token: %d %s", rec.Code, rec.Body)
	}
	rec := bearer("GET", playerAPI+"devices", phoneToken)
	if list := decodeBody[[]map[string]any](t, rec); rec.Code != http.StatusOK || len(list) != 2 || list[0]["createdAt"] == "" {
		t.Fatalf("devices: %d %s", rec.Code, rec.Body)
	}
	// A device may only revoke itself.
	if rec := bearer("DELETE", playerAPI+"devices/"+strconv.FormatInt(tablet.ID, 10), phoneToken); rec.Code != http.StatusForbidden {
		t.Fatalf("revoke other from device: %d %s", rec.Code, rec.Body)
	}
	if rec := bearer("DELETE", playerAPI+"devices/"+strconv.FormatInt(phone.ID, 10), phoneToken); rec.Code != http.StatusNoContent {
		t.Fatalf("revoke self: %d %s", rec.Code, rec.Body)
	}
	if rec := bearer("GET", playerAPI+"me", phoneToken); rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked token: %d", rec.Code)
	}

	// The web session revokes any of the user's devices, but not others'.
	ts.listener = listener
	if rec := ts.do("DELETE", playerAPI+"devices/"+strconv.FormatInt(tablet.ID, 10), ""); rec.Code != http.StatusNoContent {
		t.Fatalf("revoke from web: %d %s", rec.Code, rec.Body)
	}
	if rec := bearer("GET", playerAPI+"me", tabletToken); rec.Code != http.StatusUnauthorized {
		t.Fatalf("tablet after revoke: %d", rec.Code)
	}
	ts.signIn("bob", "Bob")
	bob, _ := ts.store.UpsertUser(ctx, store.Identity{Issuer: "https://auth.test", Subject: "bob", Name: "Bob"})
	other, _, err := ts.store.CreateDevice(ctx, bob.ID, "Bob's", store.PlatformAndroid)
	if err != nil {
		t.Fatal(err)
	}
	ts.listener = listener
	if rec := ts.do("DELETE", playerAPI+"devices/"+strconv.FormatInt(other.ID, 10), ""); rec.Code != http.StatusNotFound {
		t.Fatalf("revoke someone else's: %d %s", rec.Code, rec.Body)
	}
}
