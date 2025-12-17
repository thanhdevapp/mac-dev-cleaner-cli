# NCDU-Style Hierarchical Navigation - Brainstorm Report

**Date:** 2025-12-15
**Topic:** Add NCDU-style hierarchical folder navigation to Mac Dev Cleaner CLI
**Status:** Architecture designed, ready for implementation

---

## Problem Statement

Current Mac Dev Cleaner architecture:
```
Scan All → Flat []ScanResult → TUI (flat list)
```

**Limitations:**
- No drill-down into folders
- Can't explore what's inside large directories before deleting
- Flat list doesn't show hierarchical relationships
- No way to navigate back to parent folders

**Goal:** Implement NCDU-style hierarchical navigation:
- Navigate into folders with →/Enter
- Go back to parent with ←/h
- Refresh/rescan current folder with 'r'
- Show folder size breakdown at each level

---

## Requirements (User Answers)

1. **Scope:** Full NCDU - Navigate all scan results hierarchically
2. **Scan strategy:** Lazy - Scan folders only when user navigates into them
3. **Primary motivation:** Better UX - more intuitive folder exploration
4. **Architecture preference:** Keep lazy - Faster startup matters more
5. **Depth limit:** Smart limit - Allow deep navigation but warn/paginate after depth 5

---

## Research Findings

### NCDU Architecture (Actual Implementation)

**Key Discovery:** NCDU does **NOT** use lazy loading. It pre-scans everything upfront.

