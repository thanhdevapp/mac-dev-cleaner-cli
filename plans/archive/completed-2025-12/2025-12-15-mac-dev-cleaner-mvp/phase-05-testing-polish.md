# Phase 05: Testing & Polish

## Context

| Item | Link |
|------|------|
| Parent Plan | [plan.md](./plan.md) |
| Dependencies | [Phase 01](./phase-01-project-setup.md) - [Phase 04](./phase-04-safety-confirmation.md) |
| Research | All research documents |

---

## Overview

| Field | Value |
|-------|-------|
| Date | 2025-12-15 |
| Description | Add comprehensive tests, polish output, create documentation |
| Priority | P0 |
| Status | Pending |
| Est. Duration | 4 hours |

---

## Key Insights

1. **Test coverage**: Unit tests for scanner, cleaner, safety; integration tests for CLI
2. **Table-driven tests**: Go idiom for testing multiple cases efficiently
3. **Test helpers**: Use t.TempDir() for filesystem tests
4. **Linting**: golangci-lint for code quality
5. **Documentation**: README with installation, usage, examples

---

## Requirements

- [ ] Unit tests for all packages (>80% coverage)
- [ ] Integration tests for CLI commands
- [ ] golangci-lint passes
- [ ] README.md with usage examples
- [ ] --help text polished
- [ ] Build script for local testing

---

## Architecture

```
mac-dev-cleaner/
├── internal/
│   ├── scanner/
│   │   ├── scanner.go
│   │   └── scanner_test.go      # Unit tests
│   ├── cleaner/
│   │   ├── cleaner.go
│   │   ├── safety.go
│   │   ├── safety_test.go       # Unit tests
│   │   └── cleaner_test.go      # Unit tests
│   └── ui/
│       ├── formatter.go
│       └── formatter_test.go    # Unit tests
├── cmd/
│   └── *.go
├── Makefile                     # Build/test commands
├── README.md                    # Documentation
└── .golangci.yml               # Linter config
```

---

## Related Code Files

| File | Purpose |
|------|---------|
| `*_test.go` | Test files |
| `Makefile` | Build automation |
| `README.md` | User documentation |
| `.golangci.yml` | Linter configuration |

---

## Implementation Steps

### Step 1: Create comprehensive scanner tests

```go
// internal/scanner/scanner_test.go
package scanner

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestCalculateSize(t *testing.T) {
    tests := []struct {
        name      string
        setup     func(dir string) // Create test files
        wantSize  int64
        wantCount int
    }{
        {
            name: "empty directory",
            setup: func(dir string) {
                // Empty
            },
            wantSize:  0,
            wantCount: 0,
        },
        {
            name: "single file",
            setup: func(dir string) {
                os.WriteFile(filepath.Join(dir, "test.txt"), make([]byte, 100), 0644)
            },
            wantSize:  100,
            wantCount: 1,
        },
        {
            name: "nested directories",
            setup: func(dir string) {
                subdir := filepath.Join(dir, "sub", "nested")
                os.MkdirAll(subdir, 0755)
                os.WriteFile(filepath.Join(dir, "a.txt"), make([]byte, 50), 0644)
                os.WriteFile(filepath.Join(subdir, "b.txt"), make([]byte, 150), 0644)
            },
            wantSize:  200,
            wantCount: 2,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            dir := t.TempDir()
            tt.setup(dir)

            size, count, err := CalculateSize(dir)

            if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
            if size != tt.wantSize {
                t.Errorf("size = %d, want %d", size, tt.wantSize)
            }
            if count != tt.wantCount {
                t.Errorf("count = %d, want %d", count, tt.wantCount)
            }
        })
    }
}

func TestDirectoryExists(t *testing.T) {
    tests := []struct {
        name string
        path string
        want bool
    }{
        {"temp dir exists", t.TempDir(), true},
        {"nonexistent", "/nonexistent/path/12345", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := DirectoryExists(tt.path); got != tt.want {
                t.Errorf("DirectoryExists(%s) = %v, want %v", tt.path, got, tt.want)
            }
        })
    }
}

func TestGetDefaultTargets(t *testing.T) {
    targets := GetDefaultTargets()

    if len(targets) == 0 {
        t.Fatal("expected at least one target")
    }

    home := GetHomeDir()
    for _, target := range targets {
        // All paths should be absolute
        if !filepath.IsAbs(target.Path) {
            t.Errorf("path not absolute: %s", target.Path)
        }
        // All paths should be under home
        if !strings.HasPrefix(target.Path, home) {
            t.Errorf("path not under home: %s", target.Path)
        }
        // Type should be valid
        switch target.Type {
        case TypeiOS, TypeAndroid, TypeNode:
            // OK
        default:
            t.Errorf("invalid target type: %s", target.Type)
        }
    }
}

func TestTotalSize(t *testing.T) {
    results := []ScanResult{
        {Size: 100},
        {Size: 200},
        {Size: 300},
    }

    total := TotalSize(results)
    if total != 600 {
        t.Errorf("TotalSize = %d, want 600", total)
    }
}

func TestScanner_ScanByType(t *testing.T) {
    s := NewScanner()

    // Should not panic even if directories don't exist
    results := s.ScanByType(TypeiOS)
    // Results may be empty, but should not error
    _ = results
}
```

