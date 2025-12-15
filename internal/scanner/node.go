package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// NodeGlobalPaths contains global Node.js cache paths
var NodeGlobalPaths = []struct {
	Path string
	Name string
}{
	{"~/.npm", "npm Cache"},
	{"~/.pnpm-store", "pnpm Store"},
	{"~/.yarn/cache", "Yarn Cache"},
	{"~/.bun/install/cache", "Bun Cache"},
}

// SkipDirs are directories to skip when searching for node_modules
var SkipDirs = []string{
	".git",
	"node_modules",
	".Trash",
	"Library",
}

// ScanNode scans for Node.js development artifacts
func (s *Scanner) ScanNode(maxDepth int) []types.ScanResult {
	var results []types.ScanResult

	// Scan global caches
	for _, target := range NodeGlobalPaths {
		path := s.ExpandPath(target.Path)
		if !s.PathExists(path) {
			continue
		}

		size, count, err := s.calculateSize(path)
		if err != nil || size == 0 {
			continue
		}

		results = append(results, types.ScanResult{
			Path:      path,
			Type:      types.TypeNode,
			Size:      size,
			FileCount: count,
			Name:      target.Name,
		})
	}

	// Scan for project node_modules in common development directories
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

		nodeModules := s.findNodeModules(expandedDir, maxDepth)
		results = append(results, nodeModules...)
	}

	return results
}

// findNodeModules recursively finds node_modules directories
func (s *Scanner) findNodeModules(root string, maxDepth int) []types.ScanResult {
	var results []types.ScanResult

	if maxDepth <= 0 {
		return results
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return results
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip hidden and known non-project directories
		if shouldSkipDir(name) {
			continue
		}

		fullPath := filepath.Join(root, name)

		if name == "node_modules" {
			size, count, _ := s.calculateSize(fullPath)
			if size > 0 {
				// Get parent project name
				projectName := filepath.Base(root)
				results = append(results, types.ScanResult{
					Path:      fullPath,
					Type:      types.TypeNode,
					Size:      size,
					FileCount: count,
					Name:      projectName + "/node_modules",
				})
			}
			continue // Don't recurse into node_modules
		}

		// Recurse into subdirectories
		subResults := s.findNodeModules(fullPath, maxDepth-1)
		results = append(results, subResults...)
	}

	return results
}

// shouldSkipDir checks if a directory should be skipped
func shouldSkipDir(name string) bool {
	// Skip hidden directories
	if strings.HasPrefix(name, ".") {
		return true
	}

	for _, skip := range SkipDirs {
		if name == skip {
			return true
		}
	}

	return false
}
