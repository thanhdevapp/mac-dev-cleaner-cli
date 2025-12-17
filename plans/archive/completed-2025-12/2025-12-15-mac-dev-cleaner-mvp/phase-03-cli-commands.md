# Phase 03: CLI Commands & Flags

## Context

| Item | Link |
|------|------|
| Parent Plan | [plan.md](./plan.md) |
| Dependencies | [Phase 01](./phase-01-project-setup.md), [Phase 02](./phase-02-core-scanner.md) |
| Research | [Go CLI & Cobra](./research/researcher-go-cli-cobra.md) |

---

## Overview

| Field | Value |
|-------|-------|
| Date | 2025-12-15 |
| Description | Implement scan and clean commands with flags and output formatting |
| Priority | P0 |
| Status | Pending |
| Est. Duration | 3 hours |

---

## Key Insights (from Research)

1. **Command pattern**: `APPNAME COMMAND ARG --FLAG`
2. **Use RunE**: Return errors, let Cobra handle exit codes
3. **PersistentFlags**: For flags inherited by subcommands
4. **go-pretty**: Best for tables and formatted output
5. **Short + Long flags**: `-i` and `--ios` for usability

---

## Requirements

- [ ] scan command with type filters
- [ ] clean command with dry-run/confirm
- [ ] Human-readable size output
- [ ] Tabular results display
- [ ] Progress feedback during scan
- [ ] Total size summary

---

## Architecture

```
cmd/
├── root.go      # Root command, global flags
├── scan.go      # Scan command
└── clean.go     # Clean command

internal/ui/
└── formatter.go  # Output formatting utilities
```

**Command Structure:**
```
dev-cleaner
├── scan [--ios] [--android] [--node] [--all]
└── clean [--ios] [--android] [--node] [--confirm]
```

---

## Related Code Files

| File | Purpose |
|------|---------|
| `cmd/scan.go` | Scan command implementation |
| `cmd/clean.go` | Clean command implementation |
| `internal/ui/formatter.go` | Table formatting, size display |

---

## Implementation Steps

### Step 1: Create formatter in ui/formatter.go

```go
// internal/ui/formatter.go
package ui

import (
    "fmt"
    "io"
    "os"
    "strings"

    "github.com/dustin/go-humanize"
    "github.com/thanhdevapp/dev-cleaner/internal/scanner"
)

// Formatter handles CLI output formatting
type Formatter struct {
    Out io.Writer
}

// NewFormatter creates formatter writing to stdout
func NewFormatter() *Formatter {
    return &Formatter{Out: os.Stdout}
}

// PrintResults displays scan results in table format
func (f *Formatter) PrintResults(results []scanner.ScanResult) {
    if len(results) == 0 {
        fmt.Fprintln(f.Out, "No cleanable directories found.")
        return
    }

    // Print header
    fmt.Fprintln(f.Out, "")
    fmt.Fprintf(f.Out, "%-4s %-30s %-10s %s\n", "#", "Name", "Size", "Path")
    fmt.Fprintln(f.Out, strings.Repeat("-", 80))

    // Print rows
    for i, r := range results {
        name := truncate(r.Name, 28)
        size := humanize.Bytes(uint64(r.Size))
        path := truncate(r.Path, 35)

        fmt.Fprintf(f.Out, "%-4d %-30s %-10s %s\n", i+1, name, size, path)
    }

    // Print summary
    fmt.Fprintln(f.Out, strings.Repeat("-", 80))
    total := scanner.TotalSize(results)
    fmt.Fprintf(f.Out, "Total: %d items, %s\n\n", len(results), humanize.Bytes(uint64(total)))
}

// PrintSummary prints final summary after clean
func (f *Formatter) PrintSummary(cleaned int, size int64, dryRun bool) {
    sizeStr := humanize.Bytes(uint64(size))

    if dryRun {
        fmt.Fprintf(f.Out, "\n[DRY-RUN] Would clean %d items (%s)\n", cleaned, sizeStr)
        fmt.Fprintln(f.Out, "Run with --confirm to actually delete.")
    } else {
        fmt.Fprintf(f.Out, "\nCleaned %d items (%s)\n", cleaned, sizeStr)
    }
}

// PrintError prints error message
func (f *Formatter) PrintError(msg string) {
    fmt.Fprintf(f.Out, "Error: %s\n", msg)
}

// PrintInfo prints info message
func (f *Formatter) PrintInfo(msg string) {
    fmt.Fprintln(f.Out, msg)
}

// truncate shortens string with ellipsis
func truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen-3] + "..."
}

// FormatSize returns human-readable size
func FormatSize(bytes int64) string {
    return humanize.Bytes(uint64(bytes))
}
```

