# Phase 1: Python, Rust, Go, Homebrew

**Phase:** 1 of 2
**Ecosystems:** Python, Rust, Go, Homebrew
**Estimated Time:** 4-6 hours

---

## 1. Type System Updates

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/pkg/types/types.go`

#### Step 1.1: Add Type Constants (Line 7-13)

**Current:**
```go
const (
    TypeXcode   CleanTargetType = "xcode"
    TypeAndroid CleanTargetType = "android"
    TypeNode    CleanTargetType = "node"
    TypeFlutter CleanTargetType = "flutter"
    TypeCache   CleanTargetType = "cache"
)
```

**Change to:**
```go
const (
    TypeXcode    CleanTargetType = "xcode"
    TypeAndroid  CleanTargetType = "android"
    TypeNode     CleanTargetType = "node"
    TypeFlutter  CleanTargetType = "flutter"
    TypeCache    CleanTargetType = "cache"
    TypePython   CleanTargetType = "python"
    TypeRust     CleanTargetType = "rust"
    TypeGo       CleanTargetType = "go"
    TypeHomebrew CleanTargetType = "homebrew"
)
```

#### Step 1.2: Update ScanOptions (Line 25-33)

**Current:**
```go
type ScanOptions struct {
    IncludeXcode   bool
    IncludeAndroid bool
    IncludeNode    bool
    IncludeFlutter bool
    IncludeCache   bool
    MaxDepth       int
    ProjectRoot    string
}
```

**Change to:**
```go
type ScanOptions struct {
    IncludeXcode    bool
    IncludeAndroid  bool
    IncludeNode     bool
    IncludeFlutter  bool
    IncludeCache    bool
    IncludePython   bool
    IncludeRust     bool
    IncludeGo       bool
    IncludeHomebrew bool
    MaxDepth        int
    ProjectRoot     string
}
```

#### Step 1.3: Update DefaultScanOptions (Line 43-52)

**Current:**
```go
func DefaultScanOptions() ScanOptions {
    return ScanOptions{
        IncludeXcode:   true,
        IncludeAndroid: true,
        IncludeNode:    true,
        IncludeFlutter: true,
        IncludeCache:   true,
        MaxDepth:       3,
    }
}
```

**Change to:**
```go
func DefaultScanOptions() ScanOptions {
    return ScanOptions{
        IncludeXcode:    true,
        IncludeAndroid:  true,
        IncludeNode:     true,
        IncludeFlutter:  true,
        IncludeCache:    true,
        IncludePython:   true,
        IncludeRust:     true,
        IncludeGo:       true,
        IncludeHomebrew: true,
        MaxDepth:        3,
    }
}
```

---

## 2. Python Scanner

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/internal/scanner/python.go` (NEW)

