package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh/v2"
	"github.com/jackchuka/gh-dep/internal/types"
)

func (m Model) openPRInBrowser(pr types.PR) tea.Cmd {
	return func() tea.Msg {
		args := []string{"pr", "view", pr.URL, "--web"}
		_, _, _ = gh.Exec(args...)
		return nil
	}
}
