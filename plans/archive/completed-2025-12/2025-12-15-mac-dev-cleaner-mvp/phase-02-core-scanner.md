# Phase 02: Core Scanner Implementation

## Context

| Item | Link |
|------|------|
| Parent Plan | [plan.md](./plan.md) |
| Dependencies | [Phase 01](./phase-01-project-setup.md) |
| Research | [Filesystem Scanning](./research/researcher-filesystem-scanning.md) |

---

## Overview

| Field | Value |
|-------|-------|
| Date | 2025-12-15 |
| Description | Implement directory scanning, size calculation, target definitions |
| Priority | P0 |
| Status | Pending |
| Est. Duration | 4 hours |

---

## Key Insights (from Research)

1. **filepath.WalkDir** > filepath.Walk: 1.5x faster, avoids per-entry stat calls
2. **Concurrent scanning**: 3.5x speedup with 3-8 workers (but MVP can start serial)
3. **Symlink handling**: Skip symlinks to prevent infinite loops
4. **Permission errors**: Handle gracefully, continue scanning
5. **go-humanize**: `humanize.Bytes()` for SI units, `humanize.IBytes()` for IEC

---

## Requirements

- [ ] Define target types (iOS, Android, Node)
- [ ] Scan known paths for each target type
- [ ] Calculate directory sizes efficiently
- [ ] Handle permission errors gracefully
- [ ] Return structured results
- [ ] Support filtering by type

---

## Architecture

```
internal/scanner/
├── scanner.go      # Main Scanner struct and ScanAll()
├── targets.go      # Target definitions and paths
└── size.go         # Size calculation utilities
```

**Data Flow:**
```
Target Definitions → Scanner.ScanAll() → []ScanResult
                          ↓
                   calculateSize(path)
                          ↓
                   humanize.Bytes()
```

---

## Related Code Files

| File | Purpose |
|------|---------|
| `internal/scanner/scanner.go` | Scanner struct, ScanAll, ScanByType |
| `internal/scanner/targets.go` | Target definitions, path expansion |
| `internal/scanner/size.go` | Size calculation with WalkDir |

---

## Implementation Steps

### Step 1: Create types in targets.go

```go
// internal/scanner/targets.go
package scanner

import (
    "os"
    "path/filepath"
)

// TargetType categorizes cleanable directories
type TargetType string

const (
    TypeiOS     TargetType = "ios"
    TypeAndroid TargetType = "android"
    TypeNode    TargetType = "node"
)

// Target defines a cleanable directory pattern
type Target struct {
    Type        TargetType
    Name        string
    Path        string   // Absolute path or pattern
    Description string
}

// ScanResult represents a found cleanable directory
type ScanResult struct {
    Type      TargetType
    Name      string
    Path      string
    Size      int64
    FileCount int
    Error     error
}

// GetHomeDir returns user home directory
func GetHomeDir() string {
    home, _ := os.UserHomeDir()
    return home
}

// GetDefaultTargets returns predefined target paths
func GetDefaultTargets() []Target {
    home := GetHomeDir()

    return []Target{
        // iOS/Xcode
        {
            Type:        TypeiOS,
            Name:        "Xcode DerivedData",
            Path:        filepath.Join(home, "Library/Developer/Xcode/DerivedData"),
            Description: "Build artifacts and indexes",
        },
        {
            Type:        TypeiOS,
            Name:        "Xcode Caches",
            Path:        filepath.Join(home, "Library/Caches/com.apple.dt.Xcode"),
            Description: "Xcode cache files",
        },

        // Android
        {
            Type:        TypeAndroid,
            Name:        "Gradle Caches",
            Path:        filepath.Join(home, ".gradle/caches"),
            Description: "Gradle dependency cache",
        },
        {
            Type:        TypeAndroid,
            Name:        "Gradle Wrapper",
            Path:        filepath.Join(home, ".gradle/wrapper"),
            Description: "Gradle wrapper distributions",
        },

        // Node (handled separately via recursive scan)
        // node_modules requires special pattern matching
    }
}
```

