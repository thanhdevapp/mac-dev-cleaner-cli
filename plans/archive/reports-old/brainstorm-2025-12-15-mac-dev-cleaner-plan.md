# Mac Dev Cleaner - Development Plan

> **Date:** 2025-12-15
> **Status:** Brainstorming Complete
> **Tech Stack:** Go + GoReleaser + Homebrew

---

## Executive Summary

Building a CLI tool to clean development artifacts on macOS using **Go** for optimal balance of:
- Fast development with simple syntax
- Single binary distribution (no runtime required)
- Built-in cross-compilation
- Strong stdlib for file operations
- Easy Homebrew integration via GoReleaser

---

## Tech Stack Decision (Validated)

### Primary Stack
| Component | Choice | Rationale |
|-----------|--------|-----------|
| **Language** | Go 1.21+ | Fast builds, single binary, excellent stdlib |
| **CLI Framework** | Cobra | Industry standard, used by kubectl/hugo/gh |
| **TUI Library** | Bubble Tea | Modern, composable, best-in-class terminal UI |
| **Config** | Viper | Pairs with Cobra, multi-format support |
| **Release** | GoReleaser | Automates builds, GitHub releases, Homebrew |
| **Distribution** | Homebrew Tap | Best UX for macOS developers |

### Why Go Wins Here
✅ Perfect for file system operations (this tool's core function)
✅ Fast compilation, small binaries (~5-10MB compressed)
✅ Cross-platform ready if needed later
✅ Strong community, mature ecosystem
✅ No runtime dependencies for users

---

## Project Structure

```
mac-dev-cleaner/
├── cmd/
│   └── root.go              # Cobra root command
├── internal/
│   ├── scanner/
│   │   ├── scanner.go       # Core scanning logic
│   │   ├── ios.go           # iOS/Xcode specific
│   │   ├── android.go       # Android specific
│   │   └── node.go          # Node.js specific
│   ├── cleaner/
│   │   ├── cleaner.go       # Delete operations
│   │   └── safety.go        # Validation & confirmation
│   ├── ui/
│   │   ├── tui.go           # Bubble Tea TUI
│   │   └── formatter.go     # Size formatting, output
│   └── config/
│       ├── config.go        # Viper config management
│       └── defaults.go      # Default paths
├── pkg/
│   └── types/
│       └── types.go         # Shared types (CleanTarget, ScanResult)
├── .goreleaser.yaml         # GoReleaser config
├── go.mod
├── go.sum
├── main.go                  # Entry point
└── README.md
```

**Structure Principles:**
- `internal/` = private implementation (cannot be imported)
- `cmd/` = CLI command structure
- `pkg/` = public, reusable types
- Separation of concerns: scan → validate → clean

---

## Implementation Phases

### 🎯 Phase 1: MVP (P0) - Week 1-2

**Goal:** Basic functional CLI with core cleaning capabilities

**Features:**
- [x] Scan predefined directories (iOS, Android, Node)
- [x] List found directories with human-readable sizes
- [x] Interactive selection (simple prompt)
- [x] Dry-run mode (default)
- [x] Confirmation before actual deletion
- [x] Basic error handling

**Commands:**
```bash
dev-cleaner scan                    # Scan all
dev-cleaner scan --ios              # iOS only
dev-cleaner scan --android          # Android only
dev-cleaner scan --node             # Node only
dev-cleaner clean --dry-run         # Preview (default)
dev-cleaner clean --confirm         # Actually delete
```

**Technical Tasks:**
1. Initialize Go module
2. Setup Cobra CLI structure
3. Implement scanner for each type:
   - `~/Library/Developer/Xcode/DerivedData/`
   - `~/.gradle/caches/`
   - `*/node_modules/` (with depth limit)
4. Implement size calculation with `filepath.WalkDir`
5. Build confirmation prompt
6. Implement safe deletion with logging
7. Add dry-run flag logic

**Libraries:**
```go
github.com/spf13/cobra          // CLI framework
github.com/dustin/go-humanize   // Size formatting (5.2GB)
```

**Safety Checks MVP:**
- Never delete without explicit `--confirm` flag
- Validate paths don't contain system directories
- Log all deletions to `~/.dev-cleaner.log`

---

### 🚀 Phase 2: Enhanced UX (P1) - Week 3

**Goal:** Professional TUI with interactive selection

**Features:**
- [x] Bubble Tea TUI with arrow key navigation
- [x] Multi-select with spacebar
- [x] Real-time size calculation
- [x] Progress bars for scan/delete
- [x] Config file support (~/.dev-cleaner.yaml)

**TUI Flow:**
```
┌─────────────────────────────────────────────────┐
│ Mac Dev Cleaner - Select items to clean        │
├─────────────────────────────────────────────────┤
│                                                 │
│ [x] Xcode DerivedData        12.5 GB            │
│ [ ] Xcode Archives           3.2 GB             │
│ [x] Gradle Caches            8.1 GB             │
│ [x] node_modules (15 dirs)   4.7 GB             │
│                                                 │
│ Total Selected: 25.3 GB                         │
│                                                 │
│ ↑/↓: Navigate | Space: Select | Enter: Clean   │
└─────────────────────────────────────────────────┘
```

**Config File Example:**
```yaml
# ~/.dev-cleaner.yaml
paths:
  custom:
    - ~/CustomCache/
exclude:
  - "*/important-project/node_modules"
presets:
  aggressive: true  # Include Archives, build dirs
```

**Technical Tasks:**
1. Integrate Bubble Tea framework
2. Create interactive list model
3. Add checkbox selection
4. Implement progress indicators
5. Add Viper config parsing
6. Support custom paths from config

**Libraries Added:**
```go
github.com/charmbracelet/bubbletea    // TUI framework
github.com/charmbracelet/bubbles      // TUI components
github.com/spf13/viper                // Config management
```

---

### 🎨 Phase 3: Polish & Distribution (P1) - Week 4

**Goal:** Production-ready release with Homebrew distribution

**Features:**
- [x] GoReleaser setup
- [x] Homebrew tap creation
- [x] GitHub Actions CI/CD
- [x] Comprehensive README
- [x] Man page generation

**Distribution Setup:**

**1. GoReleaser Config (.goreleaser.yaml):**
```yaml
project_name: dev-cleaner

before:
  hooks:
    - go mod tidy
    - go test ./...

builds:
  - main: ./main.go
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

brews:
  - repository:
      owner: thanhdevapp
      name: homebrew-tools
    homepage: https://github.com/thanhdevapp/dev-cleaner
    description: "Clean development artifacts on macOS"
    install: |
      bin.install "dev-cleaner"

release:
  github:
    owner: thanhdevapp
    name: dev-cleaner
```

**2. Homebrew Tap Structure:**
```
homebrew-tools/
└── Formula/
    └── dev-cleaner.rb    # Auto-generated by GoReleaser
```

**3. GitHub Actions (.github/workflows/release.yml):**
```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**User Installation Flow:**
```bash
# Option 1: Homebrew (recommended)
brew tap thanhdevapp/tools
brew install dev-cleaner

# Option 2: Direct binary download
curl -sL https://github.com/thanhdevapp/dev-cleaner/releases/download/v1.0.0/dev-cleaner_darwin_arm64.tar.gz | tar xz
sudo mv dev-cleaner /usr/local/bin/
```

---

### 🔮 Phase 4: Future Enhancements (P2)

**Deferred Features:**
- Auto-detect project types in current directory
- Scheduled cleaning (cron integration)
- Export reports (JSON/CSV)
- Whitelist patterns
- GUI app (macOS native or Electron)

---

## Core Implementation Details

### 1. Scanning Logic

**File:** `internal/scanner/scanner.go`

```go
package scanner

import (
    "io/fs"
    "path/filepath"
)

type ScanResult struct {
    Path      string
    Type      string // "xcode", "android", "node"
    Size      int64
    FileCount int
}

type Scanner struct {
    maxDepth int
}

func (s *Scanner) ScanAll() ([]ScanResult, error) {
    var results []ScanResult

    // Scan each category
    xcode := s.scanXcode()
    android := s.scanAndroid()
    node := s.scanNode()

    results = append(results, xcode...)
    results = append(results, android...)
    results = append(results, node...)

    return results, nil
}

func (s *Scanner) calculateSize(path string) (int64, error) {
    var size int64

    err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
        if err != nil {
            return nil // Skip errors, continue
        }
        if !d.IsDir() {
            info, err := d.Info()
            if err == nil {
                size += info.Size()
            }
        }
        return nil
    })

    return size, err
}
```

**Optimization:**
- Use goroutines for parallel scanning
- Implement depth limits for node_modules search
- Cache results for 1 minute to avoid re-scanning

---

### 2. Safety Validation

**File:** `internal/cleaner/safety.go`

```go
package cleaner

