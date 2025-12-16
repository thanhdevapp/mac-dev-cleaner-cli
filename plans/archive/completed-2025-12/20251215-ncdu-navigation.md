# Implementation Plan: NCDU-Style Hierarchical Navigation

**Date:** 2025-12-15
**Feature:** Add NCDU-style hierarchical folder navigation
**Architecture:** Hybrid Lazy (Approach 2)
**Estimated Time:** 4-6 days
**Status:** Ready for implementation

---

## Overview

Transform Mac Dev Cleaner from flat list to hierarchical tree navigation:

**Current:**
```
Scan → Flat list → Select → Delete
```

**Target:**
```
Scan → Flat list → Enter → Tree view → Navigate (lazy scan) → Select → Delete
```

---

## Prerequisites

Before starting implementation, resolve these questions:

### Unresolved Decisions

1. **Selection behavior in tree mode:**
   - [ ] **Option A:** Auto-mark descendants when folder selected
   - [ ] **Option B:** Explicit selection per item (current behavior)
   - **Recommendation:** Option B (KISS - keep current behavior)

2. **Caching strategy:**
   - [ ] **Option A:** Preserve selections when navigating back
   - [ ] **Option B:** Clear selections on navigation
   - **Recommendation:** Option A (better UX)

3. **Size display:**
   - [ ] **Option A:** Aggregated (sum of all descendants)
   - [ ] **Option B:** Direct (files in current folder only)
   - **Recommendation:** Option A (match NCDU behavior)

4. **Escape key behavior:**
   - [ ] **Option A:** Esc = exit tree completely
   - [ ] **Option B:** Esc = go back one level
   - **Recommendation:** Option A (Esc exit, ← back)

5. **Delete in tree mode:**
   - [ ] **Option A:** Delete entire current node
   - [ ] **Option B:** Delete only selected children
   - **Recommendation:** Option B (more control)

**Action:** Get user approval on recommendations before proceeding.

---

## Phase 1: Data Structures & Types (Day 1)

**Goal:** Create TreeNode structure and extend TUI model for tree navigation.

### Task 1.1: Create TreeNode Type

**File:** `pkg/types/tree.go` (NEW)

**Code:**
```go
package types

import (
    "path/filepath"
)

// TreeNode represents a file/directory in hierarchical tree navigation
type TreeNode struct {
    Path      string          // Full path
    Name      string          // Display name (basename)
    Size      int64           // Size in bytes
    IsDir     bool            // Directory flag
    Type      CleanTargetType // xcode/android/node
    Children  []*TreeNode     // Child nodes (nil = not scanned)
    Scanned   bool            // Lazy scan flag
    Depth     int             // Current depth in tree
    FileCount int             // Number of files
}

// AddChild appends child to node's children
func (n *TreeNode) AddChild(child *TreeNode) {
    if n.Children == nil {
        n.Children = make([]*TreeNode, 0)
    }
    n.Children = append(n.Children, child)
}

// NeedsScanning returns true if node hasn't been scanned yet
func (n *TreeNode) NeedsScanning() bool {
    return !n.Scanned && n.IsDir
}

// HasChildren returns true if node has children
func (n *TreeNode) HasChildren() bool {
    return n.Children != nil && len(n.Children) > 0
}

// GetBasename returns the base name from path
func GetBasename(path string) string {
    return filepath.Base(path)
}

// ScanResultToTreeNode converts ScanResult to TreeNode (for initial transition)
func ScanResultToTreeNode(result ScanResult) *TreeNode {
    return &TreeNode{
        Path:      result.Path,
        Name:      result.Name,
        Size:      result.Size,
        IsDir:     true, // Scan results are always directories
        Type:      result.Type,
        FileCount: result.FileCount,
        Scanned:   false,
        Depth:     0,
    }
}
```

**Acceptance Criteria:**
- [ ] File compiles without errors
- [ ] All helper methods work correctly
- [ ] `go test ./pkg/types/...` passes

---

### Task 1.2: Add Tree State to TUI Model

**File:** `internal/tui/tui.go:112-129`

