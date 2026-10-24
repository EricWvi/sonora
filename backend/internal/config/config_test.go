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

const oidcYAML = "oidc:\n  issuer: https://auth.test\n  client_id: sonora\n  client_secret: s3cret\n"

func TestLoad(t *testing.T) {
	cfg, err := Load(write(t, "listen: \":9000\"\npublic_url: https://music.test\nmedia_dir: /srv/media\n"+oidcYAML))
	if err != nil {
		t.Fatal(err)
	}
	want := Config{
		Listen: ":9000", DataDir: "./data", MediaDir: "/srv/media", PublicURL: "https://music.test",
		OIDC: OIDC{Issuer: "https://auth.test", ClientID: "sonora", ClientSecret: "s3cret"},
	}
	if cfg != want {
		t.Fatalf("got %+v, want %+v", cfg, want)
	}
}

func TestClientSecretFromEnvironment(t *testing.T) {
	t.Setenv(ClientSecretEnv, "from-env")
	cfg, err := Load(write(t, "public_url: https://music.test\noidc:\n  issuer: https://auth.test\n  client_id: sonora\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OIDC.ClientSecret != "from-env" {
		t.Fatalf("client secret = %q", cfg.OIDC.ClientSecret)
	}
}

func TestLoadRejectsInvalid(t *testing.T) {
	for name, content := range map[string]string{
		"unknown key":     "public_url: https://music.test\ndev_user: eric\n" + oidcYAML,
		"path in url":     "public_url: https://music.test/sonora\n" + oidcYAML,
		"same dirs":       "public_url: https://music.test\ndata_dir: ./x\nmedia_dir: x\n" + oidcYAML,
		"empty media dir": "public_url: https://music.test\nmedia_dir: \"\"\n" + oidcYAML,
		"no oidc":         "public_url: https://music.test\n",
		"no secret":       "public_url: https://music.test\noidc:\n  issuer: https://auth.test\n  client_id: sonora\n",
	} {
		if _, err := Load(write(t, content)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
