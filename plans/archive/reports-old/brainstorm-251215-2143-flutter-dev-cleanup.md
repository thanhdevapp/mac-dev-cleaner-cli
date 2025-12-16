# Brainstorm: Flutter Dev Cleanup Features

**Date:** 2025-12-15
**Topic:** Flutter/Dart cleanup requirements for Mac Dev Cleaner
**Status:** Recommended Solution Ready

---

## Problem Statement

Mac Dev Cleaner currently supports Xcode, Android, and Node.js cleanup but **lacks Flutter/Dart support**. Flutter developers accumulate significant disk space from:
- Build artifacts (build/, .dart_tool/)
- Pub package cache (~/.pub-cache)
- Multiple Flutter SDK versions
- iOS/Android build artifacts within Flutter projects
- Gradle caches (shared with Android)

Research shows Flutter devs regularly recover **10-50GB** from cleanup operations.

---

## Current Architecture Analysis

**Existing Pattern (from android.go, node.go):**
```go
// 1. Define paths array with Name + Path
var FlutterPaths = []struct {
    Path string
    Name string
}{...}

// 2. Implement ScanFlutter() method
func (s *Scanner) ScanFlutter() []types.ScanResult

// 3. Add to ScanAll() concurrent goroutines
```

**Integration Points:**
- `pkg/types/types.go` - Add `TypeFlutter` constant
- `internal/scanner/scanner.go` - Add Flutter goroutine in ScanAll()
- `cmd/root/*.go` - Add `--flutter` flag
- `internal/scanner/flutter.go` - **NEW FILE** (main work)

---

## Flutter/Dart Cleanup Categories

### 1. **Global Cache Directories** (High Priority)
Clean these without scanning projects:

| Path | Description | Typical Size | Safety |
|------|-------------|--------------|--------|
| `~/.pub-cache` | Pub package cache | 5-15GB | Safe - redownloads on next `pub get` |
| `~/.dart_tool` | Dart tooling cache | 500MB-2GB | Safe |
| `~/Library/Caches/Flutter` | Flutter SDK caches | 1-3GB | Safe |

### 2. **Project Build Artifacts** (Medium Priority)
Scan common dev directories for Flutter projects:

| Path | Description | Typical Size | Safety |
|------|-------------|--------------|--------|
| `*/build/` | Flutter build output | 200MB-1GB per project | Safe - regenerates on next build |
| `*/.dart_tool/` | Project Dart cache | 50-200MB per project | Safe |
| `*/ios/build/` | iOS build artifacts | 500MB-2GB per project | Safe |
| `*/android/build/` | Android build artifacts | 500MB-2GB per project | Safe |

**Detection Method:** Look for `pubspec.yaml` file to identify Flutter projects (same pattern as node.go looks for node_modules)

### 3. **Multiple Flutter SDK Versions** (Low Priority - Future)
Advanced feature - detect multiple Flutter installations:
- Via `flutter --version`
- Check common paths: `~/flutter`, `~/fvm/versions/*`
- FVM (Flutter Version Management) support

---

## Evaluated Approaches

### **Approach A: Simple Global Cache Only** ⚡ FASTEST
**Implementation:**
- Only scan 3 global paths (~/.pub-cache, ~/.dart_tool, ~/Library/Caches/Flutter)
- Similar to Android scanner pattern
- No project scanning

**Pros:**
- ✅ Quick to implement (30min)
- ✅ Consistent with Android approach
- ✅ Covers 60-70% of Flutter disk usage
- ✅ No edge cases with project detection

**Cons:**
- ❌ Misses per-project build/ and .dart_tool/
- ❌ Less value than scanning projects

**Risk:** Low

---

### **Approach B: Global + Project Scanning** 🎯 RECOMMENDED
**Implementation:**
- Global cache paths (like Approach A)
- Scan project directories for `pubspec.yaml` files
- For each Flutter project found, scan:
  - `build/`
  - `.dart_tool/`
  - `ios/build/`
  - `android/build/`

**Pros:**
- ✅ Maximum disk space recovery (10-50GB)
- ✅ Consistent with node.go pattern (scans node_modules)
- ✅ Covers 95%+ of Flutter cleanup needs
- ✅ Users expect this from similar tools

**Cons:**
- ⚠️ More complex implementation (2-3 hours)
- ⚠️ Slower scanning (recursive directory traversal)
- ⚠️ Need proper depth limits (maxDepth=3)

