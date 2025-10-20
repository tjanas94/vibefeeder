//go:build prod

package static

import (
	"embed"
	"io/fs"
)

//go:embed all:dist/static
var embeddedFiles embed.FS

//go:embed dist/manifest.json
var ManifestData []byte

//go:embed dist/favicon.ico
var FaviconData []byte

//go:embed dist/icon.svg
var IconSVGData []byte

//go:embed dist/apple-touch-icon.png
var AppleTouchIconData []byte

// FS returns the embedded filesystem for production
func FS() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "dist/static")
}
