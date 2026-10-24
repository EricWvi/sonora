package store

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestSettingsArePerUserJSONObjects(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	alice, bob := newUser(t, s, "alice"), newUser(t, s, "bob")
	if got, err := s.Settings(ctx, alice); err != nil || string(got) != "{}" {
		t.Fatalf("unsaved = %s, %v", got, err)
	}
	if err := s.SaveSettings(ctx, alice, []byte(`{ "themeMode": "dark",
		"lyricFontSize": 46 }`)); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveSettings(ctx, alice, []byte(`{"themeMode":"light"}`)); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Settings(ctx, alice); string(got) != `{"themeMode":"light"}` {
		t.Fatalf("saved = %s; a save replaces the whole object", got)
	}
	if got, _ := s.Settings(ctx, bob); string(got) != "{}" {
		t.Fatalf("bob sees %s", got)
	}
	for _, bad := range []string{``, `[1]`, `"x"`, `{"a":`, `{} {}`, `{"a":"` + strings.Repeat("x", MaxSettingsBytes) + `"}`} {
		if err := s.SaveSettings(ctx, alice, []byte(bad)); !errors.Is(err, ErrInvalid) {
			t.Errorf("save %.20q: %v, want ErrInvalid", bad, err)
		}
	}
}
