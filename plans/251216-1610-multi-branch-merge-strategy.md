# Multi-Branch Merge Strategy

**Date:** 2025-12-16
**Branch Context:** feat/wails-v2-migration → dev-mvp
**Strategy:** Sequential merge approach

---

## Overview

Consolidate parallel development from multiple feature branches into dev-mvp:
1. **flutter-cleanup-support**: Multi-ecosystem scanner support (Flutter, Go, Python, Rust, Homebrew, Docker, Java/Kotlin)
2. **wails-v2-migration**: Wails v2 GUI + React Native support
3. **Cleanup**: Remove duplicate branches (feat/wails-gui, feature/react-native-support)

---

## Pre-Merge Checklist

### Current Branch Testing (feat/wails-v2-migration)
- [ ] Run all tests: `go test ./...`
- [ ] Build successfully: `go build`
- [ ] Test Wails GUI: `wails dev`
- [ ] Verify React Native scanner works
- [ ] Check no breaking changes in TUI
- [ ] Review git status for uncommitted changes

### Backup Current State
```bash
# Create backup tags
git tag backup/wails-v2-migration-pre-merge
git tag backup/flutter-cleanup-pre-merge origin/feature/flutter-cleanup-support
git tag backup/dev-mvp-pre-merge dev-mvp
git push origin --tags
```

### Environment Preparation
- [ ] Working directory clean: `git status`
- [ ] All changes committed
- [ ] Latest changes pulled: `git pull origin dev-mvp`
- [ ] Stash any WIP: `git stash save "wip-backup"`

---

## Merge Strategy: Sequential Integration

### Phase 1: Merge flutter-cleanup-support → dev-mvp

**Rationale:** Backend scanner features first, less conflict risk

```bash
# Switch to dev-mvp
git checkout dev-mvp
git pull origin dev-mvp

# Merge flutter-cleanup
git merge origin/feature/flutter-cleanup-support --no-ff -m "merge: Integrate multi-ecosystem scanner support (Flutter, Go, Python, Rust, Homebrew, Docker, Java/Kotlin)"
```

**Expected Conflicts:**
- `cmd/root/scan.go`: Scanner type definitions
- `cmd/root/clean.go`: Clean command options
- `internal/scanner/scanner.go`: Scanner interface
- `README.md`: Feature documentation
- `.gitignore`: Plans/docs directories

**Conflict Resolution:**
1. **Scanner files** (`internal/scanner/*`):
   - Keep all new scanner implementations (docker.go, flutter.go, golang.go, etc.)
   - Merge scanner.go to include all scanner types

2. **Command files** (`cmd/root/*`):
   - Merge flag definitions from both branches
   - Keep all scanner type options
   - Preserve command structure

3. **README.md**:
   - Combine feature lists
   - Keep latest version info
   - Merge installation instructions

4. **.gitignore**:
   - Accept both: plans/ and docs/ directories

**Verification:**
```bash
# Build and test
go mod tidy
go build
./dev-cleaner scan --help  # Verify all scanner types listed
./dev-cleaner scan --type flutter ~/test-project
./dev-cleaner scan --type go ~/test-project
./dev-cleaner scan --type python ~/test-project

# Run tests
go test ./...
```

**Commit if successful:**
```bash
git add .
git commit --amend --no-edit  # If conflicts resolved
git push origin dev-mvp
```

---

### Phase 2: Merge wails-v2-migration → dev-mvp

**Rationale:** Add GUI layer on top of complete backend

```bash
# Ensure Phase 1 complete
git checkout dev-mvp
git pull origin dev-mvp

# Merge wails-v2
git merge feat/wails-v2-migration --no-ff -m "merge: Integrate Wails v2 GUI with React Native support"
```

**Expected Conflicts:**
- `cmd/root/scan.go`: React Native scanner already in Phase 1
- `README.md`: Documentation updates
- `frontend/*`: Should be no conflicts (new directory)
- `app.go`: Wails-specific file

**Conflict Resolution:**
1. **cmd/root/scan.go**:
   - If React Native scanner conflicts, verify it's in both
   - Keep unified version with proper imports

2. **README.md**:
   - Combine both feature lists (backend + GUI)
   - Keep latest screenshots
   - Merge installation instructions for both CLI and GUI

3. **Frontend files**:
   - Should accept all (new directory from wails-v2)

**Verification:**
```bash
# Build CLI
go mod tidy
go build

# Test CLI with all scanners
./dev-cleaner scan --type react-native ~/test-project
./dev-cleaner scan --type flutter ~/test-project

# Build and test GUI
cd frontend && npm install
cd .. && wails dev

# Verify GUI features:
# - Scan button works
# - All scanner types available
# - Clean dialog functions
# - Settings accessible
```

**Commit if successful:**
```bash
git add .
git commit --amend --no-edit  # If conflicts resolved
git push origin dev-mvp
```

---

### Phase 3: Cleanup Duplicate Branches

```bash
# Delete local duplicate branches
git branch -d feat/wails-gui
git branch -d feature/react-native-support

# Delete remote if pushed
git push origin --delete feat/wails-gui 2>/dev/null || true

# Clean up merged branches
git remote prune origin

# Verify branch list
git branch -a
```

---

## Conflict Resolution Patterns

### Pattern 1: Scanner Type Additions
**Location:** `cmd/root/scan.go`, `pkg/types/types.go`

