package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/thanhdevapp/dev-cleaner/internal/cleaner"
	"github.com/thanhdevapp/dev-cleaner/pkg/types"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type CleanService struct {
	ctx      context.Context
	cleaner  *cleaner.Cleaner
	cleaning bool
	mu       sync.RWMutex
}

func NewCleanService(dryRun bool) (*CleanService, error) {
	c, err := cleaner.New(dryRun)
	if err != nil {
		return nil, err
	}

	return &CleanService{
		cleaner: c,
	}, nil
}

func (c *CleanService) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// Clean deletes items with progress events
func (c *CleanService) Clean(items []types.ScanResult) ([]cleaner.CleanResult, error) {
	// Validate input
	if len(items) == 0 {
		return nil, fmt.Errorf("no items to clean")
	}

	c.mu.Lock()
	if c.cleaning {
		c.mu.Unlock()
		return nil, fmt.Errorf("clean already in progress")
	}
	c.cleaning = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.cleaning = false
		c.mu.Unlock()
	}()

	if c.ctx != nil {
		runtime.EventsEmit(c.ctx, "clean:started", len(items))
	}

	results, err := c.cleaner.Clean(items)
	if err != nil {
		if c.ctx != nil {
			runtime.EventsEmit(c.ctx, "clean:error", err.Error())
		}
		return results, err
	}

	// Calculate freed space
	var freedSpace int64
	successCount := 0
	for _, r := range results {
		if r.Success {
			freedSpace += r.Size
			successCount++
		}
	}

	if c.ctx != nil {
		runtime.EventsEmit(c.ctx, "clean:complete", map[string]interface{}{
			"results":      results,
			"freedSpace":   freedSpace,
			"successCount": successCount,
		})
	}

	return results, nil
}

// IsCleaning returns status
func (c *CleanService) IsCleaning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cleaning
}
