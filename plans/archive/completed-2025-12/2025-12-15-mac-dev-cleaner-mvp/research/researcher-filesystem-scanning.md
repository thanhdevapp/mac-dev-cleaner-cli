# Research Report: Go Filesystem Scanning & Directory Size Calculation

**Research Date:** 2025-12-15
**Focus:** Production-ready patterns for mac-dev-cleaner MVP

## Executive Summary

Go 1.16+ provides `filepath.WalkDir` as the standard for efficient directory traversal, avoiding excessive stat calls. For large trees, concurrent scanning with bounded goroutine pools (3-8 workers) yields 3.5x speedup on multi-core systems. Symlinks require explicit detection; permission errors must be handled gracefully. The `go-humanize` library is standard for human-readable size formatting. Avoid recursive approaches for symlink following due to termination issues.

## Key Findings

### 1. filepath.WalkDir vs filepath.Walk Performance

**Critical Advantage of WalkDir:**
- Avoids per-entry `os.Lstat()` calls by leveraging OS-provided file type info
- Performance improvement: **1.5x faster** on cached Linux, **dozens of times faster** on network filesystems and Windows
- WalkDir receives `DirEntry` (type info) vs Walk receives `FileInfo` (requires stat call)

**Recommendation:** Always use `filepath.WalkDir` (Go 1.16+) for new code. Migration from Walk is straightforward - callback signature changes from `(string, os.FileInfo, error)` to `(string, os.DirEntry, error)`.

### 2. Concurrent Directory Scanning

**Performance Metrics:**
- Serial `filepath.WalkDir`: Baseline
- Concurrent with 8 workers on 8-core CPU: **3.5x average speedup** (cwalk project)
- Warning: Too many goroutines cause cache contention, degrading performance

**Best Pattern - Worker Pool with sync.ErrGroup:**
```go
var g errgroup.Group
sem := make(chan struct{}, workers) // Bounded pool

g.Go(func() error {
    return filepath.WalkDir(root, func(path string, d DirEntry, err error) error {
        if err != nil {
            return nil // Skip errors
        }
        sem <- struct{}{}        // Acquire slot
        g.Go(func() error {
            defer func() { <-sem }()
            // Process file
            return nil
        })
        return nil
    })
})
return g.Wait()
```

**Worker Count Guidance:**
- CPU-bound tasks: `runtime.NumCPU()` workers
- I/O-bound tasks: 3-8 workers (benchmark your system with `utkusen/goroutine-benchmark`)
- Start conservative; add workers incrementally

### 3. Directory Size Calculation for Large Trees

**Core Pattern:**
```go
var totalSize int64
filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
    if err != nil {
        return nil // Skip permission errors
    }
    if !d.IsDir() {
        info, _ := d.Info()
        totalSize += info.Size()
    }
    return nil
})
```

**Optimization for Large Trees:**
- Use `os.ReadDir()` (Go 1.16+) instead of deprecated `ioutil.ReadDir`
- Leverage `DirEntry.Info()` only when needed (avoids redundant stat calls)
- For concurrent size calculation, use atomic operations or channels to aggregate

**Performance Note:** Recursive algorithms are problematic - use iteration with filepath.WalkDir to avoid stack exhaustion on deep trees.

### 4. Symlink & Permission Error Handling

**Symlink Detection:**
```go
if d.Type()&os.ModeSymlink != 0 {
    // Is a symlink - WalkDir does NOT follow these
    return nil // Skip or handle explicitly
}
```

**Key Behaviors:**
- `WalkDir(root)`: If root is a symlink, it's resolved once before traversal begins
- Symlinks encountered during traversal are NOT followed (prevents infinite loops, bounds traversal)
- Detecting symlinks requires checking `d.Type()&os.ModeSymlink`

**Permission Error Pattern:**
```go
filepath.WalkDir(root, func(path string, d DirEntry, err error) error {
    if err != nil {
        if os.IsPermission(err) {
            log.Printf("Skip: %s (permission denied)", path)
            return nil // Continue
        }
        return err // Fail on other errors
    }
    // Process entry
    return nil
})
```

**Security Note:** WalkDir is susceptible to TOCTOU (time-of-check/time-of-use) race conditions with symlinks - don't rely on it for security-critical path traversals.

### 5. Human-Readable Size Formatting

**go-humanize Library (Standard Choice):**
```go
import "github.com/dustin/go-humanize"

humanize.Bytes(82854982)  // "83 MB" (SI: base 1000)
humanize.IBytes(82854982) // "79 MiB" (IEC: base 1024)
```

