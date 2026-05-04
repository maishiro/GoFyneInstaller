package script

import "io/fs"

// AssetProvider provides unified interface for accessing assets in both
// development (local filesystem) and production (embedded zip) modes
type AssetProvider interface {
	// GetFS returns the filesystem interface for accessing assets
	GetFS() fs.FS
}
