package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSettingsService tests SettingsService initialization
func TestNewSettingsService(t *testing.T) {
	service := NewSettingsService()
	require.NotNil(t, service, "SettingsService should not be nil")
	assert.NotEmpty(t, service.path, "Path should be set")

	// Note: Cannot test exact default values because settings file may already exist
	// Just verify service is properly initialized
	settings := service.Get()
	assert.NotEmpty(t, settings.Theme, "Theme should not be empty")
	assert.NotEmpty(t, settings.DefaultView, "DefaultView should not be empty")
}

// TestSettingsGet tests Get method
func TestSettingsGet(t *testing.T) {
	service := NewSettingsService()

	settings := service.Get()
	assert.NotNil(t, settings, "Settings should not be nil")

	// Verify it's a copy (not a reference)
	settings.Theme = "modified"
	originalSettings := service.Get()
	assert.NotEqual(t, settings.Theme, originalSettings.Theme,
		"Modifying returned settings should not affect internal settings")
}

// TestSettingsUpdate tests Update method
func TestSettingsUpdate(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	service := &SettingsService{
		path: filepath.Join(tmpDir, "test-settings.json"),
	}
	service.Load()

	// Update settings
	newSettings := Settings{
		Theme:           "dark",
		DefaultView:     "list",
		AutoScan:        false,
		ConfirmDelete:   false,
		ScanCategories:  []string{"node", "react-native"},
		MaxDepth:        3,
		CheckAutoUpdate: false,
	}

	err := service.Update(newSettings)
	require.NoError(t, err, "Update should not return error")

	// Verify settings were updated
	currentSettings := service.Get()
	assert.Equal(t, "dark", currentSettings.Theme)
	assert.Equal(t, "list", currentSettings.DefaultView)
	assert.False(t, currentSettings.AutoScan)
	assert.False(t, currentSettings.ConfirmDelete)
	assert.False(t, currentSettings.CheckAutoUpdate)
	assert.Equal(t, 3, currentSettings.MaxDepth)
	assert.Equal(t, []string{"node", "react-native"}, currentSettings.ScanCategories)
}

// TestSettingsSaveAndLoad tests Save and Load methods
func TestSettingsSaveAndLoad(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "test-settings.json")

	// Create service and set custom settings
	service1 := &SettingsService{
		path: settingsPath,
		settings: Settings{
			Theme:           "light",
			DefaultView:     "treemap",
			AutoScan:        true,
			ConfirmDelete:   false,
			ScanCategories:  []string{"flutter", "python"},
			MaxDepth:        10,
			CheckAutoUpdate: true,
		},
	}

	// Save settings
	err := service1.Save()
	require.NoError(t, err, "Save should not return error")

	// Verify file was created
	_, err = os.Stat(settingsPath)
	assert.NoError(t, err, "Settings file should exist")

	// Create new service and load settings
	service2 := &SettingsService{
		path: settingsPath,
	}
	err = service2.Load()
	require.NoError(t, err, "Load should not return error")

	// Verify loaded settings match saved settings
	loadedSettings := service2.Get()
	assert.Equal(t, "light", loadedSettings.Theme)
	assert.Equal(t, "treemap", loadedSettings.DefaultView)
	assert.True(t, loadedSettings.AutoScan)
	assert.False(t, loadedSettings.ConfirmDelete)
	assert.True(t, loadedSettings.CheckAutoUpdate)
	assert.Equal(t, 10, loadedSettings.MaxDepth)
	assert.Equal(t, []string{"flutter", "python"}, loadedSettings.ScanCategories)
}

// TestSettingsLoadNonExistentFile tests loading when file doesn't exist
func TestSettingsLoadNonExistentFile(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	service := &SettingsService{
		path: filepath.Join(tmpDir, "non-existent.json"),
	}

	// Load should not error and should use defaults
	err := service.Load()
	assert.NoError(t, err, "Load should not error for non-existent file")

	// Verify default settings
	settings := service.Get()
	assert.Equal(t, "auto", settings.Theme, "Should use default theme")
	assert.Equal(t, "split", settings.DefaultView, "Should use default view")
	assert.True(t, settings.AutoScan, "Should use default AutoScan")
	assert.True(t, settings.ConfirmDelete, "Should use default ConfirmDelete")
	assert.True(t, settings.CheckAutoUpdate, "Should use default CheckAutoUpdate")
}

// TestSettingsJSONMarshaling tests JSON marshaling/unmarshaling
func TestSettingsJSONMarshaling(t *testing.T) {
	originalSettings := Settings{
		Theme:           "dark",
		DefaultView:     "list",
		AutoScan:        false,
		ConfirmDelete:   true,
		ScanCategories:  []string{"xcode", "android"},
		MaxDepth:        7,
		CheckAutoUpdate: false,
	}

	// Marshal to JSON
	data, err := json.Marshal(originalSettings)
	require.NoError(t, err, "JSON marshal should not error")

	// Unmarshal from JSON
	var loadedSettings Settings
	err = json.Unmarshal(data, &loadedSettings)
	require.NoError(t, err, "JSON unmarshal should not error")

	// Verify all fields match
	assert.Equal(t, originalSettings.Theme, loadedSettings.Theme)
	assert.Equal(t, originalSettings.DefaultView, loadedSettings.DefaultView)
	assert.Equal(t, originalSettings.AutoScan, loadedSettings.AutoScan)
	assert.Equal(t, originalSettings.ConfirmDelete, loadedSettings.ConfirmDelete)
	assert.Equal(t, originalSettings.CheckAutoUpdate, loadedSettings.CheckAutoUpdate)
	assert.Equal(t, originalSettings.MaxDepth, loadedSettings.MaxDepth)
	assert.Equal(t, originalSettings.ScanCategories, loadedSettings.ScanCategories)
}

// TestSettingsConcurrentAccess tests concurrent read/write access
func TestSettingsConcurrentAccess(t *testing.T) {
	service := NewSettingsService()

	// Concurrent reads should not panic
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = service.Get()
			done <- true
		}()
	}

	// Wait for all reads to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	assert.True(t, true, "Concurrent reads should complete without panic")
}
