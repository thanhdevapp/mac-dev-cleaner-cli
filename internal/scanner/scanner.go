// Package scanner provides file system scanning functionality for dev artifacts
package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// Scanner handles scanning for development artifacts
type Scanner struct {
	homeDir  string
	maxDepth int
}

// New creates a new Scanner instance
func New() (*Scanner, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return &Scanner{
		homeDir:  home,
		maxDepth: 3,
	}, nil
}

// SetMaxDepth sets the maximum directory depth for scanning
func (s *Scanner) SetMaxDepth(depth int) {
	s.maxDepth = depth
}

// ScanAll scans all categories based on options
func (s *Scanner) ScanAll(opts types.ScanOptions) ([]types.ScanResult, error) {
	var results []types.ScanResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	if opts.IncludeXcode {
		wg.Add(1)
		go func() {
			defer wg.Done()
			xcodeResults := s.ScanXcode()
			mu.Lock()
			results = append(results, xcodeResults...)
			mu.Unlock()
		}()
	}

	if opts.IncludeAndroid {
		wg.Add(1)
		go func() {
			defer wg.Done()
			androidResults := s.ScanAndroid()
			mu.Lock()
			results = append(results, androidResults...)
			mu.Unlock()
		}()
	}

	if opts.IncludeNode {
		wg.Add(1)
		go func() {
			defer wg.Done()
			nodeResults := s.ScanNode(opts.MaxDepth)
			mu.Lock()
			results = append(results, nodeResults...)
			mu.Unlock()
		}()
	}

	if opts.IncludeFlutter {
		wg.Add(1)
		go func() {
			defer wg.Done()
			flutterResults := s.ScanFlutter(opts.MaxDepth)
			mu.Lock()
			results = append(results, flutterResults...)
			mu.Unlock()
		}()
	}

	if opts.IncludePython {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pythonResults := s.ScanPython(opts.MaxDepth)
			mu.Lock()
			results = append(results, pythonResults...)
			mu.Unlock()
		}()
	}

	if opts.IncludeRust {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rustResults := s.ScanRust(opts.MaxDepth)
			mu.Lock()
			results = append(results, rustResults...)
			mu.Unlock()
		}()
	}

	if opts.IncludeGo {
		wg.Add(1)
		go func() {
			defer wg.Done()
			goResults := s.ScanGo(opts.MaxDepth)
			mu.Lock()
			results = append(results, goResults...)
			mu.Unlock()
		}()
	}

	if opts.IncludeHomebrew {
		wg.Add(1)
		go func() {
			defer wg.Done()
			homebrewResults := s.ScanHomebrew()
			mu.Lock()
			results = append(results, homebrewResults...)
			mu.Unlock()
		}()
	}

	if opts.IncludeDocker {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dockerResults := s.ScanDocker()
			mu.Lock()
			results = append(results, dockerResults...)
			mu.Unlock()
		}()
	}

	if opts.IncludeJava {
		wg.Add(1)
		go func() {
			defer wg.Done()
			javaResults := s.ScanJava(opts.MaxDepth)
			mu.Lock()
			results = append(results, javaResults...)
			mu.Unlock()
		}()
	}

	wg.Wait()
	return results, nil
}

// calculateSize calculates the total size of a directory
func (s *Scanner) calculateSize(path string) (int64, int, error) {
	var size int64
	var count int

	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors, continue
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				size += info.Size()
				count++
			}
		}
		return nil
	})

	return size, count, err
}

// ExpandPath expands ~ to home directory
func (s *Scanner) ExpandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		return filepath.Join(s.homeDir, path[1:])
	}
	return path
}

// PathExists checks if a path exists
func (s *Scanner) PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ScanDirectory scans a single directory lazily and returns TreeNode with children
func (s *Scanner) ScanDirectory(path string, currentDepth int, maxDepth int) (*types.TreeNode, error) {
	// Depth limit check
	if currentDepth >= maxDepth {
		return nil, fmt.Errorf("max depth %d reached", maxDepth)
	}

	// Read directory entries
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	// Calculate total size
	totalSize, fileCount, _ := s.calculateSize(path)

	// Build TreeNode
	node := &types.TreeNode{
		Path:      path,
		Name:      types.GetBasename(path),
		Size:      totalSize,
		IsDir:     true,
		Children:  make([]*types.TreeNode, 0),
		Scanned:   true,
		Depth:     currentDepth,
		FileCount: fileCount,
	}

	// Process children
	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())

		// Skip symlinks to avoid cycles
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}

		isDir := entry.IsDir()
		var childSize int64
		var childFileCount int

		if isDir {
			// For directories, calculate size
			childSize, childFileCount, _ = s.calculateSize(childPath)
		} else {
			// For files, use file size
			childSize = info.Size()
			childFileCount = 1
		}

		child := &types.TreeNode{
			Path:      childPath,
			Name:      entry.Name(),
			Size:      childSize,
			IsDir:     isDir,
			Scanned:   false, // Lazy - not scanned yet
			Depth:     currentDepth + 1,
			FileCount: childFileCount,
		}

		node.AddChild(child)
	}

	return node, nil
}

// ScanResultToTreeNode converts ScanResult to initial TreeNode
func (s *Scanner) ScanResultToTreeNode(result types.ScanResult) (*types.TreeNode, error) {
	node := types.ScanResultToTreeNode(result)

	// Verify path exists
	if !s.PathExists(result.Path) {
		return nil, fmt.Errorf("path does not exist: %s", result.Path)
	}

	return node, nil
}