### Step 2: Create cleaner tests

```go
// internal/cleaner/cleaner_test.go
package cleaner

import (
    "os"
    "path/filepath"
    "testing"
)

func TestNewCleaner(t *testing.T) {
    c, err := NewCleaner(true, false)
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
    defer c.Close()

    if !c.DryRun {
        t.Error("expected DryRun to be true")
    }
}

func TestCleaner_Clean_DryRun(t *testing.T) {
    // Create temp directory
    dir := t.TempDir()
    testDir := filepath.Join(dir, "node_modules")
    os.MkdirAll(testDir, 0755)
    os.WriteFile(filepath.Join(testDir, "test.txt"), []byte("test"), 0644)

    // Need to make it pass validation - create under simulated structure
    home, _ := os.UserHomeDir()
    safeDir := filepath.Join(home, "Library/Developer/test_cleaner_dry")
    os.MkdirAll(safeDir, 0755)
    defer os.RemoveAll(safeDir)

    c, _ := NewCleaner(true, false)
    defer c.Close()

    result := c.Clean(safeDir, 100)

    // Dry run should succeed without deleting
    if !result.Success {
        t.Errorf("dry-run should succeed: %v", result.Error)
    }

    // Directory should still exist
    if _, err := os.Stat(safeDir); os.IsNotExist(err) {
        t.Error("directory should not be deleted in dry-run")
    }
}

func TestCleaner_Clean_ActualDelete(t *testing.T) {
    home, _ := os.UserHomeDir()
    safeDir := filepath.Join(home, "Library/Developer/test_cleaner_delete")
    os.MkdirAll(safeDir, 0755)

    c, _ := NewCleaner(false, false) // Not dry-run
    defer c.Close()

    result := c.Clean(safeDir, 100)

    if !result.Success {
        t.Errorf("delete should succeed: %v", result.Error)
    }

    // Directory should be deleted
    if _, err := os.Stat(safeDir); !os.IsNotExist(err) {
        t.Error("directory should be deleted")
        os.RemoveAll(safeDir) // Cleanup
    }
}

func TestSummarizeResults(t *testing.T) {
    results := []CleanResult{
        {Success: true, Size: 100},
        {Success: true, Size: 200},
        {Success: false, Size: 50},
    }

    success, failed, total := SummarizeResults(results)

    if success != 2 {
        t.Errorf("success = %d, want 2", success)
    }
    if failed != 1 {
        t.Errorf("failed = %d, want 1", failed)
    }
    if total != 300 {
        t.Errorf("total = %d, want 300", total)
    }
}
```

### Step 3: Create formatter tests

```go
// internal/ui/formatter_test.go
package ui

import (
    "bytes"
    "strings"
    "testing"

    "github.com/thanhdevapp/dev-cleaner/internal/scanner"
)

func TestFormatter_PrintResults_Empty(t *testing.T) {
    var buf bytes.Buffer
    f := &Formatter{Out: &buf}

    f.PrintResults(nil)

    output := buf.String()
    if !strings.Contains(output, "No cleanable directories found") {
        t.Errorf("expected empty message, got: %s", output)
    }
}

func TestFormatter_PrintResults_WithData(t *testing.T) {
    var buf bytes.Buffer
    f := &Formatter{Out: &buf}

    results := []scanner.ScanResult{
        {Name: "Test Item", Path: "/path/to/test", Size: 1024 * 1024 * 100}, // 100MB
    }

    f.PrintResults(results)

    output := buf.String()
    if !strings.Contains(output, "Test Item") {
        t.Errorf("expected item name in output: %s", output)
    }
    if !strings.Contains(output, "MB") {
        t.Errorf("expected size formatting in output: %s", output)
    }
}

func TestFormatSize(t *testing.T) {
    tests := []struct {
        bytes int64
        want  string
    }{
        {0, "0 B"},
        {1024, "1.0 kB"},
        {1024 * 1024, "1.0 MB"},
        {1024 * 1024 * 1024, "1.1 GB"}, // go-humanize uses SI
    }

    for _, tt := range tests {
        got := FormatSize(tt.bytes)
        if got != tt.want {
            t.Errorf("FormatSize(%d) = %s, want %s", tt.bytes, got, tt.want)
        }
    }
}

func TestTruncate(t *testing.T) {
    tests := []struct {
        input  string
        maxLen int
        want   string
    }{
        {"short", 10, "short"},
        {"this is a long string", 10, "this is..."},
        {"exact", 5, "exact"},
    }

    for _, tt := range tests {
        got := truncate(tt.input, tt.maxLen)
        if got != tt.want {
            t.Errorf("truncate(%s, %d) = %s, want %s", tt.input, tt.maxLen, got, tt.want)
        }
    }
}
```