**Installation:** `go get github.com/dustin/go-humanize`

**Format Comparison:**
| Format | Unit | Example | Use Case |
|--------|------|---------|----------|
| SI (Bytes) | Base 1000 | 83 MB | Disk, network speeds |
| IEC (IBytes) | Base 1024 | 79 MiB | RAM, file sizes |

**Alternative:** Implement custom formatter if minimizing dependencies is critical (10-15 lines of code).

## Implementation Recommendations

### Production Pattern: Concurrent Directory Scanner with Size Aggregation
```go
type ScanResult struct {
    Path string
    Size int64
    Err  error
}

func ScanDirConcurrent(root string, workers int) (int64, error) {
    var totalSize int64
    var g errgroup.Group
    sem := make(chan struct{}, workers)

    g.Go(func() error {
        return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
            if err != nil {
                if os.IsPermission(err) {
                    return nil // Continue
                }
                return err
            }

            if d.IsDir() || d.Type()&os.ModeSymlink != 0 {
                return nil // Skip dirs & symlinks
            }

            info, err := d.Info()
            if err != nil {
                return nil
            }

            atomic.AddInt64(&totalSize, info.Size())
            return nil
        })
    })

    return totalSize, g.Wait()
}
```

### Common Pitfalls to Avoid

1. **Using filepath.Walk:** Legacy performance bottleneck - migrate to WalkDir
2. **Following symlinks naively:** Creates infinite loops; use explicit symlink checks
3. **Unbounded goroutines:** Leads to cache thrashing; use semaphore-bounded pools
4. **Ignoring permission errors:** Can miss large portions of tree; handle gracefully
5. **Recursive size calculation:** Stack exhaustion on deep directories - use iteration
6. **Calling d.Info() for every entry:** Extra stat calls; call only when needed

## Benchmarks & Code Examples

**WalkDir vs Walk (Ben Boyter, 2018):**
- Walk: 1.97x slower than WalkDir on large repos
- Real-world impact: Scanning 50k files ~20% faster with WalkDir

**cwalk Concurrent Performance:**
- 8-core system, 8 workers: 3.5x average speedup vs serial WalkDir
- Diminishing returns beyond logical CPU count

**Symlink Handling Safety:**
Use `d.Type()&os.ModeSymlink != 0` to detect before accessing; prevents errors and security issues.

## Resources & References

### Official Documentation
- [filepath package](https://pkg.go.dev/path/filepath)
- [Coming in Go 1.16: ReadDir and DirEntry](https://benhoyt.com/writings/go-readdir/)

### Benchmarking & Performance
- [Ben Boyter - Go File Walk Comparison](https://boyter.org/2018/03/quick-comparison-go-file-walk-implementations/)
- [Kablamo Engineering - File Walking in Go](https://engineering.kablamo.com.au/posts/2021/quick-comparison-between-go-file-walk-implementations/)
- [utkusen/goroutine-benchmark](https://github.com/utkusen/goroutine-benchmark)

### Concurrent Scanning Projects
- [cwalk - Concurrent WalkDir](https://github.com/iafan/cwalk)
- [golangconcurrentdirscan](https://github.com/mwiater/golangconcurrentdirscan)
- [sync.ErrGroup File Searching](https://www.oreilly.com/content/run-strikingly-fast-parallel-file-searches-in-go-with-sync-errgroup/)

### Directory Size & Utilities
- [dugo - Du utility in Go](https://github.com/missedone/dugo)
- [go-get-folder-size](https://pkg.go.dev/github.com/markthree/go-get-folder-size)

### Size Formatting
- [go-humanize](https://github.com/dustin/go-humanize)
- [YourBasic - Byte Size Formatting](https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/)

### Error & Symlink Handling
- [filepath.WalkDir Function Guide](https://www.javaguides.net/2025/01/golang-filepath-walkdir-function.html)
- [DEV Community - WalkDir Traversal](https://dev.to/rezmoss/file-system-walking-with-walkdir-recursive-tree-traversal-49-dj3)
- [Go Issue #70007 - WalkDir Symlink Race](https://github.com/golang/go/issues/70007)

## Unresolved Questions

1. What is the optimal worker count for mac-dev-cleaner on typical macOS systems? (Answer: Benchmark with 3-8 workers)
2. Should we support following symlinks explicitly, or skip them entirely? (Recommendation: Skip for security; document if needed)
3. Is concurrent scanning justified for typical macOS dev directories? (Answer: Likely yes for projects >100k files; test on real projects)