```go
package scanner

import (
    "os"
    "path/filepath"
    "strings"

    "github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// PythonGlobalPaths contains global Python cache paths
var PythonGlobalPaths = []struct {
    Path string
    Name string
}{
    {"~/.cache/pip", "pip Cache"},
    {"~/.cache/pypoetry", "Poetry Cache"},
    {"~/.cache/pdm", "pdm Cache"},
    {"~/.cache/uv", "uv Cache"},
    {"~/.local/share/virtualenvs", "pipenv virtualenvs"},
}

// PythonProjectDirs are directories that may contain Python projects
var PythonProjectDirs = []string{
    "venv",
    ".venv",
    "env",
    ".env",
    "__pycache__",
    ".pytest_cache",
    ".tox",
    ".mypy_cache",
    ".ruff_cache",
}

// PythonMarkerFiles identify Python projects
var PythonMarkerFiles = []string{
    "requirements.txt",
    "setup.py",
    "pyproject.toml",
    "Pipfile",
    "setup.cfg",
}

// ScanPython scans for Python development artifacts
func (s *Scanner) ScanPython(maxDepth int) []types.ScanResult {
    var results []types.ScanResult

    // Scan global caches
    for _, target := range PythonGlobalPaths {
        path := s.ExpandPath(target.Path)
        if !s.PathExists(path) {
            continue
        }

        size, count, err := s.calculateSize(path)
        if err != nil || size == 0 {
            continue
        }

        results = append(results, types.ScanResult{
            Path:      path,
            Type:      types.TypePython,
            Size:      size,
            FileCount: count,
            Name:      target.Name,
        })
    }

    // Scan for Python projects in common development directories
    projectDirs := []string{
        "~/Documents",
        "~/Projects",
        "~/Development",
        "~/Developer",
        "~/Code",
        "~/repos",
        "~/workspace",
    }

    for _, dir := range projectDirs {
        expandedDir := s.ExpandPath(dir)
        if !s.PathExists(expandedDir) {
            continue
        }

        pythonArtifacts := s.findPythonArtifacts(expandedDir, maxDepth)
        results = append(results, pythonArtifacts...)
    }

    return results
}

// findPythonArtifacts recursively finds Python project artifacts
func (s *Scanner) findPythonArtifacts(root string, maxDepth int) []types.ScanResult {
    var results []types.ScanResult

    if maxDepth <= 0 {
        return results
    }

    entries, err := os.ReadDir(root)
    if err != nil {
        return results
    }

    // Check if this is a Python project
    isPythonProject := false
    for _, entry := range entries {
        if !entry.IsDir() {
            for _, marker := range PythonMarkerFiles {
                if entry.Name() == marker {
                    isPythonProject = true
                    break
                }
            }
        }
        if isPythonProject {
            break
        }
    }

    // Scan for artifacts in Python project
    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        name := entry.Name()
        fullPath := filepath.Join(root, name)

        // Skip hidden dirs (except Python-specific ones)
        if strings.HasPrefix(name, ".") && !isPythonArtifactDir(name) {
            continue
        }

        // Skip common non-project dirs
        if shouldSkipDir(name) {
            continue
        }

        // Check if this is a Python artifact directory
        if isPythonArtifactDir(name) {
            size, count, _ := s.calculateSize(fullPath)
            if size > 0 {
                projectName := filepath.Base(root)
                results = append(results, types.ScanResult{
                    Path:      fullPath,
                    Type:      types.TypePython,
                    Size:      size,
                    FileCount: count,
                    Name:      projectName + "/" + name,
                })
            }
            continue // Don't recurse into artifact dirs
        }

        // Recurse into subdirectories
        subResults := s.findPythonArtifacts(fullPath, maxDepth-1)
        results = append(results, subResults...)
    }

    return results
}

// isPythonArtifactDir checks if directory is a Python artifact
func isPythonArtifactDir(name string) bool {
    for _, artifactDir := range PythonProjectDirs {
        if name == artifactDir {
            return true
        }
    }
    return false
}
```

---

## 3. Rust Scanner

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/internal/scanner/rust.go` (NEW)

```go
package scanner

import (
    "os"
    "path/filepath"
    "strings"

    "github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// RustGlobalPaths contains global Rust/Cargo cache paths
var RustGlobalPaths = []struct {
    Path string
    Name string
}{
    {"~/.cargo/registry", "Cargo Registry"},
    {"~/.cargo/git", "Cargo Git Cache"},
}

// getCargoHome returns CARGO_HOME or default ~/.cargo
func getCargoHome() string {
    if cargoHome := os.Getenv("CARGO_HOME"); cargoHome != "" {
        return cargoHome
    }
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".cargo")
}

// ScanRust scans for Rust/Cargo development artifacts
func (s *Scanner) ScanRust(maxDepth int) []types.ScanResult {
    var results []types.ScanResult

    cargoHome := getCargoHome()

    // Scan global caches (using CARGO_HOME)
    globalPaths := []struct {
        Path string
        Name string
    }{
        {filepath.Join(cargoHome, "registry"), "Cargo Registry"},
        {filepath.Join(cargoHome, "git"), "Cargo Git Cache"},
    }

    for _, target := range globalPaths {
        if !s.PathExists(target.Path) {
            continue
        }

        size, count, err := s.calculateSize(target.Path)
        if err != nil || size == 0 {
            continue
        }

        results = append(results, types.ScanResult{
            Path:      target.Path,
            Type:      types.TypeRust,
            Size:      size,
            FileCount: count,
            Name:      target.Name,
        })
    }

    // Scan for Rust projects' target directories
    projectDirs := []string{
        "~/Documents",
        "~/Projects",
        "~/Development",
        "~/Developer",
        "~/Code",
        "~/repos",
        "~/workspace",
    }

    for _, dir := range projectDirs {
        expandedDir := s.ExpandPath(dir)
        if !s.PathExists(expandedDir) {
            continue
        }

        rustTargets := s.findRustTargets(expandedDir, maxDepth)
        results = append(results, rustTargets...)
    }

    return results
}

