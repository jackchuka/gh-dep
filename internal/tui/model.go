package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackchuka/gh-dep/internal/types"
)

type ExecutionMode int

const (
	ModeApprove ExecutionMode = iota
	ModeMerge
	ModeApproveAndMerge
)

func (m ExecutionMode) String() string {
	switch m {
	case ModeApprove:
		return "Approve Only"
	case ModeMerge:
		return "Merge Only"
	case ModeApproveAndMerge:
		return "Approve & Merge"
	default:
		return "Unknown"
	}
}

type ViewState int

const (
	ViewList ViewState = iota
	ViewExecuting
	ViewComplete
	ViewHelp
)

type ExecutionResult struct {
	PR      types.PR
	Action  string
	Success bool
	Error   error
}

type Model struct {
	prs             []types.PR
	filteredPRs     []types.PR
	selected        map[int]bool // index in filteredPRs
	cursor          int
	mode            ExecutionMode
	view            ViewState
	searchInput     textinput.Model
	searching       bool
	searchQuery     string
	executionResult []ExecutionResult
	executing       bool
	mergeMethod     string
	mergeMode       string
	requireChecks   bool
	width           int
	height          int
}

type keyMap struct {
	Up              key.Binding
	Down            key.Binding
	Select          key.Binding
	SelectAll       key.Binding
	DeselectAll     key.Binding
	ToggleMode      key.Binding
	ToggleMethod    key.Binding
	ToggleMergeMode key.Binding
	ToggleChecks    key.Binding
	Execute         key.Binding
	Search          key.Binding
	OpenBrowser     key.Binding
	Help            key.Binding
	Quit            key.Binding
	CancelSearch    key.Binding
	ConfirmSearch   key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Select: key.NewBinding(
		key.WithKeys(" ", "enter"),
		key.WithHelp("space/enter", "toggle selection"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "select all"),
	),
	DeselectAll: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "deselect all"),
	),
	ToggleMode: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "toggle mode"),
	),
	ToggleMethod: key.NewBinding(
		key.WithKeys("M"),
		key.WithHelp("M", "toggle merge method"),
	),
	ToggleMergeMode: key.NewBinding(
		key.WithKeys("D"),
		key.WithHelp("D", "toggle merge mode"),
	),
	ToggleChecks: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "toggle CI checks"),
	),
	Execute: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "execute"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	OpenBrowser: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "open in browser"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	CancelSearch: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel search"),
	),
	ConfirmSearch: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm search"),
	),
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	modeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	ciSuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	ciPendingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)

	ciFailureStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	ciUnknownStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

func NewModel(prs []types.PR, mergeMethod, mergeMode string, requireChecks bool, mode ExecutionMode) *Model {
	ti := textinput.New()
	ti.Placeholder = "Search PRs..."
	ti.CharLimit = 100

	return &Model{
		prs:           prs,
		filteredPRs:   prs,
		selected:      make(map[int]bool),
		cursor:        0,
		mode:          mode,
		view:          ViewList,
		searchInput:   ti,
		searching:     false,
		searchQuery:   "",
		mergeMethod:   mergeMethod,
		mergeMode:     mergeMode,
		requireChecks: requireChecks,
	}
}

