// Package tui provides interactive terminal UI using Bubble Tea
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/thanhdevapp/dev-cleaner/internal/cleaner"
	"github.com/thanhdevapp/dev-cleaner/internal/ui"
	"github.com/thanhdevapp/dev-cleaner/pkg/types"
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
}

// Model represents the TUI state
type Model struct {
	items    []types.ScanResult
	selected map[int]bool
	cursor   int
	width    int
	height   int
	dryRun   bool
	cleaning bool
	done     bool
	results  []cleaner.CleanResult
	err      error
	quitting bool
}

// NewModel creates a new TUI model
func NewModel(items []types.ScanResult, dryRun bool) Model {
	return Model{
		items:    items,
		selected: make(map[int]bool),
		dryRun:   dryRun,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.done {
			return m, tea.Quit
		}

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
				return m, m.performClean()
			}
		}

	case cleanResultMsg:
		m.done = true
		m.results = msg.results
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

// cleanResultMsg is sent when cleaning is complete
type cleanResultMsg struct {
	results []cleaner.CleanResult
	err     error
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

	// Title
	title := "🧹 Mac Dev Cleaner"
	if m.dryRun {
		title += " [DRY-RUN]"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// Done state
	if m.done {
		return m.renderResults(&b)
	}

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
		b.WriteString("\n\nPress any key to exit.")
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
	b.WriteString("\n\nPress any key to exit.")

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

// Run starts the TUI
func Run(items []types.ScanResult, dryRun bool) error {
	m := NewModel(items, dryRun)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
