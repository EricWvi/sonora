// Package sonora embeds the built player and admin console into the Go
// binary.
package sonora

import (
	"embed"
	"io/fs"
)

// The directories are filled by scripts/build.sh from frontend/dist and
// admin/dist. Tracked .gitkeep files keep them present so the package
// compiles before a build.
//
//go:embed all:web/player all:web/admin
var dist embed.FS

// Player returns the built player rooted at its dist directory.
func Player() fs.FS { return sub("web/player") }

// Admin returns the built admin console rooted at its dist directory.
func Admin() fs.FS { return sub("web/admin") }

func sub(dir string) fs.FS {
	s, err := fs.Sub(dist, dir)
	if err != nil {
		panic(err) // the path is a compile-time constant
	}
	return s
}
