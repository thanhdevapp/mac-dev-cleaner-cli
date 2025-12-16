package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// PythonGlobalPaths contains global Python cache paths
var PythonGlobalPaths = []struct {
	Path string
	Name string
}{
	{"~/.cache/pip", "pip Cache"},
	{"~/.cache/pypoetry", "Poetry Cache"},
	{"~/.cache/pdm", "pdm Cache"},
	{"~/.cache/uv", "uv Cache"},
	{"~/.local/share/virtualenvs", "pipenv virtualenvs"},
}

// PythonProjectDirs are directories that may contain Python projects
var PythonProjectDirs = []string{
	"venv",
	".venv",
	"env",
	".env",
	"__pycache__",
	".pytest_cache",
	".tox",
	".mypy_cache",
	".ruff_cache",
}

// PythonMarkerFiles identify Python projects
var PythonMarkerFiles = []string{
	"requirements.txt",
	"setup.py",
	"pyproject.toml",
	"Pipfile",
	"setup.cfg",
}

// ScanPython scans for Python development artifacts
func (s *Scanner) ScanPython(maxDepth int) []types.ScanResult {
	var results []types.ScanResult

	// Scan global caches
	for _, target := range PythonGlobalPaths {
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
			Type:      types.TypePython,
			Size:      size,
			FileCount: count,
			Name:      target.Name,
		})
	}

	// Scan for Python projects in common development directories
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

		pythonArtifacts := s.findPythonArtifacts(expandedDir, maxDepth)
		results = append(results, pythonArtifacts...)
	}

	return results
}

// findPythonArtifacts recursively finds Python project artifacts
func (s *Scanner) findPythonArtifacts(root string, maxDepth int) []types.ScanResult {
	var results []types.ScanResult

	if maxDepth <= 0 {
		return results
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return results
	}

	// Check if this is a Python project
	isPythonProject := false
	for _, entry := range entries {
		if !entry.IsDir() {
			for _, marker := range PythonMarkerFiles {
				if entry.Name() == marker {
					isPythonProject = true
					break
				}
			}
		}
		if isPythonProject {
			break
		}
	}

	// Scan for artifacts in Python project
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		fullPath := filepath.Join(root, name)

		// Skip hidden dirs (except Python-specific ones)
		if strings.HasPrefix(name, ".") && !isPythonArtifactDir(name) {
			continue
		}

		// Skip common non-project dirs
		if shouldSkipDir(name) {
			continue
		}

		// Check if this is a Python artifact directory
		if isPythonArtifactDir(name) {
			size, count, _ := s.calculateSize(fullPath)
			if size > 0 {
				projectName := filepath.Base(root)
				results = append(results, types.ScanResult{
					Path:      fullPath,
					Type:      types.TypePython,
					Size:      size,
					FileCount: count,
					Name:      projectName + "/" + name,
				})
			}
			continue // Don't recurse into artifact dirs
		}

		// Recurse into subdirectories
		subResults := s.findPythonArtifacts(fullPath, maxDepth-1)
		results = append(results, subResults...)
	}

	return results
}

// isPythonArtifactDir checks if directory is a Python artifact
func isPythonArtifactDir(name string) bool {
	for _, artifactDir := range PythonProjectDirs {
		if name == artifactDir {
			return true
		}
	}
	return false
}
