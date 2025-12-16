package scanner

import (
	"os"
	"path/filepath"

	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// FlutterGlobalPaths contains global Flutter/Dart cache paths
var FlutterGlobalPaths = []struct {
	Path string
	Name string
}{
	{"~/.pub-cache", "Pub Cache"},
	{"~/.dart_tool", "Dart Tool Cache"},
	{"~/Library/Caches/Flutter", "Flutter Cache"},
	{"~/Library/Caches/dart", "Dart Cache"},
}

// ScanFlutter scans for Flutter/Dart development artifacts
func (s *Scanner) ScanFlutter(maxDepth int) []types.ScanResult {
	var results []types.ScanResult

	// Scan global caches
	for _, target := range FlutterGlobalPaths {
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
			Type:      types.TypeFlutter,
			Size:      size,
			FileCount: count,
			Name:      target.Name,
		})
	}

	// Scan for Flutter projects in common development directories
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

		flutterProjects := s.findFlutterProjects(expandedDir, maxDepth)
		results = append(results, flutterProjects...)
	}

	return results
}

// findFlutterProjects recursively finds Flutter projects via pubspec.yaml
func (s *Scanner) findFlutterProjects(root string, maxDepth int) []types.ScanResult {
	var results []types.ScanResult

	if maxDepth <= 0 {
		return results
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return results
	}

	// Check if current directory is a Flutter project
	hasPubspec := false
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() == "pubspec.yaml" {
			hasPubspec = true
			break
		}
	}

	// If Flutter project, scan build artifacts
	if hasPubspec {
		projectName := filepath.Base(root)
		buildTargets := []struct {
			subPath string
			name    string
		}{
			{"build", "build/"},
			{".dart_tool", ".dart_tool/"},
			{"ios/build", "ios/build/"},
			{"android/build", "android/build/"},
			{"macos/build", "macos/build/"},
			{"linux/build", "linux/build/"},
			{"windows/build", "windows/build/"},
			{"web/build", "web/build/"},
		}

		for _, target := range buildTargets {
			buildPath := filepath.Join(root, target.subPath)
			if !s.PathExists(buildPath) {
				continue
			}

			size, count, _ := s.calculateSize(buildPath)
			if size > 0 {
				results = append(results, types.ScanResult{
					Path:      buildPath,
					Type:      types.TypeFlutter,
					Size:      size,
					FileCount: count,
					Name:      projectName + "/" + target.name,
				})
			}
		}

		// Don't recurse into Flutter project subdirectories
		return results
	}

	// Recurse into subdirectories
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
		subResults := s.findFlutterProjects(fullPath, maxDepth-1)
		results = append(results, subResults...)
	}

	return results
}
