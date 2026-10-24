//go:build !dev

package config

import (
	"strings"
	"testing"
)

func TestReleaseBuildRejectsDev(t *testing.T) {
	_, err := Load(write(t, "listen: 127.0.0.1:14560\ndev: true\n"))
	if err == nil || !strings.Contains(err.Error(), "development build") {
		t.Fatalf("err = %v, want development build error", err)
	}
}
