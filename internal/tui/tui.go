// Package tui provides interactive terminal UI using Bubble Tea
package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thanhdevapp/dev-cleaner/internal/cleaner"
	"github.com/thanhdevapp/dev-cleaner/internal/scanner"
	"github.com/thanhdevapp/dev-cleaner/internal/ui"
	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// State represents the current TUI state
type State int

const (
	StateScanning   State = iota // Initial scanning animation
	StateSelecting               // Viewing and selecting items
	StateConfirming              // Showing confirmation dialog
	StateDeleting                // Actively deleting items
	StateDone                    // Operation complete
	StateTree                    // Tree navigation view
	StateHelp                    // Help screen
)

// treeState saves tree navigation state for restoration
type treeState struct {
	parentNode *types.TreeNode
	nodeStack  []*types.TreeNode
	cursorPos  int
}

// Tips array - shown randomly to help users
var tips = []string{
	"💡 Tip: Press 'c' to quickly clean ONLY the current item (clears other selections)",
	"💡 Tip: Select multiple items with Space, then press Enter to clean them all at once",
	"💡 Tip: Use 'a' to select all items, 'n' to deselect all",
	"💡 Tip: Press '→' or 'l' to drill down into folders and explore their contents",
	"💡 Tip: In tree mode, press '←' or 'h' to go back to parent folder",
	"💡 Tip: Dry-run mode is active by default - your files are safe until you confirm",
	"💡 Tip: Press '?' anytime to see detailed help and keyboard shortcuts",
	"💡 Tip: 'c' = quick single clean, Enter = batch clean selected items",
	"💡 Tip: In tree mode, 'c' lets you delete folders at any level",
	"💡 Tip: All deletion operations are logged to ~/.dev-cleaner.log",
	"💡 Tip: Press 'Esc' in tree mode to return to main list",
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED")).
			Padding(0, 2).
			MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#7C3AED")).
				Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true)

	checkboxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true).
			MarginTop(1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	// Status bar styles
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#374151")).
			Padding(0, 1)

	statusLeftStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)

	statusCenterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F59E0B"))

	statusRightStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280"))
)

// KeyMap defines the key bindings
type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Toggle     key.Binding
	All        key.Binding
	None       key.Binding
	Confirm    key.Binding
	QuickClean key.Binding // Quick select current + confirm
	Help       key.Binding // Show help screen
	Quit       key.Binding
	// Tree navigation keys
	DrillDown key.Binding
	GoBack    key.Binding
	Refresh   key.Binding
	ExitTree  key.Binding
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
	QuickClean: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "select & clean"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	// Tree navigation
	DrillDown: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "drill down"),
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

// Model represents the TUI state
type Model struct {
	state    State
	items    []types.ScanResult
	selected map[int]bool
	cursor   int
	width    int
	height   int
	dryRun   bool
	version  string // Application version
	results  []cleaner.CleanResult
	err      error
	quitting bool

	// Progress components
	spinner  spinner.Model
	progress progress.Model
	percent  float64

	// Tree navigation state
	treeMode       bool              // True when in tree view
	currentNode    *types.TreeNode   // Current tree node
	nodeStack      []*types.TreeNode // Breadcrumb trail
	cursorStack    []int             // Cursor positions for each level
	maxDepth       int               // Max depth limit
	treeSelected   map[string]bool   // Selected items in tree
	scanning       bool              // True while scanning
	returnToTree   bool              // True if should return to tree after deletion
	savedTreeState *treeState        // Saved tree state for restoration

	// Time tracking
	startTime      time.Time     // Session start time
	deleteStart    time.Time     // Delete operation start time
	deleteDuration time.Duration // Frozen duration when deletion completes

	// Scanning progress
	scanningCategories []string        // Categories being scanned
	scanComplete       map[string]bool // Which categories are complete
	currentScanning    int             // Index of currently scanning category

	// Deletion progress
	deletingItems   []types.ScanResult // Items being deleted
	deleteComplete  map[int]bool       // Which items are complete
	deleteStatus    map[int]string     // Status for each item (success/error)
	currentDeleting int                // Index of currently deleting item
	fakeProgress    float64            // Fake progress for smooth animation

	// Help and tips
	currentTip string // Current random tip to display
	showHelp   bool   // Whether to show help screen

	// Table view
	itemsTable table.Model // Table for rendering items list
	treeTable  table.Model // Table for rendering tree view
}

// updateTableRows updates the table rows to reflect current selections
func (m *Model) updateTableRows() {
	rows := []table.Row{}
	for i, item := range m.items {
		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = "[✓]"
		}

		typeBadge := string(item.Type)
		sizeStr := ui.FormatSize(item.Size)

		rows = append(rows, table.Row{
			checkbox,
			typeBadge,
			sizeStr,
			item.Name,
			item.Path, // Full path
		})
	}
	m.itemsTable.SetRows(rows)
	m.itemsTable.SetCursor(m.cursor)
}

// updateTreeTableRows updates the tree table rows to reflect current selections
func (m *Model) updateTreeTableRows() {
	if m.currentNode == nil || !m.currentNode.HasChildren() {
		m.treeTable.SetRows([]table.Row{})
		return
	}

	rows := []table.Row{}
	for _, child := range m.currentNode.Children {
		checkbox := "[ ]"
		if m.treeSelected[child.Path] {
			checkbox = "[✓]"
		}

		// Icon based on type
		icon := "📄"
		if child.IsDir {
			if child.Scanned {
				icon = "📂"
			} else {
				icon = "📁"
			}
		}

		sizeStr := ui.FormatSize(child.Size)

		rows = append(rows, table.Row{
			checkbox,
			icon,
			sizeStr,
			child.Name,
			child.Path, // Full path
		})
	}
	m.treeTable.SetRows(rows)
	m.treeTable.SetCursor(m.cursor)
}