// NewModelWithExecutor creates a new model with execution capabilities
func NewModelWithExecutor(prs []types.PR, mergeMethod, mergeMode string, requireChecks bool, mode ExecutionMode) *Model {
	return NewModel(prs, mergeMethod, mergeMode, requireChecks, mode)
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.view == ViewHelp {
			if msg.String() == "?" || msg.String() == "q" || msg.String() == "esc" {
				m.view = ViewList
			}
			return m, nil
		}

		if m.view == ViewComplete {
			if msg.String() == "q" || msg.String() == "enter" {
				return m, tea.Quit
			}
			return m, nil
		}

		if m.executing {
			return m, nil
		}

		if m.searching {
			switch {
			case key.Matches(msg, keys.ConfirmSearch):
				m.searching = false
				m.searchQuery = m.searchInput.Value()
				m.filterPRs()
				m.cursor = 0
				return m, nil
			case key.Matches(msg, keys.CancelSearch):
				m.searching = false
				m.searchInput.SetValue("")
				m.searchQuery = ""
				m.filterPRs()
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Help):
			m.view = ViewHelp
			return m, nil

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.filteredPRs)-1 {
				m.cursor++
			}

		case key.Matches(msg, keys.Select):
			m.selected[m.cursor] = !m.selected[m.cursor]

		case key.Matches(msg, keys.SelectAll):
			for i := range m.filteredPRs {
				m.selected[i] = true
			}

		case key.Matches(msg, keys.DeselectAll):
			m.selected = make(map[int]bool)

		case key.Matches(msg, keys.ToggleMode):
			m.mode = (m.mode + 1) % 3

		case key.Matches(msg, keys.ToggleMethod):
			switch m.mergeMethod {
			case "squash":
				m.mergeMethod = "merge"
			case "merge":
				m.mergeMethod = "rebase"
			case "rebase":
				m.mergeMethod = "squash"
			}

		case key.Matches(msg, keys.ToggleMergeMode):
			if m.mergeMode == "dependabot" {
				m.mergeMode = "api"
			} else {
				m.mergeMode = "dependabot"
			}

		case key.Matches(msg, keys.ToggleChecks):
			m.requireChecks = !m.requireChecks
			m.filterPRs()
			m.cursor = 0

		case key.Matches(msg, keys.Search):
			m.searching = true
			m.searchInput.Focus()
			return m, textinput.Blink

		case key.Matches(msg, keys.OpenBrowser):
			if len(m.filteredPRs) > 0 && m.cursor < len(m.filteredPRs) {
				return m, m.openPRInBrowser(m.filteredPRs[m.cursor])
			}

		case key.Matches(msg, keys.Execute):
			if m.hasSelection() {
				m.executing = true
				m.view = ViewExecuting
				return m, m.executeSelected()
			}
		}

	case ExecutionResult:
		m.executionResult = append(m.executionResult, msg)
		return m, nil

	case executionCompleteMsg:
		m.executing = false
		m.view = ViewComplete
		return m, nil
	}

	return m, nil
}

func (m *Model) View() string {
	switch m.view {
	case ViewHelp:
		return m.renderHelp()
	case ViewExecuting:
		return m.renderExecuting()
	case ViewComplete:
		return m.renderComplete()
	default:
		return m.renderList()
	}
}

func (m *Model) renderList() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("gh-dep Interactive Mode"))
	s.WriteString("\n\n")

	// Mode and settings indicator
	s.WriteString(headerStyle.Render("Action: "))
	s.WriteString(modeStyle.Render(m.mode.String()))
	s.WriteString("  ")

	s.WriteString(headerStyle.Render("Method: "))
	s.WriteString(modeStyle.Render(m.mergeMethod))
	s.WriteString("  ")

	s.WriteString(headerStyle.Render("Mode: "))
	s.WriteString(modeStyle.Render(m.mergeMode))

	if m.requireChecks {
		s.WriteString("  ")
		s.WriteString(headerStyle.Render("Checks: "))
		s.WriteString(modeStyle.Render("required"))
	}

	s.WriteString("\n\n")

	// Search bar
	if m.searching {
		s.WriteString(headerStyle.Render("Search: "))
		s.WriteString(m.searchInput.View())
		s.WriteString("\n\n")
	} else if m.searchQuery != "" {
		s.WriteString(helpStyle.Render(fmt.Sprintf("Filtered by: %q", m.searchQuery)))
		s.WriteString("\n\n")
	}

	// PR list
	s.WriteString(headerStyle.Render(fmt.Sprintf("PRs (%d selected / %d total):", m.countSelected(), len(m.filteredPRs))))
	s.WriteString("\n\n")

	visibleStart, visibleEnd := m.getVisibleRange()

	for i := visibleStart; i < visibleEnd; i++ {
		if i >= len(m.filteredPRs) {
			break
		}

		pr := m.filteredPRs[i]
		cursor := " "
		if i == m.cursor {
			cursor = cursorStyle.Render("❯")
		}

		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = selectedStyle.Render("[✓]")
		}

		ciStatus := formatCIStatus(pr.CIStatus)

		line := fmt.Sprintf("%s %s %s %s #%d - %s",
			cursor,
			checkbox,
			ciStatus,
			pr.Repo,
			pr.Number,
			pr.Title,
		)

		if i == m.cursor {
			line = cursorStyle.Render(line)
		}

		s.WriteString(line)
		s.WriteString("\n")
	}

	// Help
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("↑/↓: navigate • space: select • a: select all • d: deselect all"))
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("m/M/D/c: toggle settings • /: search • o: open • x: execute • ?: help • q: quit"))

	return s.String()
}