**Changes:**
```go
// State represents the current TUI state
type State int

const (
    StateSelecting  State = iota // Viewing and selecting items
    StateConfirming              // Showing confirmation dialog
    StateDeleting                // Actively deleting items
    StateDone                    // Operation complete
    StateTree                    // NEW: Tree navigation view
)

// Model represents the TUI state
type Model struct {
    state    State
    items    []types.ScanResult
    selected map[int]bool
    cursor   int
    width    int
    height   int
    dryRun   bool
    results  []cleaner.CleanResult
    err      error
    quitting bool

    // Progress components
    spinner  spinner.Model
    progress progress.Model
    percent  float64

    // NEW: Tree navigation state
    treeMode     bool              // True when in tree view
    currentNode  *types.TreeNode   // Current tree node
    nodeStack    []*types.TreeNode // Breadcrumb trail
    maxDepth     int               // Max depth limit (default: 5)
    treeSelected map[string]bool   // Selected items in tree (path -> bool)
    scanning     bool              // True while scanning folder
}
```

**Update NewModel():**
```go
func NewModel(items []types.ScanResult, dryRun bool) Model {
    s := spinner.New()
    s.Spinner = spinner.Dot
    s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

    p := progress.New(progress.WithDefaultGradient())

    return Model{
        state:        StateSelecting,
        items:        items,
        selected:     make(map[int]bool),
        dryRun:       dryRun,
        spinner:      s,
        progress:     p,
        // NEW
        treeMode:     false,
        nodeStack:    make([]*types.TreeNode, 0),
        maxDepth:     5,
        treeSelected: make(map[string]bool),
        scanning:     false,
    }
}
```

**Acceptance Criteria:**
- [ ] Model compiles with new fields
- [ ] NewModel() initializes tree fields correctly
- [ ] Existing TUI functionality unchanged

---

### Task 1.3: Add Tree Key Bindings

**File:** `internal/tui/tui.go:71-110`

**Changes:**
```go
// KeyMap defines the key bindings
type KeyMap struct {
    Up       key.Binding
    Down     key.Binding
    Toggle   key.Binding
    All      key.Binding
    None     key.Binding
    Confirm  key.Binding
    Quit     key.Binding
    // NEW: Tree navigation keys
    DrillDown key.Binding // Enter/→ in tree mode
    GoBack    key.Binding // ←/h in tree mode
    Refresh   key.Binding // r - rescan current folder
    ExitTree  key.Binding // Esc - exit tree view
}

var keys = KeyMap{
    Up: key.NewBinding(
        key.WithKeys("up", "k"),
        key.WithHelp("↑/k", "up"),
    ),
    Down: key.NewBinding(
        key.WithKeys("down", "j"),
        key.WithHelp("↓/j", "down"),
    ),
    Toggle: key.NewBinding(
        key.WithKeys(" "),
        key.WithHelp("space", "toggle"),
    ),
    All: key.NewBinding(
        key.WithKeys("a"),
        key.WithHelp("a", "select all"),
    ),
    None: key.NewBinding(
        key.WithKeys("n"),
        key.WithHelp("n", "select none"),
    ),
    Confirm: key.NewBinding(
        key.WithKeys("enter"),
        key.WithHelp("enter", "clean selected"),
    ),
    Quit: key.NewBinding(
        key.WithKeys("q", "ctrl+c"),
        key.WithHelp("q", "quit"),
    ),
    // NEW
    DrillDown: key.NewBinding(
        key.WithKeys("right", "enter"),
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

**Acceptance Criteria:**
- [ ] Key bindings compile correctly
- [ ] No conflicts with existing keys
- [ ] Help text displays correctly

---

## Phase 2: Lazy Scanner Implementation (Days 2-3)

**Goal:** Add lazy directory scanning capability to scanner.

### Task 2.1: Add ScanDirectory Method

**File:** `internal/scanner/scanner.go`

**Add method:**
```go
// ScanDirectory scans a single directory lazily and returns TreeNode with children
func (s *Scanner) ScanDirectory(path string, currentDepth int, maxDepth int) (*types.TreeNode, error) {
    // Depth limit check
    if currentDepth >= maxDepth {
        return nil, fmt.Errorf("max depth %d reached", maxDepth)
    }

    // Read directory entries
    entries, err := os.ReadDir(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
    }

    // Calculate total size
    totalSize, fileCount, err := s.calculateSize(path)
    if err != nil {
        totalSize = 0
        fileCount = 0
    }

    // Build TreeNode
    node := &types.TreeNode{
        Path:      path,
        Name:      types.GetBasename(path),
        Size:      totalSize,
        IsDir:     true,
        Children:  make([]*types.TreeNode, 0),
        Scanned:   true,
        Depth:     currentDepth,
        FileCount: fileCount,
    }

    // Process children
    for _, entry := range entries {
        childPath := filepath.Join(path, entry.Name())

        // Skip symlinks to avoid cycles
        info, err := entry.Info()
        if err != nil {
            continue
        }
        if info.Mode()&os.ModeSymlink != 0 {
            continue
        }

        isDir := entry.IsDir()
        var childSize int64
        var childFileCount int

        if isDir {
            // For directories, calculate size
            childSize, childFileCount, _ = s.calculateSize(childPath)
        } else {
            // For files, use file size
            childSize = info.Size()
            childFileCount = 1
        }

        child := &types.TreeNode{
            Path:      childPath,
            Name:      entry.Name(),
            Size:      childSize,
            IsDir:     isDir,
            Scanned:   false, // Lazy - not scanned yet
            Depth:     currentDepth + 1,
            FileCount: childFileCount,
        }

        node.AddChild(child)
    }

    return node, nil
}
```

**Acceptance Criteria:**
- [ ] Method scans directory correctly
- [ ] Returns TreeNode with children
- [ ] Handles errors gracefully (permissions, symlinks)
- [ ] Respects maxDepth limit
- [ ] Calculates sizes accurately

---

### Task 2.2: Add Utility Methods

**File:** `internal/scanner/scanner.go`

**Add methods:**
```go
// FastCalculateSize calculates size without full tree traversal
// Used for lazy tree nodes to show size before drilling down
func (s *Scanner) FastCalculateSize(path string) (int64, int, error) {
    return s.calculateSize(path)
}

