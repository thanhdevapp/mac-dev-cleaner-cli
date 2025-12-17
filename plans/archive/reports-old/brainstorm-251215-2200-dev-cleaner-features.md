# Dev Cleaner Feature Research - Additional Ecosystem Support

**Date:** 2025-12-15
**Topic:** Research additional features for dev tool cleaner (Node.js, Bun, Python, Rust, Go, etc.)
**Status:** Completed

---

## Executive Summary

Current tool supports: Xcode, Android, Node.js (npm/yarn/pnpm/bun caches + node_modules), Flutter/Dart.

Research reveals 8+ additional ecosystems with significant cache/artifact footprint. Priority features identified across 4 tiers based on ecosystem popularity, disk impact, and implementation complexity.

---

## Current State Analysis

### Implemented Features
- **iOS/Xcode:** DerivedData, Archives, Caches, CoreSimulator, CocoaPods
- **Android:** Gradle caches, wrapper, SDK system-images
- **Node.js:** npm/yarn/pnpm/bun caches + project node_modules (depth-limited scan)
- **Flutter/Dart:** .pub-cache, .dart_tool, build artifacts, platform-specific builds

### Implementation Architecture
- Parallel scanning with goroutines + mutex synchronization
- Category-specific scanner modules (xcode.go, android.go, node.go, flutter.go)
- Depth-limited recursive search for project artifacts (maxDepth=3)
- Size calculation with file counting
- Skip logic for .git, node_modules, .Trash, Library

---

## Competitive Analysis

### Existing Multi-Language Cleaners