// updateTableColumns updates table column widths based on terminal width
func (m *Model) updateTableColumns() {
	if m.width == 0 {
		return // No width info yet
	}

	// Fixed column widths: checkbox(3) + category/type(12 or 4) + size(10) + name(30) + borders/padding(~10)
	fixedWidth := 3 + 12 + 10 + 30 + 10
	pathWidth := m.width - fixedWidth
	if pathWidth < 30 {
		pathWidth = 30 // Minimum path width
	}

	// Update main table columns
	mainCols := []table.Column{
		{Title: "", Width: 3},             // Checkbox
		{Title: "Category", Width: 12},    // Type badge
		{Title: "Size", Width: 10},        // Formatted size
		{Title: "Name", Width: 30},        // Item name
		{Title: "Path", Width: pathWidth}, // Dynamic path width
	}
	m.itemsTable.SetColumns(mainCols)

	// Update tree table columns (slightly different fixed widths)
	treeFixedWidth := 3 + 4 + 10 + 30 + 10
	treePathWidth := m.width - treeFixedWidth
	if treePathWidth < 30 {
		treePathWidth = 30
	}

	treeCols := []table.Column{
		{Title: "", Width: 3},                 // Checkbox
		{Title: "Type", Width: 4},             // Icon
		{Title: "Size", Width: 10},            // Formatted size
		{Title: "Name", Width: 30},            // Item name
		{Title: "Path", Width: treePathWidth}, // Dynamic path width
	}
	m.treeTable.SetColumns(treeCols)
}

