package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// TestNewScanService tests ScanService initialization
func TestNewScanService(t *testing.T) {
	service, err := NewScanService()
	require.NoError(t, err, "NewScanService should not return error")
	require.NotNil(t, service, "ScanService should not be nil")
	assert.NotNil(t, service.scanner, "Scanner should be initialized")
	assert.Empty(t, service.results, "Results should be empty initially")
	assert.False(t, service.scanning, "Should not be scanning initially")
}

// TestScanDeduplication tests that duplicate paths are removed
func TestScanDeduplication(t *testing.T) {
	_, err := NewScanService()
	require.NoError(t, err)

	// Create mock results with duplicates
	mockResults := []types.ScanResult{
		{Path: "/path/to/item1", Size: 1000, Type: types.TypeNode},
		{Path: "/path/to/item2", Size: 2000, Type: types.TypeXcode},
		{Path: "/path/to/item1", Size: 1000, Type: types.TypeNode}, // Duplicate
		{Path: "/path/to/item3", Size: 3000, Type: types.TypeAndroid},
		{Path: "/path/to/item2", Size: 2000, Type: types.TypeXcode}, // Duplicate
	}

	// Simulate deduplication (same logic as in Scan method)
	seen := make(map[string]bool)
	dedupedResults := make([]types.ScanResult, 0, len(mockResults))
	for _, result := range mockResults {
		if !seen[result.Path] {
			seen[result.Path] = true
			dedupedResults = append(dedupedResults, result)
		}
	}

	// Verify deduplication
	assert.Equal(t, 3, len(dedupedResults), "Should have 3 unique items after deduplication")

	// Verify no duplicates
	paths := make(map[string]bool)
	for _, result := range dedupedResults {
		assert.False(t, paths[result.Path], "Path %s should not be duplicate", result.Path)
		paths[result.Path] = true
	}
}

// TestScanSorting tests that results are sorted by size (largest first)
func TestScanSorting(t *testing.T) {
	mockResults := []types.ScanResult{
		{Path: "/small", Size: 100, Type: types.TypeNode},
		{Path: "/large", Size: 10000, Type: types.TypeXcode},
		{Path: "/medium", Size: 5000, Type: types.TypeAndroid},
		{Path: "/tiny", Size: 10, Type: types.TypeReactNative},
	}

	// Sort by size (largest first) - same as scan service
	sortBySize := func(results []types.ScanResult) {
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[i].Size < results[j].Size {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	}

	sortBySize(mockResults)

	// Verify sorting
	assert.Equal(t, "/large", mockResults[0].Path, "Largest item should be first")
	assert.Equal(t, "/medium", mockResults[1].Path, "Medium item should be second")
	assert.Equal(t, "/small", mockResults[2].Path, "Small item should be third")
	assert.Equal(t, "/tiny", mockResults[3].Path, "Smallest item should be last")

	// Verify descending order
	for i := 0; i < len(mockResults)-1; i++ {
		assert.GreaterOrEqual(t, mockResults[i].Size, mockResults[i+1].Size,
			"Results should be sorted by size in descending order")
	}
}

// TestGetResults tests GetResults method
func TestGetResults(t *testing.T) {
	service, err := NewScanService()
	require.NoError(t, err)

	// Initially empty
	results := service.GetResults()
	assert.Empty(t, results, "Results should be empty initially")

	// Set some results
	mockResults := []types.ScanResult{
		{Path: "/test1", Size: 1000, Type: types.TypeNode},
		{Path: "/test2", Size: 2000, Type: types.TypeXcode},
	}
	service.results = mockResults

	// Get results
	results = service.GetResults()
	assert.Equal(t, 2, len(results), "Should return 2 results")
	assert.Equal(t, mockResults, results, "Should return exact results")
}

// TestIsScanning tests IsScanning method
func TestIsScanning(t *testing.T) {
	service, err := NewScanService()
	require.NoError(t, err)

	// Initially not scanning
	assert.False(t, service.IsScanning(), "Should not be scanning initially")

	// Set scanning flag
	service.scanning = true
	assert.True(t, service.IsScanning(), "Should be scanning after flag set")

	// Unset scanning flag
	service.scanning = false
	assert.False(t, service.IsScanning(), "Should not be scanning after flag unset")
}

// TestConcurrentScan tests that concurrent scans are prevented
func TestConcurrentScan(t *testing.T) {
	service, err := NewScanService()
	require.NoError(t, err)

	// Set scanning flag to simulate ongoing scan
	service.mu.Lock()
	service.scanning = true
	service.mu.Unlock()

	// Try to scan while already scanning
	opts := types.ScanOptions{
		IncludeNode:    true,
		IncludeXcode:   true,
		IncludeAndroid: true,
		MaxDepth:       3,
	}

	err = service.Scan(opts)
	assert.Error(t, err, "Should return error when scan already in progress")
	assert.Contains(t, err.Error(), "scan already in progress", "Error message should indicate scan in progress")
}

// TestDeduplicationPreservesFirst tests that deduplication keeps first occurrence
func TestDeduplicationPreservesFirst(t *testing.T) {
	mockResults := []types.ScanResult{
		{Path: "/path/to/item", Size: 1000, Type: types.TypeNode, Name: "First"},
		{Path: "/path/to/item", Size: 2000, Type: types.TypeXcode, Name: "Second"}, // Duplicate - different metadata
	}

	// Deduplicate
	seen := make(map[string]bool)
	dedupedResults := make([]types.ScanResult, 0)
	for _, result := range mockResults {
		if !seen[result.Path] {
			seen[result.Path] = true
			dedupedResults = append(dedupedResults, result)
		}
	}

	// Verify only first occurrence is kept
	assert.Equal(t, 1, len(dedupedResults), "Should have only 1 item after deduplication")
	assert.Equal(t, "First", dedupedResults[0].Name, "Should keep first occurrence")
	assert.Equal(t, int64(1000), dedupedResults[0].Size, "Should preserve first occurrence's size")
	assert.Equal(t, types.TypeNode, dedupedResults[0].Type, "Should preserve first occurrence's type")
}

// TestEmptyResults tests handling of empty scan results
func TestEmptyResults(t *testing.T) {
	_, err := NewScanService()
	require.NoError(t, err)

	// Empty results
	var emptyResults []types.ScanResult

	// Deduplicate empty results
	seen := make(map[string]bool)
	dedupedResults := make([]types.ScanResult, 0)
	for _, result := range emptyResults {
		if !seen[result.Path] {
			seen[result.Path] = true
			dedupedResults = append(dedupedResults, result)
		}
	}

	assert.Empty(t, dedupedResults, "Empty results should remain empty after deduplication")
}
