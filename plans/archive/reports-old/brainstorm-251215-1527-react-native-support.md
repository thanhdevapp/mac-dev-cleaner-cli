# React Native Support - Implementation Plan

> **Date:** 2025-12-15
> **Status:** Approved - Ready for Implementation
> **Approach:** Option 1 - Separate Category with `--react-native` flag
> **Version:** v1.1.0

---

## Executive Summary

Add React Native-specific cache scanning to Mac Dev Cleaner CLI via dedicated `--react-native` / `--rn` flag. Targets Metro bundler cache, Haste maps, RN packager cache, and temp files in `$TMPDIR`. Follows existing architecture pattern (separate scanner, new type, dedicated flag).

**Key Decision:** Separate category for clear control, avoiding overlap confusion with Node/iOS/Android flags.

---

## Problem Statement

**User Need:** RN developers accumulate significant cache bloat from Metro bundler, Haste maps, and RN packager. Tools like `react-native-clean-project` exist but require npm install and manual execution per project.

**Gap:** Mac Dev Cleaner CLI currently scans Node/iOS/Android artifacts but misses RN-specific caches in `$TMPDIR`.

**Impact:** RN caches can grow to 500MB-2GB, not detected by current tool.

---

## Solution: Option 1 - Dedicated `--react-native` Flag

### Architecture

**New Type:**
```go
// pkg/types/types.go
const (
    TypeXcode       CleanTargetType = "xcode"
    TypeAndroid     CleanTargetType = "android"
    TypeNode        CleanTargetType = "node"
    TypeReactNative CleanTargetType = "react-native" // NEW
)
```

**New Scanner:**
```go
// internal/scanner/react_native.go
package scanner

// ReactNativeCachePaths defines RN-specific cache locations in TMPDIR
var ReactNativeCachePaths = []CachePattern{
    {Pattern: "metro-*", Name: "Metro Bundler Cache"},
    {Pattern: "haste-map-*", Name: "Haste Map Cache"},
    {Pattern: "react-native-packager-cache-*", Name: "RN Packager Cache"},
    {Pattern: "react-*", Name: "React Native Temp Files"},
}

// ScanReactNative scans for React Native caches in TMPDIR
func (s *Scanner) ScanReactNative() []types.ScanResult
```

**New CLI Flag:**
```go
// cmd/root/scan.go
var scanReactNative bool

scanCmd.Flags().BoolVar(&scanReactNative, "react-native", false, "Scan React Native caches")
scanCmd.Flags().BoolVar(&scanReactNative, "rn", false, "Alias for --react-native")
```

### What Gets Scanned

**Phase 1 (MVP):** Global RN caches in `$TMPDIR`

| Pattern | Location | Description | Est. Size |
|---------|----------|-------------|-----------|
| `metro-*` | `$TMPDIR/metro-*` | Metro bundler cache | 100-500MB |
| `haste-map-*` | `$TMPDIR/haste-map-*` | Haste file map cache | 50-200MB |
| `react-native-packager-cache-*` | `$TMPDIR/react-native-packager-cache-*` | RN packager cache | 50-300MB |
| `react-*` | `$TMPDIR/react-*` | React temp files | 10-100MB |

**Total Potential:** 200MB - 1.2GB per system

**Not Included (Covered by Existing Scanners):**
- ✅ `node_modules` → `--node` flag
- ✅ `ios/Pods` → `--ios` flag
- ✅ `android/.gradle` → `--android` flag
- ✅ Xcode DerivedData → `--ios` flag

---

## Implementation Details

### Files to Create

**1. `internal/scanner/react_native.go`**
```go
package scanner

import (
    "os"
    "path/filepath"
    "github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// CachePattern represents a cache pattern to match
type CachePattern struct {
    Pattern string
    Name    string
}

// ReactNativeCachePaths contains RN-specific cache locations
var ReactNativeCachePaths = []CachePattern{
    {Pattern: "metro-*", Name: "Metro Bundler Cache"},
    {Pattern: "haste-map-*", Name: "Haste Map Cache"},
    {Pattern: "react-native-packager-cache-*", Name: "RN Packager Cache"},
    {Pattern: "react-*", Name: "React Native Temp Files"},
}

// ScanReactNative scans for React Native caches in TMPDIR
func (s *Scanner) ScanReactNative() []types.ScanResult {
    var results []types.ScanResult
    tmpDir := os.TempDir()

    for _, cache := range ReactNativeCachePaths {
        pattern := filepath.Join(tmpDir, cache.Pattern)
        matches, err := filepath.Glob(pattern)
        if err != nil {
            continue
        }

        for _, match := range matches {
            // Skip if not a directory
            info, err := os.Stat(match)
            if err != nil || !info.IsDir() {
                continue
            }

            size, count, err := s.calculateSize(match)
            if err != nil || size == 0 {
                continue
            }

            results = append(results, types.ScanResult{
                Path:      match,
                Type:      types.TypeReactNative,
                Size:      size,
                FileCount: count,
                Name:      cache.Name,
            })
        }
    }

    return results
}
```

