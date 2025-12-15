// Package cleaner provides safe file deletion functionality
package cleaner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// Cleaner handles safe deletion of directories
type Cleaner struct {
	dryRun  bool
	logger  *log.Logger
	logFile *os.File
}

// New creates a new Cleaner instance
func New(dryRun bool) (*Cleaner, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	logPath := filepath.Join(home, ".dev-cleaner.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	logger := log.New(logFile, "", log.LstdFlags)

	return &Cleaner{
		dryRun:  dryRun,
		logger:  logger,
		logFile: logFile,
	}, nil
}

// Close closes the log file
func (c *Cleaner) Close() error {
	if c.logFile != nil {
		return c.logFile.Close()
	}
	return nil
}

// SetDryRun sets the dry-run mode
func (c *Cleaner) SetDryRun(dryRun bool) {
	c.dryRun = dryRun
}

// Logger returns the cleaner's logger instance
func (c *Cleaner) Logger() *log.Logger {
	return c.logger
}

// CleanResult represents the result of a clean operation
type CleanResult struct {
	Path        string
	Size        int64
	Success     bool
	Error       error
	WasDryRun   bool
}

// Clean deletes the specified paths after validation
func (c *Cleaner) Clean(results []types.ScanResult) ([]CleanResult, error) {
	var cleanResults []CleanResult

	for _, result := range results {
		// Validate path safety
		if err := ValidatePath(result.Path); err != nil {
			cleanResults = append(cleanResults, CleanResult{
				Path:    result.Path,
				Size:    result.Size,
				Success: false,
				Error:   err,
			})
			continue
		}

		if c.dryRun {
			c.logger.Printf("[DRY-RUN] Would delete: %s (%.2f MB)\n", result.Path, float64(result.Size)/(1024*1024))
			cleanResults = append(cleanResults, CleanResult{
				Path:      result.Path,
				Size:      result.Size,
				Success:   true,
				WasDryRun: true,
			})
		} else {
			c.logger.Printf("[DELETE] Removing: %s (%.2f MB)\n", result.Path, float64(result.Size)/(1024*1024))
			
			if err := os.RemoveAll(result.Path); err != nil {
				c.logger.Printf("[ERROR] Failed to delete %s: %v\n", result.Path, err)
				cleanResults = append(cleanResults, CleanResult{
					Path:    result.Path,
					Size:    result.Size,
					Success: false,
					Error:   err,
				})
			} else {
				c.logger.Printf("[SUCCESS] Deleted: %s at %s\n", result.Path, time.Now().Format(time.RFC3339))
				cleanResults = append(cleanResults, CleanResult{
					Path:    result.Path,
					Size:    result.Size,
					Success: true,
				})
			}
		}
	}

	return cleanResults, nil
}

// TotalSize calculates total size from results
func TotalSize(results []types.ScanResult) int64 {
	var total int64
	for _, r := range results {
		total += r.Size
	}
	return total
}

// FormatSize formats bytes to human-readable format
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