```go
// If both branches add scanner types
const (
    ScannerTypeNodeModules  ScannerType = "node_modules"
    ScannerTypeFlutter      ScannerType = "flutter"      // Phase 1
    ScannerTypeGo           ScannerType = "go"           // Phase 1
    ScannerTypePython       ScannerType = "python"       // Phase 1
    ScannerTypeReactNative  ScannerType = "react-native" // Phase 2
    // Keep all, merge alphabetically
)
```

### Pattern 2: README Feature Lists
**Location:** `README.md`

```markdown
## Features
- Multi-ecosystem support:
  - Node.js (node_modules)
  - Flutter/Dart (.dart_tool, build/)
  - Go (vendor/, pkg/mod cache)
  - Python (venv/, __pycache__)
  - React Native (.gradle, node_modules)
  - ... (combine all from both)
- TUI interface (CLI)
- GUI interface (Wails v2)  <!-- Phase 2 addition -->
```

### Pattern 3: Import Conflicts
**Location:** Various Go files

```go
// If import lists conflict, merge and sort
import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/your/project/internal/scanner"  // Ensure all needed
    "github.com/your/project/pkg/types"
)
```

---

## Verification Checklist

### Post-Phase 1 (flutter-cleanup merged)
- [ ] Build succeeds: `go build`
- [ ] All tests pass: `go test ./...`
- [ ] Scanner types complete: `./dev-cleaner scan --help`
- [ ] Each scanner works:
  - [ ] Flutter: `./dev-cleaner scan --type flutter ~/test`
  - [ ] Go: `./dev-cleaner scan --type go ~/test`
  - [ ] Python: `./dev-cleaner scan --type python ~/test`
  - [ ] Rust: `./dev-cleaner scan --type rust ~/test`
  - [ ] Homebrew: `./dev-cleaner scan --type homebrew`
  - [ ] Docker: `./dev-cleaner scan --type docker`
  - [ ] Java: `./dev-cleaner scan --type java ~/test`
- [ ] TUI functions properly
- [ ] Clean command works: `./dev-cleaner clean --help`

### Post-Phase 2 (wails-v2 merged)
- [ ] Build succeeds: `go build`
- [ ] All tests pass: `go test ./...`
- [ ] React Native scanner works: `./dev-cleaner scan --type react-native ~/test`
- [ ] Frontend builds: `cd frontend && npm run build`
- [ ] GUI launches: `wails dev`
- [ ] GUI features:
  - [ ] Scan runs successfully
  - [ ] Results display correctly
  - [ ] Clean dialog opens
  - [ ] Settings accessible
  - [ ] All scanner types in dropdown
- [ ] CLI still works independently
- [ ] No regression in existing features

### Post-Phase 3 (Cleanup)
- [ ] Duplicate branches deleted: `git branch -a | grep -E "(wails-gui|react-native-support)"`
- [ ] Remote refs cleaned: `git remote prune origin`
- [ ] Tags created: `git tag | grep backup`

---

## Rollback Procedures

### If Phase 1 fails:
```bash
git merge --abort
git reset --hard backup/dev-mvp-pre-merge
git push origin dev-mvp --force-with-lease
```

### If Phase 2 fails:
```bash
git merge --abort
# Phase 1 changes preserved
# Only wails-v2 merge aborted
```

### Complete rollback:
```bash
# Restore all to pre-merge state
git checkout dev-mvp
git reset --hard backup/dev-mvp-pre-merge
git push origin dev-mvp --force-with-lease

# Restore feature branches if needed
git checkout feat/wails-v2-migration
git reset --hard backup/wails-v2-migration-pre-merge
```

---

## Post-Merge Tasks

### Documentation Updates
- [ ] Update `README.md` with complete feature list
- [ ] Update `docs/codebase-summary.md`
- [ ] Add migration notes to `docs/` if needed
- [ ] Update `CHANGELOG.md` (if exists)

### Testing
- [ ] Run full test suite: `go test ./... -v`
- [ ] Integration tests with real projects
- [ ] Performance benchmarks
- [ ] GUI smoke tests

### Communication
- [ ] Update PR/issue status
- [ ] Notify team of merged features
- [ ] Document known issues

### Cleanup
- [ ] Remove backup tags: `git tag -d backup/*`
- [ ] Archive old plans if needed
- [ ] Update project board/tracker

---

## Expected Timeline

**Phase 1:** 30-60 min (conflicts + testing)
**Phase 2:** 30-60 min (conflicts + GUI testing)
**Phase 3:** 5-10 min (cleanup)
**Total:** ~2 hours with thorough testing

---

## Emergency Contacts

If major issues arise:
1. Abort merge immediately
2. Review conflict sections
3. Consider alternative strategies:
   - Cherry-pick specific commits
   - Rebase instead of merge
   - Create integration branch for testing

---

## Notes

- Keep terminal output logs during merge
- Screenshot GUI before/after Phase 2
- Test on clean Go module cache
- Verify all scanner imports resolve
- Check for duplicate function definitions
- Ensure version numbers consistent

---

## Success Criteria

✅ All scanners functional (10 types total)
✅ GUI launches and operates
✅ CLI maintains full functionality
✅ No test failures
✅ No build errors
✅ Documentation updated
✅ Duplicate branches removed

---

## References

- Branch graph: `git log --oneline --graph --all --decorate`
- Diff preview: `git diff dev-mvp origin/feature/flutter-cleanup-support`
- Conflict markers: `git diff --name-only --diff-filter=U`
