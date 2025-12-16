# Multi-Ecosystem Support Implementation Plan

**Plan ID:** 251216-0027-multi-ecosystem-support
**Created:** 2025-12-16
**Status:** Ready for Implementation

---

## Executive Summary

Add support for 6 new development ecosystems to Mac Dev Cleaner CLI:
- **Phase 1:** Python, Rust, Go, Homebrew (file-based scanning)
- **Phase 2:** Docker, Java/Kotlin (CLI integration + file scanning)

**Expected Impact:** 20-60 GB additional disk space reclaim per developer

---

## Current Architecture Overview

### Scanner Pattern (existing)

```
internal/scanner/
  - scanner.go     # ScanAll() with parallel goroutines + mutex
  - xcode.go       # XcodePaths[] + ScanXcode()
  - node.go        # NodeGlobalPaths[] + ScanNode(maxDepth)
  - flutter.go     # FlutterGlobalPaths[] + ScanFlutter(maxDepth)
  - android.go     # AndroidPaths[] + ScanAndroid()
```

Each scanner follows this pattern:
1. Define `{Type}Paths` or `{Type}GlobalPaths` struct array
2. Implement `Scan{Type}()` or `Scan{Type}(maxDepth int)` method
3. Expand `~` paths, check existence with `PathExists()`
4. Calculate size with `calculateSize()`
5. Return `[]types.ScanResult`

### Key Files to Modify

| File | Changes |
|------|---------|
| `pkg/types/types.go` | Add new type constants, ScanOptions fields |
| `internal/scanner/scanner.go` | Add goroutines in ScanAll() |
| `cmd/root/scan.go` | Add CLI flags, update help text |
| `cmd/root/clean.go` | Add CLI flags, update help text |
| `internal/cleaner/safety.go` | Update blocklist if needed |
| `internal/tui/tui.go` | Add type badges for new ecosystems |
| `README.md` | Document new ecosystems |

---

## Implementation Phases

### Phase 1: File-Based Ecosystems

| Ecosystem | Disk Impact | Complexity | Files to Create |
|-----------|-------------|------------|-----------------|
| Python | 5-15 GB | Medium | `internal/scanner/python.go` |
| Rust | 10-50 GB | Low | `internal/scanner/rust.go` |
| Go | 2-10 GB | Low | `internal/scanner/golang.go` |
| Homebrew | 1-5 GB | Low | `internal/scanner/homebrew.go` |

**Estimated Time:** 4-6 hours

### Phase 2: CLI-Integrated Ecosystems

| Ecosystem | Disk Impact | Complexity | Files to Create |
|-----------|-------------|------------|-----------------|
| Docker | 10-50+ GB | Medium | `internal/scanner/docker.go` |
| Java/Kotlin | 5-20 GB | Low | `internal/scanner/java.go` |

**Estimated Time:** 3-4 hours

---

## Detailed Phase Specifications

- **[Phase 1: Python, Rust, Go, Homebrew](./phase-01.md)**
- **[Phase 2: Docker, Java/Kotlin](./phase-02.md)**

---

## Type System Changes (pkg/types/types.go)

### New Constants

```go
const (
    TypeXcode    CleanTargetType = "xcode"
    TypeAndroid  CleanTargetType = "android"
    TypeNode     CleanTargetType = "node"
    TypeFlutter  CleanTargetType = "flutter"
    TypeCache    CleanTargetType = "cache"
    // NEW:
    TypePython   CleanTargetType = "python"
    TypeRust     CleanTargetType = "rust"
    TypeGo       CleanTargetType = "go"
    TypeHomebrew CleanTargetType = "homebrew"
    TypeDocker   CleanTargetType = "docker"
    TypeJava     CleanTargetType = "java"
)
```

### Updated ScanOptions

```go
type ScanOptions struct {
    IncludeXcode    bool
    IncludeAndroid  bool
    IncludeNode     bool
    IncludeFlutter  bool
    IncludeCache    bool
    // NEW:
    IncludePython   bool
    IncludeRust     bool
    IncludeGo       bool
    IncludeHomebrew bool
    IncludeDocker   bool
    IncludeJava     bool
    MaxDepth        int
    ProjectRoot     string
}
```

