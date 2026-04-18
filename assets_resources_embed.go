//go:build !debug

package main

import (
	"embed"
	"io/fs"
)

// Only embed installers.zip for production release builds
// Individual asset files are in assets/ folder for configuration loading
//
//go:embed installers.zip assets/installer-config.yaml assets/readme.txt
var embeddedAssets embed.FS

// EmbeddedAssetProvider wraps embed.FS for production use
type EmbeddedAssetProvider struct {
	fs embed.FS
}

// NewAssetProvider creates a new asset provider for production mode
func NewAssetProvider(embeddedFS embed.FS) *EmbeddedAssetProvider {
	return &EmbeddedAssetProvider{fs: embeddedFS}
}

// GetFS returns the embedded filesystem
func (p *EmbeddedAssetProvider) GetFS() fs.FS {
	return p.fs
}