### Step 2: Create size calculation in size.go

```go
// internal/scanner/size.go
package scanner

import (
    "io/fs"
    "os"
    "path/filepath"
)

// CalculateSize returns total size of directory in bytes
func CalculateSize(path string) (size int64, fileCount int, err error) {
    err = filepath.WalkDir(path, func(p string, d fs.DirEntry, walkErr error) error {
        // Handle errors (permissions, etc)
        if walkErr != nil {
            if os.IsPermission(walkErr) {
                return nil // Skip, continue
            }
            return nil // Skip other errors too
        }

        // Skip symlinks
        if d.Type()&os.ModeSymlink != 0 {
            return nil
        }

        // Count files and sum sizes
        if !d.IsDir() {
            info, err := d.Info()
            if err == nil {
                size += info.Size()
                fileCount++
            }
        }

        return nil
    })

    return size, fileCount, err
}

// DirectoryExists checks if path exists and is a directory
func DirectoryExists(path string) bool {
    info, err := os.Stat(path)
    if err != nil {
        return false
    }
    return info.IsDir()
}
```

### Step 3: Create main scanner in scanner.go

```go
// internal/scanner/scanner.go
package scanner

import (
    "io/fs"
    "os"
    "path/filepath"
    "strings"
)

// Scanner handles directory scanning operations
type Scanner struct {
    MaxNodeDepth int // Max depth for node_modules search
}

// NewScanner creates scanner with defaults
func NewScanner() *Scanner {
    return &Scanner{
        MaxNodeDepth: 3,
    }
}

// ScanAll scans all target types
func (s *Scanner) ScanAll() []ScanResult {
    var results []ScanResult

    // Scan predefined targets
    for _, target := range GetDefaultTargets() {
        if DirectoryExists(target.Path) {
            size, count, err := CalculateSize(target.Path)
            results = append(results, ScanResult{
                Type:      target.Type,
                Name:      target.Name,
                Path:      target.Path,
                Size:      size,
                FileCount: count,
                Error:     err,
            })
        }
    }

    // Scan for node_modules
    nodeResults := s.scanNodeModules()
    results = append(results, nodeResults...)

    return results
}

// ScanByType scans only specified type
func (s *Scanner) ScanByType(targetType TargetType) []ScanResult {
    var results []ScanResult

    if targetType == TypeNode {
        return s.scanNodeModules()
    }

    for _, target := range GetDefaultTargets() {
        if target.Type == targetType && DirectoryExists(target.Path) {
            size, count, err := CalculateSize(target.Path)
            results = append(results, ScanResult{
                Type:      target.Type,
                Name:      target.Name,
                Path:      target.Path,
                Size:      size,
                FileCount: count,
                Error:     err,
            })
        }
    }

    return results
}

// scanNodeModules finds node_modules directories
func (s *Scanner) scanNodeModules() []ScanResult {
    var results []ScanResult
    home := GetHomeDir()

    // Common project locations
    searchRoots := []string{
        filepath.Join(home, "Projects"),
        filepath.Join(home, "Developer"),
        filepath.Join(home, "Code"),
        filepath.Join(home, "Documents"),
    }

    for _, root := range searchRoots {
        if !DirectoryExists(root) {
            continue
        }

        found := s.findNodeModules(root)
        results = append(results, found...)
    }

    return results
}

// findNodeModules searches for node_modules within depth limit
func (s *Scanner) findNodeModules(root string) []ScanResult {
    var results []ScanResult
    rootDepth := strings.Count(root, string(os.PathSeparator))

    filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return nil // Skip errors
        }

        // Check depth
        currentDepth := strings.Count(path, string(os.PathSeparator)) - rootDepth
        if currentDepth > s.MaxNodeDepth {
            if d.IsDir() {
                return filepath.SkipDir
            }
            return nil
        }

        // Found node_modules
        if d.IsDir() && d.Name() == "node_modules" {
            size, count, _ := CalculateSize(path)

            // Get parent project name
            parentDir := filepath.Dir(path)
            projectName := filepath.Base(parentDir)

            results = append(results, ScanResult{
                Type:      TypeNode,
                Name:      "node_modules (" + projectName + ")",
                Path:      path,
                Size:      size,
                FileCount: count,
            })

            return filepath.SkipDir // Don't recurse into node_modules
        }

        return nil
    })

    return results
}

// TotalSize returns sum of all result sizes
func TotalSize(results []ScanResult) int64 {
    var total int64
    for _, r := range results {
        total += r.Size
    }
    return total
}
```