// IsDirectory checks if path is a directory
func (s *Scanner) IsDirectory(path string) bool {
    info, err := os.Stat(path)
    if err != nil {
        return false
    }
    return info.IsDir()
}

// ScanResultToTreeNode converts ScanResult to initial TreeNode
func (s *Scanner) ScanResultToTreeNode(result types.ScanResult) (*types.TreeNode, error) {
    node := types.ScanResultToTreeNode(result)

    // Verify path exists
    if !s.PathExists(result.Path) {
        return nil, fmt.Errorf("path does not exist: %s", result.Path)
    }

    return node, nil
}
```

**Acceptance Criteria:**
- [ ] All methods work correctly
- [ ] Size calculations match existing behavior
- [ ] Error handling robust

---

### Task 2.3: Add Scanner Tests

**File:** `internal/scanner/scanner_test.go`

**Add tests:**
```go
func TestScanDirectory(t *testing.T) {
    s, err := New()
    if err != nil {
        t.Fatalf("Failed to create scanner: %v", err)
    }

    // Create test directory structure
    tmpDir := t.TempDir()
    os.Mkdir(filepath.Join(tmpDir, "folder1"), 0755)
    os.Mkdir(filepath.Join(tmpDir, "folder2"), 0755)
    os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)

    // Test scanning
    node, err := s.ScanDirectory(tmpDir, 0, 5)
    if err != nil {
        t.Fatalf("ScanDirectory failed: %v", err)
    }

    // Verify results
    if !node.Scanned {
        t.Error("Node should be marked as scanned")
    }

    if len(node.Children) != 3 {
        t.Errorf("Expected 3 children, got %d", len(node.Children))
    }
}

func TestScanDirectoryMaxDepth(t *testing.T) {
    s, err := New()
    if err != nil {
        t.Fatalf("Failed to create scanner: %v", err)
    }

    tmpDir := t.TempDir()

    // Test max depth enforcement
    _, err = s.ScanDirectory(tmpDir, 5, 5)
    if err == nil {
        t.Error("Expected max depth error")
    }
}