**Risk:** Medium (manageable with existing patterns)

---

### **Approach C: Full Featured with FVM Support** 🚀 FUTURE
**Implementation:**
- Everything from Approach B
- Plus: Detect multiple Flutter SDK versions
- Plus: Show Flutter SDK version per project
- Plus: FVM-aware scanning

**Pros:**
- ✅ Most comprehensive solution
- ✅ Handles advanced Flutter workflows

**Cons:**
- ❌ Over-engineered for MVP (violates YAGNI)
- ❌ Complex FVM detection logic
- ❌ Requires external commands (flutter --version)
- ❌ Much longer implementation (5-8 hours)

**Risk:** High - premature optimization

---

## Recommended Solution: Approach B

**Rationale:**
1. **Balanced value** - captures 95%+ of cleanup opportunities
2. **Proven pattern** - mirrors existing node.go implementation
3. **User expectations** - similar tools (flutter_clear pkg) do this
4. **Maintainable** - straightforward, no external dependencies
5. **Follows KISS** - just enough complexity, no more

---

## Implementation Details (Approach B)

### File: `internal/scanner/flutter.go`

```go
package scanner

import (
    "os"
    "path/filepath"
    "github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// FlutterGlobalPaths - global Flutter/Dart caches
var FlutterGlobalPaths = []struct {
    Path string
    Name string
}{
    {"~/.pub-cache", "Pub Cache"},
    {"~/.dart_tool", "Dart Tool Cache"},
    {"~/Library/Caches/Flutter", "Flutter SDK Cache"},
}

// ScanFlutter scans for Flutter/Dart artifacts
func (s *Scanner) ScanFlutter(maxDepth int) []types.ScanResult {
    var results []types.ScanResult

    // 1. Scan global caches
    for _, target := range FlutterGlobalPaths {
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
            Type:      types.TypeFlutter,
            Size:      size,
            FileCount: count,
            Name:      target.Name,
        })
    }

    // 2. Scan for Flutter projects
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

        flutterProjects := s.findFlutterProjects(expandedDir, maxDepth)
        results = append(results, flutterProjects...)
    }

    return results
}

// findFlutterProjects finds Flutter projects by pubspec.yaml
func (s *Scanner) findFlutterProjects(root string, maxDepth int) []types.ScanResult {
    var results []types.ScanResult

    if maxDepth <= 0 {
        return results
    }

    entries, err := os.ReadDir(root)
    if err != nil {
        return results
    }

    // Check if current directory is a Flutter project
    if s.PathExists(filepath.Join(root, "pubspec.yaml")) {
        // Scan build artifacts
        buildPaths := []struct {
            subPath string
            name    string
        }{
            {"build", "build"},
            {".dart_tool", ".dart_tool"},
            {"ios/build", "ios/build"},
            {"android/build", "android/build"},
        }

        projectName := filepath.Base(root)

        for _, bp := range buildPaths {
            fullPath := filepath.Join(root, bp.subPath)
            if !s.PathExists(fullPath) {
                continue
            }

            size, count, _ := s.calculateSize(fullPath)
            if size > 0 {
                results = append(results, types.ScanResult{
                    Path:      fullPath,
                    Type:      types.TypeFlutter,
                    Size:      size,
                    FileCount: count,
                    Name:      projectName + "/" + bp.name,
                })
            }
        }

        // Don't recurse into Flutter project subdirectories
        return results
    }

    // Recurse into subdirectories
    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        name := entry.Name()
        if shouldSkipDir(name) {
            continue
        }

        fullPath := filepath.Join(root, name)
        subResults := s.findFlutterProjects(fullPath, maxDepth-1)
        results = append(results, subResults...)
    }

    return results
}
```

### Changes to `pkg/types/types.go`
```go
const (
    TypeXcode   CleanTargetType = "xcode"
    TypeAndroid CleanTargetType = "android"
    TypeNode    CleanTargetType = "node"
    TypeFlutter CleanTargetType = "flutter"  // ADD THIS
    TypeCache   CleanTargetType = "cache"
)

type ScanOptions struct {
    IncludeXcode   bool
    IncludeAndroid bool
    IncludeNode    bool
    IncludeFlutter bool  // ADD THIS
    IncludeCache   bool
    MaxDepth       int
    ProjectRoot    string
}
```