From [NCDU Manual](https://dev.yorhel.nl/ncdu/man) and [NCDU Architecture](https://blog.csdn.net/qq_62784677/article/details/147313969):

- **Scanning:** Depth-first search (DFS) using opendir(), readdir(), stat()
- **Pre-scan:** Entire tree scanned before UI shows (not lazy)
- **Parallel scanning:** v2.5+ supports `-t8` flag (8 threads)
- **Async rendering:** Shows progress while scanning
- **Memory optimization:** v2.6+ binary export format for massive trees

**NCDU's approach:**
```
Scan Everything (DFS) → Build Full Tree → Fast Navigation
```

### Bubble Tea Ecosystem

From [Bubble Tea GitHub](https://github.com/charmbracelet/bubbletea) and [tree-bubble](https://github.com/savannahostrowski/tree-bubble):

- **tree-bubble:** TUI tree component for Bubble Tea (`github.com/savannahostrowski/tree-bubble`)
- **File picker example:** [bubbletea/examples/file-picker](https://github.com/charmbracelet/bubbletea/blob/main/examples/file-picker/main.go)
- **Architecture:** Elm-based model-update-view pattern
- **Tree rendering:** "Root model receives all messages, relayed down tree to child models"

**No ready-made lazy-loading tree solutions found** - need custom implementation.

---

## Evaluated Approaches

### Approach 1: True NCDU Clone (Pre-scan + Hierarchical Tree)

**Architecture:**
```
Scanner → Build FileTree → TUI renders tree → Navigate
```

**Flow:**
1. Scan all categories (Xcode/Android/Node) upfront with parallel threads
2. Build hierarchical FileTree structure
3. TUI renders current level, tracks navigation stack
4. →/Enter: Push to stack, show children (instant - pre-computed)
5. ←/h: Pop stack, show parent

**Data Structure:**
```go
type FileTree struct {
    Root     *FileNode
    Current  *FileNode
    Stack    []*FileNode // Breadcrumb trail
}

type FileNode struct {
    Path      string
    Name      string
    Size      int64
    IsDir     bool
    Type      CleanTargetType
    Children  map[string]*FileNode
    Parent    *FileNode
}
```

**Pros:**
- ✅ True NCDU experience
- ✅ Instant navigation (pre-computed tree)
- ✅ Works with existing parallel scanner (go routines already in scanner.go:43-73)
- ✅ Accurate size calculations (aggregated from children)

**Cons:**
- ❌ Slow startup for massive projects (scan all before showing TUI)
- ❌ High memory for millions of files
- ❌ Major refactor: New FileTree structure, rewrite TUI navigation
- ❌ Over-engineered for dev cleaner use case (not analyzing entire disk)

**Complexity:** HIGH (7-10 days)
**YAGNI Violation:** Scans everything when users explore ~10% of results

---

### Approach 2: Hybrid Lazy (Scan targets → Lazy drill-down) **[RECOMMENDED]**

**Architecture:**
```
Scanner → Top-level targets → TUI flat list → Lazy scan on drill-down
```

**Flow:**
1. **Initial scan:** Scan top-level targets only (current behavior - fast)
   - `~/Library/Developer/Xcode/DerivedData/*`
   - `~/.gradle/caches/*`
   - `node_modules` folders in known project dirs
2. **TUI State:** Show flat list (current StateSelecting)
3. **Enter on directory item:**
   - Switch to StateTree
   - Lazy scan: Call `ScanDirectory(path)` on-demand
   - Build TreeNode for current folder
   - Show children
4. **Tree navigation:**
   - →/Enter: Scan children if `node.Children == nil`, drill down
   - ←/h: Go back to parent in stack
   - r: Clear `node.Children`, rescan
5. **Smart depth limit:**
   - Track current depth
   - After depth 5: Show warning "Deep tree detected. Press 'c' to continue or 'b' to go back"
   - Paginate children if >100 items

**Data Structure:**
```go
// New: TreeNode for lazy navigation
type TreeNode struct {
    Path      string
    Name      string
    Size      int64
    IsDir     bool
    Type      CleanTargetType
    Children  []*TreeNode // nil = not scanned yet
    Scanned   bool
    Depth     int
}

// TUI states (extend existing)
const (
    StateSelecting  State = iota // Current: flat list view
    StateConfirming              // Current: confirmation dialog
    StateDeleting                // Current: deleting items
    StateDone                    // Current: operation complete
    StateTree                    // NEW: tree navigation view
)

// TUI Model additions
type Model struct {
    // ... existing fields ...

    // Tree navigation state
    currentNode  *TreeNode
    nodeStack    []*TreeNode  // Breadcrumb trail
    treeView     bool         // true when in StateTree
}
```

**New Scanner Methods:**
```go
// Add to scanner/scanner.go
func (s *Scanner) ScanDirectory(path string, maxDepth int) (*TreeNode, error) {
    // Lazy scan single directory, return tree node with children
}

func (s *Scanner) CalculateSize(path string) (int64, error) {
    // Fast size calc without full tree (for lazy nodes)
}
```

**New TUI Key Bindings:**
```go
// In StateList (existing):
Enter on dir item → Switch to StateTree, lazy scan children

// In StateTree (new):
→/Enter → Drill into folder (scan children if needed)
←/h     → Go back to parent
r       → Refresh/rescan current node
Esc     → Return to StateList (flat list)
Space   → Toggle selection (works in tree too)
d       → Delete selected node and descendants
```

**Pros:**
- ✅ Fast startup (scan only known targets - current speed)
- ✅ True lazy loading (scan on-demand)
- ✅ Minimal scanner refactor (add `ScanDirectory()`)
- ✅ Best UX: Quick results + deep exploration when needed
- ✅ Memory efficient (only load explored branches)
- ✅ Smart depth limits prevent scanning 20+ levels accidentally

**Cons:**
- ⚠️ Slower navigation first time (scan on each drill-down)
- ⚠️ Complex state management (flat list + tree mode)
- ⚠️ Need caching strategy for rescanned folders

**Complexity:** MEDIUM (4-6 days)
**KISS:** Solves actual user need without over-engineering

---

### Approach 3: Fake NCDU (Virtual Tree from Flat Results)

**Architecture:**
```
Scanner → Flat results → TUI groups by path hierarchy
```

**Flow:**
1. Scan all (current flow - upfront)
2. Group `[]ScanResult` by path prefix
3. TUI renders virtual tree (no real tree structure)
4. Navigate by filtering/grouping results

**Example:**
```go
// Virtual grouping
Results:
  /Users/me/Projects/app1/node_modules
  /Users/me/Projects/app2/node_modules
  /Users/me/Library/Xcode/DerivedData/App-xxx

Grouped view at /Users/me:
  Projects/ (2 items)
  Library/ (1 item)

Enter on "Projects/" → Filter results by prefix "/Users/me/Projects"
```

**Pros:**
- ✅ No scanner refactor
- ✅ Fast navigation (in-memory grouping)
- ✅ Smallest code change

**Cons:**
- ❌ Not true tree (can't explore arbitrary subdirs)
- ❌ Limited to originally scanned paths
- ❌ Fake UX (users expect real exploration, get filtered list)
- ❌ Size calculations wrong (grouping != actual folder sizes)

**Complexity:** LOW (2-3 days)
**Rejected:** Too limited, doesn't meet "true exploration" requirement

---

## Final Recommendation: Approach 2 (Hybrid Lazy)

**Why?**

1. **Matches requirements:**
   - ✅ Full hierarchical navigation
   - ✅ Lazy scanning (fast startup)
   - ✅ Better UX (intuitive exploration)
   - ✅ Smart depth limits

2. **YAGNI compliant:**
   - Doesn't scan everything upfront (users explore ~10% of results)
   - No over-engineered full tree structure
   - Implements only what's needed

3. **KISS:**
   - Extends existing architecture (no rewrite)
   - Clear state separation (flat list vs tree)
   - Simple lazy loading (scan on Enter)

4. **DRY:**
   - Reuses existing scanner methods (`calculateSize`, `PathExists`)
   - Reuses TUI components (spinner, progress, styles)
   - Reuses key bindings (↑/↓/Space/Enter)

---

## Implementation Plan

### Phase 1: Data Structures (1 day)

**Files to create:**
- `pkg/types/tree.go` - TreeNode struct, navigation stack

**Changes:**
```go
// pkg/types/tree.go (NEW)
type TreeNode struct {
    Path      string
    Name      string
    Size      int64
    IsDir     bool
    Type      CleanTargetType
    Children  []*TreeNode
    Scanned   bool
    Depth     int
}

func (n *TreeNode) AddChild(child *TreeNode)
func (n *TreeNode) NeedsScanning() bool
func (n *TreeNode) HasChildren() bool
```

**Files to modify:**
- `internal/tui/tui.go:112-129` - Add tree state to Model

```go
type Model struct {
    // ... existing ...

    // Tree navigation
    treeMode     bool
    currentNode  *TreeNode
    nodeStack    []*TreeNode
    maxDepth     int // Default: 5
}
```

---

### Phase 2: Lazy Scanner (1-2 days)

**Files to modify:**
- `internal/scanner/scanner.go` - Add lazy scanning methods

**New methods:**
```go
// Lazy scan single directory
func (s *Scanner) ScanDirectory(path string, currentDepth int, maxDepth int) (*TreeNode, error) {
    if currentDepth >= maxDepth {
        return &TreeNode{/* truncated */}, ErrMaxDepthReached
    }

    // Read directory entries
    entries, err := os.ReadDir(path)

    // Build TreeNode with children
    node := &TreeNode{
        Path:     path,
        Children: make([]*TreeNode, 0),
        Scanned:  true,
        Depth:    currentDepth,
    }

    for _, entry := range entries {
        childPath := filepath.Join(path, entry.Name())
        childSize, _ := s.calculateSize(childPath)

        child := &TreeNode{
            Path:    childPath,
            Name:    entry.Name(),
            Size:    childSize,
            IsDir:   entry.IsDir(),
            Scanned: false, // Lazy - not scanned yet
            Depth:   currentDepth + 1,
        }
        node.AddChild(child)
    }

    return node, nil
}

// Convert ScanResult to TreeNode (for initial flat list → tree transition)
func (s *Scanner) ResultToTreeNode(result types.ScanResult) (*TreeNode, error)
```

**Edge cases:**
- Permission denied → Skip, log warning
- Symlinks → Detect cycles, skip if already visited
- Max depth reached → Return node with `Children: nil`, show "..." indicator

---

### Phase 3: TUI Tree View (2 days)

**Files to modify:**
- `internal/tui/tui.go` - Add StateTree rendering + navigation

**New state:**
```go
const (
    StateSelecting  State = iota
    StateConfirming
    StateDeleting
    StateDone
    StateTree       // NEW
)
```

**New key bindings (extend KeyMap:71-110):**
```go
type KeyMap struct {
    // ... existing ...
    DrillDown  key.Binding // Enter/→ in tree mode
    GoBack     key.Binding // ←/h in tree mode
    Refresh    key.Binding // r
    ExitTree   key.Binding // Esc
}

var keys = KeyMap{
    // ... existing ...
    DrillDown: key.NewBinding(
        key.WithKeys("enter", "right"),
        key.WithHelp("→/enter", "drill down"),
    ),
    GoBack: key.NewBinding(
        key.WithKeys("left", "h"),
        key.WithHelp("←/h", "go back"),
    ),
    Refresh: key.NewBinding(
        key.WithKeys("r"),
        key.WithHelp("r", "refresh"),
    ),
    ExitTree: key.NewBinding(
        key.WithKeys("esc"),
        key.WithHelp("esc", "exit tree"),
    ),
}
```

**Update() handler (extend tui.go:155-247):**
```go
case StateSelecting:
    switch {
    // ... existing navigation ...

    case key.Matches(msg, keys.Confirm):
        // NEW: Check if current item is directory
        if m.items[m.cursor].IsDir() {
            return m, m.enterTreeMode()
        }
        // ... existing confirm logic ...
    }

case StateTree: // NEW
    switch {
    case key.Matches(msg, keys.ExitTree):
        m.treeMode = false
        m.state = StateSelecting
        return m, nil

    case key.Matches(msg, keys.GoBack):
        if len(m.nodeStack) > 0 {
            m.currentNode = m.nodeStack[len(m.nodeStack)-1]
            m.nodeStack = m.nodeStack[:len(m.nodeStack)-1]
        }
        return m, nil

    case key.Matches(msg, keys.DrillDown):
        selectedNode := m.currentNode.Children[m.cursor]
        if !selectedNode.IsDir {
            return m, nil // Can't drill into file
        }

        // Check depth limit
        if selectedNode.Depth >= m.maxDepth {
            return m, m.showDepthWarning()
        }

        // Lazy scan if needed
        if !selectedNode.Scanned {
            return m, m.scanNode(selectedNode)
        }

        // Navigate
        m.nodeStack = append(m.nodeStack, m.currentNode)
        m.currentNode = selectedNode
        m.cursor = 0
        return m, nil

    case key.Matches(msg, keys.Refresh):
        return m, m.rescanNode(m.currentNode)

    case key.Matches(msg, keys.Up):
        if m.cursor > 0 { m.cursor-- }

    case key.Matches(msg, keys.Down):
        if m.cursor < len(m.currentNode.Children)-1 {
            m.cursor++
        }

    case key.Matches(msg, keys.Toggle):
        // Mark child for deletion
        m.selected[m.cursor] = !m.selected[m.cursor]
    }
```

**View() renderer (extend tui.go:282-317):**
```go
case StateTree:
    return m.renderTreeView(&b)
```

**New render function:**
```go
func (m Model) renderTreeView(b *strings.Builder) string {
    // Breadcrumb
    breadcrumb := m.buildBreadcrumb()
    b.WriteString(helpStyle.Render(breadcrumb))
    b.WriteString("\n\n")

    // Current folder info
    folderInfo := fmt.Sprintf("📁 %s (%s, %d items)",
        m.currentNode.Name,
        ui.FormatSize(m.currentNode.Size),
        len(m.currentNode.Children),
    )
    b.WriteString(titleStyle.Render(folderInfo))
    b.WriteString("\n\n")

    // Children list
    for i, child := range m.currentNode.Children {
        cursor := "  "
        if i == m.cursor {
            cursor = cursorStyle.Render("▸ ")
        }

        checkbox := "[ ]"
        if m.selected[i] {
            checkbox = checkboxStyle.Render("[✓]")
        }

        icon := "📄"
        if child.IsDir {
            if child.Scanned {
                icon = "📂"
            } else {
                icon = "📁" // Unopened folder
            }
        }

        sizeStr := ui.FormatSize(child.Size)

        line := fmt.Sprintf("%s%s %s %10s  %s",
            cursor,
            checkbox,
            icon,
            sizeStr,
            child.Name,
        )

        if i == m.cursor {
            b.WriteString(selectedItemStyle.Render(line))
        } else {
            b.WriteString(itemStyle.Render(line))
        }
        b.WriteString("\n")
    }

    // Depth warning if near limit
    if m.currentNode.Depth >= m.maxDepth - 1 {
        warning := fmt.Sprintf("⚠️  Depth %d/%d - Approaching limit",
            m.currentNode.Depth, m.maxDepth)
        b.WriteString(errorStyle.Render(warning))
        b.WriteString("\n")
    }

    // Help
    help := "\n↑/↓: Navigate • →/Enter: Drill down • ←/h: Go back • r: Refresh • Space: Toggle • Esc: Exit tree • q: Quit"
    b.WriteString(helpStyle.Render(help))

    return b.String()
}

func (m Model) buildBreadcrumb() string {
    parts := []string{}
    for _, node := range m.nodeStack {
        parts = append(parts, node.Name)
    }
    parts = append(parts, m.currentNode.Name)
    return "📍 " + strings.Join(parts, " > ")
}
```

**Async commands:**
```go
type scanNodeMsg struct {
    node *TreeNode
    err  error
}

func (m Model) scanNode(node *TreeNode) tea.Cmd {
    return func() tea.Msg {
        s, _ := scanner.New()
        scanned, err := s.ScanDirectory(node.Path, node.Depth, m.maxDepth)
        if err != nil {
            return scanNodeMsg{err: err}
        }
        node.Children = scanned.Children
        node.Scanned = true
        return scanNodeMsg{node: node}
    }
}

func (m Model) rescanNode(node *TreeNode) tea.Cmd {
    return func() tea.Msg {
        node.Scanned = false
        node.Children = nil
        return m.scanNode(node)
    }
}

func (m Model) enterTreeMode() tea.Cmd {
    return func() tea.Msg {
        // Convert current ScanResult to TreeNode
        item := m.items[m.cursor]
        s, _ := scanner.New()
        node, err := s.ResultToTreeNode(item)
        if err != nil {
            return scanNodeMsg{err: err}
        }

        // Scan children
        scanned, err := s.ScanDirectory(item.Path, 0, m.maxDepth)
        if err != nil {
            return scanNodeMsg{err: err}
        }

        return scanNodeMsg{node: scanned}
    }
}
```

---

### Phase 4: Testing + Polish (1-2 days)

**Test cases:**
1. **Lazy scanning:**
   - Enter folder → Verify children scanned only once
   - Go back → Verify cache preserved
   - Refresh → Verify rescan clears cache

2. **Depth limits:**
   - Navigate to depth 5 → Verify warning shown
   - Attempt depth 6 → Verify blocked (or paginated)

3. **Edge cases:**
   - Empty directory → Show "Empty folder"
   - Permission denied → Show "Access denied" error
   - Symlink loop → Detect, skip gracefully
   - Very large folder (1000+ items) → Paginate (show first 100)

4. **Selection in tree mode:**
   - Toggle item → Mark for deletion
   - Delete node → Remove all descendants
   - Verify selection persists when navigating

5. **Performance:**
   - Large folder (10K files) → Scan completes in <2s
   - Deep tree (10 levels) → Navigate without lag

**Polish:**
- Add loading spinner while scanning folder
- Show "Scanning..." status during lazy scan
- Add progress bar for large folders
- Colorize breadcrumb trail
- Add file type icons (using lipgloss)

---

## Implementation Risks

### Risk 1: Scan Performance on Large Folders
**Issue:** Scanning folders with 10K+ files may block TUI (no async scanning yet)

**Mitigation:**
- Async scanning: Return tea.Cmd from `scanNode()`
- Show spinner while scanning
- Timeout after 30s, show "Scan too large, press Enter to continue"

### Risk 2: Memory Usage for Deep Trees
**Issue:** Keeping all scanned nodes in memory could consume GBs for deep trees

**Mitigation:**
- Cache eviction: Clear `node.Children` after navigating away (keep only current + stack)
- Smart caching: Keep last 10 visited nodes
- Add flag `--tree-cache-size` to control memory

### Risk 3: Complex State Management
**Issue:** Managing flat list + tree state + navigation stack could cause bugs

**Mitigation:**
- Clear separation: StateSelecting (flat) vs StateTree (hierarchical)
- Single source of truth: `m.currentNode` for tree, `m.items` for list
- Unit tests for state transitions

### Risk 4: Size Calculations Wrong
**Issue:** Lazy scanning means parent size != sum(children sizes) initially

**Mitigation:**
- Recalculate on scan: After scanning children, update parent size
- Show "Calculating..." during size aggregation
- Cache sizes per node

---

## Success Metrics

1. **Startup performance:** Initial scan ≤ 3s (same as current)
2. **Navigation performance:** Drill-down scan ≤ 2s for folders <10K files
3. **Memory efficiency:** Use ≤ 100MB for typical tree (depth 5, 1K nodes)
4. **UX quality:** Users can explore 5+ levels deep without confusion
5. **Code quality:** TUI code ≤ 800 lines (current: 482, budget: +318)

---

## Next Steps

1. **Get approval:** Review this brainstorm report with user
2. **Create implementation plan:** Detailed task breakdown with file changes
3. **Phase 1 implementation:** Data structures + scanner methods
4. **Phase 2 implementation:** TUI tree view + navigation
5. **Testing:** Verify edge cases, performance, UX
6. **Documentation:** Update README with tree navigation features

---

## Unresolved Questions

1. **Selection behavior:** When user selects folder in tree mode, should it mark all descendants for deletion automatically? Or require explicit selection per file?

2. **Caching strategy:** Should rescanning a folder preserve selection state? (User selects 5 items, goes back, drills down again - selections preserved?)

3. **Size display:** Show aggregated size (all descendants) or direct size (files only in this folder)? NCDU shows aggregated - should we match?

4. **Escape key behavior:** Esc in tree mode - go back one level OR exit tree completely? (Proposed: Esc = exit tree, ← = back one level)

5. **Integration with clean:** How does deletion work in tree mode? Delete entire current node? Only selected children? Need confirmation dialog per-node?

---

## Sources

Research sources used in this brainstorm:

- [NCDU Manual](https://dev.yorhel.nl/ncdu/man)
- [How NCDU Works - CSDN](https://blog.csdn.net/qq_62784677/article/details/147313969)
- [Bubble Tea GitHub](https://github.com/charmbracelet/bubbletea)
- [tree-bubble: TUI Tree Component](https://github.com/savannahostrowski/tree-bubble)
- [Bubble Tea File Picker Example](https://github.com/charmbracelet/bubbletea/blob/main/examples/file-picker/main.go)
- [Tips for Building Bubble Tea Programs](https://leg100.github.io/en/posts/building-bubbletea-programs/)

---

**Report generated:** 2025-12-15
**Estimated implementation time:** 4-6 days
**Complexity:** Medium
**Recommendation:** Proceed with Approach 2 (Hybrid Lazy)
