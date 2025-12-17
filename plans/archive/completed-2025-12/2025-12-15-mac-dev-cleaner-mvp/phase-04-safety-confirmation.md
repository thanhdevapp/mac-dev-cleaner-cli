# Phase 04: Safety Validation & Confirmation

## Context

| Item | Link |
|------|------|
| Parent Plan | [plan.md](./plan.md) |
| Dependencies | [Phase 01](./phase-01-project-setup.md), [Phase 02](./phase-02-core-scanner.md), [Phase 03](./phase-03-cli-commands.md) |
| Research | [Brainstorm Report](../reports/brainstorm-2025-12-15-mac-dev-cleaner-plan.md) |

---

## Overview

| Field | Value |
|-------|-------|
| Date | 2025-12-15 |
| Description | Implement safe deletion with path validation, logging, and error handling |
| Priority | P0 (Critical for safety) |
| Status | Pending |
| Est. Duration | 3 hours |

---

## Key Insights (from Research)

1. **Never delete system paths**: Whitelist approach, only delete under home directory
2. **Dry-run default**: Always preview before actual deletion
3. **Logging**: Log all operations for recovery reference
4. **Confirmation flow**: Multiple checkpoints before destructive action
5. **Error handling**: Partial failures should not crash, continue with remaining

---

## Requirements

- [ ] Path validation against dangerous locations
- [ ] Only allow deletion under user home directory
- [ ] Log all deletion operations
- [ ] Graceful error handling (continue on partial failure)
- [ ] Clear confirmation prompts
- [ ] Ability to cancel at any point

---

## Architecture

```
internal/cleaner/
├── cleaner.go    # Cleaner struct, Delete operations
└── safety.go     # Path validation, dangerous path checks
```

**Safety Flow:**
```
User Selection → ValidatePath() → ConfirmDelete() → Delete() → Log()
                     ↓                                  ↓
              REJECT if unsafe              Continue on error
```

---

## Related Code Files

| File | Purpose |
|------|---------|
| `internal/cleaner/safety.go` | Path validation, dangerous paths list |
| `internal/cleaner/cleaner.go` | Delete operations with logging |

---

## Implementation Steps

### Step 1: Create safety validation in safety.go

```go
// internal/cleaner/safety.go
package cleaner

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// DangerousPaths that should NEVER be deleted
var DangerousPaths = []string{
    "/",
    "/System",
    "/Library",
    "/Applications",
    "/Users",
    "/bin",
    "/sbin",
    "/usr",
    "/var",
    "/private",
    "/etc",
    "/tmp",
    "/cores",
}

// SafePathPrefixes - paths must start with one of these
var SafePathPrefixes = []string{
    "Library/Developer",
    "Library/Caches",
    ".gradle",
    ".android",
    ".npm",
    ".pnpm-store",
    ".yarn",
    "node_modules",
}

// PathValidationError represents a validation failure
type PathValidationError struct {
    Path   string
    Reason string
}

func (e *PathValidationError) Error() string {
    return fmt.Sprintf("unsafe path %s: %s", e.Path, e.Reason)
}

// ValidatePath checks if path is safe to delete
func ValidatePath(path string) error {
    // Normalize path
    absPath, err := filepath.Abs(path)
    if err != nil {
        return &PathValidationError{Path: path, Reason: "cannot resolve path"}
    }

    // Clean path (removes .., etc)
    absPath = filepath.Clean(absPath)

    // Check against dangerous paths
    for _, dangerous := range DangerousPaths {
        if absPath == dangerous {
            return &PathValidationError{Path: absPath, Reason: "system directory"}
        }
        // Also check if trying to delete parent of dangerous path
        if strings.HasPrefix(dangerous, absPath+"/") {
            return &PathValidationError{Path: absPath, Reason: "contains system directories"}
        }
    }

    // Must be under home directory
    home, err := os.UserHomeDir()
    if err != nil {
        return &PathValidationError{Path: absPath, Reason: "cannot determine home directory"}
    }

    if !strings.HasPrefix(absPath, home) {
        return &PathValidationError{Path: absPath, Reason: "not under home directory"}
    }

    // Path relative to home must match safe prefixes
    relPath := strings.TrimPrefix(absPath, home+"/")
    isSafe := false
    for _, prefix := range SafePathPrefixes {
        if strings.HasPrefix(relPath, prefix) || strings.Contains(relPath, "/"+prefix) {
            isSafe = true
            break
        }
    }

    if !isSafe {
        return &PathValidationError{Path: absPath, Reason: "not a known safe location"}
    }

    // Path must exist
    info, err := os.Stat(absPath)
    if err != nil {
        if os.IsNotExist(err) {
            return &PathValidationError{Path: absPath, Reason: "path does not exist"}
        }
        return &PathValidationError{Path: absPath, Reason: err.Error()}
    }

    // Must be a directory (we don't delete individual files)
    if !info.IsDir() {
        return &PathValidationError{Path: absPath, Reason: "not a directory"}
    }

    return nil
}

// ValidatePaths checks multiple paths, returns first error
func ValidatePaths(paths []string) error {
    for _, path := range paths {
        if err := ValidatePath(path); err != nil {
            return err
        }
    }
    return nil
}

// IsSafePath returns true if path is safe to delete
func IsSafePath(path string) bool {
    return ValidatePath(path) == nil
}
```

