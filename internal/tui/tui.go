// Package tui provides interactive terminal UI using Bubble Tea
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
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
)

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
	Up      key.Binding
	Down    key.Binding
	Toggle  key.Binding
	All     key.Binding
	None    key.Binding
	Confirm key.Binding
	Quit    key.Binding
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
	treeMode     bool              // True when in tree view
	currentNode  *types.TreeNode   // Current tree node
	nodeStack    []*types.TreeNode // Breadcrumb trail
	cursorStack  []int             // Cursor positions for each level
	maxDepth     int               // Max depth limit
	treeSelected map[string]bool   // Selected items in tree
	scanning     bool              // True while scanning

	// Time tracking
	startTime   time.Time // Session start time
	deleteStart time.Time // Delete operation start time

	// Scanning progress
	scanningCategories []string // Categories being scanned
	scanComplete       map[string]bool // Which categories are complete
	currentScanning    int // Index of currently scanning category
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
	}

	// Start in scanning state if we have items
	initialState := StateSelecting
	if len(items) > 0 && len(categories) > 0 {
		initialState = StateScanning
	}

	return Model{
		state:              initialState,
		items:              items,
		selected:           make(map[int]bool),
		dryRun:             dryRun,
		version:            version,
		spinner:            s,
		progress:           p,
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
	}
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
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case deleteProgressMsg:
		m.percent = msg.percent
		cmd := m.progress.SetPercent(m.percent)
		return m, cmd

	case tea.KeyMsg:
		// Handle based on current state
		switch m.state {
		case StateDone:
			// 'q' to quit, any other key to rescan and continue
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				m.quitting = true
				return m, tea.Quit
			}
			// Rescan and return to selection
			return m, m.rescanItems()

		case StateConfirming:
			switch msg.String() {
			case "y", "Y":
				m.state = StateDeleting
				m.percent = 0
				m.deleteStart = time.Now()
				return m, tea.Batch(m.performClean(), m.progress.SetPercent(0))
			case "n", "N", "esc":
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
				if m.countSelected() > 0 {
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
			}
		}

	case cleanResultMsg:
		m.state = StateDone
		m.results = msg.results
		m.err = msg.err
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
		return m, nil

	case rescanItemsMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		// Reset state and show new items
		m.items = msg.items
		m.selected = make(map[int]bool)
		m.cursor = 0
		m.state = StateSelecting
		m.results = nil
		m.err = nil
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

		opts := types.ScanOptions{
			MaxDepth:       3,
			IncludeXcode:   true,
			IncludeAndroid: true,
			IncludeNode:    true,
			IncludeFlutter: true,
		}

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

// performClean starts the cleaning process
func (m Model) performClean() tea.Cmd {
	return func() tea.Msg {
		c, err := cleaner.New(m.dryRun)
		if err != nil {
			return cleanResultMsg{err: err}
		}
		defer c.Close()

		var toClean []types.ScanResult
		for i, item := range m.items {
			if m.selected[i] {
				toClean = append(toClean, item)
			}
		}

		results, err := c.Clean(toClean)
		return cleanResultMsg{results: results, err: err}
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
		b.WriteString("🗑️  Deleting selected items...\n\n")
		b.WriteString(m.progress.View())
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press q to cancel"))
		content = b.String()

	case StateConfirming:
		content = m.renderConfirmation(&b)

	case StateTree:
		content = m.renderTreeView(&b)

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

			line := fmt.Sprintf("%s%s %s %s  %s",
				cursor,
				checkbox,
				icon,
				sizeStyle.Render(fmt.Sprintf("%10s", sizeStr)),
				child.Name,
			)

			if i == m.cursor {
				b.WriteString(selectedItemStyle.Render(line))
			} else {
				b.WriteString(itemStyle.Render(line))
			}
			b.WriteString("\n")
		}
	}

	// Depth info
	if m.currentNode.Depth >= m.maxDepth-1 {
		warning := fmt.Sprintf("\n⚠️  Depth %d/%d - Approaching limit",
			m.currentNode.Depth+1, m.maxDepth)
		b.WriteString(errorStyle.Render(warning))
	}

	// Help
	help := "\n\n↑/↓: Navigate • →/l: Drill down • ←/h: Go back • r: Refresh • Space: Toggle • Esc: Exit • q: Quit"
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

// renderConfirmation shows the confirmation dialog
func (m Model) renderConfirmation(b *strings.Builder) string {
	selectedCount := m.countSelected()
	selectedSize := m.selectedSize()

	// Confirmation box style
	confirmBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#F59E0B")).
		Padding(1, 2).
		Width(50)

	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Bold(true)

	confirmMsg := fmt.Sprintf(
		"%s\n\n"+
			"  Items: %d\n"+
			"  Size:  %s\n\n"+
			"  Press [y] to confirm, [n] to cancel",
		warningStyle.Render("⚠️  Confirm Deletion"),
		selectedCount,
		ui.FormatSize(selectedSize),
	)

	b.WriteString(confirmBoxStyle.Render(confirmMsg))
	return b.String()
}

// renderSelection shows the item selection list
func (m Model) renderSelection(b *strings.Builder) string {
	// Items list
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("▸ ")
		}

		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = checkboxStyle.Render("[✓]")
		}

		// Type badge
		typeBadge := m.getTypeBadge(item.Type)

		// Size with color
		sizeStr := ui.FormatSize(item.Size)
		sizeStyle := lipgloss.NewStyle().Width(10).Align(lipgloss.Right)
		if item.Size > 1024*1024*1024 {
			sizeStyle = sizeStyle.Foreground(lipgloss.Color("#EF4444")).Bold(true)
		} else if item.Size > 100*1024*1024 {
			sizeStyle = sizeStyle.Foreground(lipgloss.Color("#F59E0B"))
		} else {
			sizeStyle = sizeStyle.Foreground(lipgloss.Color("#10B981"))
		}

		line := fmt.Sprintf("%s%s %s %s  %s",
			cursor,
			checkbox,
			typeBadge,
			sizeStyle.Render(sizeStr),
			item.Name,
		)

		if i == m.cursor {
			b.WriteString(selectedItemStyle.Render(line))
		} else {
			b.WriteString(itemStyle.Render(line))
		}
		b.WriteString("\n")
	}

	// Status bar
	selectedCount := m.countSelected()
	selectedSize := m.selectedSize()
	status := fmt.Sprintf("\n📊 Selected: %d items • %s", selectedCount, ui.FormatSize(selectedSize))
	b.WriteString(statusStyle.Render(status))

	// Help
	help := "\n\n↑/↓: Navigate • Space: Toggle • a: All • n: None • Enter: Clean • q: Quit"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

func (m Model) renderResults(b *strings.Builder) string {
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\nPress any key to rescan, q to quit.")
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
	b.WriteString("\n\nPress any key to rescan, q to quit.")

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

		// Right: Key hints
		right = "↑↓:nav space:toggle a:all n:none enter:clean q:quit"

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
		right = "→:drill ←:back r:refresh esc:exit q:quit"

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

		// Right: Total time + hints
		right = fmt.Sprintf("Total: %ds • any key:rescan q:quit", int(elapsed.Seconds()))
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
