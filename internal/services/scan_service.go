package services

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/thanhdevapp/dev-cleaner/internal/scanner"
	"github.com/thanhdevapp/dev-cleaner/pkg/types"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type ScanService struct {
	ctx      context.Context
	scanner  *scanner.Scanner
	results  []types.ScanResult
	scanning bool
	mu       sync.RWMutex
}

func NewScanService() (*ScanService, error) {
	s, err := scanner.New()
	if err != nil {
		return nil, err
	}

	return &ScanService{
		scanner: s,
		results: make([]types.ScanResult, 0),
	}, nil
}

func (s *ScanService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// Scan performs full scan with events
func (s *ScanService) Scan(opts types.ScanOptions) error {
	s.mu.Lock()
	if s.scanning {
		s.mu.Unlock()
		return fmt.Errorf("scan already in progress")
	}
	s.scanning = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.scanning = false
		s.mu.Unlock()
	}()

	// Emit start event
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "scan:started")
	}

	// Perform scan
	results, err := s.scanner.ScanAll(opts)
	if err != nil {
		fmt.Printf("❌ Scan error: %v\n", err)
		if s.ctx != nil {
			runtime.EventsEmit(s.ctx, "scan:error", err.Error())
		}
		return err
	}

	fmt.Printf("📊 Scan found %d results\n", len(results))

	// Sort by size (largest first) using sort.Slice for O(n log n)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Size > results[j].Size
	})

	// Update results atomically
	s.mu.Lock()
	s.results = results
	s.mu.Unlock()

	// Emit complete event
	fmt.Printf("📡 Emitting scan:complete event with %d results\n", len(results))
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "scan:complete", results)
	}
	return nil
}

// GetResults returns cached results
func (s *ScanService) GetResults() []types.ScanResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.results
}

// IsScanning returns scan status
func (s *ScanService) IsScanning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scanning
}