### Step 2: Create cleaner with logging in cleaner.go

```go
// internal/cleaner/cleaner.go
package cleaner

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"
)

// CleanResult represents result of a single clean operation
type CleanResult struct {
    Path    string
    Size    int64
    Success bool
    Error   error
}

// Cleaner handles deletion operations
type Cleaner struct {
    DryRun  bool
    Verbose bool
    Logger  *log.Logger
    logFile *os.File
}

// NewCleaner creates cleaner with logging
func NewCleaner(dryRun, verbose bool) (*Cleaner, error) {
    c := &Cleaner{
        DryRun:  dryRun,
        Verbose: verbose,
    }

    // Setup log file
    if err := c.setupLogger(); err != nil {
        // Non-fatal: continue without logging
        fmt.Fprintf(os.Stderr, "Warning: could not setup logging: %v\n", err)
    }

    return c, nil
}

// setupLogger initializes the log file
func (c *Cleaner) setupLogger() error {
    home, err := os.UserHomeDir()
    if err != nil {
        return err
    }

    logPath := filepath.Join(home, ".dev-cleaner.log")

    // Open log file (append mode)
    f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }

    c.logFile = f
    c.Logger = log.New(f, "", 0) // No prefix, we add timestamp manually

    return nil
}

// Close closes the log file
func (c *Cleaner) Close() {
    if c.logFile != nil {
        c.logFile.Close()
    }
}

// log writes to log file with timestamp
func (c *Cleaner) log(format string, args ...interface{}) {
    if c.Logger == nil {
        return
    }

    timestamp := time.Now().Format("2006-01-02 15:04:05")
    msg := fmt.Sprintf(format, args...)
    c.Logger.Printf("[%s] %s", timestamp, msg)
}

// Clean deletes a single path after validation
func (c *Cleaner) Clean(path string, size int64) CleanResult {
    result := CleanResult{
        Path: path,
        Size: size,
    }

    // Validate path
    if err := ValidatePath(path); err != nil {
        result.Error = err
        c.log("REJECTED %s: %v", path, err)
        return result
    }

    // Dry-run mode
    if c.DryRun {
        c.log("DRY-RUN %s (%d bytes)", path, size)
        result.Success = true
        return result
    }

    // Actually delete
    c.log("DELETE START %s (%d bytes)", path, size)

    if err := os.RemoveAll(path); err != nil {
        result.Error = err
        c.log("DELETE FAILED %s: %v", path, err)
        return result
    }

    c.log("DELETE SUCCESS %s (%d bytes)", path, size)
    result.Success = true
    return result
}

// CleanAll deletes multiple paths, continuing on errors
func (c *Cleaner) CleanAll(paths []string, sizes map[string]int64) []CleanResult {
    var results []CleanResult

    for _, path := range paths {
        size := sizes[path]
        result := c.Clean(path, size)
        results = append(results, result)

        // Print progress
        if c.Verbose || !result.Success {
            if result.Error != nil {
                fmt.Printf("  Error: %s - %v\n", path, result.Error)
            } else if c.DryRun {
                fmt.Printf("  [DRY-RUN] %s\n", path)
            } else {
                fmt.Printf("  Deleted: %s\n", path)
            }
        }
    }

    return results
}

// SummarizeResults returns counts and total size
func SummarizeResults(results []CleanResult) (success, failed int, totalSize int64) {
    for _, r := range results {
        if r.Success {
            success++
            totalSize += r.Size
        } else {
            failed++
        }
    }
    return
}
```

### Step 3: Create unit tests for safety

