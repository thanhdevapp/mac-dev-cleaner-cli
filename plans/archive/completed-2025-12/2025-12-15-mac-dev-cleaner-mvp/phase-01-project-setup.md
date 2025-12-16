# Phase 01: Project Setup & Structure

## Context

| Item | Link |
|------|------|
| Parent Plan | [plan.md](./plan.md) |
| Dependencies | None |
| Research | [Go CLI & Cobra](./research/researcher-go-cli-cobra.md) |

---

## Overview

| Field | Value |
|-------|-------|
| Date | 2025-12-15 |
| Description | Initialize Go module, create project structure, setup Cobra CLI |
| Priority | P0 |
| Status | Pending |
| Est. Duration | 2 hours |

---

## Key Insights (from Research)

1. **Cobra scaffolding**: Use `cobra-cli` for initial setup, creates consistent structure
2. **Project layout**: `cmd/` for CLI, `internal/` for private implementation
3. **Flag precedence**: CLI flags > env vars > config > defaults (Viper pattern)
4. **Error handling**: Use `RunE` (not `Run`) to return errors properly

---

## Requirements

- [ ] Go 1.21+ installed
- [ ] Cobra CLI generator (`cobra-cli`)
- [ ] Module initialized with proper path
- [ ] All directories created
- [ ] Basic root command working

---

## Architecture

```
mac-dev-cleaner/
├── main.go                     # Entry: calls cmd.Execute()
├── cmd/
│   └── root.go                # Root command, global flags
├── internal/
│   ├── scanner/               # Scanning logic (Phase 2)
│   ├── cleaner/               # Deletion logic (Phase 4)
│   └── ui/                    # Formatting (Phase 3)
├── go.mod
├── go.sum
└── .gitignore
```

---

## Related Code Files

| File | Purpose |
|------|---------|
| `main.go` | Entry point, minimal code |
| `cmd/root.go` | Root command with global flags |
| `go.mod` | Module definition |
| `.gitignore` | Ignore build artifacts |

---

## Implementation Steps

### Step 1: Install Prerequisites

```bash
# Verify Go version
go version  # Requires 1.21+

# Install cobra-cli
go install github.com/spf13/cobra-cli@latest
```

### Step 2: Initialize Module

```bash
mkdir -p mac-dev-cleaner
cd mac-dev-cleaner

go mod init github.com/thanhdevapp/dev-cleaner
```

### Step 3: Create Directory Structure

```bash
mkdir -p cmd internal/scanner internal/cleaner internal/ui
```

### Step 4: Create main.go

```go
// main.go
package main

import "github.com/thanhdevapp/dev-cleaner/cmd"

func main() {
    cmd.Execute()
}
```

### Step 5: Create cmd/root.go

```go
// cmd/root.go
package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var (
    version = "0.1.0"
    dryRun  bool
    verbose bool
)

var rootCmd = &cobra.Command{
    Use:     "dev-cleaner",
    Short:   "Clean macOS development artifacts",
    Long:    `dev-cleaner scans and removes development caches, build artifacts, and node_modules to free disk space.`,
    Version: version,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func init() {
    // Global flags
    rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", true, "Preview without deleting (default: true)")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}
```

### Step 6: Install Dependencies

```bash
go get github.com/spf13/cobra@v1.8.0
go get github.com/dustin/go-humanize@v1.0.1
go mod tidy
```

### Step 7: Create .gitignore

```gitignore
# Binaries
dev-cleaner
*.exe

# Build
dist/

# IDE
.idea/
.vscode/

# OS
.DS_Store
```

### Step 8: Verify Setup

```bash
go build -o dev-cleaner .
./dev-cleaner --help
./dev-cleaner --version
```

Expected output:
```
dev-cleaner scans and removes development caches...

Usage:
  dev-cleaner [command]

Available Commands:
  help        Help about any command

Flags:
      --dry-run     Preview without deleting (default true)
  -h, --help        help for dev-cleaner
  -v, --verbose     Verbose output
      --version     version for dev-cleaner
```

---

## Todo List

- [ ] Verify Go 1.21+ installed
- [ ] Initialize Go module
- [ ] Create directory structure
- [ ] Create main.go
- [ ] Create cmd/root.go with global flags
- [ ] Install Cobra and go-humanize
- [ ] Create .gitignore
- [ ] Verify build and --help output
- [ ] Initial git commit

---

## Success Criteria

| Criteria | Metric |
|----------|--------|
| Build succeeds | `go build` exits 0 |
| Help displays | `--help` shows usage |
| Version displays | `--version` shows 0.1.0 |
| Global flags work | `--dry-run`, `--verbose` recognized |
| No lint errors | `go vet ./...` passes |

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Go not installed | High | Low | Check before starting |
| Cobra API changed | Medium | Low | Pin version in go.mod |
| Module path conflict | Low | Low | Use unique GitHub path |

---

## Security Considerations

- No secrets or credentials in code
- .gitignore excludes sensitive files
- Module uses public GitHub path

---

## Next Steps

After Phase 01 complete:
1. Proceed to [Phase 02: Core Scanner](./phase-02-core-scanner.md)
2. Implement scanner interface and target definitions
