# Mac Dev Cleaner CLI - MVP Implementation Plan

> **Date:** 2025-12-15
> **Status:** ✅ Completed
> **Tech Stack:** Go 1.21+ | Cobra | Lipgloss

---

## Overview

CLI tool to clean macOS development artifacts (Xcode DerivedData, Gradle caches, node_modules). Single binary, no runtime dependencies, Homebrew-distributable.

**MVP Scope (P0):**
- Scan predefined directories
- Human-readable size display
- Interactive selection
- Dry-run mode (default)
- Confirmation before deletion
- Safety validation

---

## Project Structure

```
mac-dev-cleaner/
├── main.go                 # Entry point
├── cmd/
│   ├── root.go            # Root command
│   ├── scan.go            # Scan command
│   └── clean.go           # Clean command
├── internal/
│   ├── scanner/
│   │   ├── scanner.go     # Core scanning logic
│   │   ├── xcode.go       # Xcode targets
│   │   ├── android.go     # Android targets
│   │   └── node.go        # Node.js targets
│   ├── cleaner/
│   │   ├── cleaner.go     # Delete operations
│   │   └── safety.go      # Path validation
│   └── ui/
│       └── formatter.go   # Lipgloss output formatting
├── pkg/types/
│   └── types.go           # Shared types
├── Makefile               # Build automation
├── go.mod
└── go.sum
```

---

## Implementation Phases

| Phase | Name                  | Status      | File                                                                 |
| ----- | --------------------- | ----------- | -------------------------------------------------------------------- |
| 01    | Project Setup         | ✅ Completed | [phase-01-project-setup.md](./phase-01-project-setup.md)             |
| 02    | Core Scanner          | ✅ Completed | [phase-02-core-scanner.md](./phase-02-core-scanner.md)               |
| 03    | CLI Commands          | ✅ Completed | [phase-03-cli-commands.md](./phase-03-cli-commands.md)               |
| 04    | Safety & Confirmation | ✅ Completed | [phase-04-safety-confirmation.md](./phase-04-safety-confirmation.md) |
| 05    | Testing & Polish      | ✅ Completed | [phase-05-testing-polish.md](./phase-05-testing-polish.md)           |

---

## MVP Commands

```bash
dev-cleaner scan                # Scan all targets
dev-cleaner scan --ios          # iOS/Xcode only
dev-cleaner scan --android      # Android/Gradle only
dev-cleaner scan --node         # Node.js only

dev-cleaner clean               # Interactive clean (dry-run default)
dev-cleaner clean --confirm     # Actually delete
dev-cleaner clean --ios --confirm
```

---

## Target Directories

| Type    | Path                                     | Description      |
| ------- | ---------------------------------------- | ---------------- |
| iOS     | `~/Library/Developer/Xcode/DerivedData/` | Build artifacts  |
| iOS     | `~/Library/Caches/com.apple.dt.Xcode/`   | Xcode caches     |
| Android | `~/.gradle/caches/`                      | Gradle caches    |
| Android | `~/.gradle/wrapper/`                     | Gradle wrappers  |
| Node    | `*/node_modules/` (depth 3)              | NPM dependencies |

---

## Success Criteria

- [ ] Scan completes <5s for ~100 projects
- [ ] Binary size <10MB
- [ ] Zero accidental deletions (dry-run default)
- [ ] All target directories detected
- [ ] Clear size reporting

---

## Research References

- [Go CLI & Cobra](./research/researcher-go-cli-cobra.md)
- [Filesystem Scanning](./research/researcher-filesystem-scanning.md)

---

## Dependencies

```go
github.com/spf13/cobra v1.8.0
github.com/dustin/go-humanize v1.0.1
```
