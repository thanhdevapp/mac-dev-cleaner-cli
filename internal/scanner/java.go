package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// JavaGlobalPaths contains global Java/JVM cache paths
var JavaGlobalPaths = []struct {
	Path string
	Name string
}{
	// Maven
	{"~/.m2/repository", "Maven Local Repository"},
	// Gradle (note: ~/.gradle/caches already in Android scanner)
	{"~/.gradle/wrapper", "Gradle Wrapper Distributions"},
	{"~/.gradle/daemon", "Gradle Daemon Logs"},
}

// JavaMarkerFiles identify Java/Kotlin projects
var JavaMarkerFiles = map[string]string{
	"pom.xml":          "maven",
	"build.gradle":     "gradle",
	"build.gradle.kts": "gradle",
	"settings.gradle":  "gradle",
}

// ScanJava scans for Java/Kotlin development artifacts
func (s *Scanner) ScanJava(maxDepth int) []types.ScanResult {
	var results []types.ScanResult

	// Scan global caches
	for _, target := range JavaGlobalPaths {
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
			Type:      types.TypeJava,
			Size:      size,
			FileCount: count,
			Name:      target.Name,
		})
	}

	// Scan for Java projects in common development directories
	projectDirs := []string{
		"~/Documents",
		"~/Projects",
		"~/Development",
		"~/Developer",
		"~/Code",
		"~/repos",
		"~/workspace",
		"~/IdeaProjects", // IntelliJ default
	}

	for _, dir := range projectDirs {
		expandedDir := s.ExpandPath(dir)
		if !s.PathExists(expandedDir) {
			continue
		}

		javaArtifacts := s.findJavaArtifacts(expandedDir, maxDepth)
		results = append(results, javaArtifacts...)
	}

	return results
}

// findJavaArtifacts recursively finds Java project build artifacts
func (s *Scanner) findJavaArtifacts(root string, maxDepth int) []types.ScanResult {
	var results []types.ScanResult

	if maxDepth <= 0 {
		return results
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return results
	}

	// Check if this is a Java project
	projectType := ""
	hasBuildDir := false
	hasTargetDir := false

	for _, entry := range entries {
		name := entry.Name()

		if !entry.IsDir() {
			if pType, ok := JavaMarkerFiles[name]; ok {
				projectType = pType
			}
		} else {
			if name == "build" {
				hasBuildDir = true
			}
			if name == "target" {
				hasTargetDir = true
			}
		}
	}

	// Add build artifacts if Java project
	if projectType != "" {
		projectName := filepath.Base(root)

		// Maven: target directory
		if projectType == "maven" && hasTargetDir {
			targetPath := filepath.Join(root, "target")
			size, count, _ := s.calculateSize(targetPath)
			if size > 0 {
				results = append(results, types.ScanResult{
					Path:      targetPath,
					Type:      types.TypeJava,
					Size:      size,
					FileCount: count,
					Name:      projectName + "/target (Maven)",
				})
			}
		}

		// Gradle: build directory
		if projectType == "gradle" && hasBuildDir {
			buildPath := filepath.Join(root, "build")
			size, count, _ := s.calculateSize(buildPath)
			if size > 0 {
				results = append(results, types.ScanResult{
					Path:      buildPath,
					Type:      types.TypeJava,
					Size:      size,
					FileCount: count,
					Name:      projectName + "/build (Gradle)",
				})
			}
		}

		// Also check for .gradle directory in project root
		dotGradlePath := filepath.Join(root, ".gradle")
		if s.PathExists(dotGradlePath) {
			size, count, _ := s.calculateSize(dotGradlePath)
			if size > 0 {
				results = append(results, types.ScanResult{
					Path:      dotGradlePath,
					Type:      types.TypeJava,
					Size:      size,
					FileCount: count,
					Name:      projectName + "/.gradle",
				})
			}
		}

		// Don't recurse into Java projects
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

		// Skip build/target directories without marker files
		if name == "build" || name == "target" {
			continue
		}

		fullPath := filepath.Join(root, name)
		subResults := s.findJavaArtifacts(fullPath, maxDepth-1)
		results = append(results, subResults...)
	}

	return results
}
