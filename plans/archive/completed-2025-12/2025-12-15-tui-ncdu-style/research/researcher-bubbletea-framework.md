# Bubble Tea Framework Research

## Overview

Bubble Tea is a Go TUI framework implementing the **Elm Architecture** (functional, message-driven pattern). Perfect for building interactive CLI apps like ncdu-style disk cleaner with selection, progress bars, and confirmations.

## Architecture: Model-View-Update (MVU)

Three core methods define every Bubble Tea app:

```go
type Model interface {
    Init() tea.Cmd                          // Initial state + startup commands
    Update(msg tea.Msg) (Model, tea.Cmd)   // Handle events, update state
    View() string                           // Render UI as string
}
```

**Flow**: User input → tea.KeyMsg → Update() → new Model → View() → re-render terminal

All state immutably lives in the model struct. Commands enable side effects (API calls, timers, file I/O).

## Event Handling: Keyboard & Beyond

```go
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            return m, tea.Quit
        case "up", "k":
            m.cursor--
        case "down", "j":
            m.cursor++
        case " ":  // Space to toggle checkbox
            m.items[m.cursor].checked = !m.items[m.cursor].checked
        case "enter":
            return m, m.confirmDelete()
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }
    return m, nil
}
```

Key types: `tea.KeyMsg`, `tea.WindowSizeMsg`, `tea.TimerMsg`, custom types via channels.

## List Component with Checkboxes

**1. Custom Item Type:**
```go
type Item struct {
    path    string
    size    int64
    checked bool
}

func (i Item) FilterValue() string { return i.path }
func (i Item) Title() string       { return i.path }
func (i Item) Description() string { return fmt.Sprintf("%.2f GB", float64(i.size)/1e9) }
```

**2. Custom Delegate (renderer):**
```go
type Delegate struct{}

func (d Delegate) Render(w io.Writer, m list.Model, idx int, item list.Item) {
    i := item.(Item)
    checkbox := "☐"
    if i.checked { checkbox = "☑" }

    selected := idx == m.Index()
    cursor := " "
    if selected { cursor = ">" }

    fmt.Fprintf(w, "%s %s %s", cursor, checkbox, i.Title())
}

func (d Delegate) Height() int   { return 1 }
func (d Delegate) Spacing() int  { return 0 }
func (d Delegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
```

**3. List in Model:**
```go
type Model struct {
    list   list.Model
    items  []Item
}

func (m *Model) toggleItem() {
    if selected := m.list.SelectedItem(); selected != nil {
        idx := m.list.Index()
        m.items[idx].checked = !m.items[idx].checked
    }
}
```

## Progress Bars & Loading States

```go
import "github.com/charmbracelet/bubbles/progress"

type Model struct {
    progress progress.Model
    percent  float64
}

func (m Model) View() string {
    return m.progress.ViewAs(m.percent)
}

// In Update:
case DeleteProgressMsg:
    m.percent = msg.percent
    if m.percent >= 1.0 {
        m.state = Complete
    }
```

Progress animates automatically. Use custom commands for background deletion:

```go
func deleteCommand(items []Item) tea.Cmd {
    return func() tea.Msg {
        for i, item := range items {
            if item.checked {
                os.RemoveAll(item.path)
            }
            return DeleteProgressMsg{percent: float64(i) / float64(len(items))}
        }
        return DeleteCompleteMsg{}
    }
}
```

## Styling with Lipgloss

```go
import "github.com/charmbracelet/lipgloss"

var (
    selectedStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#ff00ff")).
        Background(lipgloss.Color("#333333")).
        Bold(true)

    headerStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#00aa00")).
        Bold(true).
        Padding(0, 2)

    deleteStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#ff0000")).
        BorderStyle(lipgloss.RoundedBorder()).
        Padding(1, 2)
)

func (m Model) View() string {
    header := headerStyle.Render("Dev Cleaner - Select Items")
    list := m.list.View()
    return lipgloss.JoinVertical(lipgloss.Left, header, list)
}
```

**Color options**: ANSI 16 (`Color("5")`), 256 (`Color("201")`), hex (`Color("#FF00FF")`).

## Confirmation Dialog Pattern

```go
type State string
const (
    Listing    State = "listing"
    Confirming State = "confirming"
    Deleting   State = "deleting"
)

type Model struct {
    state    State
    selected []Item
}

func (m Model) View() string {
    switch m.state {
    case Listing:
        return m.list.View()
    case Confirming:
        msg := fmt.Sprintf("Delete %d items? (y/n)", len(m.selected))
        return deleteStyle.Render(msg)
    case Deleting:
        return progressStyle.Render(m.progress.View())
    }
    return ""
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    if m.state == Confirming {
        if km, ok := msg.(tea.KeyMsg); ok {
            if km.String() == "y" {
                m.state = Deleting
                return m, m.deleteCmd()
            }
            if km.String() == "n" {
                m.state = Listing
            }
        }
    }
    // ... handle other states
}
```

## Key Features for Dev Cleaner

| Need | Solution |
|------|----------|
| Multi-select | Toggle `.checked` on space |
| Sort by size | Custom sort in model |
| Real-time scanning | Fire commands from Init() |
| Progress during delete | Progress component + Cmd |
| Quit confirmation | State machine (Listing → Confirming → Deleting) |
| Disk usage tree | Custom render with nesting |
| Keyboard nav | List component handles up/down |

## Example Projects

Bubble Tea repo includes 51+ examples:
- `list-fancy`: Custom styled lists
- `progress-animated`: Progress bars
- `mouse`: Input handling
- `table`: Tabular selection
- `credit-card-form`: Multi-step flows with confirmation

## Unresolved Questions

- Should we use `bubbles/list` or build custom render for tree hierarchy (disk paths)?
- Command execution during deletion: use goroutines or Bubble Tea Cmd pattern?
- Progress granularity: per-file or per-category deletion?

---

**Sources:**
- [Bubble Tea GitHub](https://github.com/charmbracelet/bubbletea)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Lipgloss Styling](https://github.com/charmbracelet/lipgloss)
- [List Component Docs](https://pkg.go.dev/github.com/charmbracelet/bubbles/list)
- [Processing Input Tutorial](https://dev.to/andyhaskell/processing-user-input-in-bubble-tea-with-a-menu-component-222i)