### Updated DefaultScanOptions

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
        IncludeDocker:   true,
        IncludeJava:     true,
        MaxDepth:        3,
    }
}
```

---

## CLI Flag Changes

### scan.go Additions

```go
var (
    scanIOS      bool
    scanAndroid  bool
    scanNode     bool
    scanFlutter  bool
    // NEW:
    scanPython   bool
    scanRust     bool
    scanGo       bool
    scanHomebrew bool
    scanDocker   bool
    scanJava     bool
    scanAll      bool
    scanTUI      bool
)

func init() {
    // ... existing flags ...
    // NEW:
    scanCmd.Flags().BoolVar(&scanPython, "python", false, "Scan Python caches (pip, poetry, venv, __pycache__)")
    scanCmd.Flags().BoolVar(&scanRust, "rust", false, "Scan Rust/Cargo caches and target directories")
    scanCmd.Flags().BoolVar(&scanGo, "go", false, "Scan Go build and module caches")
    scanCmd.Flags().BoolVar(&scanHomebrew, "homebrew", false, "Scan Homebrew caches")
    scanCmd.Flags().BoolVar(&scanDocker, "docker", false, "Scan Docker images, containers, volumes")
    scanCmd.Flags().BoolVar(&scanJava, "java", false, "Scan Maven/Gradle caches")
}
```

### clean.go Additions

Same pattern - add corresponding clean flags.

---

## Scanner Integration (scanner.go)

### ScanAll() Additions

```go
func (s *Scanner) ScanAll(opts types.ScanOptions) ([]types.ScanResult, error) {
    // ... existing goroutines ...

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

    if opts.IncludeDocker {
        wg.Add(1)
        go func() {
            defer wg.Done()
            dockerResults := s.ScanDocker()
            mu.Lock()
            results = append(results, dockerResults...)
            mu.Unlock()
        }()
    }

    if opts.IncludeJava {
        wg.Add(1)
        go func() {
            defer wg.Done()
            javaResults := s.ScanJava(opts.MaxDepth)
            mu.Lock()
            results = append(results, javaResults...)
            mu.Unlock()
        }()
    }

    wg.Wait()
    return results, nil
}
```

---

## TUI Updates (tui.go)

### NewModel() - Category Badges

Add new ecosystem detection in scanning animation:

```go
// Add to category detection in NewModel():
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
if typesSeen[types.TypeDocker] {
    categories = append(categories, "Docker")
}
if typesSeen[types.TypeJava] {
    categories = append(categories, "Java")
}
```

### getTypeBadge() - Colors

```go
func (m Model) getTypeBadge(t types.CleanTargetType) string {
    style := lipgloss.NewStyle().Width(10).Bold(true)
    switch t {
    // ... existing cases ...
    case types.TypePython:
        return style.Foreground(lipgloss.Color("#3776AB")).Render(string(t)) // Python blue
    case types.TypeRust:
        return style.Foreground(lipgloss.Color("#DEA584")).Render(string(t)) // Rust orange
    case types.TypeGo:
        return style.Foreground(lipgloss.Color("#00ADD8")).Render(string(t)) // Go cyan
    case types.TypeHomebrew:
        return style.Foreground(lipgloss.Color("#FBB040")).Render(string(t)) // Homebrew yellow
    case types.TypeDocker:
        return style.Foreground(lipgloss.Color("#2496ED")).Render(string(t)) // Docker blue
    case types.TypeJava:
        return style.Foreground(lipgloss.Color("#ED8B00")).Render(string(t)) // Java orange
    default:
        return style.Render(string(t))
    }
}
```

### rescanItems() Update

```go
func (m Model) rescanItems() tea.Cmd {
    return func() tea.Msg {
        s, err := scanner.New()
        if err != nil {
            return rescanItemsMsg{err: err}
        }

        opts := types.ScanOptions{
            MaxDepth:        3,
            IncludeXcode:    true,
            IncludeAndroid:  true,
            IncludeNode:     true,
            IncludeFlutter:  true,
            // NEW:
            IncludePython:   true,
            IncludeRust:     true,
            IncludeGo:       true,
            IncludeHomebrew: true,
            IncludeDocker:   true,
            IncludeJava:     true,
        }

        results, err := s.ScanAll(opts)
        // ... rest unchanged ...
    }
}
```

---

## Safety Considerations

### Path Validation (safety.go)

No changes needed - existing validation covers:
- System paths blocklist
- Protected patterns (.ssh, .aws, etc.)
- Home directory requirement
- Absolute path requirement

### New Ecosystem-Specific Safety

For Docker: Verify daemon availability before CLI operations
For each scanner: Only target known cache/artifact paths

---

## Testing Strategy

### Manual Testing Checklist

For each new ecosystem:
1. Run `dev-cleaner scan --{ecosystem}` with real caches present
2. Verify correct paths detected
3. Verify correct sizes calculated
4. Test with missing directories (graceful skip)
5. Test dry-run deletion
6. Test actual deletion with `--confirm`

### Edge Cases

- Environment variables not set (Go, Rust, Docker)
- Docker daemon not running
- Empty cache directories
- Permission denied errors
- Symbolic links (skip to avoid cycles)

### Automated Test Files to Create

```
internal/scanner/python_test.go
internal/scanner/rust_test.go
internal/scanner/golang_test.go
internal/scanner/homebrew_test.go
internal/scanner/docker_test.go
internal/scanner/java_test.go
```

---

## README.md Updates

Add new ecosystems to:
1. Overview section
2. Usage examples
3. Scanned Directories section

### New Scanned Directories Content

```markdown
### Python
- `~/.cache/pip/` (pip cache)
- `~/.cache/pypoetry/` (Poetry cache)
- `~/.cache/uv/` (uv cache)
- `*/__pycache__/` (bytecode cache)
- `*/venv/`, `*/.venv/` (virtual environments)
- `*/.pytest_cache/` (pytest cache)
- `*/.tox/` (tox environments)