**2. `internal/scanner/react_native_test.go`**
```go
package scanner

import (
    "os"
    "path/filepath"
    "testing"
)

func TestScanReactNative(t *testing.T) {
    s, err := New()
    if err != nil {
        t.Fatalf("Failed to create scanner: %v", err)
    }

    // Create test cache directory
    tmpDir := t.TempDir()
    testCache := filepath.Join(tmpDir, "metro-test-cache")
    os.MkdirAll(testCache, 0755)

    // Create test file
    testFile := filepath.Join(testCache, "test.txt")
    os.WriteFile(testFile, []byte("test data"), 0644)

    // Mock TMPDIR
    oldTmpDir := os.TempDir
    defer func() { os.TempDir = oldTmpDir }()
    os.TempDir = func() string { return tmpDir }

    results := s.ScanReactNative()

    if len(results) == 0 {
        t.Error("Expected to find RN caches, got 0")
    }
}
```

### Files to Modify

**1. `pkg/types/types.go`**
```diff
const (
    TypeXcode   CleanTargetType = "xcode"
    TypeAndroid CleanTargetType = "android"
    TypeNode    CleanTargetType = "node"
+   TypeReactNative CleanTargetType = "react-native"
)

type ScanOptions struct {
    IncludeXcode   bool
    IncludeAndroid bool
    IncludeNode    bool
+   IncludeReactNative bool
    MaxDepth       int
    ProjectRoot    string
}

func DefaultScanOptions() ScanOptions {
    return ScanOptions{
        IncludeXcode:   true,
        IncludeAndroid: true
        IncludeNode:    true,
+       IncludeReactNative: true,
        MaxDepth:       3,
    }
}
```

**2. `internal/scanner/scanner.go`**
```diff
func (s *Scanner) ScanAll(opts types.ScanOptions) ([]types.ScanResult, error) {
    var results []types.ScanResult
    var mu sync.Mutex
    var wg sync.WaitGroup

    if opts.IncludeXcode { /* ... */ }
    if opts.IncludeAndroid { /* ... */ }
    if opts.IncludeNode { /* ... */ }

+   if opts.IncludeReactNative {
+       wg.Add(1)
+       go func() {
+           defer wg.Done()
+           rnResults := s.ScanReactNative()
+           mu.Lock()
+           results = append(results, rnResults...)
+           mu.Unlock()
+       }()
+   }

    wg.Wait()
    return results, nil
}
```

**3. `cmd/root/scan.go`**
```diff
var (
    scanIOS     bool
    scanAndroid bool
    scanNode    bool
+   scanReactNative bool
    scanAll     bool
    scanTUI     bool
)

func init() {
    rootCmd.AddCommand(scanCmd)

    scanCmd.Flags().BoolVar(&scanIOS, "ios", false, "Scan iOS/Xcode artifacts only")
    scanCmd.Flags().BoolVar(&scanAndroid, "android", false, "Scan Android/Gradle artifacts only")
    scanCmd.Flags().BoolVar(&scanNode, "node", false, "Scan Node.js artifacts only")
+   scanCmd.Flags().BoolVar(&scanReactNative, "react-native", false, "Scan React Native caches")
+   scanCmd.Flags().BoolVar(&scanReactNative, "rn", false, "Alias for --react-native")
    scanCmd.Flags().BoolVar(&scanAll, "all", true, "Scan all categories (default)")
    scanCmd.Flags().BoolVar(&scanTUI, "tui", true, "Launch interactive TUI (default)")
    scanCmd.Flags().BoolP("no-tui", "T", false, "Disable TUI, show text output")
}

func runScan(cmd *cobra.Command, args []string) {
    // ... existing code ...

    // If any specific flag is set, use only those
-   if scanIOS || scanAndroid || scanNode {
+   if scanIOS || scanAndroid || scanNode || scanReactNative {
        opts.IncludeXcode = scanIOS
        opts.IncludeAndroid = scanAndroid
        opts.IncludeNode = scanNode
+       opts.IncludeReactNative = scanReactNative
    } else {
        // Default: scan all
        opts.IncludeXcode = true
        opts.IncludeAndroid = true
        opts.IncludeNode = true
+       opts.IncludeReactNative = true
    }

    // ... rest of function ...
}
```

