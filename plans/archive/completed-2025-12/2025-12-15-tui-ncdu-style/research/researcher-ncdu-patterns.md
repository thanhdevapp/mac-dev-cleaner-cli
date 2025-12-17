# Research Report: NCDU UX Patterns & TUI Best Practices

**Date:** 2025-12-15 | **Research Type:** UX/UI Pattern Analysis | **Target:** File Management TUI Design

## Executive Summary

NCDU (NCurses Disk Usage) represents a mature, proven model for intuitive file management TUI interaction. Core patterns: vim-style navigation (hjkl), space-based selection, single-keystroke actions, and minimal visual clutter. For mac-dev-cleaner TUI implementation with BubbleTea, adopt NCDU's modal-less flat hierarchy, vim keys, visual feedback via colors/symbols, and state-driven confirmation flows. Key takeaway: **keyboard-first, single-action-per-key, modal confirmations for destructive ops**.

## NCDU Keyboard Patterns (Battle-Tested)

### Navigation Shortcuts
| Key | Action | Note |
|-----|--------|------|
| `↑/k` | Move up | Vim-style alternative |
| `↓/j` | Move down | Consistent vim mappings |
| `→/Enter` | Open directory | Recurse into selection |
| `←/h` | Go to parent | Pop directory level |
| `q` | Quit | Exit immediately (no confirmation) |
| `/` | Search/filter | Enter search mode |
| `g` | Toggle graph | Show/hide bar chart visualization |
| `a` | Toggle apparent size | Switch between apparent/actual disk usage |
| `e` | Show excluded files | Toggle hidden files visibility |
| `i` | Show item info | Display metadata for selection |
| `r` | Recalculate | Rescan current directory |

### Selection & Deletion
- `Space` - Toggle item selection (no state change to cursor)
- `d` - Delete selected (triggers confirmation prompt: "Delete? (y/n)")
- `n` - Sort by name (toggle asc/desc)
- `s` - Sort by size (toggle asc/desc)
- `t` - Toggle dirs-first in sort

**Key insight:** Destructive ops (deletion) always prompt; non-destructive navigation is immediate.

## File Selection UI Patterns (TUI Standard)

### Multi-Selection Conventions
1. **Toggle selection** - Space bar without moving cursor position
2. **Range selection** - Shift+Arrow to extend from "anchor" item to current
3. **Select all** - Ctrl+A for bulk operations
4. **Visual indicators** - Square brackets `[ ]` or `[x]` prefix per item

### Visual Feedback Hierarchy
```
Not selected:    Item Name
Selected:        [x] Item Name
Focused:         > [x] Item Name    (cursor indicator)
Hovered:         (inverse video or highlight color)
```

**Best practice:** Use *color + symbol* not just color (accessibility for colorblind users).

## Confirmation Dialog Design

### Pattern 1: Inline Prompt (NCDU approach)
```
Delete /path/to/item (1.5 GB)? (y/n) _
```
- Appears at status line
- Wait for single-key response
- No escape hatch needed (unlike modal)
- Intuitive for power users

### Pattern 2: Modal Box (BubbleTea-friendly)
```
┌─────────────────────────┐
│ Confirm Deletion        │
├─────────────────────────┤
│ Delete 5 items (2.3GB)? │
│                         │
│ [Yes] [No]              │
└─────────────────────────┘
```
Advantage: Explicit buttons, visual weight signals importance. Disadvantage: screen real estate.

### Pattern 3: State Machine with Visual Feedback
```
State 1: Show selection count → "Delete 3 items? Press Y/N"
State 2: Once Y pressed → "Deleting... [████░░░░░░] 40%"
State 3: Complete → "Deleted 2.1 GB ✓"
```

## Progress & Error Handling

### Long-Running Operations
- Use indeterminate spinner for unpredictable duration: `⠋ Scanning...`
- Use progress bar for known-total scans: `[████████░░░░░░░░░░] 42/100`
- Update frequency: 100-200ms refresh (not per-file)

