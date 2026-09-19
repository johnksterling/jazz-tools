package web

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var distFS embed.FS

// Assets returns the filesystem rooted inside the embedded dist directory.
func Assets() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
