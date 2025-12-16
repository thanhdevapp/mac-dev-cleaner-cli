# Mac Dev Cleaner - Phase 2: NCDU-Style TUI

> **Date:** 2025-12-15
> **Status:** ✅ Completed
> **Tech Stack:** Go 1.21+ | Bubble Tea v0.25+ | Bubbles v0.18+ | Lipgloss
> **Depends On:** [MVP Plan](../2025-12-15-mac-dev-cleaner-mvp/plan.md) (Completed)

---

## Overview

Interactive TUI for Mac Dev Cleaner using Bubble Tea framework. Implements NCDU-style keyboard navigation, multi-select with checkboxes, progress visualization, and multiple confirmation modes for safe deletion.

**Key Features:**
- Vim-style navigation (j/k) + arrow keys
- Multi-select with Space, select all (a/A)
- Visual checkboxes `[x]` / `[ ]`
- 4 confirmation modes: Safe, Interactive, Dry-run, Force
- Real-time progress bars for scan/delete
- NCDU-proven UX patterns

---

## Architecture

```
State Machine:
  Scanning --> Selecting --> Confirming --> Deleting --> Done
      |           |              |              |
      v           v              v              v
   spinner     list+kbd      modal/inline   progress

TUI Package Structure:
  internal/tui/
    ├── app.go           # Main Bubble Tea model
    ├── state.go         # State machine types
    ├── keys.go          # Key bindings
    ├── list.go          # List component + delegate
    ├── confirm.go       # Confirmation dialogs
    ├── progress.go      # Progress bar wrapper
    └── styles.go        # Lipgloss styles
```

---

## Implementation Phases

| Phase | Name                     | Priority | Status | File                             |
| ----- | ------------------------ | -------- | ------ | -------------------------------- |
| 01    | Bubble Tea Setup         | P0       | ✅ Done | (Already in tui.go from MVP)     |
| 02    | List + Checkboxes        | P0       | ✅ Done | (Already in tui.go from MVP)     |
| 03    | Keyboard Shortcuts       | P0       | ✅ Done | (j/k, space, a, n, enter, q)     |
| 04    | Confirmation Dialogs     | P0       | ✅ Done | State machine + y/n confirmation |
| 05    | Progress + Visual Polish | P1       | ✅ Done | progress.Model + styled views    |
| 06    | Integration + Testing    | P1       | ✅ Done | Build + tests passing            |

---

## CLI Integration

```bash
# New TUI mode (this plan)
dev-cleaner tui              # Full interactive TUI
dev-cleaner clean --tui      # TUI selection, then delete

# Existing CLI (from MVP)
dev-cleaner scan
dev-cleaner clean --confirm
```

---

## Research References

- [Bubble Tea Framework](./research/researcher-bubbletea-framework.md)
- [NCDU Patterns](./research/researcher-ncdu-patterns.md)

---

## Dependencies

```go
// New dependencies for TUI
github.com/charmbracelet/bubbletea v0.25+
github.com/charmbracelet/bubbles v0.18+
github.com/charmbracelet/lipgloss v0.10+

// Existing from MVP
github.com/spf13/cobra v1.8.0
github.com/dustin/go-humanize v1.0.1
```

---

## Success Criteria

- [ ] TUI launches <500ms
- [ ] Scan results display <2s for 100 items
- [ ] Keyboard response <50ms (no lag)
- [ ] All 4 confirmation modes work
- [ ] Zero accidental deletions
- [ ] Works in 80x24 terminal minimum
