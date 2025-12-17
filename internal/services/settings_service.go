package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Settings struct {
	Theme          string   `json:"theme"`          // "light" | "dark" | "auto"
	DefaultView    string   `json:"defaultView"`    // "list" | "treemap" | "split"
	AutoScan       bool     `json:"autoScan"`       // Scan on launch
	ConfirmDelete  bool     `json:"confirmDelete"`  // Show confirm dialog
	ScanCategories []string `json:"scanCategories"` // ["xcode", "android", "node"]
	MaxDepth       int      `json:"maxDepth"`       // Tree depth limit
	CheckAutoUpdate bool    `json:"checkAutoUpdate"` // Check for updates on startup
}

type SettingsService struct {
	settings Settings
	path     string
	mu       sync.RWMutex
}

func NewSettingsService() *SettingsService {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".dev-cleaner-gui.json")

	s := &SettingsService{
		path: path,
	}
	s.Load()
	return s
}

func (s *SettingsService) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		// Set defaults
		s.settings = Settings{
			Theme:           "auto",
			DefaultView:     "split",
			AutoScan:        true,
			ConfirmDelete:   true,
			ScanCategories:  []string{"xcode", "android", "node"},
			MaxDepth:        5,
			CheckAutoUpdate: true,
		}
		return nil
	}

	return json.Unmarshal(data, &s.settings)
}

func (s *SettingsService) Save() error {
	s.mu.RLock()
	data, _ := json.MarshalIndent(s.settings, "", "  ")
	s.mu.RUnlock()

	return os.WriteFile(s.path, data, 0644)
}

func (s *SettingsService) Get() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

func (s *SettingsService) Update(settings Settings) error {
	s.mu.Lock()
	s.settings = settings
	s.mu.Unlock()
	return s.Save()
}