### Step 4: Create Makefile

```makefile
# Makefile
.PHONY: build test lint clean install

# Variables
BINARY_NAME=dev-cleaner
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build
build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME) .

# Install locally
install: build
	cp $(BINARY_NAME) /usr/local/bin/

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Run the tool
run:
	go run . $(ARGS)

# All checks before commit
check: fmt lint test
	@echo "All checks passed!"
```

### Step 5: Create golangci-lint config

```yaml
# .golangci.yml
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - misspell

linters-settings:
  errcheck:
    check-blank: true
  misspell:
    locale: US

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
```

### Step 6: Create README.md

```markdown
# Mac Dev Cleaner

A CLI tool to clean development artifacts on macOS and free up disk space.

## Features

- Scan Xcode DerivedData and caches
- Scan Gradle caches and wrappers
- Find and clean node_modules directories
- Safe dry-run mode by default
- Human-readable size display
- Audit logging

## Installation

### Homebrew (Recommended)

```bash
brew tap thanhdevapp/tools
brew install dev-cleaner
```

### From Source

```bash
go install github.com/thanhdevapp/dev-cleaner@latest
```

### Binary Download

Download from [Releases](https://github.com/thanhdevapp/dev-cleaner/releases).

## Usage

### Scan for Cleanable Items

```bash
# Scan all types
dev-cleaner scan

# Scan specific types
dev-cleaner scan --ios        # Xcode artifacts only
dev-cleaner scan --android    # Gradle caches only
dev-cleaner scan --node       # node_modules only
```

### Clean Items

```bash
# Interactive selection (dry-run by default)
dev-cleaner clean

# Clean specific types
dev-cleaner clean --ios

# Actually delete (requires confirmation)
dev-cleaner clean --confirm
dev-cleaner clean --ios --confirm
```

### Options

| Flag | Short | Description |
|------|-------|-------------|
| `--ios` | `-i` | Target iOS/Xcode artifacts |
| `--android` | `-a` | Target Android/Gradle artifacts |
| `--node` | `-n` | Target node_modules |
| `--confirm` | | Actually delete (not dry-run) |
| `--verbose` | `-v` | Verbose output |
| `--help` | `-h` | Show help |

## Target Directories

### iOS/Xcode
- `~/Library/Developer/Xcode/DerivedData/`
- `~/Library/Caches/com.apple.dt.Xcode/`

### Android
- `~/.gradle/caches/`
- `~/.gradle/wrapper/`

### Node.js
- `*/node_modules/` (searched in ~/Projects, ~/Developer, ~/Code, ~/Documents)

## Safety

- **Dry-run by default**: Use `--confirm` to actually delete
- **Path validation**: Only deletes known-safe directories
- **Confirmation prompt**: Type "yes" to confirm deletion
- **Audit logging**: All operations logged to `~/.dev-cleaner.log`

## License

MIT
```

### Step 7: Run all checks

```bash
# Install linter if needed
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run all checks
make check

# Or individually
make fmt
make lint
make test
make test-coverage
```

---

## Todo List

- [ ] Create scanner_test.go with comprehensive tests
- [ ] Create cleaner_test.go
- [ ] Create safety_test.go (expand from Phase 4)
- [ ] Create formatter_test.go
- [ ] Create Makefile
- [ ] Create .golangci.yml
- [ ] Create README.md
- [ ] Run `make check` - fix any issues
- [ ] Verify test coverage >80%
- [ ] Manual end-to-end testing

---

## Success Criteria

| Criteria | Metric |
|----------|--------|
| Tests pass | `go test ./...` exits 0 |
| Coverage >80% | `go test -cover` |
| Lint passes | `golangci-lint run` exits 0 |
| Build succeeds | `go build` exits 0 |
| Help readable | `--help` shows clear usage |
| README complete | Installation, usage, examples |

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Test flakiness | Low | Medium | Use t.TempDir(), isolate tests |
| Missing edge cases | Medium | Medium | Add more test cases over time |
| Coverage gaps | Low | Low | Focus on critical paths |

---

## Security Considerations

- Tests should not touch real user data
- Use temporary directories for filesystem tests
- Clean up test artifacts
- No secrets in test code

---

## Next Steps

After Phase 05 complete:
1. **MVP Complete!** Ready for local use
2. Create GitHub repository
3. Plan Phase 2: Enhanced UX with TUI
4. Plan Phase 3: Distribution with GoReleaser/Homebrew