// NewModel creates a new TUI model
func NewModel(items []types.ScanResult, dryRun bool, version string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	p := progress.New(progress.WithDefaultGradient())

	// Determine which categories to show in scanning animation
	categories := []string{}
	if len(items) > 0 {
		// Check which types exist in results
		typesSeen := make(map[types.CleanTargetType]bool)
		for _, item := range items {
			typesSeen[item.Type] = true
		}
		if typesSeen[types.TypeXcode] {
			categories = append(categories, "Xcode")
		}
		if typesSeen[types.TypeAndroid] {
			categories = append(categories, "Android")
		}
		if typesSeen[types.TypeNode] {
			categories = append(categories, "Node.js")
		}
		if typesSeen[types.TypeFlutter] {
			categories = append(categories, "Flutter")
		}
		if typesSeen[types.TypePython] {
			categories = append(categories, "Python")
		}
		if typesSeen[types.TypeRust] {
			categories = append(categories, "Rust")
		}
		if typesSeen[types.TypeGo] {
			categories = append(categories, "Go")
		}
		if typesSeen[types.TypeHomebrew] {
			categories = append(categories, "Homebrew")
		}
		if typesSeen[types.TypeDocker] {
			categories = append(categories, "Docker")
		}
		if typesSeen[types.TypeJava] {
			categories = append(categories, "Java")
		}
	}

	// Start in scanning state if we have items
	initialState := StateSelecting
	if len(items) > 0 && len(categories) > 0 {
		initialState = StateScanning
	}

	// Pick a random tip
	randomTip := tips[time.Now().UnixNano()%int64(len(tips))]

	// Create table columns
	columns := []table.Column{
		{Title: "", Width: 3},          // Checkbox
		{Title: "Category", Width: 12}, // Type badge
		{Title: "Size", Width: 10},     // Formatted size
		{Title: "Name", Width: 30},     // Item name (shorter to make room for path)
		{Title: "Path", Width: 50},     // Full path
	}

	// Create table rows from items
	rows := []table.Row{}
	for _, item := range items {
		checkbox := "[ ]"
		typeBadge := string(item.Type)
		sizeStr := ui.FormatSize(item.Size)

		rows = append(rows, table.Row{
			checkbox,
			typeBadge,
			sizeStr,
			item.Name,
			item.Path, // Full path
		})
	}

	// Initialize table with dynamic height to show all items
	tableHeight := len(items)
	if tableHeight < 5 {
		tableHeight = 5 // Minimum height
	}
	if tableHeight > 30 {
		tableHeight = 30 // Cap at 30 to prevent huge tables
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	// Apply table styles
	ts := table.DefaultStyles()
	ts.Header = ts.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		BorderBottom(true).
		Bold(false)
	ts.Selected = ts.Selected.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7C3AED")).
		Bold(false)
	t.SetStyles(ts)

	// Create tree table (same columns as main table but with icon instead of type)
	treeColumns := []table.Column{
		{Title: "", Width: 3},      // Checkbox
		{Title: "Type", Width: 4},  // Icon (📁/📂/📄)
		{Title: "Size", Width: 10}, // Formatted size
		{Title: "Name", Width: 30}, // Item name (shorter to make room for path)
		{Title: "Path", Width: 50}, // Full path
	}

	treeT := table.New(
		table.WithColumns(treeColumns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(20), // Tree view can be taller
	)
	treeT.SetStyles(ts)

	m := Model{
		state:    initialState,
		items:    items,
		selected: make(map[int]bool),
		dryRun:   dryRun,
		version:  version,
		spinner:  s,
		progress: p,
		// Tree navigation
		treeMode:     false,
		nodeStack:    make([]*types.TreeNode, 0),
		maxDepth:     5,
		treeSelected: make(map[string]bool),
		scanning:     false,
		// Time tracking
		startTime: time.Now(),
		// Scanning animation
		scanningCategories: categories,
		scanComplete:       make(map[string]bool),
		currentScanning:    0,
		// Deletion progress
		deletingItems:   []types.ScanResult{},
		deleteComplete:  make(map[int]bool),
		deleteStatus:    make(map[int]string),
		currentDeleting: 0,
		// Help and tips
		currentTip: randomTip,
		showHelp:   false,
		// Table view
		itemsTable: t,
		treeTable:  treeT,
	}

	// Initialize table rows
	m.updateTableRows()

	return m
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	if m.state == StateScanning {
		return tea.Batch(m.spinner.Tick, m.tickScanning())
	}
	return m.spinner.Tick
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 10
		m.updateTableColumns() // Update table columns to fit new width
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case deletionTickMsg:
		// Keep UI responsive during deletion - tick spinner, update progress, and schedule next tick
		if m.state == StateDeleting {
			var spinnerCmd tea.Cmd
			m.spinner, spinnerCmd = m.spinner.Update(m.spinner.Tick())

			// Increment fake progress for smooth animation (small increments until real progress catches up)
			targetProgress := float64(m.currentDeleting) / float64(len(m.deletingItems))
			if m.fakeProgress < targetProgress {
				m.fakeProgress = targetProgress // Snap to real progress
			} else if m.fakeProgress < 0.99 && m.fakeProgress < targetProgress+0.05 {
				// Small fake increment to show activity
				m.fakeProgress += 0.002
			}

			// Update progress bar with combined progress
			displayProgress := m.fakeProgress
			if m.percent > displayProgress {
				displayProgress = m.percent
			}
			progressCmd := m.progress.SetPercent(displayProgress)

			return m, tea.Batch(spinnerCmd, progressCmd, m.tickDeletion())
		}
		return m, nil

	case deleteProgressMsg:
		m.percent = msg.percent
		cmd := m.progress.SetPercent(m.percent)
		return m, cmd

	case tea.KeyMsg:
		// Handle based on current state
		switch m.state {
		case StateDone:
			switch msg.String() {
			case "q", "ctrl+c":
				// Quit application
				m.quitting = true
				return m, tea.Quit

			case "r", "enter":
				// Rescan and refresh results - return to view immediately for better UX
				// Check if we should return to tree mode
				if m.returnToTree && m.savedTreeState != nil {
					// Restore tree state immediately
					m.state = StateTree
					m.treeMode = true
					m.currentNode = m.savedTreeState.parentNode
					m.nodeStack = m.savedTreeState.nodeStack
					m.cursor = 0 // Reset cursor to top
					m.scanning = true
					m.returnToTree = false
					m.savedTreeState = nil
					m.updateTreeTableRows()

					// Trigger rescan in background (non-blocking)
					return m, m.rescanNode(m.currentNode)
				}

				// Normal rescan - transition to selecting state immediately
				m.state = StateSelecting
				m.results = nil
				m.err = nil
				m.scanning = true
				m.updateTableRows()

				// Trigger rescan in background (non-blocking)
				return m, m.rescanItems()

			case "esc":
				// Go back to selection without rescanning
				if m.returnToTree && m.savedTreeState != nil {
					// Restore tree state without rescanning
					m.state = StateTree
					m.treeMode = true
					m.currentNode = m.savedTreeState.parentNode
					m.nodeStack = m.savedTreeState.nodeStack
					m.cursor = m.savedTreeState.cursorPos
					m.returnToTree = false
					m.savedTreeState = nil
					m.updateTreeTableRows()
					return m, nil
				}

				// Go back to main list without rescanning
				m.state = StateSelecting
				m.results = nil
				m.err = nil
				return m, nil
			}

		case StateConfirming:
			switch msg.String() {
			case "y", "Y":
				m.state = StateDeleting
				m.percent = 0
				m.deleteStart = time.Now()

				// Prepare deletion list
				m.deletingItems = []types.ScanResult{}
				for i, item := range m.items {
					if m.selected[i] {
						m.deletingItems = append(m.deletingItems, item)
					}
				}
				m.deleteComplete = make(map[int]bool)
				m.deleteStatus = make(map[int]string)
				m.currentDeleting = 0

				// Start deletion with spinner, progress updates, and continuous tick
				return m, tea.Batch(
					m.spinner.Tick,
					m.progress.SetPercent(0),
					m.tickDeletion(), // Start continuous UI refresh
					m.performClean(),
				)
			case "n", "N", "esc":
				// Check if we came from tree mode
				if m.returnToTree && m.savedTreeState != nil {
					// Return to tree mode
					m.state = StateTree
					m.treeMode = true
					m.currentNode = m.savedTreeState.parentNode
					m.nodeStack = m.savedTreeState.nodeStack
					m.cursor = m.savedTreeState.cursorPos
					m.returnToTree = false
					m.savedTreeState = nil
					return m, nil
				}
				// Normal return to selection
				m.state = StateSelecting
				return m, nil
			}
			return m, nil

		case StateDeleting:
			// Ignore key presses while deleting
			if key.Matches(msg, keys.Quit) {
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil

		case StateHelp:
			// Any key to exit help
			m.state = StateSelecting
			return m, nil

		case StateSelecting:
			switch {
			case key.Matches(msg, keys.Quit):
				m.quitting = true
				return m, tea.Quit

			case key.Matches(msg, keys.Help):
				m.state = StateHelp
				return m, nil

			case key.Matches(msg, keys.Up):
				if m.cursor > 0 {
					m.cursor--
					m.updateTableRows()
				}

			case key.Matches(msg, keys.Down):
				if m.cursor < len(m.items)-1 {
					m.cursor++
					m.updateTableRows()
				}

			case key.Matches(msg, keys.Toggle):
				m.selected[m.cursor] = !m.selected[m.cursor]
				m.updateTableRows()

			case key.Matches(msg, keys.All):
				for i := range m.items {
					m.selected[i] = true
				}
				m.updateTableRows()

			case key.Matches(msg, keys.None):
				m.selected = make(map[int]bool)
				m.updateTableRows()

			case key.Matches(msg, keys.Confirm):
				if m.countSelected() > 0 {
					m.state = StateConfirming
					return m, nil
				}

			case key.Matches(msg, keys.QuickClean):
				// Quick clean ONLY current item (clear all other selections)
				if m.cursor < len(m.items) {
					// Clear all previous selections
					m.selected = make(map[int]bool)
					// Select ONLY current item
					m.selected[m.cursor] = true
					// Go to confirmation
					m.state = StateConfirming
					return m, nil
				}

			case key.Matches(msg, keys.DrillDown):
				// Enter tree mode for current item
				if m.cursor < len(m.items) {
					m.state = StateTree
					m.treeMode = true
					m.scanning = true
					return m, m.enterTreeMode()
				}
			}

		case StateTree:
			switch {
			case key.Matches(msg, keys.Quit):
				m.quitting = true
				return m, tea.Quit

			case key.Matches(msg, keys.Help):
				m.state = StateHelp
				return m, nil

			case key.Matches(msg, keys.ExitTree):
				m.exitTreeMode()
				return m, nil

			case key.Matches(msg, keys.GoBack):
				m.goBackInTree()
				m.updateTreeTableRows()
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
					m.updateTreeTableRows()
				}

			case key.Matches(msg, keys.Down):
				if m.currentNode != nil && m.currentNode.HasChildren() {
					if m.cursor < len(m.currentNode.Children)-1 {
						m.cursor++
						m.updateTreeTableRows()
					}
				}

			case key.Matches(msg, keys.Toggle):
				if m.currentNode != nil && m.currentNode.HasChildren() {
					if m.cursor < len(m.currentNode.Children) {
						child := m.currentNode.Children[m.cursor]
						m.treeSelected[child.Path] = !m.treeSelected[child.Path]
						m.updateTreeTableRows()
					}
				}

			case key.Matches(msg, keys.QuickClean):
				// Delete current item at current level (like normal file managers)
				if m.currentNode != nil && m.currentNode.HasChildren() {
					if m.cursor < len(m.currentNode.Children) {
						child := m.currentNode.Children[m.cursor]

						// Create a single-item deletion
						m.deletingItems = []types.ScanResult{{
							Path:      child.Path,
							Type:      types.TypeNode,
							Size:      child.Size,
							FileCount: child.FileCount,
							Name:      child.Name,
						}}

						// Setup deletion state
						m.deleteComplete = make(map[int]bool)
						m.deleteStatus = make(map[int]string)
						m.currentDeleting = 0
						m.selected = map[int]bool{0: true}

						// Save tree state to return after deletion
						m.returnToTree = true
						m.savedTreeState = &treeState{
							parentNode: m.currentNode,
							nodeStack:  m.nodeStack,
							cursorPos:  m.cursor,
						}

						m.state = StateConfirming
						return m, nil
					}
				}
			}
		}

	case deleteItemProgressMsg:
		// Update item status
		m.deleteComplete[msg.index] = true
		if msg.status == "error" {
			m.deleteStatus[msg.index] = "error"
		} else {
			m.deleteStatus[msg.index] = "success"
		}

		// Update progress
		m.currentDeleting++
		if len(m.deletingItems) > 0 {
			m.percent = float64(m.currentDeleting) / float64(len(m.deletingItems))
		}

		// Continue with next item or finish
		return m, tea.Batch(
			m.spinner.Tick, // Keep spinner animating
			m.progress.SetPercent(m.percent),
			m.performClean(), // Delete next item or finish
		)

	case cleanResultMsg:
		m.state = StateDone
		m.results = msg.results
		m.err = msg.err
		// Freeze the deletion duration so timer stops counting
		m.deleteDuration = time.Since(m.deleteStart)
		m.percent = 1.0 // Ensure progress shows 100%
		return m, nil

	case scanNodeMsg:
		m.scanning = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if m.currentNode == nil {
			m.currentNode = msg.node
			m.nodeStack = make([]*types.TreeNode, 0)
		}
		m.cursor = 0
		m.updateTreeTableRows()
		return m, nil

	case rescanItemsMsg:
		if msg.err != nil {
			m.err = msg.err
			m.scanning = false
			return m, nil
		}
		// Reset state and show new items
		m.items = msg.items
		m.selected = make(map[int]bool)
		m.cursor = 0
		m.state = StateSelecting
		m.results = nil
		m.err = nil
		m.scanning = false
		m.updateTableRows()
		return m, nil

	case scanProgressMsg:
		if m.state != StateScanning {
			return m, nil
		}

		// Mark current category as complete
		if m.currentScanning < len(m.scanningCategories) {
			m.scanComplete[m.scanningCategories[m.currentScanning]] = true
			m.currentScanning++
		}

		// If all categories scanned, transition to selecting
		if m.currentScanning >= len(m.scanningCategories) {
			m.state = StateSelecting
			m.updateTableRows()
			return m, nil
		}

		// Continue scanning animation
		return m, m.tickScanning()
	}

	return m, nil
}

// cleanResultMsg is sent when cleaning is complete
type cleanResultMsg struct {
	results []cleaner.CleanResult
	err     error
}

// deleteProgressMsg is sent to update progress bar
type deleteProgressMsg struct {
	percent float64
}

// scanNodeMsg is sent when folder scan completes
type scanNodeMsg struct {
	node *types.TreeNode
	err  error
}

// rescanItemsMsg is sent when items rescan completes
type rescanItemsMsg struct {
	items []types.ScanResult
	err   error
}

// scanProgressMsg is sent to advance scanning animation
type scanProgressMsg struct{}

// deleteItemProgressMsg is sent when an item deletion starts/completes
type deleteItemProgressMsg struct {
	index  int
	status string // "start", "success", "error"
	err    error
}

// deletionTickMsg for UI refresh during deletion
type deletionTickMsg struct{}

// tickDeletion sends periodic UI refresh messages during deletion
func (m Model) tickDeletion() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return deletionTickMsg{}
	})
}

