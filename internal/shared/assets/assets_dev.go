//go:build !prod

package assets

// URL returns the asset path unchanged in development mode.
// No cache busting is applied in development.
func URL(path string) string {
	return path
}
