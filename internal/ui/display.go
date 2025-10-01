package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/cli/go-gh/v2/pkg/term"
	"github.com/jackchuka/gh-dep/internal/types"
)

// UI encapsulates display state and configuration
type UI struct {
	multiRepo bool
	json      bool
}

// New creates a new UI instance
// Automatically detects if PRs span multiple repos
func New(prs []types.PR, json bool) *UI {
	return &UI{
		multiRepo: isMultiRepo(prs),
		json:      json,
	}
}

// NewFromGroups creates a new UI instance from grouped PRs
func NewFromGroups(groups map[string][]types.PR, json bool) *UI {
	return &UI{
		multiRepo: isMultiRepoGroups(groups),
		json:      json,
	}
}

// DisplayList prints PRs in a flat list format (JSON or table-like)
func (u *UI) DisplayList(prs []types.PR) error {
	if u.json {
		return u.displayListJSON(prs)
	}

	if len(prs) == 0 {
		return nil
	}

	isTTY := term.IsTerminal(os.Stdout)
	termWidth, _, _ := term.FromEnv().Size()
	table := tableprinter.New(os.Stdout, isTTY, termWidth)

	table.AddHeader([]string{"REPO", "PR", "TITLE"})
	for _, pr := range prs {
		table.AddField(pr.Repo)
		table.AddField("#" + strconv.Itoa(pr.Number))
		table.AddField(pr.Title)
		table.EndRow()
	}

	return table.Render()
}

// DisplayGroups prints PRs grouped by package@version (JSON or hierarchical format)
func (u *UI) DisplayGroups(groups map[string][]types.PR) error {
	if u.json {
		return u.displayGroupsJSON(groups)
	}

	isTTY := term.IsTerminal(os.Stdout)
	termWidth, _, _ := term.FromEnv().Size()

	// Sort groups alphabetically for consistent output
	sortedKeys := make([]string, 0, len(groups))
	for k := range groups {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	// Create single table for all groups
	table := tableprinter.New(os.Stdout, isTTY, termWidth)
	table.AddHeader([]string{"GROUP", "REPO", "PR", "URL"})

	for _, key := range sortedKeys {
		groupPRs := groups[key]

		// Sort PRs within group by repo name, then by PR number
		sort.Slice(groupPRs, func(i, j int) bool {
			if groupPRs[i].Repo != groupPRs[j].Repo {
				return groupPRs[i].Repo < groupPRs[j].Repo
			}
			return groupPRs[i].Number < groupPRs[j].Number
		})

		for i, pr := range groupPRs {
			// Group name only on first row of each group
			if i == 0 {
				table.AddField(key)
			} else {
				table.AddField("")
			}

			repoParts := strings.Split(pr.Repo, "/")
			repoShort := repoParts[len(repoParts)-1]
			table.AddField(repoShort)

			table.AddField("#" + strconv.Itoa(pr.Number))
			table.AddField(pr.URL)
			table.EndRow()
		}
	}

	return table.Render()
}

// PrintAction prints a standardized action message for a PR
// Examples:
//   - approve #123
//   - [owner/repo] approve #123
//   - merge #123 via dependabot
//   - [owner/repo] skipped #123: CI checks not passing
func (u *UI) PrintAction(action string, pr types.PR, details ...string) {
	prefix := ""
	if u.multiRepo {
		prefix = fmt.Sprintf("[%s] ", pr.Repo)
	}

	message := fmt.Sprintf("%s #%d", action, pr.Number)
	if len(details) > 0 {
		message += ": " + details[0]
	}

	fmt.Printf("%s%s\n", prefix, message)
}

// PrintError prints a standardized error message for a PR action
func (u *UI) PrintError(action string, pr types.PR, err error) {
	prefix := ""
	if u.multiRepo {
		prefix = fmt.Sprintf("[%s] ", pr.Repo)
	}

	fmt.Printf("%sfailed to %s #%d: %v\n", prefix, action, pr.Number, err)
}

func isMultiRepo(prs []types.PR) bool {
	if len(prs) == 0 {
		return false
	}

	firstRepo := prs[0].Repo
	for _, pr := range prs {
		if pr.Repo != firstRepo {
			return true
		}
	}
	return false
}

func isMultiRepoGroups(groups map[string][]types.PR) bool {
	for _, prs := range groups {
		if len(prs) > 0 {
			firstRepo := prs[0].Repo
			for _, pr := range prs {
				if pr.Repo != firstRepo {
					return true
				}
			}
		}
	}
	return false
}

func (u *UI) displayListJSON(prs []types.PR) error {
	data, err := json.MarshalIndent(prs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func (u *UI) displayGroupsJSON(groups map[string][]types.PR) error {
	data, err := json.MarshalIndent(groups, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