### Step 2: Create scan command in cmd/scan.go

```go
// cmd/scan.go
package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
    "github.com/thanhdevapp/dev-cleaner/internal/scanner"
    "github.com/thanhdevapp/dev-cleaner/internal/ui"
)

var (
    scanIOS     bool
    scanAndroid bool
    scanNode    bool
)

var scanCmd = &cobra.Command{
    Use:   "scan",
    Short: "Scan for cleanable directories",
    Long:  `Scan scans your system for development artifacts that can be safely cleaned.`,
    Example: `  dev-cleaner scan           # Scan all types
  dev-cleaner scan --ios     # Scan iOS/Xcode only
  dev-cleaner scan --android # Scan Android/Gradle only
  dev-cleaner scan --node    # Scan node_modules only`,
    RunE: runScan,
}

func init() {
    rootCmd.AddCommand(scanCmd)

    scanCmd.Flags().BoolVarP(&scanIOS, "ios", "i", false, "Scan iOS/Xcode artifacts")
    scanCmd.Flags().BoolVarP(&scanAndroid, "android", "a", false, "Scan Android/Gradle artifacts")
    scanCmd.Flags().BoolVarP(&scanNode, "node", "n", false, "Scan node_modules")
}

func runScan(cmd *cobra.Command, args []string) error {
    s := scanner.NewScanner()
    f := ui.NewFormatter()

    f.PrintInfo("Scanning for development artifacts...")

    var results []scanner.ScanResult

    // Determine what to scan
    scanAll := !scanIOS && !scanAndroid && !scanNode

    if scanAll {
        results = s.ScanAll()
    } else {
        if scanIOS {
            results = append(results, s.ScanByType(scanner.TypeiOS)...)
        }
        if scanAndroid {
            results = append(results, s.ScanByType(scanner.TypeAndroid)...)
        }
        if scanNode {
            results = append(results, s.ScanByType(scanner.TypeNode)...)
        }
    }

    // Check for errors in results
    for _, r := range results {
        if r.Error != nil && verbose {
            fmt.Printf("Warning: %s - %v\n", r.Path, r.Error)
        }
    }

    f.PrintResults(results)

    return nil
}
```

### Step 3: Create clean command in cmd/clean.go

```go
// cmd/clean.go
package cmd

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"

    "github.com/spf13/cobra"
    "github.com/thanhdevapp/dev-cleaner/internal/scanner"
    "github.com/thanhdevapp/dev-cleaner/internal/ui"
)

var (
    cleanIOS     bool
    cleanAndroid bool
    cleanNode    bool
    confirm      bool
)

var cleanCmd = &cobra.Command{
    Use:   "clean",
    Short: "Clean selected development artifacts",
    Long: `Clean removes selected development artifacts from your system.
By default, runs in dry-run mode (preview only).
Use --confirm to actually delete files.`,
    Example: `  dev-cleaner clean                  # Interactive selection (dry-run)
  dev-cleaner clean --confirm        # Interactive selection (delete)
  dev-cleaner clean --ios --confirm  # Clean iOS only`,
    RunE: runClean,
}

func init() {
    rootCmd.AddCommand(cleanCmd)

    cleanCmd.Flags().BoolVarP(&cleanIOS, "ios", "i", false, "Clean iOS/Xcode artifacts")
    cleanCmd.Flags().BoolVarP(&cleanAndroid, "android", "a", false, "Clean Android/Gradle artifacts")
    cleanCmd.Flags().BoolVarP(&cleanNode, "node", "n", false, "Clean node_modules")
    cleanCmd.Flags().BoolVar(&confirm, "confirm", false, "Actually delete (not dry-run)")
}