func (m *Model) renderExecuting() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Executing..."))
	s.WriteString("\n\n")

	for _, result := range m.executionResult {
		status := successStyle.Render("✓")
		if !result.Success {
			status = errorStyle.Render("✗")
		}

		msg := fmt.Sprintf("%s %s %s #%d",
			status,
			result.Action,
			result.PR.Repo,
			result.PR.Number,
		)

		if !result.Success {
			msg += errorStyle.Render(fmt.Sprintf(" - %v", result.Error))
		}

		s.WriteString(msg)
		s.WriteString("\n")
	}

	if !m.executing {
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("Press enter or q to exit"))
	}

	return s.String()
}

func (m *Model) renderComplete() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Execution Complete"))
	s.WriteString("\n\n")

	successCount := 0
	failCount := 0

	for _, result := range m.executionResult {
		status := successStyle.Render("✓")
		if !result.Success {
			status = errorStyle.Render("✗")
			failCount++
		} else {
			successCount++
		}

		msg := fmt.Sprintf("%s %s %s #%d",
			status,
			result.Action,
			result.PR.Repo,
			result.PR.Number,
		)

		if !result.Success {
			msg += errorStyle.Render(fmt.Sprintf(" - %v", result.Error))
		}

		s.WriteString(msg)
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(headerStyle.Render(fmt.Sprintf("Summary: %d succeeded, %d failed", successCount, failCount)))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("Press enter or q to exit"))

	return s.String()
}

func (m *Model) renderHelp() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Help - Keyboard Shortcuts"))
	s.WriteString("\n\n")

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"↑/k", "Move cursor up"},
		{"↓/j", "Move cursor down"},
		{"space/enter", "Toggle selection of current item"},
		{"a", "Select all visible PRs"},
		{"d", "Deselect all PRs"},
		{"m", "Toggle action mode (Approve → Merge → Approve & Merge)"},
		{"M", "Toggle merge method (squash → merge → rebase)"},
		{"D", "Toggle merge mode (dependabot → api)"},
		{"c", "Toggle CI checks requirement"},
		{"/", "Enter search mode"},
		{"esc", "Cancel search"},
		{"o", "Open current PR in browser"},
		{"x", "Execute selected actions"},
		{"?", "Show/hide this help screen"},
		{"q", "Quit the application"},
	}

	for _, sh := range shortcuts {
		s.WriteString(fmt.Sprintf("  %s - %s\n",
			selectedStyle.Render(fmt.Sprintf("%-15s", sh.key)),
			sh.desc,
		))
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("Press ? or q to return"))

	return s.String()
}

func (m *Model) hasSelection() bool {
	for _, selected := range m.selected {
		if selected {
			return true
		}
	}
	return false
}

func formatCIStatus(status string) string {
	switch status {
	case "success":
		return ciSuccessStyle.Render("✓")
	case "pending":
		return ciPendingStyle.Render("●")
	case "failure", "error":
		return ciFailureStyle.Render("✗")
	default:
		return ciUnknownStyle.Render("-")
	}
}

func (m *Model) countSelected() int {
	count := 0
	for _, selected := range m.selected {
		if selected {
			count++
		}
	}
	return count
}

func (m *Model) filterPRs() {
	query := strings.ToLower(m.searchQuery)
	var filtered []types.PR

	for _, pr := range m.prs {
		// Filter by search query if present
		if m.searchQuery != "" {
			matchesSearch := strings.Contains(strings.ToLower(pr.Title), query) ||
				strings.Contains(strings.ToLower(pr.Repo), query) ||
				strings.Contains(fmt.Sprintf("%d", pr.Number), query)
			if !matchesSearch {
				continue
			}
		}

		// Filter by CI status if requireChecks is enabled
		if m.requireChecks && pr.CIStatus != "success" {
			continue
		}

		filtered = append(filtered, pr)
	}

	m.filteredPRs = filtered

	// Clear selections that are no longer visible
	newSelected := make(map[int]bool)
	m.selected = newSelected
}

func (m *Model) getVisibleRange() (int, int) {
	if m.height == 0 {
		return 0, len(m.filteredPRs)
	}

	// Reserve space for header, footer, etc
	maxVisible := max(m.height-15, 5)

	start := max(m.cursor-maxVisible/2, 0)

	end := start + maxVisible
	if end > len(m.filteredPRs) {
		end = len(m.filteredPRs)
		start = max(end-maxVisible, 0)
	}

	return start, end
}

type executionCompleteMsg struct{}
