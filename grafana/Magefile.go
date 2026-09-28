//go:build mage
// +build mage

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	// mage:import
	build "github.com/grafana/grafana-plugin-sdk-go/build"
)

// Default builds all targets and copies the binary to sub-plugin directories.
var Default = BuildAndCopy

// BuildAndCopy builds the backend for all platforms and copies binaries to sub-plugin directories.
func BuildAndCopy() error {
	build.BuildAll()
	return copyBinariesToSubPlugins()
}

// CopyBinaries copies the already-built backend binary to sub-plugin directories.
// Use this after a platform-specific build e.g. `mage build:linux && mage copyBinaries`.
func CopyBinaries() error {
	return copyBinariesToSubPlugins()
}

func copyBinariesToSubPlugins() error {
	if err := os.MkdirAll("dist/datasources/ssaiot-datasource", 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	// Copy every gpx_* binary from dist/ to the datasource sub-plugin directory
	// so Grafana can find the backend regardless of OS/arch.
	matches, err := filepath.Glob("dist/gpx_*")
	if err != nil {
		return err
	}
	for _, src := range matches {
		dst := filepath.Join("dist/datasources/ssaiot-datasource", filepath.Base(src))
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	fmt.Printf("copied %s → %s\n", src, dst)
	return nil
}