// findRustTargets recursively finds Rust target directories
func (s *Scanner) findRustTargets(root string, maxDepth int) []types.ScanResult {
    var results []types.ScanResult

    if maxDepth <= 0 {
        return results
    }

    entries, err := os.ReadDir(root)
    if err != nil {
        return results
    }

    // Check if this directory contains Cargo.toml (is a Rust project)
    hasCargoToml := false
    hasTargetDir := false
    for _, entry := range entries {
        if !entry.IsDir() && entry.Name() == "Cargo.toml" {
            hasCargoToml = true
        }
        if entry.IsDir() && entry.Name() == "target" {
            hasTargetDir = true
        }
    }

    // If Rust project with target, add it
    if hasCargoToml && hasTargetDir {
        targetPath := filepath.Join(root, "target")
        size, count, _ := s.calculateSize(targetPath)
        if size > 0 {
            projectName := filepath.Base(root)
            results = append(results, types.ScanResult{
                Path:      targetPath,
                Type:      types.TypeRust,
                Size:      size,
                FileCount: count,
                Name:      projectName + "/target",
            })
        }
        // Don't recurse into Rust projects
        return results
    }

    // Recurse into subdirectories
    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        name := entry.Name()

        // Skip hidden directories
        if strings.HasPrefix(name, ".") {
            continue
        }

        // Skip common non-project dirs
        if shouldSkipDir(name) {
            continue
        }

        // Skip target directories without Cargo.toml
        if name == "target" {
            continue
        }

        fullPath := filepath.Join(root, name)
        subResults := s.findRustTargets(fullPath, maxDepth-1)
        results = append(results, subResults...)
    }

    return results
}
```

---

## 4. Go Scanner

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/internal/scanner/golang.go` (NEW)

```go
package scanner

import (
    "os"
    "path/filepath"

    "github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// getGOCACHE returns GOCACHE path or default
func getGOCACHE() string {
    if gocache := os.Getenv("GOCACHE"); gocache != "" {
        return gocache
    }
    // macOS default
    home, _ := os.UserHomeDir()
    return filepath.Join(home, "Library", "Caches", "go-build")
}

// getGOMODCACHE returns GOMODCACHE path or default
func getGOMODCACHE() string {
    if gomodcache := os.Getenv("GOMODCACHE"); gomodcache != "" {
        return gomodcache
    }
    // Default: $GOPATH/pkg/mod or ~/go/pkg/mod
    gopath := os.Getenv("GOPATH")
    if gopath == "" {
        home, _ := os.UserHomeDir()
        gopath = filepath.Join(home, "go")
    }
    return filepath.Join(gopath, "pkg", "mod")
}

// ScanGo scans for Go development artifacts
func (s *Scanner) ScanGo(maxDepth int) []types.ScanResult {
    var results []types.ScanResult

    // Go build cache
    gocache := getGOCACHE()
    if s.PathExists(gocache) {
        size, count, err := s.calculateSize(gocache)
        if err == nil && size > 0 {
            results = append(results, types.ScanResult{
                Path:      gocache,
                Type:      types.TypeGo,
                Size:      size,
                FileCount: count,
                Name:      "Go Build Cache",
            })
        }
    }

    // Go module cache
    gomodcache := getGOMODCACHE()
    if s.PathExists(gomodcache) {
        size, count, err := s.calculateSize(gomodcache)
        if err == nil && size > 0 {
            results = append(results, types.ScanResult{
                Path:      gomodcache,
                Type:      types.TypeGo,
                Size:      size,
                FileCount: count,
                Name:      "Go Module Cache",
            })
        }
    }

    // Go test cache (same location as build cache typically)
    gotestcache := os.Getenv("GOTESTCACHE")
    if gotestcache != "" && gotestcache != gocache && s.PathExists(gotestcache) {
        size, count, err := s.calculateSize(gotestcache)
        if err == nil && size > 0 {
            results = append(results, types.ScanResult{
                Path:      gotestcache,
                Type:      types.TypeGo,
                Size:      size,
                FileCount: count,
                Name:      "Go Test Cache",
            })
        }
    }

    return results
}
```

