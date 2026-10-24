//go:build dev

package config

import "testing"

func TestDevBuildAcceptsDevOnLoopback(t *testing.T) {
	cfg, err := Load(write(t, "listen: 127.0.0.1:14560\ndev: true\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Dev || cfg.PublicURL != "" {
		t.Fatalf("cfg = %+v", cfg)
	}
	for _, listen := range []string{":14560", "0.0.0.0:14560", "192.168.1.2:14560"} {
		if _, err := Load(write(t, "listen: \""+listen+"\"\ndev: true\n")); err == nil {
			t.Errorf("dev accepted listen %s", listen)
		}
	}
}