import (
    "fmt"
    "strings"
)

var dangerousPaths = []string{
    "/System",
    "/Library/System",
    "/usr/bin",
    "/usr/lib",
    "/bin",
    "/sbin",
}

func ValidatePath(path string) error {
    // Check against dangerous paths
    for _, dangerous := range dangerousPaths {
        if strings.HasPrefix(path, dangerous) {
            return fmt.Errorf("refusing to delete system path: %s", path)
        }
    }

    // Must be in home directory or known safe locations
    if !strings.HasPrefix(path, os.Getenv("HOME")) {
        return fmt.Errorf("path outside home directory: %s", path)
    }

    return nil
}
```

---

### 3. Dry-Run Implementation

**File:** `internal/cleaner/cleaner.go`

```go
package cleaner

import (
    "log"
    "os"
)

type Cleaner struct {
    dryRun bool
    logger *log.Logger
}

func (c *Cleaner) Clean(paths []string) error {
    for _, path := range paths {
        if err := ValidatePath(path); err != nil {
            return err
        }

        if c.dryRun {
            c.logger.Printf("[DRY-RUN] Would delete: %s\n", path)
        } else {
            c.logger.Printf("[DELETE] Removing: %s\n", path)
            if err := os.RemoveAll(path); err != nil {
                return err
            }
        }
    }
    return nil
}
```

---

## Dependencies & Tools

### Required Go Modules

```go
// go.mod
module github.com/thanhdevapp/dev-cleaner

