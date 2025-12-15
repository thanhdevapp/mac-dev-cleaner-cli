# Phase 2: Docker, Java/Kotlin

**Phase:** 2 of 2
**Ecosystems:** Docker, Java/Kotlin (Maven + Gradle)
**Estimated Time:** 3-4 hours
**Prerequisites:** Phase 1 complete

---

## 1. Type System Updates

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/pkg/types/types.go`

#### Step 1.1: Add Docker and Java Constants (after Homebrew constant)

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
    // Phase 2:
    TypeDocker   CleanTargetType = "docker"
    TypeJava     CleanTargetType = "java"
)
```

#### Step 1.2: Update ScanOptions

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
    // Phase 2:
    IncludeDocker   bool
    IncludeJava     bool
    MaxDepth        int
    ProjectRoot     string
}
```

#### Step 1.3: Update DefaultScanOptions

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

## 2. Docker Scanner

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/internal/scanner/docker.go` (NEW)

```go
package scanner

import (
    "encoding/json"
    "os/exec"
    "strings"

    "github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// DockerSystemDF represents docker system df output
type DockerSystemDF struct {
    Type        string `json:"Type"`
    TotalCount  int    `json:"TotalCount"`
    Active      int    `json:"Active"`
    Size        string `json:"Size"`
    Reclaimable string `json:"Reclaimable"`
}

// parseDockerSize converts Docker size strings like "1.5GB" to bytes
func parseDockerSize(sizeStr string) int64 {
    sizeStr = strings.TrimSpace(sizeStr)
    if sizeStr == "" || sizeStr == "0B" {
        return 0
    }

    // Remove any parenthetical info like "(100%)"
    if idx := strings.Index(sizeStr, " "); idx > 0 {
        sizeStr = sizeStr[:idx]
    }

    var multiplier int64 = 1
    var value float64

    // Determine unit
    sizeStr = strings.ToUpper(sizeStr)
    if strings.HasSuffix(sizeStr, "KB") {
        multiplier = 1024
        sizeStr = strings.TrimSuffix(sizeStr, "KB")
    } else if strings.HasSuffix(sizeStr, "MB") {
        multiplier = 1024 * 1024
        sizeStr = strings.TrimSuffix(sizeStr, "MB")
    } else if strings.HasSuffix(sizeStr, "GB") {
        multiplier = 1024 * 1024 * 1024
        sizeStr = strings.TrimSuffix(sizeStr, "GB")
    } else if strings.HasSuffix(sizeStr, "TB") {
        multiplier = 1024 * 1024 * 1024 * 1024
        sizeStr = strings.TrimSuffix(sizeStr, "TB")
    } else if strings.HasSuffix(sizeStr, "B") {
        sizeStr = strings.TrimSuffix(sizeStr, "B")
    }

    // Parse numeric value
    _, err := json.Unmarshal([]byte(sizeStr), &value)
    if err != nil {
        // Try parsing as float directly
        var f float64
        if _, err := fmt.Sscanf(sizeStr, "%f", &f); err == nil {
            value = f
        }
    }

    return int64(value * float64(multiplier))
}

// isDockerAvailable checks if Docker daemon is running
func isDockerAvailable() bool {
    cmd := exec.Command("docker", "info")
    err := cmd.Run()
    return err == nil
}

// ScanDocker scans for Docker artifacts using docker CLI
func (s *Scanner) ScanDocker() []types.ScanResult {
    var results []types.ScanResult

    // Check if Docker is available
    if !isDockerAvailable() {
        // Docker not installed or not running - skip silently
        return results
    }

    // Get Docker disk usage
    cmd := exec.Command("docker", "system", "df", "--format", "{{json .}}")
    output, err := cmd.Output()
    if err != nil {
        return results
    }

    // Parse each line of JSON output
    lines := strings.Split(strings.TrimSpace(string(output)), "\n")
    for _, line := range lines {
        if line == "" {
            continue
        }

        var df DockerSystemDF
        if err := json.Unmarshal([]byte(line), &df); err != nil {
            continue
        }

        // Only include if there's reclaimable space
        reclaimSize := parseDockerSize(df.Reclaimable)
        if reclaimSize == 0 {
            continue
        }

        // Create result for each Docker resource type
        var name string
        switch df.Type {
        case "Images":
            name = "Docker Images (unused)"
        case "Containers":
            name = "Docker Containers (stopped)"
        case "Local Volumes":
            name = "Docker Volumes (unused)"
        case "Build Cache":
            name = "Docker Build Cache"
        default:
            name = "Docker " + df.Type
        }

        results = append(results, types.ScanResult{
            Path:      "docker:" + strings.ToLower(strings.ReplaceAll(df.Type, " ", "-")),
            Type:      types.TypeDocker,
            Size:      reclaimSize,
            FileCount: df.TotalCount - df.Active,
            Name:      name,
        })
    }

    return results
}
```