// tickScanning sends a message to advance scanning animation
func (m Model) tickScanning() tea.Cmd {
	return tea.Tick(time.Millisecond*600, func(t time.Time) tea.Msg {
		return scanProgressMsg{}
	})
}

// rescanItems rescans all items and returns to selection
func (m Model) rescanItems() tea.Cmd {
	return func() tea.Msg {
		s, err := scanner.New()
		if err != nil {
			return rescanItemsMsg{err: err}
		}

		opts := types.DefaultScanOptions()

		results, err := s.ScanAll(opts)
		if err != nil {
			return rescanItemsMsg{err: err}
		}

		// Sort by size
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[j].Size > results[i].Size {
					results[i], results[j] = results[j], results[i]
				}
			}
		}

		return rescanItemsMsg{items: results}
	}
}

// enterTreeMode transitions from flat list to tree view
func (m Model) enterTreeMode() tea.Cmd {
	return func() tea.Msg {
		if m.cursor >= len(m.items) {
			return scanNodeMsg{err: fmt.Errorf("invalid cursor position")}
		}

		item := m.items[m.cursor]

		s, err := scanner.New()
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
		m.exitTreeMode()
		return
	}

	// Pop node and cursor from stacks
	m.currentNode = m.nodeStack[len(m.nodeStack)-1]
	m.nodeStack = m.nodeStack[:len(m.nodeStack)-1]

	// Restore cursor position
	if len(m.cursorStack) > 0 {
		m.cursor = m.cursorStack[len(m.cursorStack)-1]
		m.cursorStack = m.cursorStack[:len(m.cursorStack)-1]
	} else {
		m.cursor = 0
	}
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

	if !selectedNode.IsDir {
		return nil
	}

	if selectedNode.Depth >= m.maxDepth {
		return nil
	}

	if selectedNode.NeedsScanning() {
		m.scanning = true
		// Save cursor position before scanning
		m.cursorStack = append(m.cursorStack, m.cursor)
		return m.scanNode(selectedNode)
	}

	// Save cursor position before navigating
	m.cursorStack = append(m.cursorStack, m.cursor)
	m.nodeStack = append(m.nodeStack, m.currentNode)
	m.currentNode = selectedNode
	m.cursor = 0
	m.updateTreeTableRows()

	return nil
}

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

		node.Children = scanned.Children
		node.Scanned = true

		return scanNodeMsg{node: node}
	}
}

