//go:build prod

package static

import (
	"embed"
	"io/fs"
)

//go:embed all:dist/static
var embeddedFiles embed.FS

// FS returns the embedded filesystem for production
func FS() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "dist/static")
}