### Important: Docker Cleaner Integration

For Docker, cleanup requires special handling since we can't use `os.RemoveAll()`.

#### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/internal/cleaner/cleaner.go`

**Add Docker handling in Clean() method (after ValidatePath check):**

```go
func (c *Cleaner) Clean(results []types.ScanResult) ([]CleanResult, error) {
    var cleanResults []CleanResult

    for _, result := range results {
        // Handle Docker paths specially
        if strings.HasPrefix(result.Path, "docker:") {
            cleanResult := c.cleanDocker(result)
            cleanResults = append(cleanResults, cleanResult)
            continue
        }

        // ... existing ValidatePath and os.RemoveAll logic ...
    }

    return cleanResults, nil
}

// cleanDocker handles Docker resource cleanup via CLI
func (c *Cleaner) cleanDocker(result types.ScanResult) CleanResult {
    resourceType := strings.TrimPrefix(result.Path, "docker:")

    if c.dryRun {
        c.logger.Printf("[DRY-RUN] Would clean Docker %s (%.2f MB)\n", resourceType, float64(result.Size)/(1024*1024))
        return CleanResult{
            Path:      result.Path,
            Size:      result.Size,
            Success:   true,
            WasDryRun: true,
        }
    }

    var cmd *exec.Cmd
    switch resourceType {
    case "images":
        cmd = exec.Command("docker", "image", "prune", "-a", "-f")
    case "containers":
        cmd = exec.Command("docker", "container", "prune", "-f")
    case "local-volumes":
        cmd = exec.Command("docker", "volume", "prune", "-f")
    case "build-cache":
        cmd = exec.Command("docker", "builder", "prune", "-a", "-f")
    default:
        return CleanResult{
            Path:    result.Path,
            Size:    result.Size,
            Success: false,
            Error:   fmt.Errorf("unknown docker resource type: %s", resourceType),
        }
    }

    c.logger.Printf("[DELETE] Running: %s\n", strings.Join(cmd.Args, " "))

    if err := cmd.Run(); err != nil {
        c.logger.Printf("[ERROR] Docker cleanup failed: %v\n", err)
        return CleanResult{
            Path:    result.Path,
            Size:    result.Size,
            Success: false,
            Error:   err,
        }
    }

    c.logger.Printf("[SUCCESS] Docker %s cleaned\n", resourceType)
    return CleanResult{
        Path:    result.Path,
        Size:    result.Size,
        Success: true,
    }
}
```

#### Update imports in cleaner.go:

```go
import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"

    "github.com/thanhdevapp/dev-cleaner/pkg/types"
)
```

---

## 3. Java Scanner

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/internal/scanner/java.go` (NEW)

```go
package scanner

