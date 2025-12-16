package main

import (
	"context"
	"log"

	"github.com/thanhdevapp/dev-cleaner/internal/cleaner"
	"github.com/thanhdevapp/dev-cleaner/internal/services"
	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

type App struct {
	ctx             context.Context
	scanService     *services.ScanService
	treeService     *services.TreeService
	cleanService    *services.CleanService
	settingsService *services.SettingsService
}

func NewApp() *App {
	a := &App{}
	var err error

	// Initialize scan service
	a.scanService, err = services.NewScanService()
	if err != nil {
		log.Printf("❌ Failed to create ScanService: %v", err)
	} else {
		log.Println("✅ ScanService initialized")
	}

	// Initialize tree service
	a.treeService, err = services.NewTreeService()
	if err != nil {
		log.Printf("❌ Failed to create TreeService: %v", err)
	} else {
		log.Println("✅ TreeService initialized")
	}

	// Initialize clean service
	a.cleanService, err = services.NewCleanService(false)
	if err != nil {
		log.Printf("❌ Failed to create CleanService: %v", err)
	} else {
		log.Println("✅ CleanService initialized")
	}

	// Initialize settings service
	a.settingsService = services.NewSettingsService()
	log.Println("✅ SettingsService initialized")

	log.Println("🎉 All services initialized successfully!")
	return a
}

func (a *App) startup(ctx context.Context) {
	log.Println("🚀 OnStartup called - injecting context...")
	a.ctx = ctx

	if a.scanService != nil {
		a.scanService.SetContext(ctx)
	}
	if a.treeService != nil {
		a.treeService.SetContext(ctx)
	}
	if a.cleanService != nil {
		a.cleanService.SetContext(ctx)
	}
}

func (a *App) shutdown(ctx context.Context) {
	log.Println("👋 OnShutdown called")
}

// ScanService methods exposed to frontend
func (a *App) Scan(opts types.ScanOptions) error {
	if a.scanService == nil {
		return nil
	}
	return a.scanService.Scan(opts)
}

func (a *App) GetScanResults() []types.ScanResult {
	if a.scanService == nil {
		return []types.ScanResult{}
	}
	return a.scanService.GetResults()
}

func (a *App) IsScanning() bool {
	if a.scanService == nil {
		return false
	}
	return a.scanService.IsScanning()
}

// TreeService methods exposed to frontend
func (a *App) GetTreeNode(path string, depth int) (*types.TreeNode, error) {
	if a.treeService == nil {
		return nil, nil
	}
	return a.treeService.GetTreeNode(path, depth)
}

func (a *App) ClearTreeCache() {
	if a.treeService != nil {
		a.treeService.ClearCache()
	}
}

// CleanService methods exposed to frontend
func (a *App) Clean(items []types.ScanResult) ([]cleaner.CleanResult, error) {
	if a.cleanService == nil {
		return []cleaner.CleanResult{}, nil
	}
	return a.cleanService.Clean(items)
}

func (a *App) IsCleaning() bool {
	if a.cleanService == nil {
		return false
	}
	return a.cleanService.IsCleaning()
}

// SettingsService methods exposed to frontend
func (a *App) GetSettings() services.Settings {
	if a.settingsService == nil {
		return services.Settings{}
	}
	return a.settingsService.Get()
}

func (a *App) UpdateSettings(settings services.Settings) error {
	if a.settingsService == nil {
		return nil
	}
	return a.settingsService.Update(settings)
}
