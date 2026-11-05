// Package config loads the YAML configuration file.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/url"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"sonora/internal/devmode"
)

// ClientSecretEnv overrides oidc.client_secret so the secret can stay out
// of the config file.
const ClientSecretEnv = "SONORA_OIDC_CLIENT_SECRET"

// Config is the application configuration.
type Config struct {
	// Listen is the HTTP listen address, e.g. ":14560".
	Listen string `yaml:"listen"`
	// DataDir holds the SQLite database and cover images.
	DataDir string `yaml:"data_dir"`
	// MediaDir is the volume the service stores uploaded audio in. Only the
	// service writes here; files placed by hand are ignored.
	MediaDir string `yaml:"media_dir"`
	// PublicURL is the origin browsers use to reach the service, e.g.
	// "https://music.example.com". Requests that change state must come
	// from this origin. It also forms the OIDC redirect URI
	// <public_url>/auth/callback.
	PublicURL string `yaml:"public_url"`
	// OIDC signs listeners in to the player.
	OIDC OIDC `yaml:"oidc"`
	// Netease connects the NetEase Cloud Music API service.
	Netease Netease `yaml:"netease"`
	// Dev signs in with fixed accounts, skipping the admin setup token and
	// OIDC. Only a binary built with -tags dev accepts it.
	Dev bool `yaml:"dev"`
}

// OIDC configures the identity provider (Authelia).
type OIDC struct {
	Issuer       string `yaml:"issuer"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

// Netease configures the NetEase Cloud Music integration.
type Netease struct {
	// APIURL is the base URL of the API service container, reachable only
	// from Sonora. Empty turns NetEase off.
	APIURL string `yaml:"api_url"`
	// ImportGroup is the identity provider group whose members may import
	// songs into the library. Empty lets nobody import.
	ImportGroup string `yaml:"import_group"`
}

// Default returns the configuration used when no file is present.
func Default() Config {
	return Config{
		Listen:   ":14560",
		DataDir:  "./data",
		MediaDir: "./media",
	}
}

// Load reads the YAML file at path on top of the defaults.
// A missing file is not an error; the defaults are used. Unknown keys are
// rejected so a misplaced setting cannot be ignored.
func Load(path string) (Config, error) {
	cfg := Default()
	raw, err := os.ReadFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return cfg, fmt.Errorf("read config: %w", err)
	default:
		dec := yaml.NewDecoder(bytes.NewReader(raw))
		dec.KnownFields(true)
		if err := dec.Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
			return cfg, fmt.Errorf("parse config %s: %w", path, err)
		}
	}
	if secret := os.Getenv(ClientSecretEnv); secret != "" {
		cfg.OIDC.ClientSecret = secret
	}
	return cfg, cfg.validate()
}

func (c Config) validate() error {
	if c.Listen == "" {
		return errors.New("config: listen must not be empty")
	}
	if c.DataDir == "" || c.MediaDir == "" {
		return errors.New("config: data_dir and media_dir must not be empty")
	}
	if filepath.Clean(c.DataDir) == filepath.Clean(c.MediaDir) {
		return errors.New("config: data_dir and media_dir must be different directories")
	}
	if c.Netease.APIURL != "" {
		u, err := url.Parse(c.Netease.APIURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return errors.New("config: netease.api_url must be a URL such as http://netease-api:3000")
		}
	}
	if c.Dev {
		if !devmode.Available {
			return errors.New("config: dev is only accepted by a development build (go run -tags dev)")
		}
		host, _, err := net.SplitHostPort(c.Listen)
		if ip := net.ParseIP(host); err != nil || ip == nil || !ip.IsLoopback() {
			return errors.New("config: dev requires listen on a loopback address such as 127.0.0.1:14560")
		}
		return nil
	}
	u, err := url.Parse(c.PublicURL)
	if c.PublicURL == "" || err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" ||
		(u.Path != "" && u.Path != "/") {
		return errors.New("config: public_url must be an origin such as https://music.example.com")
	}
	if c.OIDC.Issuer == "" || c.OIDC.ClientID == "" || c.OIDC.ClientSecret == "" {
		return fmt.Errorf("config: oidc.issuer, oidc.client_id and oidc.client_secret (or %s) are required unless dev is set", ClientSecretEnv)
	}
	return nil
}