func runClean(cmd *cobra.Command, args []string) error {
    s := scanner.NewScanner()
    f := ui.NewFormatter()

    f.PrintInfo("Scanning for development artifacts...")

    var results []scanner.ScanResult

    // Determine what to scan
    scanAll := !cleanIOS && !cleanAndroid && !cleanNode

    if scanAll {
        results = s.ScanAll()
    } else {
        if cleanIOS {
            results = append(results, s.ScanByType(scanner.TypeiOS)...)
        }
        if cleanAndroid {
            results = append(results, s.ScanByType(scanner.TypeAndroid)...)
        }
        if cleanNode {
            results = append(results, s.ScanByType(scanner.TypeNode)...)
        }
    }

    if len(results) == 0 {
        f.PrintInfo("No cleanable directories found.")
        return nil
    }

    // Display results
    f.PrintResults(results)

    // Get selection from user
    selected, err := promptSelection(results)
    if err != nil {
        return err
    }

    if len(selected) == 0 {
        f.PrintInfo("No items selected.")
        return nil
    }

    // Calculate total size
    var totalSize int64
    for _, r := range selected {
        totalSize += r.Size
    }

    // Confirm deletion
    isDryRun := !confirm
    if !isDryRun {
        if !promptConfirmDelete(len(selected), totalSize) {
            f.PrintInfo("Cancelled.")
            return nil
        }
    }

    // Perform clean (actual cleaning in Phase 4)
    // For now, just show what would be cleaned
    for _, r := range selected {
        if isDryRun {
            fmt.Printf("[DRY-RUN] Would delete: %s (%s)\n", r.Path, ui.FormatSize(r.Size))
        } else {
            fmt.Printf("[DELETE] %s (%s)\n", r.Path, ui.FormatSize(r.Size))
            // Actual deletion implemented in Phase 4
        }
    }

    f.PrintSummary(len(selected), totalSize, isDryRun)

    return nil
}

// promptSelection asks user to select items
func promptSelection(results []scanner.ScanResult) ([]scanner.ScanResult, error) {
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Enter items to clean (e.g., 1,2,3 or 'all' or 'q' to quit): ")
    input, err := reader.ReadString('\n')
    if err != nil {
        return nil, err
    }

    input = strings.TrimSpace(strings.ToLower(input))

    if input == "q" || input == "quit" {
        return nil, nil
    }

    if input == "all" || input == "a" {
        return results, nil
    }

    // Parse comma-separated numbers
    var selected []scanner.ScanResult
    parts := strings.Split(input, ",")

    for _, p := range parts {
        p = strings.TrimSpace(p)
        idx, err := strconv.Atoi(p)
        if err != nil {
            continue
        }

        // Convert to 0-indexed
        idx--
        if idx >= 0 && idx < len(results) {
            selected = append(selected, results[idx])
        }
    }

    return selected, nil
}

// promptConfirmDelete asks for final confirmation
func promptConfirmDelete(count int, size int64) bool {
    reader := bufio.NewReader(os.Stdin)

    fmt.Printf("\nAbout to delete %d items (%s). This cannot be undone.\n", count, ui.FormatSize(size))
    fmt.Print("Type 'yes' to confirm: ")

    input, err := reader.ReadString('\n')
    if err != nil {
        return false
    }

    return strings.TrimSpace(strings.ToLower(input)) == "yes"
}
```

### Step 4: Update root.go for verbose flag access

```go
// Update cmd/root.go - add export for verbose
// In the var block:
var (
    version = "0.1.0"
    dryRun  bool
    verbose bool  // This needs to be accessible to other commands
)

// Or make it accessible via getter:
func IsVerbose() bool {
    return verbose
}
```

### Step 5: Verify compilation and test

```bash
# Build
go build -o dev-cleaner .

# Test scan
./dev-cleaner scan
./dev-cleaner scan --ios
./dev-cleaner scan --help

# Test clean
./dev-cleaner clean
./dev-cleaner clean --ios
./dev-cleaner clean --help
```

---

## Todo List

- [ ] Create internal/ui/formatter.go
- [ ] Create cmd/scan.go with filters
- [ ] Create cmd/clean.go with interactive selection
- [ ] Add --confirm flag logic
- [ ] Add dry-run output
- [ ] Test all command combinations
- [ ] Verify help text displays correctly

---

## Success Criteria

| Criteria | Metric |
|----------|--------|
| scan command works | Shows results table |
| Type filters work | `--ios`, `--android`, `--node` |
| clean prompts | Interactive selection works |
| dry-run default | No deletion without `--confirm` |
| Help displays | `--help` shows examples |
| Size formatting | Human-readable (GB, MB) |

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Input parsing errors | Medium | Medium | Validate input, handle errors |
| Table alignment broken | Low | Low | Test with long paths |
| Verbose flag not shared | Low | Medium | Use getter function |

---

## Security Considerations

- Never auto-delete without user input
- Require explicit --confirm for deletion
- Type "yes" confirmation for actual deletes
- Show full paths before deletion

---

## Next Steps

After Phase 03 complete:
1. Proceed to [Phase 04: Safety & Confirmation](./phase-04-safety-confirmation.md)
2. Implement actual deletion with safety checks
3. Add logging
