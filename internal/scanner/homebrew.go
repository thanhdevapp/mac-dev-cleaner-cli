package scanner

import (
	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// HomebrewPaths contains Homebrew cache paths
var HomebrewPaths = []struct {
	Path string
	Name string
}{
	// User cache
	{"~/Library/Caches/Homebrew", "Homebrew Cache"},
	// Apple Silicon Homebrew
	{"/opt/homebrew/Library/Caches/Homebrew", "Homebrew Cache (ARM)"},
	// Intel Homebrew
	{"/usr/local/Homebrew/Library/Caches/Homebrew", "Homebrew Cache (Intel)"},
}

// ScanHomebrew scans for Homebrew caches
func (s *Scanner) ScanHomebrew() []types.ScanResult {
	var results []types.ScanResult

	for _, target := range HomebrewPaths {
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
			Type:      types.TypeHomebrew,
			Size:      size,
			FileCount: count,
			Name:      target.Name,
		})
	}

	return results
}