#### clean-dev-dirs (Rust)
- **Languages:** Rust (target/), Node.js (node_modules/), Python (__pycache__, venv), Go (vendor/)
- **Features:** Parallel scanning, smart filtering, interactive mode, dry-run, progress indicators, detailed stats
- **Link:** [clean-dev-dirs](https://github.com/TomPlanche/clean-dev-dirs)

#### devclean (Tauri GUI + CLI)
- **Languages:** Node.js, Rust (extensible architecture)
- **Features:** Desktop GUI + CLI, visual interface for cache management
- **Link:** [devclean](https://github.com/HuakunShen/devclean)

#### vyrti/cleaner (Parallel Scanner)
- **Languages:** .terraform, target, node_modules, __pycache__
- **Features:** Ultra-fast parallel scanning, instant drive-wide search
- **Link:** [vyrti/cleaner](https://github.com/vyrti/cleaner)

### Gap Analysis
**Missing ecosystems in current tool:**
1. Python (pip, poetry, pdm, uv, venv, __pycache__)
2. Rust (cargo cache, target directories)
3. Go (GOCACHE, GOMODCACHE, vendor)
4. Docker (images, containers, volumes, build cache)
5. Java/Kotlin (Maven .m2, Gradle cache)
6. Ruby (gems, bundler cache)
7. PHP (Composer cache)
8. Database tools (Redis dumps, PostgreSQL logs)

---

## Proposed Feature Additions

### Tier 1: High Priority (Large Impact, Common Use)

#### 1. Python Ecosystem
**Disk Impact:** 5-15 GB typical
**Complexity:** Medium

**Cache Locations:**
```
~/.cache/pip/                          # pip cache
~/.cache/pypoetry/                     # poetry cache
~/.cache/pdm/                          # pdm cache
~/.cache/uv/                           # uv cache (modern tool)
~/.local/share/virtualenv/             # virtualenv cache
*/__pycache__/                         # bytecode cache (project)
*/.pytest_cache/                       # pytest cache
*/.tox/                                # tox virtual environments
*/.mypy_cache/                         # mypy type checker cache
*/venv/, */env/, */.venv/              # virtual environments
```

**Detection Logic:**
- Marker files: `requirements.txt`, `setup.py`, `pyproject.toml`, `Pipfile`, `poetry.lock`
- Global caches: Fixed paths
- Project caches: Depth-limited scan for __pycache__, venv dirs

**Built-in Commands:**
- `pip cache purge` - clear pip cache
- `poetry cache clear --all .` - clear poetry cache
- `uv cache clean` - clear uv cache

**References:**
- [pip cache docs](https://pip.pypa.io/en/stable/cli/pip_cache/)
- [Poetry cache config](https://python-poetry.org/docs/configuration/)
- [uv package manager](https://astral.sh/blog/uv)
- [PyClean tool](https://pypi.org/project/pyclean/)

---

#### 2. Docker Artifacts
**Disk Impact:** 10-50 GB typical (can exceed 100GB)
**Complexity:** Low-Medium

**Cache Locations:**
```
~/Library/Containers/com.docker.docker/Data/vms/0/data/Docker.raw   # Docker VM (Mac)
# Programmatically via docker CLI:
docker system df                       # Show space usage
docker image ls -a                     # List all images
docker container ls -a                 # List all containers
docker volume ls                       # List volumes
docker buildx du                       # Build cache usage
```

**Cleanup Commands:**
```bash
docker system prune -a                 # Remove all unused data
docker image prune -a                  # Remove unused images
docker container prune                 # Remove stopped containers
docker volume prune                    # Remove unused volumes
docker builder prune                   # Remove build cache
```

**Implementation Strategy:**
- Execute `docker system df --format json` to get size breakdown
- Present as scan results (images, containers, volumes, build cache)
- Clean action runs `docker system prune` with confirmation
- Safety: Preserve running containers, warn about data loss

**References:**
- [Docker cache guide - Blacksmith](https://www.blacksmith.sh/blog/a-guide-to-disk-space-management-with-docker-how-to-clear-your-cache)
- [Docker clear cache - Depot](https://depot.dev/blog/docker-clear-cache)
- [Reclaim disk space - Medium](https://medium.com/@alexeysamoshkin/reclaim-disk-space-by-removing-stale-and-unused-docker-data-a4c3bd1e4001)

---

#### 3. Rust/Cargo Ecosystem
**Disk Impact:** 10-50 GB typical (target dirs accumulate fast)
**Complexity:** Low-Medium

**Cache Locations:**
```
~/.cargo/registry/                     # Package registry cache
~/.cargo/git/                          # Git dependencies cache
*/target/                              # Build artifacts (per project)
```

**Detection Logic:**
- Marker files: `Cargo.toml`, `Cargo.lock`
- Global caches: ~/.cargo/registry, ~/.cargo/git
- Project caches: Find target/ in depth-limited scan

**Built-in Commands:**
```bash
cargo clean                            # Remove target directory
cargo cache --autoclean                # Auto-clean with cargo-cache tool
cargo cache -a                         # Show cache info
```

**Implementation Strategy:**
- Scan ~/.cargo/{registry,git} for global caches
- Recursively find Cargo.toml projects and their target/ dirs
- Option to preserve executables (cargo-cleaner feature)
- Cargo auto-cleans old cache items (last-use tracker)

**References:**
- [Cargo cache cleaning - Rust Blog](https://blog.rust-lang.org/2023/12/11/cargo-cache-cleaning/)
- [cargo-cache tool](https://github.com/matthiaskrgr/cargo-cache)
- [Freeing gigabytes - thisDaveJ](https://thisdavej.com/freeing-up-gigabytes-reclaiming-disk-space-from-rust-cargo-builds/)
- [Cleaning up Rust - Heath Stewart](https://heaths.dev/rust/2025/03/01/cleaning-up-rust.html)

---

### Tier 2: Medium Priority (Moderate Impact)

#### 4. Go Ecosystem
**Disk Impact:** 2-10 GB typical
**Complexity:** Low

**Cache Locations:**
```
~/Library/Caches/go-build/             # Build cache (Mac)
~/go/pkg/mod/                          # Module cache (GOMODCACHE)
*/vendor/                              # Vendored dependencies
```

**Built-in Commands:**
```bash
go clean -cache                        # Clear build cache
go clean -modcache                     # Clear module cache
go clean -testcache                    # Clear test cache
```

**Detection Logic:**
- Check GOCACHE env var (default: ~/Library/Caches/go-build on Mac)
- Check GOMODCACHE env var (default: ~/go/pkg/mod)
- Find go.mod projects and scan vendor/ dirs

**References:**
- [How to clean Go - Leapcell](https://leapcell.io/blog/how-to-clean-go-a-guide-to-keeping-your-go-environment-tidy)

---

#### 5. Java/Kotlin Ecosystem
**Disk Impact:** 5-20 GB typical
**Complexity:** Low-Medium

**Maven Cache:**
```
~/.m2/repository/                      # Maven local repository
~/.m2/.build-cache/                    # Maven build cache (3.9+)
```

**Gradle Cache:**
```
~/.gradle/caches/                      # Gradle caches
~/.gradle/wrapper/                     # Gradle wrapper distributions
~/.gradle/daemon/                      # Gradle daemon logs
```

**Detection Logic:**
- Maven: Find pom.xml projects
- Gradle: Find build.gradle/build.gradle.kts projects
- Scan global cache dirs

**Built-in Commands:**
```bash
# Maven
mvn dependency:purge-local-repository  # Purge and re-download
rm -rf ~/.m2/repository                # Manual deletion

# Gradle
gradle clean                           # Clean project build dir
gradle cleanBuildCache                 # Clean build cache
```

**References:**
- [Clearing Maven cache - Baeldung](https://www.baeldung.com/maven-clear-cache)
- [Maven cache clearing - GeeksforGeeks](https://www.geeksforgeeks.org/advance-java/clearing-the-maven-cache/)
- [Maven build cache extension](https://maven.apache.org/extensions/maven-build-cache-extension/)

---

#### 6. Homebrew (Mac-specific)
**Disk Impact:** 1-5 GB typical
**Complexity:** Low

**Cache Locations:**
```
~/Library/Caches/Homebrew/             # Formula downloads
/Library/Caches/Homebrew/              # System cache
$(brew --cache)                        # Get cache location
```

**Built-in Commands:**
```bash
brew cleanup -s                        # Cleanup with scrubbing
brew autoremove                        # Remove unused dependencies
brew cleanup --prune=all               # Remove all cached downloads
```

**References:**
- [Free up storage - David Manske](https://davidemanske.com/free-up-storage-on-mac-from-homebrew-and-docker-images/)
- [Mac cleanup guide - Bomberbot](https://www.bomberbot.com/git/maximizing-disk-space-on-your-mac-the-ultimate-cleanup-guide-for-developers/)

---

### Tier 3: Lower Priority (Niche but Valuable)

#### 7. Ruby Ecosystem
**Disk Impact:** 1-5 GB typical
**Complexity:** Low

**Cache Locations:**
```
~/.bundle/cache/                       # Bundler cache
~/.gem/                                # RubyGems cache
vendor/bundle/                         # Project-local bundle (if --deployment)
```

**Built-in Commands:**
```bash
gem cleanup                            # Remove old gem versions
bundle clean                           # Remove unused gems
```

---

#### 8. PHP Ecosystem
**Disk Impact:** 1-3 GB typical
**Complexity:** Low

**Cache Locations:**
```
~/.composer/cache/                     # Composer cache
vendor/                                # Project dependencies
```

**Built-in Commands:**
```bash
composer clear-cache                   # Clear composer cache
rm -rf vendor                          # Remove project deps
```

---

#### 9. iOS Dependency Managers
**Disk Impact:** 500MB - 2GB
**Complexity:** Low (already partially implemented)

**Additional Locations:**
```
~/Library/Caches/org.carthage.CarthageKit/  # Carthage cache
~/.carthage/                                # Carthage build artifacts
Carthage/Build/                             # Project build artifacts
```

**Built-in Commands:**
```bash
pod cache clean --all                  # CocoaPods (already known)
# Carthage has no built-in clean - must manually delete
```

**References:**
- [Understanding Xcode space - Kodeco](https://www.kodeco.com/19998365-understanding-and-managing-xcode-space/page/3)
- [Clear CocoaPods cache](https://gist.github.com/mbinna/4202236)
- [Xcode cleanup tools](https://www.medevel.com/cleaning-xcode-clutter-best-free-tools/)

---

### Tier 4: Edge Cases (Optional)

#### 10. Database Development Tools
```
~/Library/Application Support/Postgres/  # PostgreSQL data
~/.redis/                                # Redis dumps
```

#### 11. Terraform
```
.terraform/                              # Per-project cache
~/.terraform.d/plugin-cache/             # Plugin cache
```

---

## Implementation Recommendations

### Architecture Additions

#### 1. New Scanner Modules (Follow Existing Pattern)
```go
// internal/scanner/python.go
func (s *Scanner) ScanPython(maxDepth int) []types.ScanResult

// internal/scanner/docker.go
func (s *Scanner) ScanDocker() []types.ScanResult

// internal/scanner/rust.go
func (s *Scanner) ScanRust(maxDepth int) []types.ScanResult

// internal/scanner/go.go
func (s *Scanner) ScanGo(maxDepth int) []types.ScanResult

// internal/scanner/java.go
func (s *Scanner) ScanJava(maxDepth int) []types.ScanResult

// internal/scanner/homebrew.go (Mac-specific)
func (s *Scanner) ScanHomebrew() []types.ScanResult
```

#### 2. Type System Expansion
```go
// pkg/types/types.go
const (
    TypePython    = "python"
    TypeDocker    = "docker"
    TypeRust      = "rust"
    TypeGo        = "go"
    TypeJava      = "java"
    TypeHomebrew  = "homebrew"
    TypeRuby      = "ruby"
    TypePHP       = "php"
)
```

#### 3. CLI Flag Additions
```go
// cmd/root/scan.go
--python       Scan Python caches
--docker       Scan Docker artifacts
--rust         Scan Rust/Cargo caches
--go           Scan Go build/module caches
--java         Scan Maven/Gradle caches
--homebrew     Scan Homebrew caches (Mac)
--all          Scan all categories (default)
```

#### 4. Docker-Specific Implementation
```go
// Use exec.Command to run docker CLI
cmd := exec.Command("docker", "system", "df", "--format", "json")
// Parse JSON output to build ScanResults
// Handle case where Docker is not installed
```

---

## Priority Rollout Plan

### Phase 1 (High Value, Low Complexity)
1. **Python** - Huge ecosystem, straightforward paths
2. **Rust/Cargo** - Large target dirs, simple detection
3. **Go** - Built-in clean commands, env vars
4. **Homebrew** - Mac developers, simple brew CLI integration

### Phase 2 (High Value, Moderate Complexity)
5. **Docker** - Massive disk usage, requires CLI integration
6. **Java/Maven/Gradle** - Enterprise developers, large caches

### Phase 3 (Nice to Have)
7. **Ruby** - Smaller user base but straightforward
8. **PHP** - Smaller user base
9. **Carthage** (iOS) - Complete iOS dependency coverage

### Phase 4 (Optional)
10. Database/Terraform caches

---

## Risk Assessment

### Technical Risks
1. **Docker CLI dependency:** Requires Docker installed + running
   - Mitigation: Graceful degradation, check if docker command exists

2. **Path variations:** Cache locations vary by OS/config
   - Mitigation: Check env vars (GOCACHE, CARGO_HOME, etc.)

3. **Permission errors:** Some caches may require elevated permissions
   - Mitigation: Skip with warning, log errors

4. **Active process conflicts:** Cleaning while builds running
   - Mitigation: Detect lock files, warn users, dry-run mode

### User Experience Risks
1. **Information overload:** Too many scan categories
   - Mitigation: Keep --all as default, allow category filtering

2. **Accidental deletion:** Users delete needed caches
   - Mitigation: Dry-run default (already implemented), clear warnings

---

## Success Metrics

1. **Coverage:** Support 8-10 ecosystems (current: 4)
2. **Disk reclamation:** Average cleanup increases from ~30GB to 50-100GB
3. **User adoption:** Track which categories get most usage
4. **Safety:** Zero reports of data loss from incorrect deletions

---

## Competitive Differentiation

### Current Advantages
- Mac-first design (DevCleaner-style)
- Safe dry-run defaults
- Category filtering
- Size-sorted results

### Post-Implementation Advantages
- **Most comprehensive:** 10+ ecosystems vs competitors' 3-4
- **Mac + CLI + TUI:** DevCleaner (Mac GUI only), clean-dev-dirs (CLI only)
- **Ecosystem-specific smarts:** Detect active projects, preserve executables
- **Docker integration:** Unique among dev cleaners

---

## Unresolved Questions

1. **Should we support Windows/Linux paths?**
   - Current: Mac-only (~/Library paths)
   - Decision: Keep Mac focus or expand scope?

2. **Database caches - too niche?**
   - PostgreSQL/Redis are dev tools but less universal
   - Decision: Phase 4 or skip entirely?

3. **TUI roadmap integration?**
   - New categories need TUI design
   - Decision: CLI first, TUI after? Or parallel?

4. **Config file for custom paths?**
   - Users may have non-standard cache locations
   - Decision: Priority vs fixed paths with env var support?

---

## Sources

- [clean-dev-dirs (Rust)](https://github.com/TomPlanche/clean-dev-dirs)
- [devclean (GUI)](https://github.com/HuakunShen/devclean)
- [How to Clean Go - Leapcell](https://leapcell.io/blog/how-to-clean-go-a-guide-to-keeping-your-go-environment-tidy)
- [vyrti/cleaner](https://github.com/vyrti/cleaner)
- [Docker Cache Management - Blacksmith](https://www.blacksmith.sh/blog/a-guide-to-disk-space-management-with-docker-how-to-clear-your-cache)
- [Mac Developer Cleanup - Indie Hackers](https://www.indiehackers.com/post/how-i-cleaned-up-space-on-my-mac-as-a-developer-4905133a31)
- [Mac Cleanup Guide - Bomberbot](https://www.bomberbot.com/git/maximizing-disk-space-on-your-mac-the-ultimate-cleanup-guide-for-developers/)
- [Homebrew/Docker Cleanup - David Manske](https://davidemanske.com/free-up-storage-on-mac-from-homebrew-and-docker-images/)
- [MacOS Disk Space - Pawel Urbanek](https://pawelurbanek.com/macos-free-disk-space)
- [Docker Clear Cache - Depot](https://depot.dev/blog/docker-clear-cache)
- [Docker Cache Optimization - CyberPanel](https://cyberpanel.net/blog/clear-docker-cache)
- [Reclaim Docker Space - Medium](https://medium.com/@alexeysamoshkin/reclaim-disk-space-by-removing-stale-and-unused-docker-data-a4c3bd1e4001)
- [Cargo Cache Cleaning - Rust Blog](https://blog.rust-lang.org/2023/12/11/cargo-cache-cleaning/)
- [cargo clean - Cargo Book](https://doc.rust-lang.org/cargo/commands/cargo-clean.html)
- [Cargo Cache Issues - GitHub](https://github.com/rust-lang/cargo/issues/5885)
- [cargo-cache Tool](https://github.com/matthiaskrgr/cargo-cache)
- [Freeing Gigabytes - thisDaveJ](https://thisdavej.com/freeing-up-gigabytes-reclaiming-disk-space-from-rust-cargo-builds/)
- [Cleaning Up Rust - Heath Stewart](https://heaths.dev/rust/2025/03/01/cleaning-up-rust.html)
- [pip cache - pip docs](https://pip.pypa.io/en/stable/cli/pip_cache/)
- [Poetry Cache Config](https://python-poetry.org/docs/configuration/)
- [uv Python Package Manager](https://astral.sh/blog/uv)
- [pip Cache Management - IT trip](https://en.ittrip.xyz/python/pip-cache-management)
- [pyclean Tool](https://pypi.org/project/pyclean/)
- [Understanding Xcode Space - Kodeco](https://www.kodeco.com/19998365-understanding-and-managing-xcode-space/page/3)
- [Xcode Quick Fix - Developer Insider](https://developerinsider.co/clean-xcode-cache-quick-fix/)
- [Reduce Xcode Space - Medium](https://medium.com/@aykutkardes/understanding-and-fast-way-to-reduce-xcode-space-75e07acf6b1e)
- [Cleaning Xcode Clutter](https://www.medevel.com/cleaning-xcode-clutter-best-free-tools/)
- [Reset Xcode - GitHub Gist](https://gist.github.com/maciekish/66b6deaa7bc979d0a16c50784e16d697)
- [Swift Package Manager Caching - Uptech](https://www.uptech.team/blog/swift-package-manager)
- [Cleaning Old Xcode Files - Caesar Wirth](https://cjwirth.com/tech/cleaning-up-old-xcode-files)
- [Clear CocoaPods Cache - GitHub Gist](https://gist.github.com/mbinna/4202236)
- [Clearing Maven Cache - Baeldung](https://www.baeldung.com/maven-clear-cache)
- [Maven Cache Cleanup - CloudBees](https://docs.cloudbees.com/docs/cloudbees-ci-kb/latest/troubleshooting-guides/how-to-clean-up-maven-cache)
- [Maven Cache - Cloudsmith](https://support.cloudsmith.com/hc/en-us/articles/18028180425873-How-can-I-clear-my-Maven-cache)
- [Maven Cache - GeeksforGeeks](https://www.geeksforgeeks.org/advance-java/clearing-the-maven-cache/)
- [Maven Build Cache Extension](https://maven.apache.org/extensions/maven-build-cache-extension/)
- [Maven Purge Dependencies - Intertech](https://www.intertech.com/maven-purge-old-dependencies-from-local-repository/)
- [Clear Maven Cache Mac](https://devwithus.com/clear-maven-cache-mac/)
