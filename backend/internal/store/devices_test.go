package store

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestDeviceTokenResolvesUntilRevoked(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	u, err := s.UpsertUser(ctx, Identity{Issuer: "https://auth.test", Subject: "alice", Name: "Alice", Groups: []string{"sonora-import"}})
	if err != nil {
		t.Fatal(err)
	}
	other := newUser(t, s, "bob")

	d, token, err := s.CreateDevice(ctx, u.ID, "  Pixel  ", PlatformAndroid)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(token, DeviceTokenPrefix) || d.Name != "Pixel" || d.CreatedAt == "" {
		t.Fatalf("device = %+v, token = %q", d, token)
	}
	got, user, err := s.DeviceUser(ctx, token)
	if err != nil || got.ID != d.ID || user.ID != u.ID || !slices.Equal(user.Groups, []string{"sonora-import"}) {
		t.Fatalf("DeviceUser = %+v, %+v, %v", got, user, err)
	}
	if _, _, err := s.DeviceUser(ctx, token+"x"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("wrong token: %v", err)
	}

	list, err := s.ListDevices(ctx, u.ID)
	if err != nil || len(list) != 1 || list[0].ID != d.ID {
		t.Fatalf("ListDevices = %+v, %v", list, err)
	}
	// Someone else cannot revoke it.
	if err := s.DeleteDevice(ctx, other, d.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("delete by other: %v", err)
	}
	if err := s.DeleteDevice(ctx, u.ID, d.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.DeviceUser(ctx, token); !errors.Is(err, ErrNotFound) {
		t.Fatalf("revoked token: %v", err)
	}
}

func TestCreateDeviceValidates(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	id := newUser(t, s, "alice")
	for name, tc := range map[string]struct{ name, platform string }{
		"empty name":   {" ", PlatformAndroid},
		"long name":    {strings.Repeat("长", MaxDeviceName+1), PlatformAndroid},
		"mac platform": {"MacBook", "mac"},
	} {
		if _, _, err := s.CreateDevice(ctx, id, tc.name, tc.platform); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
