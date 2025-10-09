//go:build !prod

package static

import (
	"io/fs"
)

// FS returns nil for development (not used)
func FS() (fs.FS, error) {
	return nil, nil
}