import (
    "os"
    "path/filepath"
    "strings"

    "github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// JavaGlobalPaths contains global Java/JVM cache paths
var JavaGlobalPaths = []struct {
    Path string
    Name string
}{
    // Maven
    {"~/.m2/repository", "Maven Local Repository"},
    // Gradle (note: ~/.gradle/caches already in Android scanner)
    {"~/.gradle/wrapper", "Gradle Wrapper Distributions"},
    {"~/.gradle/daemon", "Gradle Daemon Logs"},
}

// JavaMarkerFiles identify Java/Kotlin projects
var JavaMarkerFiles = map[string]string{
    "pom.xml":           "maven",
    "build.gradle":      "gradle",
    "build.gradle.kts":  "gradle",
    "settings.gradle":   "gradle",
}

// ScanJava scans for Java/Kotlin development artifacts
func (s *Scanner) ScanJava(maxDepth int) []types.ScanResult {
    var results []types.ScanResult

    // Scan global caches
    for _, target := range JavaGlobalPaths {
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
            Type:      types.TypeJava,
            Size:      size,
            FileCount: count,
            Name:      target.Name,
        })
    }

    // Scan for Java projects in common development directories
    projectDirs := []string{
        "~/Documents",
        "~/Projects",
        "~/Development",
        "~/Developer",
        "~/Code",
        "~/repos",
        "~/workspace",
        "~/IdeaProjects", // IntelliJ default
    }

    for _, dir := range projectDirs {
        expandedDir := s.ExpandPath(dir)
        if !s.PathExists(expandedDir) {
            continue
        }

        javaArtifacts := s.findJavaArtifacts(expandedDir, maxDepth)
        results = append(results, javaArtifacts...)
    }

    return results
}

// findJavaArtifacts recursively finds Java project build artifacts
func (s *Scanner) findJavaArtifacts(root string, maxDepth int) []types.ScanResult {
    var results []types.ScanResult

    if maxDepth <= 0 {
        return results
    }

    entries, err := os.ReadDir(root)
    if err != nil {
        return results
    }

    // Check if this is a Java project
    projectType := ""
    hasBuildDir := false
    hasTargetDir := false

    for _, entry := range entries {
        name := entry.Name()

        if !entry.IsDir() {
            if pType, ok := JavaMarkerFiles[name]; ok {
                projectType = pType
            }
        } else {
            if name == "build" {
                hasBuildDir = true
            }
            if name == "target" {
                hasTargetDir = true
            }
        }
    }

    // Add build artifacts if Java project
    if projectType != "" {
        projectName := filepath.Base(root)

        // Maven: target directory
        if projectType == "maven" && hasTargetDir {
            targetPath := filepath.Join(root, "target")
            size, count, _ := s.calculateSize(targetPath)
            if size > 0 {
                results = append(results, types.ScanResult{
                    Path:      targetPath,
                    Type:      types.TypeJava,
                    Size:      size,
                    FileCount: count,
                    Name:      projectName + "/target (Maven)",
                })
            }
        }

        // Gradle: build directory
        if projectType == "gradle" && hasBuildDir {
            buildPath := filepath.Join(root, "build")
            size, count, _ := s.calculateSize(buildPath)
            if size > 0 {
                results = append(results, types.ScanResult{
                    Path:      buildPath,
                    Type:      types.TypeJava,
                    Size:      size,
                    FileCount: count,
                    Name:      projectName + "/build (Gradle)",
                })
            }
        }

        // Also check for .gradle directory in project root
        dotGradlePath := filepath.Join(root, ".gradle")
        if s.PathExists(dotGradlePath) {
            size, count, _ := s.calculateSize(dotGradlePath)
            if size > 0 {
                results = append(results, types.ScanResult{
                    Path:      dotGradlePath,
                    Type:      types.TypeJava,
                    Size:      size,
                    FileCount: count,
                    Name:      projectName + "/.gradle",
                })
            }
        }

        // Don't recurse into Java projects
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

        // Skip build/target directories without marker files
        if name == "build" || name == "target" {
            continue
        }

        fullPath := filepath.Join(root, name)
        subResults := s.findJavaArtifacts(fullPath, maxDepth-1)
        results = append(results, subResults...)
    }

    return results
}
```

---

## 4. Scanner Integration

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/internal/scanner/scanner.go`