### Changes to `internal/scanner/scanner.go`
```go
func (s *Scanner) ScanAll(opts types.ScanOptions) ([]types.ScanResult, error) {
    var results []types.ScanResult
    var mu sync.Mutex
    var wg sync.WaitGroup

    // ... existing Xcode, Android, Node scanners ...

    // ADD THIS:
    if opts.IncludeFlutter {
        wg.Add(1)
        go func() {
            defer wg.Done()
            flutterResults := s.ScanFlutter(opts.MaxDepth)
            mu.Lock()
            results = append(results, flutterResults...)
            mu.Unlock()
        }()
    }

    wg.Wait()
    return results, nil
}
```

### Changes to `cmd/root/scan.go` and `cmd/root/clean.go`
```go
// Add flag
scanCmd.Flags().BoolVar(&scanOpts.IncludeFlutter, "flutter", false, "Scan Flutter/Dart artifacts")
```

---

## Expected Disk Space Recovery

Based on research:
- **Small projects (1-5 Flutter projects):** 2-8GB
- **Medium projects (5-15 Flutter projects):** 10-20GB
- **Heavy users (15+ projects + FVM):** 30-50GB

**Most common:** 10-15GB for typical Flutter developers

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Slow scanning | Medium | Low | Use maxDepth=3 (same as Node) |
| False positives | Low | Low | Require pubspec.yaml detection |
| Breaking changes | Low | High | All deletions are safe - regenerated on next build |
| User confusion | Low | Low | Clear naming: "ProjectName/build" |

---

## Success Metrics

**Must Have:**
- ✅ Detects ~/.pub-cache and shows size
- ✅ Finds Flutter projects via pubspec.yaml
- ✅ Scans build/, .dart_tool/, ios/build/, android/build/
- ✅ Matches naming pattern of other scanners

**Nice to Have:**
- ✅ Concurrent scanning (goroutines)
- ✅ Respects maxDepth limits
- ✅ Skips hidden directories

**Future:**
- FVM support (Approach C)
- Flutter SDK version display
- Smart cleanup (preserve recent builds)

---

## Next Steps

1. **Create `internal/scanner/flutter.go`** - implement ScanFlutter() method
2. **Update `pkg/types/types.go`** - add TypeFlutter constant
3. **Update `internal/scanner/scanner.go`** - add Flutter to ScanAll()
4. **Update `cmd/root/*.go`** - add --flutter flag
5. **Update `README.md`** - document Flutter support
6. **Test on real Flutter projects** - verify detection and sizing
7. **Update help text** - `dev-cleaner scan --flutter`

**Estimated Time:** 2-3 hours implementation + 1 hour testing

---

## Trade-offs & Decisions

**Decision 1: Scan project directories?**
- ✅ **Yes** - users expect this, provides maximum value

**Decision 2: Include ios/build and android/build?**
- ✅ **Yes** - these are Flutter-generated, safe to clean

**Decision 3: Support FVM now?**
- ❌ **No** - YAGNI principle, add later if users request

**Decision 4: Use same project dirs as Node scanner?**
- ✅ **Yes** - DRY principle, consistent behavior

---

## Unresolved Questions

None - recommended solution is clear and actionable.

---

## Sources

Research sources for Flutter cleanup best practices:
- [Flutter Stole 48GB from My MacBook](https://bwnyasse.net/2025/08/flutter-stole-48gb-from-my-macbook-and-how-i-got-it-back/)
- [Flutter Build Directories Are Eating Your SSD](https://medium.com/easy-flutter/flutter-build-directories-are-eating-your-ssd-heres-how-to-fight-back-3e4adf22058b)
- [flutter_clear Dart package](https://pub.dev/packages/flutter_clear)
- [Complete Guide to Cleaning Up Gradle and Flutter Caches on Windows](https://www.devsecopsnow.com/complete-guide-to-cleaning-up-gradle-and-flutter-caches-on-windows/)
- [How to Clean Up Storage as a Mobile Developer](https://medium.com/@forstman/how-to-clean-up-storage-as-a-mobile-developer-react-native-flutter-5728f1b8c6a2)
- [dart pub cache documentation](https://dart.dev/tools/pub/cmd/pub-cache)
- [Using Flutter Clean Pub Cache](https://www.dhiwise.com/post/guide-to-managing-pub-cache-with-flutter-clean-pub-cache)
- [How to Clear Flutter Project Build Cache](https://sourcebae.com/blog/how-to-clear-flutter-project-build-cache/)
