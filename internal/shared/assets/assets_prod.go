//go:build prod

package assets

import (
	"encoding/json"
	"sync"

	static "github.com/tjanas94/vibefeeder"
)

var (
	manifest     map[string]string
	manifestOnce sync.Once
)

// loadManifest ensures the manifest is loaded only once
func loadManifest() {
	manifestOnce.Do(func() {
		manifest = make(map[string]string)

		if err := json.Unmarshal(static.ManifestData, &manifest); err != nil {
			// In case of error, manifest stays empty (fallback)
			manifest = make(map[string]string)
		}
	})
}

// URL returns the versioned URL for a given asset path.
// In production, it appends ?v=hash query parameter for cache busting.
func URL(path string) string {
	loadManifest()

	if hash, exists := manifest[path]; exists {
		return path + "?v=" + hash
	}

	// Fallback to original path if asset not in manifest
	return path
}
