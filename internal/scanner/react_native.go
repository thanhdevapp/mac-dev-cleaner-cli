package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// CachePattern represents a cache pattern to match in TMPDIR
type CachePattern struct {
	Pattern string
	Name    string
}

// ReactNativeCachePaths contains RN-specific cache locations in TMPDIR
var ReactNativeCachePaths = []CachePattern{
	{Pattern: "metro-*", Name: "Metro Bundler Cache"},
	{Pattern: "haste-map-*", Name: "Haste Map Cache"},
	{Pattern: "react-native-packager-cache-*", Name: "RN Packager Cache"},
	{Pattern: "react-*", Name: "React Native Temp Files"},
}

// ScanReactNative scans for React Native caches in TMPDIR
func (s *Scanner) ScanReactNative() []types.ScanResult {
	results := make([]types.ScanResult, 0)
	tmpDir := os.TempDir()

	for _, cache := range ReactNativeCachePaths {
		pattern := filepath.Join(tmpDir, cache.Pattern)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, match := range matches {
			// Skip if not a directory
			info, err := os.Stat(match)
			if err != nil || !info.IsDir() {
				continue
			}

			size, count, err := s.calculateSize(match)
			if err != nil || size == 0 {
				continue
			}

			results = append(results, types.ScanResult{
				Path:      match,
				Type:      types.TypeReactNative,
				Size:      size,
				FileCount: count,
				Name:      cache.Name,
			})
		}
	}

	// Also scan project-specific builds
	projectResults := s.ScanReactNativeProjects()
	results = append(results, projectResults...)

	return results
}

// isReactNativeProject checks if a directory is a React Native project
func (s *Scanner) isReactNativeProject(path string) bool {
	packageJSONPath := filepath.Join(path, "package.json")

	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return false
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}

	// Check for react-native dependency
	if _, ok := pkg.Dependencies["react-native"]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies["react-native"]; ok {
		return true
	}

	return false
}

// ScanReactNativeProjects scans for React Native project-specific build artifacts
func (s *Scanner) ScanReactNativeProjects() []types.ScanResult {
	results := make([]types.ScanResult, 0)

	// Search for React Native projects in common directories
	searchDirs := []string{
		"~/Documents",
		"~/Projects",
		"~/Development",
		"~/Developer",
		"~/Code",
		"~/repos",
		"~/workspace",
	}

	for _, dir := range searchDirs {
		expandedDir := s.ExpandPath(dir)
		if !s.PathExists(expandedDir) {
			continue
		}

		projects := s.findReactNativeProjects(expandedDir, 3)
		for _, projectPath := range projects {
			projectResults := s.scanReactNativeProjectBuilds(projectPath)
			results = append(results, projectResults...)
		}
	}

	return results
}

// findReactNativeProjects recursively finds React Native projects
func (s *Scanner) findReactNativeProjects(root string, maxDepth int) []string {
	var projects []string

	if maxDepth <= 0 {
		return projects
	}

	// Check if current directory is a React Native project
	if s.isReactNativeProject(root) {
		projects = append(projects, root)
		return projects // Don't recurse into RN project subdirectories
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return projects
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
		subProjects := s.findReactNativeProjects(fullPath, maxDepth-1)
		projects = append(projects, subProjects...)
	}

	return projects
}

// scanReactNativeProjectBuilds scans build directories in a React Native project
func (s *Scanner) scanReactNativeProjectBuilds(projectPath string) []types.ScanResult {
	results := make([]types.ScanResult, 0)
	projectName := filepath.Base(projectPath)

	// Directories to scan in React Native projects
	buildDirs := []struct {
		Path string
		Name string
	}{
		{filepath.Join(projectPath, "ios", "build"), "iOS Build"},
		{filepath.Join(projectPath, "android", "build"), "Android Build"},
		{filepath.Join(projectPath, "android", "app", "build"), "Android App Build"},
		{filepath.Join(projectPath, "android", ".gradle"), "Project Gradle Cache"},
	}

	for _, buildDir := range buildDirs {
		if !s.PathExists(buildDir.Path) {
			continue
		}

		size, count, err := s.calculateSize(buildDir.Path)
		if err != nil || size == 0 {
			continue
		}

		results = append(results, types.ScanResult{
			Path:      buildDir.Path,
			Type:      types.TypeReactNative,
			Size:      size,
			FileCount: count,
			Name:      projectName + " - " + buildDir.Name,
		})
	}

	return results
}