go 1.21

require (
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.2
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/bubbles v0.18.0
    github.com/dustin/go-humanize v1.0.1
)
```

### Development Tools

```bash
# Install Go
brew install go

# Install GoReleaser
brew install goreleaser/tap/goreleaser

# Testing tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

---

## Risk Assessment & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| **Accidental system file deletion** | Critical | Low | Multi-layer validation, dry-run default, whitelist approach |
| **Performance on large directories** | Medium | Medium | Depth limits, concurrent scanning, progress indicators |
| **Cross-platform path differences** | Low | Low | Use `filepath` package, detect OS at runtime |
| **GoReleaser Homebrew tap setup** | Low | Medium | Follow official docs, test locally before release |
| **User unfamiliar with CLI** | Medium | Medium | Comprehensive help text, examples in README |

---

## Success Metrics

### MVP Success Criteria
- ✅ Successfully scans and identifies >90% of common dev artifacts
- ✅ Dry-run mode prevents accidental deletions
- ✅ Binary size < 10MB compressed
- ✅ Scan completes in <5s for ~100 projects
- ✅ Zero system file deletions (100% safety)

### P1 Success Criteria
- ✅ TUI provides intuitive selection experience
- ✅ Homebrew installation works first try
- ✅ Config file allows customization without code changes

---

## Development Timeline