**4. `cmd/root/clean.go`** (similar changes)
```diff
var (
    cleanIOS     bool
    cleanAndroid bool
    cleanNode    bool
+   cleanReactNative bool
    cleanConfirm bool
    cleanDryRun  bool
)

// Add flags similar to scan.go
```

**5. `README.md`**
```diff
- **Node.js** - node_modules, npm/yarn/pnpm/bun caches
+ **Node.js** - node_modules, npm/yarn/pnpm/bun caches
+ **React Native** - Metro bundler, Haste maps, packager caches

### Scan for Cleanable Items

```bash
# Scan all categories
dev-cleaner scan

# Scan specific category
dev-cleaner scan --ios
dev-cleaner scan --android
dev-cleaner scan --node
+dev-cleaner scan --react-native  # or --rn

+# Combine flags for RN projects
+dev-cleaner scan --rn --ios --android --node
```

### Scanned Directories

+### React Native
+- `$TMPDIR/metro-*` - Metro bundler cache
+- `$TMPDIR/haste-map-*` - Haste map cache
+- `$TMPDIR/react-native-packager-cache-*` - RN packager cache
+- `$TMPDIR/react-*` - React Native temp files
```

---

## Testing Strategy

### Unit Tests

**Test Cases:**
1. ✅ Scan empty TMPDIR → returns 0 results
2. ✅ Scan with Metro cache → detects Metro cache
3. ✅ Scan with multiple RN caches → detects all
4. ✅ Scan with non-directory matches → ignores files
5. ✅ Scan with empty directories → excludes 0-byte dirs
6. ✅ Calculate size accurately → matches expected size

**Run:**
```bash
go test ./internal/scanner/... -v -run TestReactNative
```

### Integration Tests

**Manual Testing:**
1. Create test caches:
```bash
cd $(mktemp -d)
mkdir -p metro-test haste-map-test react-native-packager-cache-test
dd if=/dev/zero of=metro-test/file1.bin bs=1M count=10
dd if=/dev/zero of=haste-map-test/file2.bin bs=1M count=5
```

2. Run scan:
```bash
dev-cleaner scan --rn
```

3. Verify output shows test caches

4. Clean:
```bash
dev-cleaner clean --rn --confirm
```

5. Verify caches deleted

### Cross-Platform Testing

- ✅ macOS (Darwin): TMPDIR = `/var/folders/...`
- ✅ Linux: TMPDIR = `/tmp/`
- ✅ Both architectures: amd64, arm64

---

## Usage Examples

**Scan RN caches only:**
```bash
dev-cleaner scan --react-native
dev-cleaner scan --rn  # alias
```

**Scan full RN project (recommended):**
```bash
dev-cleaner scan --rn --ios --android --node
```

**Clean RN caches:**
```bash
dev-cleaner clean --rn --confirm
```

**Dry-run (default):**
```bash
dev-cleaner clean --rn
# Shows what would be deleted without actually deleting
```

**Combine with TUI:**
```bash
dev-cleaner scan --rn
# Opens TUI with RN caches for interactive selection
```

---

## Implementation Checklist

### Phase 1: Core Implementation
- [ ] Add `TypeReactNative` to `pkg/types/types.go`
- [ ] Update `ScanOptions` struct
- [ ] Create `internal/scanner/react_native.go`
- [ ] Implement `ScanReactNative()` function
- [ ] Add RN scanning to `ScanAll()` in `scanner.go`
- [ ] Add `--react-native` and `--rn` flags to scan command
- [ ] Add flags to clean command
- [ ] Update flag logic in `runScan()` and `runClean()`

### Phase 2: Testing
- [ ] Create `react_native_test.go`
- [ ] Write unit tests (6 test cases)
- [ ] Run `go test ./...` - all pass
- [ ] Manual testing on macOS
- [ ] Manual testing on Linux (optional)
- [ ] Cross-architecture testing (amd64, arm64)

### Phase 3: Documentation
- [ ] Update README.md Overview section
- [ ] Add RN section to "Scanned Directories"
- [ ] Update Usage examples
- [ ] Add RN-specific examples
- [ ] Update help text in scan.go

### Phase 4: Polish
- [ ] Update `.goreleaser.yaml` description
- [ ] Add RN to Roadmap completion
- [ ] Verify version bump to v1.1.0
- [ ] Update CHANGELOG (if exists)