#### Add Docker and Java Goroutines (after Homebrew goroutine)

```go
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
```

---

## 5. CLI Integration

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/cmd/root/scan.go`

#### Step 5.1: Add Flag Variables

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
    // Phase 2:
    scanDocker   bool
    scanJava     bool
    scanAll      bool
    scanTUI      bool
)
```

#### Step 5.2: Update Long Description

Add to Categories Scanned:
```
  • Docker (unused images, containers, volumes, build cache)
  • Java/Kotlin (Maven .m2, Gradle caches, build directories)
```

Add to Examples:
```
  dev-cleaner scan --docker            # Scan Docker artifacts
  dev-cleaner scan --java              # Scan Java/Maven/Gradle
```

#### Step 5.3: Register Flags in init()

```go
    scanCmd.Flags().BoolVar(&scanDocker, "docker", false, "Scan Docker images, containers, volumes")
    scanCmd.Flags().BoolVar(&scanJava, "java", false, "Scan Maven/Gradle caches and build dirs")
```

#### Step 5.4: Update runScan() Options Logic

```go
    specificFlagSet := scanIOS || scanAndroid || scanNode || scanFlutter ||
        scanPython || scanRust || scanGo || scanHomebrew ||
        scanDocker || scanJava

    if specificFlagSet {
        opts.IncludeXcode = scanIOS
        opts.IncludeAndroid = scanAndroid
        opts.IncludeNode = scanNode
        opts.IncludeFlutter = scanFlutter
        opts.IncludePython = scanPython
        opts.IncludeRust = scanRust
        opts.IncludeGo = scanGo
        opts.IncludeHomebrew = scanHomebrew
        opts.IncludeDocker = scanDocker
        opts.IncludeJava = scanJava
    } else {
        // Default: scan all
        opts = types.DefaultScanOptions()
    }
```

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/cmd/root/clean.go`

Apply same pattern as scan.go.

---

## 6. TUI Updates

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/internal/tui/tui.go`

#### Step 6.1: Update NewModel() Category Detection

Add after Homebrew:
```go
        if typesSeen[types.TypeDocker] {
            categories = append(categories, "Docker")
        }
        if typesSeen[types.TypeJava] {
            categories = append(categories, "Java")
        }
```

#### Step 6.2: Update getTypeBadge()

Add cases:
```go
    case types.TypeDocker:
        return style.Foreground(lipgloss.Color("#2496ED")).Render(string(t)) // Docker blue
    case types.TypeJava:
        return style.Foreground(lipgloss.Color("#ED8B00")).Render(string(t)) // Java orange
```

#### Step 6.3: Update rescanItems()

Update opts:
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
            IncludeDocker:   true,
            IncludeJava:     true,
        }
```

---

## 7. Safety Validation Update

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/internal/cleaner/safety.go`

#### Update ValidatePath for Docker Special Paths

```go
// ValidatePath checks if a path is safe to delete
func ValidatePath(path string) error {
    // Allow Docker pseudo-paths
    if strings.HasPrefix(path, "docker:") {
        return nil
    }

    // ... rest of existing validation ...
}
```

---

## 8. Testing

### Test Commands

```bash
# Build
cd /Users/macmini/Documents/Startup/mac-dev-cleaner-cli
go build -o dev-cleaner .

# Test Docker (requires Docker running)
./dev-cleaner scan --docker --no-tui

# Test Java
./dev-cleaner scan --java --no-tui

# Test all ecosystems
./dev-cleaner scan

# Test Docker cleanup (dry-run)
./dev-cleaner clean --docker

# Test Docker cleanup (actual)
./dev-cleaner clean --docker --confirm
```

### Expected Results

