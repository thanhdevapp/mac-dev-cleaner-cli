package services

import (
	"context"
	"sync"

	"github.com/thanhdevapp/dev-cleaner/internal/scanner"
	"github.com/thanhdevapp/dev-cleaner/pkg/types"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type TreeService struct {
	ctx     context.Context
	scanner *scanner.Scanner
	cache   map[string]*types.TreeNode
	mu      sync.RWMutex
}

func NewTreeService() (*TreeService, error) {
	s, err := scanner.New()
	if err != nil {
		return nil, err
	}

	return &TreeService{
		scanner: s,
		cache:   make(map[string]*types.TreeNode),
	}, nil
}

func (t *TreeService) SetContext(ctx context.Context) {
	t.ctx = ctx
}

// GetTreeNode lazily scans directory
func (t *TreeService) GetTreeNode(path string, depth int) (*types.TreeNode, error) {
	// Check cache
	t.mu.RLock()
	if node, exists := t.cache[path]; exists && node.Scanned {
		t.mu.RUnlock()
		return node, nil
	}
	t.mu.RUnlock()

	// Scan directory
	node, err := t.scanner.ScanDirectory(path, depth, 5)
	if err != nil {
		return nil, err
	}

	// Cache node
	t.mu.Lock()
	t.cache[path] = node
	t.mu.Unlock()

	// Emit event
	if t.ctx != nil {
		runtime.EventsEmit(t.ctx, "tree:updated", node)
	}
	return node, nil
}

// ClearCache clears all cached nodes
func (t *TreeService) ClearCache() {
	t.mu.Lock()
	t.cache = make(map[string]*types.TreeNode)
	t.mu.Unlock()

	if t.ctx != nil {
		runtime.EventsEmit(t.ctx, "tree:cleared")
	}
}
