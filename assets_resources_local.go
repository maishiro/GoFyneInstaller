//go:build debug

package main

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

// Embed all assets for development mode (local reference)
// This is replaced by production embedding in assets_resources_embed.go
//
//go:embed assets/*
var embeddedAssets embed.FS

// LocalAssetProvider uses local filesystem for development
type LocalAssetProvider struct {
	fs fs.FS
}

// NewAssetProvider creates a new asset provider for development mode
func NewAssetProvider(embeddedFS embed.FS) *LocalAssetProvider {
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exeDir := filepath.Dir(exePath)
	assetsPath := filepath.Join(exeDir, "assets")
	return &LocalAssetProvider{fs: os.DirFS(assetsPath)}
}

// GetFS returns the local filesystem
func (p *LocalAssetProvider) GetFS() fs.FS {
	return p.fs
}
