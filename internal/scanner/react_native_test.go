package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

func TestScanReactNative(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create scanner: %v", err)
	}

	// Note: This test will scan actual TMPDIR
	// In a real environment, RN caches may or may not exist
	results := s.ScanReactNative()

	// Just verify it returns a slice (may be empty)
	if results == nil {
		t.Error("Expected non-nil results slice")
	}

	// Verify all results have correct type
	for _, result := range results {
		if result.Type != types.TypeReactNative {
			t.Errorf("Expected type %s, got %s", types.TypeReactNative, result.Type)
		}
		if result.Size == 0 {
			t.Errorf("Expected non-zero size for %s", result.Path)
		}
	}
}

func TestScanReactNativeWithMockCache(t *testing.T) {
	// Create temporary directory to simulate TMPDIR
	tmpDir := t.TempDir()

	// Create mock cache directories
	testCaches := []string{
		"metro-test-cache",
		"haste-map-test",
		"react-native-packager-cache-test",
		"react-test",
	}

	for _, cache := range testCaches {
		cachePath := filepath.Join(tmpDir, cache)
		if err := os.MkdirAll(cachePath, 0755); err != nil {
			t.Fatalf("Failed to create test cache: %v", err)
		}

		// Create test file
		testFile := filepath.Join(cachePath, "test.txt")
		if err := os.WriteFile(testFile, []byte("test data for cache"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Mock os.TempDir by temporarily creating scanner with different approach
	// In real test, we'd use interface or dependency injection
	// For now, this verifies the logic works

	s, err := New()
	if err != nil {
		t.Fatalf("Failed to create scanner: %v", err)
	}

	// Test that scanner can calculate size correctly
	size, count, err := s.calculateSize(filepath.Join(tmpDir, "metro-test-cache"))
	if err != nil {
		t.Errorf("Failed to calculate size: %v", err)
	}
	if size == 0 {
		t.Error("Expected non-zero size")
	}
	if count == 0 {
		t.Error("Expected non-zero file count")
	}
}

func TestReactNativeCachePatterns(t *testing.T) {
	// Verify cache patterns are defined
	if len(ReactNativeCachePaths) == 0 {
		t.Error("Expected non-empty ReactNativeCachePaths")
	}

	// Verify all patterns have name and pattern
	for i, cache := range ReactNativeCachePaths {
		if cache.Pattern == "" {
			t.Errorf("Cache pattern %d has empty Pattern", i)
		}
		if cache.Name == "" {
			t.Errorf("Cache pattern %d has empty Name", i)
		}
	}

	// Verify expected patterns exist
	expectedPatterns := []string{"metro-*", "haste-map-*", "react-native-packager-cache-*", "react-*"}
	for _, expected := range expectedPatterns {
		found := false
		for _, cache := range ReactNativeCachePaths {
			if cache.Pattern == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected pattern %s not found in ReactNativeCachePaths", expected)
		}
	}
}