---

## 5. Homebrew Scanner

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/internal/scanner/homebrew.go` (NEW)

```go
package scanner

import (
    "github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// HomebrewPaths contains Homebrew cache paths
var HomebrewPaths = []struct {
    Path string
    Name string
}{
    // User cache
    {"~/Library/Caches/Homebrew", "Homebrew Cache"},
    // Apple Silicon Homebrew
    {"/opt/homebrew/Library/Caches/Homebrew", "Homebrew Cache (ARM)"},
    // Intel Homebrew
    {"/usr/local/Homebrew/Library/Caches/Homebrew", "Homebrew Cache (Intel)"},
}

// ScanHomebrew scans for Homebrew caches
func (s *Scanner) ScanHomebrew() []types.ScanResult {
    var results []types.ScanResult

    for _, target := range HomebrewPaths {
        path := s.ExpandPath(target.Path)
        if !s.PathExists(path) {
            continue
        }

        size, count, err := s.calculateSize(path)
        if err != nil || size == 0 {
            continue
        }

        results = append(results, types.ScanResult{
            Path:      path,
            Type:      types.TypeHomebrew,
            Size:      size,
            FileCount: count,
            Name:      target.Name,
        })
    }

    return results
}
```

---

## 6. Scanner Integration

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/internal/scanner/scanner.go`

#### Step 6.1: Add Goroutines in ScanAll() (after line 85)

**Add after Flutter goroutine (around line 85):**

```go
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
```

---

## 7. CLI Integration - scan.go

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/cmd/root/scan.go`

#### Step 7.1: Add Flag Variables (after line 19)

**Current (line 14-21):**
```go
var (
    scanIOS     bool
    scanAndroid bool
    scanNode    bool
    scanFlutter bool
    scanAll     bool
    scanTUI     bool
)
```

**Change to:**
```go
var (
    scanIOS      bool
    scanAndroid  bool
    scanNode     bool
    scanFlutter  bool
    scanPython   bool
    scanRust     bool
    scanGo       bool
    scanHomebrew bool
    scanAll      bool
    scanTUI      bool
)
```

#### Step 7.2: Update Command Long Description (line 27-61)

**Replace Categories Scanned section:**
```go
Long: `Scan your system for development artifacts that can be cleaned.

By default, scans all supported categories and opens interactive TUI
for browsing, selection, and cleanup. The TUI provides tree navigation,
keyboard shortcuts, and real-time deletion progress.

Categories Scanned:
  • Xcode (DerivedData, Archives, CoreSimulator, CocoaPods)
  • Android (Gradle caches, SDK system images)
  • Node.js (node_modules, npm/yarn/pnpm/bun caches)
  • Flutter (build artifacts, .pub-cache, .dart_tool)
  • Python (pip/poetry/uv caches, venv, __pycache__)
  • Rust (Cargo registry/git, target directories)
  • Go (build cache, module cache)
  • Homebrew (download caches)

Examples:
  dev-cleaner scan                    # Scan all, launch TUI (default)
  dev-cleaner scan --ios              # Scan iOS/Xcode only
  dev-cleaner scan --android          # Scan Android only
  dev-cleaner scan --node             # Scan Node.js only
  dev-cleaner scan --flutter          # Scan Flutter only
  dev-cleaner scan --python           # Scan Python only
  dev-cleaner scan --rust             # Scan Rust/Cargo only
  dev-cleaner scan --go               # Scan Go only
  dev-cleaner scan --homebrew         # Scan Homebrew only
  dev-cleaner scan --no-tui           # Text output without TUI

Flags:
  --ios             Scan iOS/Xcode artifacts only
  --android         Scan Android/Gradle artifacts only
  --node            Scan Node.js artifacts only
  --flutter         Scan Flutter/Dart artifacts only
  --python          Scan Python caches and virtualenvs
  --rust            Scan Rust/Cargo caches and targets
  --go              Scan Go build and module caches
  --homebrew        Scan Homebrew caches
  --no-tui, -T      Disable TUI, show simple text output
  --all             Scan all categories (default: true)

TUI Features:
  • Navigate with arrow keys or vim bindings (k/j/h/l)
  • Select items with Space, 'a' for all, 'n' for none
  • Quick clean single item with 'c'
  • Batch clean selected items with Enter
  • Drill down into folders with → or 'l'
  • Press '?' for detailed help`,
```

#### Step 7.3: Register New Flags in init() (after line 71)

**Add after existing flags:**
```go
    scanCmd.Flags().BoolVar(&scanPython, "python", false, "Scan Python caches (pip, poetry, venv, __pycache__)")
    scanCmd.Flags().BoolVar(&scanRust, "rust", false, "Scan Rust/Cargo caches and target directories")
    scanCmd.Flags().BoolVar(&scanGo, "go", false, "Scan Go build and module caches")
    scanCmd.Flags().BoolVar(&scanHomebrew, "homebrew", false, "Scan Homebrew caches")
```

#### Step 7.4: Update runScan() Options Logic (line 84-101)

**Current:**
```go
    // If any specific flag is set, use only those
    if scanIOS || scanAndroid || scanNode || scanFlutter {
        opts.IncludeXcode = scanIOS
        opts.IncludeAndroid = scanAndroid
        opts.IncludeNode = scanNode
        opts.IncludeFlutter = scanFlutter
    } else {
        // Default: scan all
        opts.IncludeXcode = true
        opts.IncludeAndroid = true
        opts.IncludeNode = true
        opts.IncludeFlutter = true
    }
```

**Change to:**
```go
    // If any specific flag is set, use only those
    specificFlagSet := scanIOS || scanAndroid || scanNode || scanFlutter ||
        scanPython || scanRust || scanGo || scanHomebrew

    if specificFlagSet {
        opts.IncludeXcode = scanIOS
        opts.IncludeAndroid = scanAndroid
        opts.IncludeNode = scanNode
        opts.IncludeFlutter = scanFlutter
        opts.IncludePython = scanPython
        opts.IncludeRust = scanRust
        opts.IncludeGo = scanGo
        opts.IncludeHomebrew = scanHomebrew
    } else {
        // Default: scan all
        opts.IncludeXcode = true
        opts.IncludeAndroid = true
        opts.IncludeNode = true
        opts.IncludeFlutter = true
        opts.IncludePython = true
        opts.IncludeRust = true
        opts.IncludeGo = true
        opts.IncludeHomebrew = true
    }
```

---

## 8. CLI Integration - clean.go

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/cmd/root/clean.go`

#### Step 8.1: Add Flag Variables (after line 24)

**Current:**
```go
var (
    dryRun       bool
    confirmFlag  bool
    cleanIOS     bool
    cleanAndroid bool
    cleanNode    bool
    cleanFlutter bool
    useTUI       bool
)
```

**Change to:**
```go
var (
    dryRun        bool
    confirmFlag   bool
    cleanIOS      bool
    cleanAndroid  bool
    cleanNode     bool
    cleanFlutter  bool
    cleanPython   bool
    cleanRust     bool
    cleanGo       bool
    cleanHomebrew bool
    useTUI        bool
)
```

#### Step 8.2: Update Command Long Description (line 32-77)

Add new ecosystems to examples and flags list.

#### Step 8.3: Register New Flags in init() (after line 89)

**Add:**
```go
    cleanCmd.Flags().BoolVar(&cleanPython, "python", false, "Clean Python caches")
    cleanCmd.Flags().BoolVar(&cleanRust, "rust", false, "Clean Rust/Cargo caches")
    cleanCmd.Flags().BoolVar(&cleanGo, "go", false, "Clean Go caches")
    cleanCmd.Flags().BoolVar(&cleanHomebrew, "homebrew", false, "Clean Homebrew caches")
```

#### Step 8.4: Update runClean() Options Logic (line 117-127)

**Current:**
```go
    if cleanIOS || cleanAndroid || cleanNode || cleanFlutter {
        opts.IncludeXcode = cleanIOS
        opts.IncludeAndroid = cleanAndroid
        opts.IncludeNode = cleanNode
        opts.IncludeFlutter = cleanFlutter
    } else {
        opts.IncludeXcode = true
        opts.IncludeAndroid = true
        opts.IncludeNode = true
        opts.IncludeFlutter = true
    }
```

**Change to:**
```go
    specificFlagSet := cleanIOS || cleanAndroid || cleanNode || cleanFlutter ||
        cleanPython || cleanRust || cleanGo || cleanHomebrew

    if specificFlagSet {
        opts.IncludeXcode = cleanIOS
        opts.IncludeAndroid = cleanAndroid
        opts.IncludeNode = cleanNode
        opts.IncludeFlutter = cleanFlutter
        opts.IncludePython = cleanPython
        opts.IncludeRust = cleanRust
        opts.IncludeGo = cleanGo
        opts.IncludeHomebrew = cleanHomebrew
    } else {
        opts.IncludeXcode = true
        opts.IncludeAndroid = true
        opts.IncludeNode = true
        opts.IncludeFlutter = true
        opts.IncludePython = true
        opts.IncludeRust = true
        opts.IncludeGo = true
        opts.IncludeHomebrew = true
    }
```

---

## 9. TUI Updates

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/internal/tui/tui.go`

#### Step 9.1: Update NewModel() Category Detection (line 253-265)

**Add after TypeFlutter check (around line 265):**
```go
        if typesSeen[types.TypePython] {
            categories = append(categories, "Python")
        }
        if typesSeen[types.TypeRust] {
            categories = append(categories, "Rust")
        }
        if typesSeen[types.TypeGo] {
            categories = append(categories, "Go")
        }
        if typesSeen[types.TypeHomebrew] {
            categories = append(categories, "Homebrew")
        }
```

#### Step 9.2: Update getTypeBadge() (line 1495-1508)

**Add cases before default:**
```go
    case types.TypePython:
        return style.Foreground(lipgloss.Color("#3776AB")).Render(string(t)) // Python blue
    case types.TypeRust:
        return style.Foreground(lipgloss.Color("#DEA584")).Render(string(t)) // Rust orange
    case types.TypeGo:
        return style.Foreground(lipgloss.Color("#00ADD8")).Render(string(t)) // Go cyan
    case types.TypeHomebrew:
        return style.Foreground(lipgloss.Color("#FBB040")).Render(string(t)) // Homebrew yellow
```

#### Step 9.3: Update rescanItems() (line 695-726)

**Update opts to include new ecosystems:**
```go
        opts := types.ScanOptions{
            MaxDepth:        3,
            IncludeXcode:    true,
            IncludeAndroid:  true,
            IncludeNode:     true,
            IncludeFlutter:  true,
            IncludePython:   true,
            IncludeRust:     true,
            IncludeGo:       true,
            IncludeHomebrew: true,
        }
```

---

## 10. Testing

### Test Commands

```bash
# Build
cd /Users/macmini/Documents/Startup/mac-dev-cleaner-cli
go build -o dev-cleaner .

# Test individual ecosystems
./dev-cleaner scan --python --no-tui
./dev-cleaner scan --rust --no-tui
./dev-cleaner scan --go --no-tui
./dev-cleaner scan --homebrew --no-tui

# Test all ecosystems (TUI)
./dev-cleaner scan

# Test with dry-run
./dev-cleaner clean --python

# Verify path safety
# Should NOT show system paths
./dev-cleaner scan --go --no-tui | grep -v "^$"
```

### Expected Results

| Ecosystem | Expected Paths |
|-----------|----------------|
| Python | `~/.cache/pip`, `~/.cache/pypoetry`, `*/venv`, `*/__pycache__` |
| Rust | `~/.cargo/registry`, `~/.cargo/git`, `*/target` (with Cargo.toml) |
| Go | `~/Library/Caches/go-build`, `~/go/pkg/mod` |
| Homebrew | `~/Library/Caches/Homebrew`, `/opt/homebrew/.../Caches` |

---

## 11. Verification Checklist

- [ ] `go build` succeeds without errors
- [ ] `go test ./...` passes
- [ ] `dev-cleaner scan` shows all 8 ecosystems in TUI
- [ ] `dev-cleaner scan --python` shows only Python results
- [ ] `dev-cleaner scan --rust` shows only Rust results
- [ ] `dev-cleaner scan --go` shows only Go results
- [ ] `dev-cleaner scan --homebrew` shows only Homebrew results
- [ ] TUI color badges display correctly for new types
- [ ] Dry-run mode works for new ecosystems
- [ ] No system paths are detected/offered for deletion
- [ ] Missing cache directories are handled gracefully
