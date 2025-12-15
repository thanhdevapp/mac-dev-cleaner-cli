// Package scanner provides file system scanning functionality for dev artifacts
package scanner

import (
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