| Phase | Duration | Deliverable |
|-------|----------|-------------|
| **MVP** | 1-2 weeks | Working CLI with basic scan/clean |
| **Enhanced UX** | 1 week | TUI + config support |
| **Distribution** | 1 week | Homebrew tap + releases |
| **Total** | **3-4 weeks** | Production-ready v1.0.0 |

**Assumptions:**
- Part-time development (~10-15 hrs/week)
- Basic Go familiarity (learning as you go)
- No major blockers or scope changes

---

## Getting Started Checklist

### Immediate Next Steps

- [ ] **Day 1: Project Setup**
  - [ ] `mkdir mac-dev-cleaner && cd mac-dev-cleaner`
  - [ ] `go mod init github.com/thanhdevapp/dev-cleaner`
  - [ ] Create project structure (cmd/, internal/, pkg/)
  - [ ] Install Cobra: `go get github.com/spf13/cobra`

- [ ] **Day 2-3: Core Scanner**
  - [ ] Implement `internal/scanner/scanner.go`
  - [ ] Add iOS/Xcode path scanning
  - [ ] Test size calculation accuracy
  - [ ] Write unit tests

- [ ] **Day 4-5: Cleaner Logic**
  - [ ] Implement `internal/cleaner/cleaner.go`
  - [ ] Add safety validation
  - [ ] Implement dry-run mode
  - [ ] Add logging to `~/.dev-cleaner.log`

- [ ] **Day 6-7: CLI Integration**
  - [ ] Wire Cobra commands
  - [ ] Add flags: `--dry-run`, `--ios`, `--android`, `--node`
  - [ ] Test end-to-end flow

- [ ] **Week 2: Testing & Refinement**
  - [ ] Add comprehensive unit tests
  - [ ] Test on real development directories
  - [ ] Refine output formatting
  - [ ] Write README with examples

---

## Example Commands (Final Product)

```bash
# Scan everything, show what would be cleaned
dev-cleaner scan

# Scan iOS only, show sizes
dev-cleaner scan --ios

# Interactive TUI for selection
dev-cleaner clean

# Clean iOS caches with confirmation
dev-cleaner clean --ios --confirm

# Dry-run (preview only)
dev-cleaner clean --dry-run --all

# Use custom config
dev-cleaner clean --config ~/.my-cleaner.yaml
```

---

## Key Libraries Documentation

| Library | Purpose | Docs |
|---------|---------|------|
| **Cobra** | CLI framework | https://github.com/spf13/cobra |
| **Viper** | Config management | https://github.com/spf13/viper |
| **Bubble Tea** | TUI framework | https://github.com/charmbracelet/bubbletea |
| **GoReleaser** | Release automation | https://goreleaser.com/intro/ |

---

## Unresolved Questions

1. **Should node_modules search be recursive or limited to common project roots?**
   - Recursive = comprehensive but slow
   - Limited depth = faster but might miss some
   - **Recommendation:** Max depth of 3 levels, configurable

2. **Should we include Cocoapods cache (`~/Library/Caches/CocoaPods/`)?**
   - **Recommendation:** Yes, add to iOS preset

3. **Log retention policy?**
   - **Recommendation:** Keep last 100 operations, auto-rotate

4. **Should --confirm require typing "yes" or just flag presence?**
   - **Recommendation:** Flag presence for CLI, typed "yes" for TUI

---

## Final Recommendation

**Proceed with Go + Cobra + GoReleaser stack.**

This provides:
- Fast development iteration
- Professional-grade CLI UX
- Painless distribution via Homebrew
- Easy maintenance and extension
- Strong safety guarantees

**Start with Phase 1 (MVP) immediately.** Get basic scan/clean working, validate the approach with real usage, then iterate to Phase 2 TUI.

**Total estimated time to production:** 3-4 weeks part-time development.

---

**Next Action:** Run `/plan` to create detailed implementation plan, or start coding directly if ready.
