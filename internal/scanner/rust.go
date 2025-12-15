package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// RustGlobalPaths contains global Rust/Cargo cache paths
var RustGlobalPaths = []struct {
	Path string
	Name string
}{
	{"~/.cargo/registry", "Cargo Registry"},
	{"~/.cargo/git", "Cargo Git Cache"},
}

// getCargoHome returns CARGO_HOME or default ~/.cargo
func getCargoHome() string {
	if cargoHome := os.Getenv("CARGO_HOME"); cargoHome != "" {
		return cargoHome
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cargo")
}

// ScanRust scans for Rust/Cargo development artifacts
func (s *Scanner) ScanRust(maxDepth int) []types.ScanResult {
	var results []types.ScanResult

	cargoHome := getCargoHome()

	// Scan global caches (using CARGO_HOME)
	globalPaths := []struct {
		Path string
		Name string
	}{
		{filepath.Join(cargoHome, "registry"), "Cargo Registry"},
		{filepath.Join(cargoHome, "git"), "Cargo Git Cache"},
	}

	for _, target := range globalPaths {
		if !s.PathExists(target.Path) {
			continue
		}

		size, count, err := s.calculateSize(target.Path)
		if err != nil || size == 0 {
			continue
		}

		results = append(results, types.ScanResult{
			Path:      target.Path,
			Type:      types.TypeRust,
			Size:      size,
			FileCount: count,
			Name:      target.Name,
		})
	}

	// Scan for Rust projects' target directories
	projectDirs := []string{
		"~/Documents",
		"~/Projects",
		"~/Development",
		"~/Developer",
		"~/Code",
		"~/repos",
		"~/workspace",
	}

	for _, dir := range projectDirs {
		expandedDir := s.ExpandPath(dir)
		if !s.PathExists(expandedDir) {
			continue
		}

		rustTargets := s.findRustTargets(expandedDir, maxDepth)
		results = append(results, rustTargets...)
	}

	return results
}

// findRustTargets recursively finds Rust target directories
func (s *Scanner) findRustTargets(root string, maxDepth int) []types.ScanResult {
	var results []types.ScanResult

	if maxDepth <= 0 {
		return results
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return results
	}

	// Check if this directory contains Cargo.toml (is a Rust project)
	hasCargoToml := false
	hasTargetDir := false
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() == "Cargo.toml" {
			hasCargoToml = true
		}
		if entry.IsDir() && entry.Name() == "target" {
			hasTargetDir = true
		}
	}

	// If Rust project with target, add it
	if hasCargoToml && hasTargetDir {
		targetPath := filepath.Join(root, "target")
		size, count, _ := s.calculateSize(targetPath)
		if size > 0 {
			projectName := filepath.Base(root)
			results = append(results, types.ScanResult{
				Path:      targetPath,
				Type:      types.TypeRust,
				Size:      size,
				FileCount: count,
				Name:      projectName + "/target",
			})
		}
		// Don't recurse into Rust projects
		return results
	}

	// Recurse into subdirectories
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip hidden directories
		if strings.HasPrefix(name, ".") {
			continue
		}

		// Skip common non-project dirs
		if shouldSkipDir(name) {
			continue
		}

		// Skip target directories without Cargo.toml
		if name == "target" {
			continue
		}

		fullPath := filepath.Join(root, name)
		subResults := s.findRustTargets(fullPath, maxDepth-1)
		results = append(results, subResults...)
	}

	return results
}