### Rust/Cargo
- `~/.cargo/registry/` (package registry)
- `~/.cargo/git/` (git dependencies)
- `~/.rustup/toolchains/` (Rust toolchains)
- `*/target/` (build artifacts, with Cargo.toml nearby)

### Go
- `~/go/pkg/mod/` (module cache)
- `~/.cache/go-build/` (build cache)

### Homebrew
- `~/Library/Caches/Homebrew/` (download cache)
- `/opt/homebrew/Library/Caches/` (Apple Silicon)
- `/usr/local/Homebrew/Library/Caches/` (Intel)

### Docker
- Unused images, containers, volumes
- Build cache
- (Uses `docker system prune` for cleanup)

### Java/Kotlin
- `~/.m2/repository/` (Maven local repo)
- `~/.gradle/caches/` (Gradle caches)
- `~/.gradle/wrapper/` (Gradle wrapper)
- `*/target/` (Maven build, with pom.xml nearby)
- `*/build/` (Gradle build, with build.gradle nearby)
```

---

## Implementation Order

### Day 1: Phase 1a (Types + Python + Rust)

1. Update `pkg/types/types.go`
2. Create `internal/scanner/python.go`
3. Create `internal/scanner/rust.go`
4. Update `internal/scanner/scanner.go`
5. Update `cmd/root/scan.go`
6. Update `cmd/root/clean.go`
7. Test Python and Rust scanning

### Day 1: Phase 1b (Go + Homebrew)

1. Create `internal/scanner/golang.go`
2. Create `internal/scanner/homebrew.go`
3. Update scanner.go with new goroutines
4. Test Go and Homebrew scanning

### Day 2: Phase 2 (Docker + Java) + TUI + Docs

1. Create `internal/scanner/docker.go`
2. Create `internal/scanner/java.go`
3. Update `internal/tui/tui.go`
4. Update README.md
5. End-to-end testing

---

## Success Metrics

| Metric | Target |
|--------|--------|
| New ecosystems | 6 |
| Additional disk reclaim potential | 20-60 GB |
| Scan time per ecosystem | <100ms |
| Zero system file deletions | Pass |
| TUI works with 10+ categories | Pass |

---

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Docker CLI not available | Check `docker ps` before scanning; graceful skip |
| Env vars not set (GOMODCACHE, CARGO_HOME) | Use default paths as fallback |
| Permission errors on caches | Skip with warning, continue scanning |
| TUI too crowded with categories | Categories auto-detected from results |
| Accidental deletion | Existing dry-run default + path validation |

---

## Files Created by This Plan

```
plans/251216-0027-multi-ecosystem-support/
  - plan.md (this file)
  - phase-01.md
  - phase-02.md
```

---

## Unresolved Questions

1. **Rustup toolchains:** Include `~/.rustup/toolchains/` or too risky?
   - Recommendation: Include but warn about removing active toolchain

2. **Docker volumes with data:** Flag for data-preserving mode?
   - Recommendation: Skip for MVP, add in future

3. **Virtual environment detection:** Include all venv or only inactive?
   - Recommendation: Include all; user can select
