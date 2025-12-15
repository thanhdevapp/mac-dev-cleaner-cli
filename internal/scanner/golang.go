package scanner

import (
	"os"
	"path/filepath"

	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// getGOCACHE returns GOCACHE path or default
func getGOCACHE() string {
	if gocache := os.Getenv("GOCACHE"); gocache != "" {
		return gocache
	}
	// macOS default
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "Caches", "go-build")
}

// getGOMODCACHE returns GOMODCACHE path or default
func getGOMODCACHE() string {
	if gomodcache := os.Getenv("GOMODCACHE"); gomodcache != "" {
		return gomodcache
	}
	// Default: $GOPATH/pkg/mod or ~/go/pkg/mod
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, _ := os.UserHomeDir()
		gopath = filepath.Join(home, "go")
	}
	return filepath.Join(gopath, "pkg", "mod")
}

// ScanGo scans for Go development artifacts
func (s *Scanner) ScanGo(maxDepth int) []types.ScanResult {
	var results []types.ScanResult

	// Go build cache
	gocache := getGOCACHE()
	if s.PathExists(gocache) {
		size, count, err := s.calculateSize(gocache)
		if err == nil && size > 0 {
			results = append(results, types.ScanResult{
				Path:      gocache,
				Type:      types.TypeGo,
				Size:      size,
				FileCount: count,
				Name:      "Go Build Cache",
			})
		}
	}

	// Go module cache
	gomodcache := getGOMODCACHE()
	if s.PathExists(gomodcache) {
		size, count, err := s.calculateSize(gomodcache)
		if err == nil && size > 0 {
			results = append(results, types.ScanResult{
				Path:      gomodcache,
				Type:      types.TypeGo,
				Size:      size,
				FileCount: count,
				Name:      "Go Module Cache",
			})
		}
	}

	// Go test cache (same location as build cache typically)
	gotestcache := os.Getenv("GOTESTCACHE")
	if gotestcache != "" && gotestcache != gocache && s.PathExists(gotestcache) {
		size, count, err := s.calculateSize(gotestcache)
		if err == nil && size > 0 {
			results = append(results, types.ScanResult{
				Path:      gotestcache,
				Type:      types.TypeGo,
				Size:      size,
				FileCount: count,
				Name:      "Go Test Cache",
			})
		}
	}

	return results
}