// rescanNode refreshes a node's children
func (m Model) rescanNode(node *types.TreeNode) tea.Cmd {
	return func() tea.Msg {
		node.Scanned = false
		node.Children = nil
		return m.scanNode(node)()
	}
}

// prepareTreeDeletion converts tree selections to flat list and transitions to confirmation
func (m *Model) prepareTreeDeletion() tea.Cmd {
	// Collect selected items from tree
	var selectedItems []types.ScanResult
	for path, selected := range m.treeSelected {
		if selected {
			// Find the node in the tree
			node := m.findNodeByPath(m.currentNode, path)
			if node != nil {
				selectedItems = append(selectedItems, types.ScanResult{
					Path:      node.Path,
					Type:      types.TypeNode, // Generic type for tree items
					Size:      node.Size,
					FileCount: node.FileCount,
					Name:      node.Name,
				})
			}
		}
	}

	// If we have selected items, prepare for deletion
	if len(selectedItems) > 0 {
		// Exit tree mode
		m.state = StateConfirming
		m.treeMode = false
		m.currentNode = nil
		m.nodeStack = make([]*types.TreeNode, 0)
		m.scanning = false

		// Replace items with tree selections for deletion
		m.items = selectedItems
		m.selected = make(map[int]bool)
		// Mark all as selected
		for i := range m.items {
			m.selected[i] = true
		}
	}

	return nil
}

// findNodeByPath recursively finds a node by path
func (m Model) findNodeByPath(root *types.TreeNode, path string) *types.TreeNode {
	if root == nil {
		return nil
	}
	if root.Path == path {
		return root
	}
	if root.HasChildren() {
		for _, child := range root.Children {
			if found := m.findNodeByPath(child, path); found != nil {
				return found
			}
		}
	}
	return nil
}

// performClean deletes a single item and returns a command to continue
func (m Model) performClean() tea.Cmd {
	// Check if all items are processed
	if m.currentDeleting >= len(m.deletingItems) {
		// All done, collect results and finish
		var results []cleaner.CleanResult
		for i, item := range m.deletingItems {
			success := m.deleteComplete[i] && m.deleteStatus[i] != "error"
			var err error
			if !success && m.deleteStatus[i] == "error" {
				err = fmt.Errorf("deletion failed")
			}
			results = append(results, cleaner.CleanResult{
				Path:      item.Path,
				Size:      item.Size,
				Success:   success,
				Error:     err,
				WasDryRun: m.dryRun,
			})
		}
		return func() tea.Msg {
			return cleanResultMsg{results: results, err: nil}
		}
	}

	// Delete current item
	idx := m.currentDeleting
	item := m.deletingItems[idx]

	return func() tea.Msg {
		c, err := cleaner.New(m.dryRun)
		if err != nil {
			return deleteItemProgressMsg{
				index:  idx,
				status: "error",
				err:    err,
			}
		}
		defer c.Close()

		// Validate path safety
		if err := cleaner.ValidatePath(item.Path); err != nil {
			return deleteItemProgressMsg{
				index:  idx,
				status: "error",
				err:    err,
			}
		}

		// Send start message first (for immediate UI update)
		time.Sleep(200 * time.Millisecond) // Initial delay to show "deleting" state

		// Perform deletion
		if m.dryRun {
			c.Logger().Printf("[DRY-RUN] Would delete: %s (%.2f MB)\n", item.Path, float64(item.Size)/(1024*1024))
			// Longer delay for visual feedback in dry-run
			time.Sleep(300 * time.Millisecond)
			return deleteItemProgressMsg{
				index:  idx,
				status: "success",
			}
		} else {
			c.Logger().Printf("[DELETE] Removing: %s (%.2f MB)\n", item.Path, float64(item.Size)/(1024*1024))

			if err := os.RemoveAll(item.Path); err != nil {
				c.Logger().Printf("[ERROR] Failed to delete %s: %v\n", item.Path, err)
				return deleteItemProgressMsg{
					index:  idx,
					status: "error",
					err:    err,
				}
			}

			c.Logger().Printf("[SUCCESS] Deleted: %s at %s\n", item.Path, time.Now().Format(time.RFC3339))
			// Delay to show success state
			time.Sleep(200 * time.Millisecond)
			return deleteItemProgressMsg{
				index:  idx,
				status: "success",
			}
		}
	}
}

// View implements tea.Model
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Title with version
	title := fmt.Sprintf("🧹 Mac Dev Cleaner v%s", m.version)
	if m.dryRun {
		title += " [DRY-RUN]"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// Render based on current state
	var content string
	switch m.state {
	case StateScanning:
		content = m.renderScanning(&b)

	case StateDone:
		content = m.renderResults(&b)

	case StateDeleting:
		content = m.renderDeleting(&b)

	case StateConfirming:
		content = m.renderConfirmation(&b)

	case StateTree:
		content = m.renderTreeView(&b)

	case StateHelp:
		content = m.renderHelp(&b)

	case StateSelecting:
		content = m.renderSelection(&b)

	default:
		content = b.String()
	}

	// Add status bar at bottom
	statusBar := m.renderStatusBar()
	return content + "\n\n" + statusBar
}

