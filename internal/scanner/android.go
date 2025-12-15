package scanner

import (
	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// AndroidPaths contains default Android-related paths to scan
var AndroidPaths = []struct {
	Path string
	Name string
}{
	{"~/.gradle/caches", "Gradle Caches"},
	{"~/.gradle/wrapper", "Gradle Wrapper"},
	{"~/.android/cache", "Android SDK Cache"},
	{"~/.android/build-cache", "Android Build Cache"},
	{"~/Library/Android/sdk/system-images", "Android System Images"},
}

// ScanAndroid scans for Android development artifacts
func (s *Scanner) ScanAndroid() []types.ScanResult {
	var results []types.ScanResult

	for _, target := range AndroidPaths {
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
			Type:      types.TypeAndroid,
			Size:      size,
			FileCount: count,
			Name:      target.Name,
		})
	}

	return results
}