```go
// internal/cleaner/safety_test.go
package cleaner

import (
    "os"
    "path/filepath"
    "testing"
)

func TestValidatePath_DangerousPaths(t *testing.T) {
    dangerous := []string{
        "/",
        "/System",
        "/Library",
        "/Applications",
        "/usr",
        "/bin",
    }

    for _, path := range dangerous {
        err := ValidatePath(path)
        if err == nil {
            t.Errorf("Expected error for dangerous path %s", path)
        }
    }
}

func TestValidatePath_OutsideHome(t *testing.T) {
    paths := []string{
        "/tmp/test",
        "/var/log",
        "/opt/something",
    }

    for _, path := range paths {
        err := ValidatePath(path)
        if err == nil {
            t.Errorf("Expected error for path outside home: %s", path)
        }
    }
}

func TestValidatePath_SafePaths(t *testing.T) {
    home, _ := os.UserHomeDir()

    // Create temp dirs for testing
    testDirs := []string{
        filepath.Join(home, "Library/Developer/test_derived"),
        filepath.Join(home, ".gradle/test_cache"),
    }

    for _, dir := range testDirs {
        os.MkdirAll(dir, 0755)
        defer os.RemoveAll(dir)

        err := ValidatePath(dir)
        if err != nil {
            t.Errorf("Expected safe path %s to pass: %v", dir, err)
        }
    }
}

func TestValidatePath_NonExistent(t *testing.T) {
    home, _ := os.UserHomeDir()
    path := filepath.Join(home, "Library/Developer/nonexistent_test_12345")

    err := ValidatePath(path)
    if err == nil {
        t.Error("Expected error for non-existent path")
    }
}

func TestIsSafePath(t *testing.T) {
    // Dangerous paths should return false
    if IsSafePath("/System") {
        t.Error("Expected /System to be unsafe")
    }

    // Non-existent paths should return false
    if IsSafePath("/nonexistent") {
        t.Error("Expected non-existent path to be unsafe")
    }
}
```

### Step 4: Update clean command to use cleaner

```go
// Update cmd/clean.go - runClean function
// Replace the deletion loop with:

func runClean(cmd *cobra.Command, args []string) error {
    // ... (scanning code stays the same)

    // Create cleaner
    isDryRun := !confirm
    cleaner, err := cleaner.NewCleaner(isDryRun, verbose)
    if err != nil {
        return err
    }
    defer cleaner.Close()

    // Build sizes map
    sizes := make(map[string]int64)
    var paths []string
    for _, r := range selected {
        paths = append(paths, r.Path)
        sizes[r.Path] = r.Size
    }

    // Perform clean
    fmt.Println("\nCleaning...")
    results := cleaner.CleanAll(paths, sizes)

    // Summarize
    success, failed, totalSize := cleaner.SummarizeResults(results)

    if failed > 0 {
        fmt.Printf("\nWarning: %d items failed\n", failed)
    }

    f.PrintSummary(success, totalSize, isDryRun)

    return nil
}
```

### Step 5: Test safety validation

```bash
# Run unit tests
go test ./internal/cleaner/...

# Manual testing
./dev-cleaner clean
# Try selecting items, verify dry-run works
# Check ~/.dev-cleaner.log for entries
```

---

## Todo List

- [ ] Create internal/cleaner/safety.go
- [ ] Create internal/cleaner/cleaner.go
- [ ] Add dangerous paths list
- [ ] Implement ValidatePath function
- [ ] Implement Cleaner with logging
- [ ] Add unit tests for safety
- [ ] Wire cleaner to clean command
- [ ] Test dry-run mode
- [ ] Test actual deletion (carefully!)
- [ ] Verify log file created

---

## Success Criteria

| Criteria | Metric |
|----------|--------|
| Dangerous paths blocked | /System, /Library, etc rejected |
| Outside home blocked | /tmp, /var, etc rejected |
| Safe paths allowed | DerivedData, .gradle, node_modules |
| Logging works | ~/.dev-cleaner.log created |
| Dry-run prevents deletion | No files deleted |
| Errors don't crash | Continue on partial failure |

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Validation bypass | Critical | Very Low | Multiple validation layers |
| Log file write fails | Low | Low | Continue without logging |
| Partial deletion fails | Medium | Low | Log progress, show errors |
| Path traversal attack | High | Very Low | Clean paths, validate prefixes |

---

## Security Considerations

- **Defense in depth**: Multiple validation checks
- **Whitelist approach**: Only known-safe paths allowed
- **Path normalization**: Clean paths before validation
- **Home directory check**: Must be under $HOME
- **Directory-only**: No individual file deletion
- **Audit logging**: All operations logged

---

## Next Steps

After Phase 04 complete:
1. Proceed to [Phase 05: Testing & Polish](./phase-05-testing-polish.md)
2. Add comprehensive tests
3. Polish output and documentation