func TestScanDirectoryPermissionDenied(t *testing.T) {
    s, err := New()
    if err != nil {
        t.Fatalf("Failed to create scanner: %v", err)
    }

    // Test permission denied handling
    _, err = s.ScanDirectory("/root", 0, 5)
    if err == nil {
        t.Error("Expected permission error")
    }
}
```

**Acceptance Criteria:**
- [ ] All tests pass
- [ ] Coverage > 80%
- [ ] Edge cases handled

---

## Phase 3: TUI Tree View (Days 3-4)

**Goal:** Implement tree navigation UI and event handling.

### Task 3.1: Add Tree State Messages

**File:** `internal/tui/tui.go`

**Add after line 258:**
```go
// scanNodeMsg is sent when folder scan completes
type scanNodeMsg struct {
    node *types.TreeNode
    err  error
}

// scanProgressMsg updates scan status
type scanProgressMsg struct {
    scanning bool
    message  string
}
```

**Acceptance Criteria:**
- [ ] Messages compile correctly
- [ ] Type-safe message handling

---

### Task 3.2: Implement Enter Tree Mode Command

**File:** `internal/tui/tui.go`

**Add method:**
```go
// enterTreeMode transitions from flat list to tree view
func (m Model) enterTreeMode() tea.Cmd {
    return func() tea.Msg {
        // Get current selected item
        if m.cursor >= len(m.items) {
            return scanNodeMsg{err: fmt.Errorf("invalid cursor position")}
        }

        item := m.items[m.cursor]

        // Convert to TreeNode
        s, err := scanner.New()
        if err != nil {
            return scanNodeMsg{err: err}
        }

        node, err := s.ScanResultToTreeNode(item)
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

**Acceptance Criteria:**
- [ ] Transitions correctly
- [ ] Handles errors gracefully
- [ ] Async command works

---

### Task 3.3: Implement Tree Navigation Commands

**File:** `internal/tui/tui.go`

**Add methods:**
```go
// scanNode scans a tree node's children lazily
func (m Model) scanNode(node *types.TreeNode) tea.Cmd {
    return func() tea.Msg {
        s, err := scanner.New()
        if err != nil {
            return scanNodeMsg{err: err}
        }

        scanned, err := s.ScanDirectory(node.Path, node.Depth, m.maxDepth)
        if err != nil {
            return scanNodeMsg{err: err}
        }

        // Update original node
        node.Children = scanned.Children
        node.Scanned = true

        return scanNodeMsg{node: node}
    }
}

// rescanNode refreshes a node's children
func (m Model) rescanNode(node *types.TreeNode) tea.Cmd {
    return func() tea.Msg {
        // Clear existing children
        node.Scanned = false
        node.Children = nil

        // Rescan
        return m.scanNode(node)()
    }
}

// exitTreeMode returns to flat list view
func (m *Model) exitTreeMode() {
    m.state = StateSelecting
    m.treeMode = false
    m.currentNode = nil
    m.nodeStack = make([]*types.TreeNode, 0)
    m.cursor = 0
    m.scanning = false
}

// goBackInTree navigates to parent node
func (m *Model) goBackInTree() {
    if len(m.nodeStack) == 0 {
        // Already at root, exit tree mode
        m.exitTreeMode()
        return
    }

    // Pop from stack
    m.currentNode = m.nodeStack[len(m.nodeStack)-1]
    m.nodeStack = m.nodeStack[:len(m.nodeStack)-1]
    m.cursor = 0
}

// drillDownInTree navigates into child node
func (m *Model) drillDownInTree() tea.Cmd {
    if m.currentNode == nil || !m.currentNode.HasChildren() {
        return nil
    }

    if m.cursor >= len(m.currentNode.Children) {
        return nil
    }

    selectedNode := m.currentNode.Children[m.cursor]

    // Can't drill into file
    if !selectedNode.IsDir {
        return nil
    }

    // Check depth limit
    if selectedNode.Depth >= m.maxDepth {
        // Show warning
        return nil
    }

    // Lazy scan if needed
    if selectedNode.NeedsScanning() {
        m.scanning = true
        return m.scanNode(selectedNode)
    }

    // Navigate
    m.nodeStack = append(m.nodeStack, m.currentNode)
    m.currentNode = selectedNode
    m.cursor = 0

    return nil
}
```

**Acceptance Criteria:**
- [ ] Navigation works correctly
- [ ] Stack management correct
- [ ] Lazy scanning triggers properly

---

### Task 3.4: Update Update() Handler

**File:** `internal/tui/tui.go:155-247`

**Modify StateSelecting case:**
```go
case StateSelecting:
    switch {
    case key.Matches(msg, keys.Quit):
        m.quitting = true
        return m, tea.Quit

    case key.Matches(msg, keys.Up):
        if m.cursor > 0 {
            m.cursor--
        }

    case key.Matches(msg, keys.Down):
        if m.cursor < len(m.items)-1 {
            m.cursor++
        }

    case key.Matches(msg, keys.Toggle):
        m.selected[m.cursor] = !m.selected[m.cursor]

    case key.Matches(msg, keys.All):
        for i := range m.items {
            m.selected[i] = true
        }

    case key.Matches(msg, keys.None):
        m.selected = make(map[int]bool)

    case key.Matches(msg, keys.Confirm):
        // NEW: Check if entering tree mode
        if m.cursor < len(m.items) {
            item := m.items[m.cursor]
            // If it's a directory and user wants to explore
            if key.Matches(msg, keys.DrillDown) {
                m.state = StateTree
                m.treeMode = true
                return m, m.enterTreeMode()
            }
        }

        // Original confirm logic
        if m.countSelected() > 0 {
            m.state = StateConfirming
            return m, nil
        }
    }
```

**Add StateTree case:**
```go
case StateTree:
    switch {
    case key.Matches(msg, keys.Quit):
        m.quitting = true
        return m, tea.Quit

    case key.Matches(msg, keys.ExitTree):
        m.exitTreeMode()
        return m, nil

    case key.Matches(msg, keys.GoBack):
        m.goBackInTree()
        return m, nil

    case key.Matches(msg, keys.DrillDown):
        return m, m.drillDownInTree()

    case key.Matches(msg, keys.Refresh):
        if m.currentNode != nil {
            m.scanning = true
            return m, m.rescanNode(m.currentNode)
        }
        return m, nil

    case key.Matches(msg, keys.Up):
        if m.cursor > 0 {
            m.cursor--
        }

    case key.Matches(msg, keys.Down):
        if m.currentNode != nil && m.currentNode.HasChildren() {
            if m.cursor < len(m.currentNode.Children)-1 {
                m.cursor++
            }
        }

    case key.Matches(msg, keys.Toggle):
        if m.currentNode != nil && m.currentNode.HasChildren() {
            if m.cursor < len(m.currentNode.Children) {
                child := m.currentNode.Children[m.cursor]
                m.treeSelected[child.Path] = !m.treeSelected[child.Path]
            }
        }

    case key.Matches(msg, keys.Confirm):
        // Confirm deletion in tree mode
        if m.countTreeSelected() > 0 {
            m.state = StateConfirming
            return m, nil
        }
    }
```

**Handle scanNodeMsg:**
```go
case scanNodeMsg:
    m.scanning = false
    if msg.err != nil {
        m.err = msg.err
        return m, nil
    }

    // First scan (entering tree mode)
    if m.currentNode == nil {
        m.currentNode = msg.node
        m.nodeStack = make([]*types.TreeNode, 0)
    }

    return m, nil
```

**Acceptance Criteria:**
- [ ] All key bindings work
- [ ] State transitions correct
- [ ] Message handling robust

---

### Task 3.5: Implement Tree View Rendering

**File:** `internal/tui/tui.go:282-317`

**Add case to View():**
```go
case StateTree:
    return m.renderTreeView(&b)
```

**Add renderTreeView method:**
```go
// renderTreeView renders the tree navigation view
func (m Model) renderTreeView(b *strings.Builder) string {
    if m.currentNode == nil {
        b.WriteString(errorStyle.Render("No tree node selected"))
        return b.String()
    }

    // Title
    title := "🧹 Mac Dev Cleaner - Tree View"
    if m.dryRun {
        title += " [DRY-RUN]"
    }
    b.WriteString(titleStyle.Render(title))
    b.WriteString("\n\n")

    // Breadcrumb
    breadcrumb := m.buildBreadcrumb()
    b.WriteString(helpStyle.Render(breadcrumb))
    b.WriteString("\n\n")

    // Current folder info
    selectedCount := m.countTreeSelected()
    folderInfo := fmt.Sprintf("📁 %s  •  %s  •  %d items",
        m.currentNode.Name,
        ui.FormatSize(m.currentNode.Size),
        len(m.currentNode.Children),
    )
    if selectedCount > 0 {
        folderInfo += fmt.Sprintf("  •  %d selected", selectedCount)
    }
    b.WriteString(statusStyle.Render(folderInfo))
    b.WriteString("\n\n")

    // Scanning indicator
    if m.scanning {
        b.WriteString(statusStyle.Render("🔍 Scanning folder...\n\n"))
    }

    // Children list
    if !m.currentNode.HasChildren() {
        b.WriteString(helpStyle.Render("  (Empty folder)"))
    } else {
        for i, child := range m.currentNode.Children {
            cursor := "  "
            if i == m.cursor {
                cursor = cursorStyle.Render("▸ ")
            }

            checkbox := "[ ]"
            if m.treeSelected[child.Path] {
                checkbox = checkboxStyle.Render("[✓]")
            }

            // Icon based on type and scan status
            icon := m.getTreeIcon(child)

            // Size with color
            sizeStr := ui.FormatSize(child.Size)
            sizeStyle := m.getSizeStyle(child.Size)

            // File count for directories
            fileCountStr := ""
            if child.IsDir && child.FileCount > 0 {
                fileCountStr = fmt.Sprintf(" (%d files)", child.FileCount)
            }

            line := fmt.Sprintf("%s%s %s %s  %s%s",
                cursor,
                checkbox,
                icon,
                sizeStyle.Render(fmt.Sprintf("%10s", sizeStr)),
                child.Name,
                fileCountStr,
            )

            if i == m.cursor {
                b.WriteString(selectedItemStyle.Render(line))
            } else {
                b.WriteString(itemStyle.Render(line))
            }
            b.WriteString("\n")
        }
    }

    // Depth warning
    if m.currentNode.Depth >= m.maxDepth-1 {
        warning := fmt.Sprintf("\n⚠️  Depth %d/%d - Approaching limit",
            m.currentNode.Depth+1, m.maxDepth)
        b.WriteString(errorStyle.Render(warning))
    }

    // Help
    help := "\n\n↑/↓: Navigate • →/Enter: Drill down • ←/h: Go back • r: Refresh • Space: Toggle • Esc: Exit tree • q: Quit"
    b.WriteString(helpStyle.Render(help))

    return b.String()
}

// buildBreadcrumb creates breadcrumb trail
func (m Model) buildBreadcrumb() string {
    if m.currentNode == nil {
        return ""
    }

    parts := []string{}
    for _, node := range m.nodeStack {
        parts = append(parts, node.Name)
    }
    parts = append(parts, m.currentNode.Name)

    return "📍 " + strings.Join(parts, " › ")
}

// getTreeIcon returns icon for tree node
func (m Model) getTreeIcon(node *types.TreeNode) string {
    if !node.IsDir {
        return "📄"
    }

    if node.Scanned {
        return "📂" // Opened folder
    }

    return "📁" // Unopened folder
}

// getSizeStyle returns styled size based on magnitude
func (m Model) getSizeStyle(size int64) lipgloss.Style {
    style := lipgloss.NewStyle().Width(10).Align(lipgloss.Right)

    if size > 1024*1024*1024 { // > 1GB
        return style.Foreground(lipgloss.Color("#EF4444")).Bold(true)
    } else if size > 100*1024*1024 { // > 100MB
        return style.Foreground(lipgloss.Color("#F59E0B"))
    }

    return style.Foreground(lipgloss.Color("#10B981"))
}

// countTreeSelected counts selected items in tree
func (m Model) countTreeSelected() int {
    count := 0
    for _, selected := range m.treeSelected {
        if selected {
            count++
        }
    }
    return count
}
```

**Acceptance Criteria:**
- [ ] Tree view renders correctly
- [ ] Breadcrumb displays path
- [ ] Icons show scan status
- [ ] Colors indicate size
- [ ] Help text accurate

---

## Phase 4: Testing & Polish (Days 5-6)

**Goal:** Comprehensive testing and UX improvements.

### Task 4.1: Manual Testing Checklist

**Test scenarios:**

1. **Basic Navigation:**
   - [ ] Scan → Enter folder → See children
   - [ ] Navigate down 5 levels
   - [ ] Go back to parent with ←
   - [ ] Exit tree with Esc

2. **Lazy Scanning:**
   - [ ] Enter unopened folder → Children load
   - [ ] Go back → Cache preserved
   - [ ] Refresh (r) → Children reload

3. **Depth Limits:**
   - [ ] Navigate to depth 4 → No warning
   - [ ] Navigate to depth 5 → Warning shown
   - [ ] Try depth 6 → Blocked

4. **Selection:**
   - [ ] Toggle items in tree
   - [ ] Selection persists on navigation
   - [ ] Confirm deletion shows correct items

5. **Edge Cases:**
   - [ ] Empty folder → Shows "(Empty folder)"
   - [ ] Permission denied → Error message
   - [ ] Very large folder (1000+ items) → Handles smoothly

6. **Performance:**
   - [ ] Large folder (10K files) → Scans in <2s
   - [ ] Deep tree (10 levels) → No lag

---

### Task 4.2: Add Loading Indicators

**File:** `internal/tui/tui.go`

**Improvements:**
```go
// Update renderTreeView to show spinner while scanning
if m.scanning {
    b.WriteString(m.spinner.View())
    b.WriteString(" Scanning folder...\n\n")
}
```

**Update Init():**
```go
func (m Model) Init() tea.Cmd {
    return tea.Batch(
        m.spinner.Tick,
    )
}
```

**Acceptance Criteria:**
- [ ] Spinner shows during scan
- [ ] Status messages clear
- [ ] No UI freezing

---

### Task 4.3: Add Error Handling UI

**File:** `internal/tui/tui.go`

**Add error display in tree view:**
```go
// In renderTreeView, after scanning check:
if m.err != nil {
    b.WriteString(errorStyle.Render(fmt.Sprintf("⚠️  Error: %v\n\n", m.err)))
}
```

**Acceptance Criteria:**
- [ ] Errors display clearly
- [ ] User can recover from errors
- [ ] No crashes on error

---

### Task 4.4: Performance Optimization

**Optimizations:**

1. **Size Calculation Caching:**
   - Cache sizes to avoid recalculation
   - Clear cache on refresh

2. **Pagination for Large Folders:**
   ```go
   const maxVisibleItems = 100

   func (m Model) getPaginatedChildren() []*types.TreeNode {
       if len(m.currentNode.Children) <= maxVisibleItems {
           return m.currentNode.Children
       }

       start := (m.cursor / maxVisibleItems) * maxVisibleItems
       end := start + maxVisibleItems
       if end > len(m.currentNode.Children) {
           end = len(m.currentNode.Children)
       }

       return m.currentNode.Children[start:end]
   }
   ```

3. **Memory Management:**
   - Clear old node children after navigating away
   - Keep only current + stack in memory

**Acceptance Criteria:**
- [ ] Large folders don't freeze UI
- [ ] Memory usage < 100MB
- [ ] Smooth scrolling

---

### Task 4.5: Documentation Updates

**Files to update:**

1. **README.md:**
   ```markdown
   ### Tree Navigation (NEW!)

   Navigate into folders interactively:

   ```bash
   dev-cleaner scan --tui
   # Press Enter on any folder to explore
   # Use arrow keys to navigate
   # Press Esc to exit tree view
   ```

   **Tree Mode Keys:**
   - `→` / `Enter` - Drill into folder
   - `←` / `h` - Go back to parent
   - `r` - Refresh current folder
   - `Space` - Toggle selection
   - `Esc` - Exit tree view
   ```

2. **Add screenshots:**
   - Tree view example
   - Breadcrumb navigation
   - Selection in tree mode

**Acceptance Criteria:**
- [ ] README updated
- [ ] Examples clear
- [ ] Screenshots added

---

## Integration Points

### Clean Command Integration

**File:** `cmd/root/clean.go`

**Changes needed:**
```go
// Update clean command to handle tree selections
if m.treeMode {
    // Collect selected items from tree
    var itemsToClean []types.ScanResult
    for path, selected := range m.treeSelected {
        if selected {
            // Convert TreeNode path to ScanResult
            // ...
        }
    }
    // Clean items
}
```

**Acceptance Criteria:**
- [ ] Clean works in tree mode
- [ ] Confirmation dialog correct
- [ ] Deletion successful

---

## Success Criteria

**Must achieve:**

1. ✅ Startup performance ≤ 3s (same as current)
2. ✅ Drill-down scan ≤ 2s for folders <10K files
3. ✅ Memory usage ≤ 100MB for typical tree
4. ✅ Users can navigate 5+ levels deep
5. ✅ TUI code ≤ 800 lines total
6. ✅ No regression in existing functionality
7. ✅ All tests pass
8. ✅ Documentation complete

---

## Risk Mitigation

### Risk 1: Performance Degradation
**Mitigation:**
- Async scanning with tea.Cmd
- Pagination for large folders
- Timeout after 30s

### Risk 2: Memory Leaks
**Mitigation:**
- Clear old nodes
- Limit cache size
- Monitor memory in tests

### Risk 3: State Bugs
**Mitigation:**
- Clear state separation
- Comprehensive unit tests
- Manual testing checklist

### Risk 4: UX Confusion
**Mitigation:**
- Clear breadcrumbs
- Help text always visible
- Intuitive key bindings

---

## Implementation Order

**Day-by-day breakdown:**

### Day 1: Foundation
- [ ] Task 1.1: TreeNode type
- [ ] Task 1.2: TUI model updates
- [ ] Task 1.3: Key bindings
- [ ] Verify compilation

### Day 2: Scanner
- [ ] Task 2.1: ScanDirectory
- [ ] Task 2.2: Utility methods
- [ ] Task 2.3: Tests
- [ ] Verify tests pass

### Day 3: TUI Part 1
- [ ] Task 3.1: Messages
- [ ] Task 3.2: Enter tree mode
- [ ] Task 3.3: Navigation commands
- [ ] Basic testing

### Day 4: TUI Part 2
- [ ] Task 3.4: Update handler
- [ ] Task 3.5: Tree view rendering
- [ ] Integration testing

### Day 5: Testing
- [ ] Task 4.1: Manual testing
- [ ] Task 4.2: Loading indicators
- [ ] Task 4.3: Error handling
- [ ] Bug fixes

### Day 6: Polish
- [ ] Task 4.4: Performance optimization
- [ ] Task 4.5: Documentation
- [ ] Final testing
- [ ] Code review

---

## Rollback Plan

If implementation fails:

1. **Revert files:**
   - `git checkout cmd/root/scan.go`
   - `git checkout internal/tui/tui.go`
   - Remove `pkg/types/tree.go`

2. **Remove features:**
   - Delete StateTree handling
   - Remove tree key bindings
   - Keep existing flat list

3. **Verify:**
   - Run tests: `go test ./...`
   - Manual test: `./dev-cleaner scan --tui`

---

## Next Steps

1. **Get approval:**
   - Review plan with user
   - Resolve unresolved questions
   - Confirm timeline

2. **Start Phase 1:**
   - Create feature branch: `git checkout -b feat/ncdu-navigation`
   - Begin Task 1.1 (TreeNode type)
   - Commit frequently

3. **Daily standup:**
   - Track progress against plan
   - Document blockers
   - Adjust timeline if needed

---

**Plan created:** 2025-12-15
**Estimated completion:** 2025-12-21 (6 days)
**Status:** Awaiting approval