| Ecosystem | Expected Items |
|-----------|----------------|
| Docker | Images (unused), Containers (stopped), Volumes (unused), Build Cache |
| Java | `~/.m2/repository`, `~/.gradle/wrapper`, `*/target` (Maven), `*/build` (Gradle) |

### Edge Cases

- Docker daemon not running - should skip silently
- No Java projects found - should show only global caches
- Empty Maven/Gradle caches - should skip
- Docker with no reclaimable space - should skip

---

## 9. Root Command Update

### File: `/Users/macmini/Documents/Startup/mac-dev-cleaner-cli/cmd/root/root.go`

Update the Long description to include all 10 ecosystems:

```go
Long: `Mac Dev Cleaner - A CLI tool to clean development project artifacts

Quickly free up disk space by removing:
  • Xcode DerivedData, Archives, and caches
  • Android Gradle caches and SDK artifacts
  • Node.js node_modules directories
  • Package manager caches (npm, yarn, pnpm, bun)
  • Flutter/Dart build artifacts and pub-cache
  • Python pip/poetry/uv caches and virtualenvs
  • Rust/Cargo registry and target directories
  • Go build and module caches
  • Homebrew download caches
  • Docker unused images, containers, volumes
  • Java/Kotlin Maven and Gradle caches

Features:
  ...
```

---

## 10. README.md Update

Add new sections for Docker and Java:

```markdown
### Docker
- Unused images (via `docker image prune`)
- Stopped containers (via `docker container prune`)
- Unused volumes (via `docker volume prune`)
- Build cache (via `docker builder prune`)

Note: Requires Docker daemon to be running.

### Java/Kotlin
- `~/.m2/repository/` (Maven local repository)
- `~/.gradle/wrapper/` (Gradle wrapper distributions)
- `~/.gradle/daemon/` (Gradle daemon logs)
- `*/target/` (Maven build directories)
- `*/build/` (Gradle build directories)
- `*/.gradle/` (Project Gradle cache)
```

---

## 11. Verification Checklist

### Docker
- [ ] `dev-cleaner scan --docker` detects Docker resources when daemon running
- [ ] `dev-cleaner scan --docker` handles Docker not installed gracefully
- [ ] `dev-cleaner clean --docker` dry-run shows correct output
- [ ] `dev-cleaner clean --docker --confirm` actually prunes Docker
- [ ] TUI shows Docker badge with correct color

### Java
- [ ] `dev-cleaner scan --java` finds ~/.m2/repository
- [ ] `dev-cleaner scan --java` finds ~/.gradle/wrapper
- [ ] `dev-cleaner scan --java` finds Maven target/ dirs (with pom.xml)
- [ ] `dev-cleaner scan --java` finds Gradle build/ dirs (with build.gradle)
- [ ] TUI shows Java badge with correct color
- [ ] Dry-run and actual deletion work

### Integration
- [ ] `go build` succeeds
- [ ] `go test ./...` passes
- [ ] All 10 ecosystems show in TUI when scanning all
- [ ] Combined flags work (e.g., `--docker --java`)
- [ ] Default scan includes Docker and Java

---

## 12. Complete Ecosystem Support Summary

After Phase 2 completion, Mac Dev Cleaner supports:

| Ecosystem | Scanner | CLI Flag | Type |
|-----------|---------|----------|------|
| Xcode | xcode.go | --ios | TypeXcode |
| Android | android.go | --android | TypeAndroid |
| Node.js | node.go | --node | TypeNode |
| Flutter | flutter.go | --flutter | TypeFlutter |
| Python | python.go | --python | TypePython |
| Rust | rust.go | --rust | TypeRust |
| Go | golang.go | --go | TypeGo |
| Homebrew | homebrew.go | --homebrew | TypeHomebrew |
| Docker | docker.go | --docker | TypeDocker |
| Java | java.go | --java | TypeJava |

**Total: 10 ecosystems**
**Estimated disk reclaim: 30-100+ GB per developer**
