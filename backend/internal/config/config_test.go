package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadRequiresPublicURL(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if err == nil || !strings.Contains(err.Error(), "public_url") {
		t.Fatalf("err = %v, want public_url error", err)
	}
}

func TestLoad(t *testing.T) {
	cfg, err := Load(write(t, "listen: \":9000\"\npublic_url: https://music.test\nmedia_dir: /srv/media\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := Config{Listen: ":9000", DataDir: "./data", MediaDir: "/srv/media", PublicURL: "https://music.test"}
	if cfg != want {
		t.Fatalf("got %+v, want %+v", cfg, want)
	}
}

func TestLoadRejectsInvalid(t *testing.T) {
	for name, content := range map[string]string{
		"unknown key":     "public_url: https://music.test\ndev_user: eric\n",
		"path in url":     "public_url: https://music.test/sonora\n",
		"same dirs":       "public_url: https://music.test\ndata_dir: ./x\nmedia_dir: x\n",
		"empty media dir": "public_url: https://music.test\nmedia_dir: \"\"\n",
	} {
		if _, err := Load(write(t, content)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
