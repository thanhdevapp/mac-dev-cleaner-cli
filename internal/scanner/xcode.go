package scanner

import (
	"os"
	"path/filepath"

	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// XcodePaths contains default Xcode-related paths to scan
var XcodePaths = []struct {
	Path string
	Name string
}{
	{"~/Library/Developer/Xcode/DerivedData", "Xcode DerivedData"},
	{"~/Library/Developer/Xcode/Archives", "Xcode Archives"},
	{"~/Library/Caches/com.apple.dt.Xcode", "Xcode Caches"},
	{"~/Library/Developer/CoreSimulator/Caches", "Simulator Caches"},
	{"~/Library/Caches/CocoaPods", "CocoaPods Cache"},
}

// ScanXcode scans for Xcode/iOS development artifacts
func (s *Scanner) ScanXcode() []types.ScanResult {
	var results []types.ScanResult

	for _, target := range XcodePaths {
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
			Type:      types.TypeXcode,
			Size:      size,
			FileCount: count,
			Name:      target.Name,
		})
	}

	// Also scan for individual DerivedData folders if parent exists
	derivedDataPath := s.ExpandPath("~/Library/Developer/Xcode/DerivedData")
	if s.PathExists(derivedDataPath) {
		entries, err := os.ReadDir(derivedDataPath)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() && entry.Name() != "ModuleCache.noindex" {
					subPath := filepath.Join(derivedDataPath, entry.Name())
					size, count, _ := s.calculateSize(subPath)
					if size > 0 {
						results = append(results, types.ScanResult{
							Path:      subPath,
							Type:      types.TypeXcode,
							Size:      size,
							FileCount: count,
							Name:      "DerivedData/" + entry.Name(),
						})
					}
				}
			}
		}
	}

	return results
}