// renderScanning renders the animated scanning progress
func (m Model) renderScanning(b *strings.Builder) string {
	b.WriteString(successStyle.Render("🔍 Scanning for development artifacts...\n\n"))

	// Show each category with status
	for i, category := range m.scanningCategories {
		var status string
		var icon string
		var style lipgloss.Style

		if m.scanComplete[category] {
			// Completed
			icon = "✓"
			status = "Complete"
			style = successStyle
		} else if i == m.currentScanning {
			// Currently scanning
			icon = m.spinner.View()
			status = "Scanning..."
			style = statusStyle
		} else {
			// Pending
			icon = "○"
			status = "Pending"
			style = helpStyle
		}

		line := fmt.Sprintf("  %s %s  %s", icon, category, status)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Please wait while we scan your system..."))

	return b.String()
}

// renderTreeView renders the tree navigation view
func (m Model) renderTreeView(b *strings.Builder) string {
	if m.currentNode == nil {
		// Show loading animation while waiting for scan
		loadingMsg := fmt.Sprintf("%s Loading directory tree...", m.spinner.View())
		b.WriteString(statusStyle.Render(loadingMsg))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Please wait while scanning directory structure..."))
		return b.String()
	}

	// Breadcrumb
	breadcrumb := m.buildBreadcrumb()
	b.WriteString(helpStyle.Render(breadcrumb))
	b.WriteString("\n\n")

	// Current folder info
	folderInfo := fmt.Sprintf("📁 %s  •  %s  •  %d items",
		m.currentNode.Name,
		ui.FormatSize(m.currentNode.Size),
		len(m.currentNode.Children),
	)
	b.WriteString(statusStyle.Render(folderInfo))
	b.WriteString("\n\n")

	// Scanning indicator
	if m.scanning {
		b.WriteString(m.spinner.View())
		b.WriteString(" Scanning folder...\n\n")
	}

	// Children list - use table
	if !m.currentNode.HasChildren() {
		b.WriteString(helpStyle.Render("  (Empty folder)"))
	} else {
		// Render tree table (already updated in Update())
		b.WriteString(m.treeTable.View())
		b.WriteString("\n")
	}

	// Depth info
	if m.currentNode.Depth >= m.maxDepth-1 {
		warning := fmt.Sprintf("\n⚠️  Depth %d/%d - Approaching limit",
			m.currentNode.Depth+1, m.maxDepth)
		b.WriteString(errorStyle.Render(warning))
	}

	// Help
	help := "\n\n↑/↓: Navigate • →/l: Drill down • ←/h: Go back • Space: Toggle • c: Quick Clean Current • Esc: Exit • q: Quit"
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

// renderDeleting shows the deletion progress with package-manager style output
func (m Model) renderDeleting(b *strings.Builder) string {
	b.WriteString(statusStyle.Render("🗑️  Cleaning up development artifacts"))
	b.WriteString("\n\n")

	// Calculate progress
	totalItems := len(m.deletingItems)
	completedItems := 0
	for _, complete := range m.deleteComplete {
		if complete {
			completedItems++
		}
	}

	// Show progress bar
	b.WriteString(m.progress.View())
	b.WriteString("\n\n")

	// Show items being deleted (package-manager style)
	maxVisible := 8
	startIdx := 0
	if totalItems > maxVisible {
		// Show last N items for scrolling effect
		startIdx = totalItems - maxVisible
		if completedItems < maxVisible {
			startIdx = 0
		} else {
			startIdx = completedItems - maxVisible/2
			if startIdx < 0 {
				startIdx = 0
			}
			if startIdx+maxVisible > totalItems {
				startIdx = totalItems - maxVisible
			}
		}
	}

	for i := startIdx; i < totalItems && i < startIdx+maxVisible; i++ {
		item := m.deletingItems[i]
		icon := "  "
		itemStyle := lipgloss.NewStyle()

		if m.deleteComplete[i] {
			if m.deleteStatus[i] == "error" {
				icon = "✗ "
				itemStyle = errorStyle
			} else {
				icon = "✓ "
				itemStyle = successStyle
			}
		} else if i == m.currentDeleting {
			icon = m.spinner.View() + " "
			itemStyle = statusStyle
		} else {
			icon = "○ "
			itemStyle = helpStyle
		}

		line := fmt.Sprintf("%s %s  %s",
			icon,
			ui.FormatSize(item.Size),
			item.Name,
		)
		b.WriteString(itemStyle.Render(line))
		b.WriteString("\n")
	}

	// Summary
	b.WriteString("\n")
	summary := fmt.Sprintf("Progress: %d/%d items", completedItems, totalItems)
	b.WriteString(helpStyle.Render(summary))

	// Help
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Please wait... Press q to cancel"))

	return b.String()
}

// renderConfirmation shows the confirmation dialog
func (m Model) renderConfirmation(b *strings.Builder) string {
	// Calculate count and size based on source
	var selectedCount int
	var selectedSize int64

	// Check if coming from tree mode (has deletingItems)
	if len(m.deletingItems) > 0 {
		// Tree mode - calculate from deletingItems
		selectedCount = len(m.deletingItems)
		for _, item := range m.deletingItems {
			selectedSize += item.Size
		}
	} else {
		// Normal selection mode
		selectedCount = m.countSelected()
		selectedSize = m.selectedSize()
	}

	// Confirmation box style - wider to show paths
	confirmBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#F59E0B")).
		Padding(1, 2).
		Width(80)

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Bold(true)

	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444"))

	sizeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	var confirmMsg strings.Builder
	confirmMsg.WriteString(warningStyle.Render("⚠️  Confirm Deletion"))
	confirmMsg.WriteString("\n\n")

	// Build list of paths to delete
	confirmMsg.WriteString("  The following items will be PERMANENTLY DELETED:\n\n")

	// Show paths - limit to first 8 items to avoid overflow
	maxDisplay := 8
	displayCount := 0

	// Check if coming from tree mode (single item deletion)
	if len(m.deletingItems) > 0 {
		// Tree mode - show deletingItems
		for i, item := range m.deletingItems {
			if displayCount >= maxDisplay {
				remaining := len(m.deletingItems) - maxDisplay
				confirmMsg.WriteString(fmt.Sprintf("  ... and %d more items\n", remaining))
				break
			}
			confirmMsg.WriteString(fmt.Sprintf("  %s %s  %s\n",
				pathStyle.Render("✗"),
				sizeStyle.Render(fmt.Sprintf("[%s]", ui.FormatSize(item.Size))),
				item.Path,
			))
			displayCount++
			_ = i
		}
	} else {
		// Normal selection mode - show selected items
		for i, item := range m.items {
			if !m.selected[i] {
				continue
			}
			if displayCount >= maxDisplay {
				remaining := selectedCount - maxDisplay
				confirmMsg.WriteString(fmt.Sprintf("  ... and %d more items\n", remaining))
				break
			}
			confirmMsg.WriteString(fmt.Sprintf("  %s %s  %s\n",
				pathStyle.Render("✗"),
				sizeStyle.Render(fmt.Sprintf("[%s]", ui.FormatSize(item.Size))),
				item.Path,
			))
			displayCount++
		}
	}

	confirmMsg.WriteString(fmt.Sprintf("\n  Total: %d items • %s\n\n", selectedCount, ui.FormatSize(selectedSize)))
	confirmMsg.WriteString("  Press [y] to confirm, [n] to cancel")

	b.WriteString(confirmBoxStyle.Render(confirmMsg.String()))
	return b.String()
}

// renderSelection shows the item selection list using table
func (m Model) renderSelection(b *strings.Builder) string {
	// Render table (already updated in Update())
	b.WriteString(m.itemsTable.View())
	b.WriteString("\n")

	// Status bar
	selectedCount := m.countSelected()
	selectedSize := m.selectedSize()
	status := fmt.Sprintf("\n📊 Selected: %d items • %s", selectedCount, ui.FormatSize(selectedSize))
	b.WriteString(statusStyle.Render(status))

	// Show random tip
	b.WriteString("\n\n")
	tipStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981")).
		Italic(true)
	b.WriteString(tipStyle.Render(m.currentTip))

	// Help
	help := "\n\n↑/↓: Navigate • Space: Toggle • a: All • n: None • c: Quick Clean Current • Enter: Clean Selected • ?: Help • q: Quit"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

// renderHelp shows comprehensive help screen
func (m Model) renderHelp(b *strings.Builder) string {
	helpBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(1, 2).
		Width(70)

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		Underline(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Bold(true)

	var help strings.Builder

	help.WriteString(headerStyle.Render("🔍 Mac Dev Cleaner - Help & Keyboard Shortcuts"))
	help.WriteString("\n\n")

	// Main List Navigation
	help.WriteString(headerStyle.Render("Main List Navigation"))
	help.WriteString("\n")
	help.WriteString(fmt.Sprintf("  %s        Move up/down\n", keyStyle.Render("↑/↓ or k/j")))
	help.WriteString(fmt.Sprintf("  %s          Toggle selection\n", keyStyle.Render("Space")))
	help.WriteString(fmt.Sprintf("  %s              Select all items\n", keyStyle.Render("a")))
	help.WriteString(fmt.Sprintf("  %s              Deselect all items\n", keyStyle.Render("n")))
	help.WriteString(fmt.Sprintf("  %s              Quick clean current item only\n", keyStyle.Render("c")))
	help.WriteString(fmt.Sprintf("  %s          Clean all selected items\n", keyStyle.Render("Enter")))
	help.WriteString(fmt.Sprintf("  %s        Drill down into folder (tree mode)\n", keyStyle.Render("→ or l")))
	help.WriteString("\n")

	// Tree Navigation
	help.WriteString(headerStyle.Render("Tree Navigation Mode"))
	help.WriteString("\n")
	help.WriteString(fmt.Sprintf("  %s        Move up/down in current folder\n", keyStyle.Render("↑/↓ or k/j")))
	help.WriteString(fmt.Sprintf("  %s        Drill deeper into subfolder\n", keyStyle.Render("→ or l")))
	help.WriteString(fmt.Sprintf("  %s        Go back to parent folder\n", keyStyle.Render("← or h")))
	help.WriteString(fmt.Sprintf("  %s          Toggle selection\n", keyStyle.Render("Space")))
	help.WriteString(fmt.Sprintf("  %s              Quick clean current item\n", keyStyle.Render("c")))
	help.WriteString(fmt.Sprintf("  %s              Refresh current folder\n", keyStyle.Render("r")))
	help.WriteString(fmt.Sprintf("  %s            Exit tree mode\n", keyStyle.Render("Esc")))
	help.WriteString("\n")

	// Important Notes
	help.WriteString(headerStyle.Render("Important Notes"))
	help.WriteString("\n")
	help.WriteString("  • 'c' key: Clears all selections and cleans ONLY current item\n")
	help.WriteString("  • Enter: Cleans ALL selected items (batch operation)\n")
	help.WriteString("  • Dry-run is ON by default - files are safe until confirmed\n")
	help.WriteString("  • All deletions are logged to ~/.dev-cleaner.log\n")
	help.WriteString("  • Tree mode: Delete items at any level, auto-refresh after\n")
	help.WriteString("\n")

	// Supported Ecosystems
	help.WriteString(headerStyle.Render("Supported Ecosystems"))
	help.WriteString("\n")
	help.WriteString("  🍎 Xcode • 🤖 Android • 📦 Node.js • 🐦 Flutter\n")
	help.WriteString("  🐍 Python • 🦀 Rust • 🐹 Go • 🍺 Homebrew\n")
	help.WriteString("  🐳 Docker • ☕ Java/Kotlin\n")
	help.WriteString("\n")

	// Tips
	help.WriteString(headerStyle.Render("Pro Tips"))
	help.WriteString("\n")
	help.WriteString("  ✨ Use 'c' for quick single-item cleanup\n")
	help.WriteString("  ✨ Use Space+Enter for batch cleanup of multiple items\n")
	help.WriteString("  ✨ Tree mode lets you explore and clean nested folders\n")
	help.WriteString("  ✨ Press any key from completion screen to rescan\n")
	help.WriteString("  ✨ Scans 10 ecosystems - can reclaim 30-100+ GB!\n")
	help.WriteString("\n")

	help.WriteString(helpStyle.Render("Press any key to return..."))

	b.WriteString(helpBoxStyle.Render(help.String()))
	return b.String()
}

func (m Model) renderResults(b *strings.Builder) string {
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("r/Enter: Rescan • Esc: Back • q: Quit"))
		return b.String()
	}

	var successCount int
	var freedSize int64
	for _, r := range m.results {
		if r.Success {
			successCount++
			freedSize += r.Size
			if r.WasDryRun {
				b.WriteString(fmt.Sprintf("  [DRY-RUN] Would delete: %s\n", r.Path))
			} else {
				b.WriteString(successStyle.Render(fmt.Sprintf("  ✓ Deleted: %s\n", r.Path)))
			}
		} else {
			b.WriteString(errorStyle.Render(fmt.Sprintf("  ✗ Failed: %s\n", r.Path)))
		}
	}

	summary := fmt.Sprintf("\n✅ Completed: %d items", successCount)
	if m.dryRun {
		summary += fmt.Sprintf(" (would free %s)", ui.FormatSize(freedSize))
	} else {
		summary += fmt.Sprintf(" (%s freed)", ui.FormatSize(freedSize))
	}
	b.WriteString(successStyle.Render(summary))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("r/Enter: Rescan • Esc: Back • q: Quit"))

	return b.String()
}

func (m Model) getTypeBadge(t types.CleanTargetType) string {
	style := lipgloss.NewStyle().Width(10).Bold(true)
	switch t {
	case types.TypeXcode:
		return style.Foreground(lipgloss.Color("#147EFB")).Render(string(t))
	case types.TypeAndroid:
		return style.Foreground(lipgloss.Color("#3DDC84")).Render(string(t))
	case types.TypeNode:
		return style.Foreground(lipgloss.Color("#68A063")).Render(string(t))
	case types.TypeFlutter:
		return style.Foreground(lipgloss.Color("#02569B")).Render(string(t))
	case types.TypePython:
		return style.Foreground(lipgloss.Color("#3776AB")).Render(string(t)) // Python blue
	case types.TypeRust:
		return style.Foreground(lipgloss.Color("#DEA584")).Render(string(t)) // Rust orange
	case types.TypeGo:
		return style.Foreground(lipgloss.Color("#00ADD8")).Render(string(t)) // Go cyan
	case types.TypeHomebrew:
		return style.Foreground(lipgloss.Color("#FBB040")).Render(string(t)) // Homebrew yellow
	case types.TypeDocker:
		return style.Foreground(lipgloss.Color("#2496ED")).Render(string(t)) // Docker blue
	case types.TypeJava:
		return style.Foreground(lipgloss.Color("#ED8B00")).Render(string(t)) // Java orange
	default:
		return style.Render(string(t))
	}
}

func (m Model) countSelected() int {
	count := 0
	for _, selected := range m.selected {
		if selected {
			count++
		}
	}
	return count
}

func (m Model) selectedSize() int64 {
	var size int64
	for i, selected := range m.selected {
		if selected && i < len(m.items) {
			size += m.items[i].Size
		}
	}
	return size
}

// renderStatusBar creates a unified status bar based on current state
func (m Model) renderStatusBar() string {
	var left, center, right string
	elapsed := time.Since(m.startTime)

	switch m.state {
	case StateScanning:
		// Left: State
		left = "[SCANNING]"

		// Center: Progress
		if m.currentScanning < len(m.scanningCategories) {
			current := m.scanningCategories[m.currentScanning]
			center = fmt.Sprintf("Scanning %s... (%d/%d)", current, m.currentScanning+1, len(m.scanningCategories))
		} else {
			center = "Almost done..."
		}

		// Right: Elapsed
		right = fmt.Sprintf("Elapsed: %ds", int(elapsed.Seconds()))

	case StateSelecting:
		// Left: State + Item count + Total size
		totalSize := int64(0)
		for _, item := range m.items {
			totalSize += item.Size
		}
		left = fmt.Sprintf("[SELECT] %d items • %s", len(m.items), ui.FormatSize(totalSize))

		// Center: Selected info
		selectedCount := m.countSelected()
		selectedSize := m.selectedSize()
		if selectedCount > 0 {
			center = fmt.Sprintf("Selected: %d/%d • %s", selectedCount, len(m.items), ui.FormatSize(selectedSize))
		} else {
			center = "No items selected"
		}

	case StateTree:
		// Left: State + Current path
		if m.currentNode != nil {
			left = fmt.Sprintf("[TREE] %s", m.currentNode.Name)

			// Center: Folder info
			center = fmt.Sprintf("%s • %d items", ui.FormatSize(m.currentNode.Size), m.currentNode.FileCount)
			if m.scanning {
				center += " • Scanning..."
			}

			// Depth indicator
			if m.currentNode.Depth > 0 {
				center += fmt.Sprintf(" • Depth: %d/%d", m.currentNode.Depth, m.maxDepth)
			}
		} else {
			left = "[TREE]"
			center = "Loading..."
		}

		// Right: Key hints
		right = "→:drill ←:back space:toggle c:quick esc:exit q:quit"

	case StateConfirming:
		// Left: State
		left = "[CONFIRM]"

		// Center: Confirmation prompt
		selectedCount := m.countSelected()
		selectedSize := m.selectedSize()
		center = fmt.Sprintf("Delete %d items (%s)?", selectedCount, ui.FormatSize(selectedSize))

		// Right: Key hints
		right = "y:yes n:no"

	case StateDeleting:
		// Left: State + Progress
		left = fmt.Sprintf("[DELETE] Progress: %.0f%%", m.percent*100)

		// Center: Items processed
		selectedCount := m.countSelected()
		processed := int(float64(selectedCount) * m.percent)
		center = fmt.Sprintf("%d/%d items", processed, selectedCount)

		// Right: Elapsed time
		deleteElapsed := time.Since(m.deleteStart)
		right = fmt.Sprintf("Elapsed: %ds", int(deleteElapsed.Seconds()))

	case StateDone:
		// Left: State
		left = "[DONE]"

		// Center: Summary
		var successCount int
		var freedSize int64
		for _, r := range m.results {
			if r.Success {
				successCount++
				freedSize += r.Size
			}
		}
		if m.dryRun {
			center = fmt.Sprintf("✓ %d items • Would free %s", successCount, ui.FormatSize(freedSize))
		} else {
			center = fmt.Sprintf("✓ %d items • %s freed", successCount, ui.FormatSize(freedSize))
		}

		// Right: Deletion time (frozen when completed) + hints
		right = fmt.Sprintf("Total: %ds • r:rescan esc:back q:quit", int(m.deleteDuration.Seconds()))
	}

	// Build status bar with sections
	leftPart := statusLeftStyle.Render(left)
	centerPart := statusCenterStyle.Render(center)
	rightPart := statusRightStyle.Render(right)

	// Calculate spacing
	leftWidth := lipgloss.Width(leftPart)
	centerWidth := lipgloss.Width(centerPart)
	rightWidth := lipgloss.Width(rightPart)

	totalContentWidth := leftWidth + centerWidth + rightWidth
	availableWidth := m.width
	if availableWidth == 0 {
		availableWidth = 80 // Default width
	}

	// Add padding between sections
	leftPadding := 2
	rightPadding := 2
	if totalContentWidth+leftPadding+rightPadding < availableWidth {
		// Center the middle section
		remainingSpace := availableWidth - totalContentWidth - leftPadding - rightPadding
		leftSpacing := strings.Repeat(" ", leftPadding)
		middleSpacing := strings.Repeat(" ", remainingSpace/2)
		rightSpacing := strings.Repeat(" ", remainingSpace-remainingSpace/2)

		content := leftPart + leftSpacing + centerPart + middleSpacing + rightPart + rightSpacing
		return statusBarStyle.Width(availableWidth).Render(content)
	}

	// If too wide, just concatenate with minimal spacing
	content := leftPart + " " + centerPart + " " + rightPart
	return statusBarStyle.Render(content)
}

// Run starts the TUI
func Run(items []types.ScanResult, dryRun bool, version string) error {
	m := NewModel(items, dryRun, version)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
