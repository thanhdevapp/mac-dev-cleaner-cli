package scanner

import (
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

	return results
}