### Step 4: Create unit tests

```go
// internal/scanner/scanner_test.go
package scanner

import (
    "os"
    "path/filepath"
    "testing"
)

func TestCalculateSize(t *testing.T) {
    // Create temp directory with known size
    tmpDir := t.TempDir()

    // Create test file (100 bytes)
    testFile := filepath.Join(tmpDir, "test.txt")
    data := make([]byte, 100)
    os.WriteFile(testFile, data, 0644)

    size, count, err := CalculateSize(tmpDir)

    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if size != 100 {
        t.Errorf("Expected size 100, got %d", size)
    }
    if count != 1 {
        t.Errorf("Expected count 1, got %d", count)
    }
}

func TestDirectoryExists(t *testing.T) {
    tmpDir := t.TempDir()

    if !DirectoryExists(tmpDir) {
        t.Error("Expected directory to exist")
    }

    if DirectoryExists("/nonexistent/path") {
        t.Error("Expected directory to not exist")
    }
}

func TestGetDefaultTargets(t *testing.T) {
    targets := GetDefaultTargets()

    if len(targets) == 0 {
        t.Error("Expected at least one target")
    }

    // Verify home path expansion
    home := GetHomeDir()
    for _, target := range targets {
        if !filepath.IsAbs(target.Path) {
            t.Errorf("Expected absolute path, got %s", target.Path)
        }
        if !strings.HasPrefix(target.Path, home) {
            t.Errorf("Expected path under home, got %s", target.Path)
        }
    }
}

func TestScannerNew(t *testing.T) {
    s := NewScanner()

    if s.MaxNodeDepth != 3 {
        t.Errorf("Expected MaxNodeDepth 3, got %d", s.MaxNodeDepth)
    }
}
```

### Step 5: Wire up and verify

```bash
# Run tests
go test ./internal/scanner/...

# Build to verify compilation
go build ./...
```

---

## Todo List

- [ ] Create internal/scanner/targets.go with types
- [ ] Create internal/scanner/size.go with CalculateSize
- [ ] Create internal/scanner/scanner.go with Scanner struct
- [ ] Implement ScanAll() method
- [ ] Implement ScanByType() method
- [ ] Implement node_modules scanning with depth limit
- [ ] Add unit tests
- [ ] Run tests, verify all pass

---

## Success Criteria

| Criteria | Metric |
|----------|--------|
| Tests pass | `go test ./internal/scanner/...` exits 0 |
| iOS targets defined | DerivedData, Xcode Caches |
| Android targets defined | Gradle caches, wrapper |
| Node scanning works | Finds node_modules with depth limit |
| Size calculation accurate | Matches `du -sh` output |
| Permission errors handled | No panics on restricted dirs |

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Slow scan on large dirs | Medium | Medium | Add timeout, progress feedback |
| Permission denied errors | Low | High | Handle gracefully, skip |
| node_modules too deep | Medium | Medium | Enforce depth limit |
| Symlink loops | High | Low | Skip symlinks explicitly |

---

## Security Considerations

- Never follow symlinks (prevents traversal attacks)
- Skip permission-denied gracefully
- Only scan under user home directory
- No execution of scanned files

---

## Next Steps

After Phase 02 complete:
1. Proceed to [Phase 03: CLI Commands](./phase-03-cli-commands.md)
2. Wire scanner to scan/clean commands
3. Add output formatting