### Error Feedback
- **Non-blocking errors:** Show in status line with color (red), auto-dismiss after 3s
  ```
  ✗ Failed to delete /path/file: Permission denied
  ```
- **Blocking errors:** Modal prompt requiring acknowledgement
  ```
  ✗ Error scanning /Volumes/External: Device not accessible [OK]
  ```

### Success Confirmation
- Brief success message with count: `✓ Cleaned 7.2 GB from 23 items`
- Auto-dismiss after 2s or on next key

## BubbleTea Implementation Patterns

### File Picker Component
```go
// Use charmbracelet/bubbles filepicker
filepicker.Model{
  Path: "~",
  AllowedTypes: []string{}, // Empty = all
  DirAllowed: true,
  FileAllowed: false,        // For this cleaner, only dirs
  ShowHidden: true,
  ShowPermissions: true,
  ShowSize: true,
}
```

### Confirmation Dialog as Submodel
```go
type Model struct {
  state State // scanning | selecting | confirming | deleting | done

  // Confirmation state
  confirmMsg string
  selectedItems []string
  confirmed bool
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  switch msg := msg.(type) {
  case tea.KeyMsg:
    if m.state == "confirming" {
      switch msg.String() {
      case "y", "Y": m.confirmed = true
      case "n", "N", "esc": m.confirmed = false
      }
    }
  }
}
```

### Visual State Rendering
```go
func (m Model) View() string {
  switch m.state {
  case "scanning":
    return fmt.Sprintf("⠋ Scanning... %d items found\n", m.itemCount)
  case "confirming":
    return fmt.Sprintf("%s\nDelete? (y/n) ", m.confirmMsg)
  case "deleting":
    return fmt.Sprintf("Deleting... [%s] %.0f%%\n", m.progressBar, m.percent)
  }
}
```

## Color & Symbol Reference (Accessibility)

| Intent | Color | Symbol | Example |
|--------|-------|--------|---------|
| Success | Green | ✓ | `✓ Cleaned 2.1 GB` |
| Error | Red | ✗ | `✗ Permission denied` |
| Warning | Yellow | ⚠ | `⚠ 500 files skipped` |
| Progress | Blue | ⠋ | `⠋ Scanning...` |
| Action | Cyan | → | `→ Select: Space` |
| Neutral | Gray | ◌ | `◌ Xcode (7.4 GB)` |

## Key Takeaways for Implementation

1. **Navigation first:** Map arrow keys + vim keys (hjkl) before mouse/tab
2. **Flat hierarchy:** No nested menus; use forward/back arrows only
3. **Single keystrokes:** No chording (Ctrl+X) except Ctrl+A for select-all
4. **Stateful confirmation:** Modal or inline prompt for destructive ops
5. **Visual + symbolic:** Never rely on color alone
6. **Immediate feedback:** Show result within 100ms of input
7. **Minimal chrome:** Status line for messages, no decorative borders
8. **Exit safety:** Optional confirmation on Ctrl+C only if selections exist

## References

- [NCDU Official](https://dev.yorhel.nl/ncdu)
- [Linux NCDU Manual](https://linux.die.net/man/1/ncdu)
- [TecMint NCDU Guide](https://www.tecmint.com/ncdu-a-ncurses-based-disk-usage-analyzer-and-tracker/)
- [BubbleTea GitHub](https://github.com/charmbracelet/bubbletea)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [TUI Best Practices - LogRocket](https://blog.logrocket.com/7-tui-libraries-interactive-terminal-apps/)
- [Keyboard Navigation Patterns - UXPin](https://www.uxpin.com/studio/blog/keyboard-navigation-patterns-complex-widgets/)
- [File Manager Selection Patterns - Syncfusion](https://ej2.syncfusion.com/vue/documentation/file-manager/multiple-selection)

**Unresolved Questions:**
- Should confirmation default to "no" or require explicit "yes"? (NCDU: implicit yes on 'y', no on 'n')
- Support mouse clicks for casual users or keyboard-only? (Recommendation: keyboard-first, mouse optional)
- How to handle Ctrl+C gracefully with unsaved selections? (Standard: prompt only if items selected)