### Phase 5: Release
- [ ] Commit changes
- [ ] Create PR (if using PR workflow)
- [ ] Tag v1.1.0
- [ ] Push tag → triggers GitHub Actions
- [ ] Verify Homebrew formula updated
- [ ] Test installation: `brew upgrade dev-cleaner`

---

## Risk Assessment

### Low Risk ✅
- **Separate category** - No impact on existing functionality
- **TMPDIR only** - No project file scanning (no accidental deletions)
- **Dry-run default** - Safe by default, requires `--confirm`
- **Cross-platform** - `os.TempDir()` handles platform differences

### Medium Risk ⚠️
- **Glob pattern matching** - Could match unintended directories
  - **Mitigation:** Use specific patterns (`metro-*`, not `me*`)
  - **Mitigation:** Verify directory before calculating size
- **Performance** - TMPDIR could have many entries
  - **Mitigation:** Parallel scanning already implemented
  - **Mitigation:** Skip files, only scan directories

### Considerations
- **Watchman cache** - Not included (requires `watchman` CLI)
  - **Decision:** Document manual command: `watchman watch-del-all`
- **Project-specific builds** - Not included in Phase 1
  - **Future:** Phase 2 could add `--deep` flag for `ios/build/`, `android/build/`

---

## Success Criteria

**MVP Complete When:**
- ✅ `--react-native` flag works
- ✅ Detects Metro cache in TMPDIR
- ✅ Detects Haste map cache
- ✅ Detects RN packager cache
- ✅ Shows accurate sizes
- ✅ Safe deletion with `--confirm`
- ✅ All tests pass
- ✅ README updated
- ✅ Works on macOS and Linux

**Metrics:**
- Scan time: <2s for TMPDIR
- Accuracy: 100% (no false positives)
- User feedback: Positive (saves 200MB-1GB+)

---

## Future Enhancements (v1.2.0+)

### Phase 2: Project Detection
```bash
dev-cleaner scan --rn --deep
```

**What it does:**
- Finds RN projects (searches for `package.json` with `react-native` dependency)
- Scans project-specific directories:
  - `<project>/ios/build/`
  - `<project>/ios/Pods/`
  - `<project>/android/build/`
  - `<project>/android/.gradle/`
  - `<project>/android/app/build/`

**Implementation:** Similar to `findNodeModules()` pattern

### Phase 3: Watchman Integration
```bash
dev-cleaner scan --rn --watchman
```

**What it does:**
- Checks if `watchman` is installed
- Queries Watchman state/cache
- Offers to clear via `watchman watch-del-all`

**Risk:** Clearing Watchman affects ALL projects, very aggressive

---

## Timeline Estimate

**Total: 4-6 hours** (assuming familiarity with codebase)

| Phase | Task | Time |
|-------|------|------|
| 1 | Core implementation | 2-3h |
| 2 | Unit tests | 1h |
| 3 | Documentation | 30min |
| 4 | Manual testing | 30min |
| 5 | Polish & release | 30min |

**Blockers:** None
**Dependencies:** None (existing architecture supports this)

---

## Questions & Decisions

### Resolved ✅
1. **Approach?** → Option 1 (Separate category)
2. **Flag name?** → `--react-native` with `--rn` alias
3. **Scope?** → Phase 1 only (global caches, no project detection)
4. **Watchman?** → Document manual command, don't auto-clear

### Open Questions
1. **Version number?** → Suggest v1.1.0 (minor feature addition)
2. **Branch strategy?** → Create `feature/react-native-support` or commit to main?
3. **Changelog?** → Create CHANGELOG.md if doesn't exist?

---

## References

**External:**
- [react-native-clean-project](https://github.com/pmadruga/react-native-clean-project) - Original tool
- [RN Cache Clearing Guide](https://gist.github.com/jarretmoses/c2e4786fd342b3444f3bc6beff32098d)
- [React Native Metro Config](https://reactnative.dev/docs/metro)

**Internal:**
- Current implementation: `internal/scanner/*.go`
- Types definition: `pkg/types/types.go`
- Scan command: `cmd/root/scan.go`

---

## Next Steps

1. **Review & Approve** - Confirm approach with team/user
2. **Create Feature Branch** - `git checkout -b feature/react-native-support`
3. **Implement Phase 1** - Follow checklist above
4. **Test Thoroughly** - Unit + manual tests
5. **Update Docs** - README + help text
6. **Create PR / Merge** - Code review
7. **Release v1.1.0** - Tag + GitHub Actions
8. **Announce** - Update README.md Roadmap

**Ready to implement? Let's go! 🚀**
