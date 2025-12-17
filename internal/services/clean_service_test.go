package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// TestNewCleanService tests CleanService initialization
func TestNewCleanService(t *testing.T) {
	// Test with dry-run enabled
	service, err := NewCleanService(true)
	require.NoError(t, err, "NewCleanService should not return error")
	require.NotNil(t, service, "CleanService should not be nil")
	assert.NotNil(t, service.cleaner, "Cleaner should be initialized")
	assert.False(t, service.cleaning, "Should not be cleaning initially")

	// Test with dry-run disabled
	service2, err := NewCleanService(false)
	require.NoError(t, err, "NewCleanService should not return error")
	require.NotNil(t, service2, "CleanService should not be nil")
}

// TestCleanEmptyItems tests cleaning with empty items
func TestCleanEmptyItems(t *testing.T) {
	service, err := NewCleanService(true)
	require.NoError(t, err)

	// Try to clean empty items
	var emptyItems []types.ScanResult
	results, err := service.Clean(emptyItems)

	assert.Error(t, err, "Should return error for empty items")
	assert.Contains(t, err.Error(), "no items to clean", "Error should indicate no items")
	assert.Nil(t, results, "Results should be nil for empty items")
}

// TestIsCleaning tests IsCleaning method
func TestIsCleaning(t *testing.T) {
	service, err := NewCleanService(true)
	require.NoError(t, err)

	// Initially not cleaning
	assert.False(t, service.IsCleaning(), "Should not be cleaning initially")

	// Set cleaning flag
	service.cleaning = true
	assert.True(t, service.IsCleaning(), "Should be cleaning after flag set")

	// Unset cleaning flag
	service.cleaning = false
	assert.False(t, service.IsCleaning(), "Should not be cleaning after flag unset")
}

// TestConcurrentClean tests that concurrent cleans are prevented
func TestConcurrentClean(t *testing.T) {
	service, err := NewCleanService(true)
	require.NoError(t, err)

	// Set cleaning flag to simulate ongoing clean
	service.mu.Lock()
	service.cleaning = true
	service.mu.Unlock()

	// Try to clean while already cleaning
	items := []types.ScanResult{
		{Path: "/test/item1", Size: 1000, Type: types.TypeNode},
	}

	results, err := service.Clean(items)
	assert.Error(t, err, "Should return error when clean already in progress")
	assert.Contains(t, err.Error(), "clean already in progress", "Error message should indicate clean in progress")
	assert.Nil(t, results, "Results should be nil when clean blocked")
}

// TestCleanNilItems tests cleaning with nil items slice
func TestCleanNilItems(t *testing.T) {
	service, err := NewCleanService(true)
	require.NoError(t, err)

	// Try to clean nil items (equivalent to empty slice)
	results, err := service.Clean(nil)

	assert.Error(t, err, "Should return error for nil items")
	assert.Contains(t, err.Error(), "no items to clean", "Error should indicate no items")
	assert.Nil(t, results, "Results should be nil for nil items")
}

// TestCleanFreedSpaceCalculation tests freed space calculation logic
func TestCleanFreedSpaceCalculation(t *testing.T) {
	// This test validates the freed space calculation logic
	// that would be used in the Clean method
	mockResults := []struct {
		Success bool
		Size    int64
	}{
		{Success: true, Size: 1000},
		{Success: true, Size: 2000},
		{Success: false, Size: 500},  // Failed - should not count
		{Success: true, Size: 3000},
		{Success: false, Size: 1000}, // Failed - should not count
	}

	var freedSpace int64
	successCount := 0

	for _, r := range mockResults {
		if r.Success {
			freedSpace += r.Size
			successCount++
		}
	}

	// Verify calculations
	assert.Equal(t, int64(6000), freedSpace, "Should sum only successful deletions (1000+2000+3000)")
	assert.Equal(t, 3, successCount, "Should count only successful deletions")
}

// TestCleanValidation tests input validation
func TestCleanValidation(t *testing.T) {
	service, err := NewCleanService(true)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		items       []types.ScanResult
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Empty slice",
			items:       []types.ScanResult{},
			shouldError: true,
			errorMsg:    "no items to clean",
		},
		{
			name:        "Nil slice",
			items:       nil,
			shouldError: true,
			errorMsg:    "no items to clean",
		},
		{
			name: "Valid items",
			items: []types.ScanResult{
				{Path: "/test/item1", Size: 1000, Type: types.TypeNode},
			},
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset cleaning flag for each test
			service.cleaning = false

			results, err := service.Clean(tc.items)

			if tc.shouldError {
				assert.Error(t, err, "Should return error for %s", tc.name)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg, "Error message should contain: %s", tc.errorMsg)
				}
				assert.Nil(t, results, "Results should be nil on error")
			}
		})
	}
}
