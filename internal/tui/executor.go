package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackchuka/gh-dep/internal/github"
	"github.com/jackchuka/gh-dep/internal/types"
)

// executeSelected returns a command that executes actions on selected PRs
func (m *Model) executeSelected() tea.Cmd {
	// Get list of selected PRs
	var selectedPRs []types.PR
	for i, pr := range m.filteredPRs {
		if m.selected[i] {
			selectedPRs = append(selectedPRs, pr)
		}
	}

	// Return a batch of commands - one for each PR
	var cmds []tea.Cmd
	for _, pr := range selectedPRs {
		cmds = append(cmds, m.executePRCmd(pr))
	}

	// Run all PR commands concurrently, but only mark completion after they all finish.
	return tea.Sequence(
		tea.Batch(cmds...),
		func() tea.Msg { return executionCompleteMsg{} },
	)
}

// executePRCmd creates a command to execute action on a single PR
func (m *Model) executePRCmd(pr types.PR) tea.Cmd {
	return func() tea.Msg {
		switch m.mode {
		case ModeApprove:
			return m.approvePR(pr)
		case ModeMerge:
			return m.mergePR(pr)
		case ModeApproveAndMerge:
			// First approve
			approveResult := m.approvePR(pr)
			if !approveResult.Success {
				return approveResult
			}
			// Then merge
			return m.mergePR(pr)
		}
		return ExecutionResult{
			PR:      pr,
			Action:  "unknown",
			Success: false,
			Error:   fmt.Errorf("unknown execution mode"),
		}
	}
}

func (m *Model) approvePR(pr types.PR) ExecutionResult {
	err := github.ApprovePR(pr.Repo, pr.Number)
	return ExecutionResult{
		PR:      pr,
		Action:  "approve",
		Success: err == nil,
		Error:   err,
	}
}

func (m *Model) mergePR(pr types.PR) ExecutionResult {
	// Check CI status if required
	if m.requireChecks {
		headSHA := pr.HeadSHA
		if headSHA == "" {
			sha, err := github.GetPRHead(pr.Repo, pr.Number)
			if err != nil {
				return ExecutionResult{
					PR:      pr,
					Action:  "merge (skipped)",
					Success: false,
					Error:   fmt.Errorf("failed to fetch PR head: %w", err),
				}
			}
			headSHA = sha
		}

		status, err := github.GetCIStatus(pr.Repo, headSHA)
		if err != nil {
			return ExecutionResult{
				PR:      pr,
				Action:  "merge (skipped)",
				Success: false,
				Error:   fmt.Errorf("failed to check CI status: %w", err),
			}
		}

		if !status.AllPassed {
			return ExecutionResult{
				PR:      pr,
				Action:  "merge (skipped)",
				Success: false,
				Error:   fmt.Errorf("CI checks not passing (state: %s)", status.State),
			}
		}
	}

	err := github.MergeViaPR(pr.Repo, pr.Number, m.mergeMethod)
	action := "merge (api)"

	return ExecutionResult{
		PR:      pr,
		Action:  action,
		Success: err == nil,
		Error:   err,
	}
}
