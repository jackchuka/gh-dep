package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackchuka/gh-dep/internal/github"
	"github.com/jackchuka/gh-dep/internal/tui"
	"github.com/spf13/cobra"
)

var (
	rootLabel           string
	rootAuthor          string
	rootLimit           int
	rootRepo            string
	rootOwner           string
	rootMergeMethod     string
	rootMergeMode       string
	rootRequireCheck    bool
	rootMode            string
	rootReviewRequested string
	rootArchived        bool
)

var rootCmd = &cobra.Command{
	Use:   "gh-dep",
	Short: "Streamline dependency PR review and merge workflow",
	Long: `gh-dep is a GitHub CLI extension that helps you manage automated
dependency update PRs by grouping, bulk approving, and bulk merging them.

When run without subcommands, launches interactive TUI mode.`,
	SilenceUsage: true,
	RunE:         runRoot,
}

func Execute() error {
	return rootCmd.Execute()
}

func runRoot(cmd *cobra.Command, args []string) error {
	searchParams := github.SearchParams{
		Owner:           rootOwner,
		Repos:           strings.Split(rootRepo, ","),
		Label:           rootLabel,
		Author:          rootAuthor,
		Limit:           rootLimit,
		ReviewRequested: rootReviewRequested,
		Archived:        rootArchived,
	}

	allPRs, err := github.SearchPRs(searchParams)
	if err != nil {
		return fmt.Errorf("failed to search PRs: %w", err)
	}

	if len(allPRs) == 0 {
		fmt.Println("No dependency PRs found")
		return nil
	}

	// Parse mode
	mode := tui.ModeApprove // default
	switch rootMode {
	case "approve":
		mode = tui.ModeApprove
	case "merge":
		mode = tui.ModeMerge
	case "approve-and-merge", "both":
		mode = tui.ModeApproveAndMerge
	}

	// Launch TUI
	model := tui.NewModel(allPRs, rootMergeMethod, rootMergeMode, rootRequireCheck, mode, searchParams)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}

func init() {
	rootCmd.Flags().IntVar(&rootLimit, "limit", 200, "Max PRs to fetch per repo")
	rootCmd.Flags().StringVar(&rootLabel, "label", "", "PR label to filter")
	rootCmd.Flags().StringVar(&rootAuthor, "author", "dependabot[bot]", "PR author to filter (use 'any' for all)")
	rootCmd.Flags().StringVarP(&rootRepo, "repo", "R", "", "Target repo(s), comma-separated")
	rootCmd.Flags().StringVar(&rootOwner, "owner", "", "Target owner (user or org)")
	rootCmd.Flags().StringVar(&rootMergeMethod, "merge-method", "squash", "Merge method: merge, squash, or rebase")
	rootCmd.Flags().StringVar(&rootMergeMode, "merge-mode", "dependabot", "Merge mode: dependabot or api")
	rootCmd.Flags().BoolVar(&rootRequireCheck, "require-checks", false, "Require CI checks to pass (API mode only)")
	rootCmd.Flags().StringVar(&rootMode, "mode", "approve", "Execution mode: approve, merge, or approve-and-merge (both)")
	rootCmd.Flags().StringVar(&rootReviewRequested, "review-requested", "", "Filter PRs by review requested from user or team (e.g., '@me' or 'username')")
	rootCmd.Flags().BoolVar(&rootArchived, "archived", false, "Include PRs from archived repositories")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(groupsCmd)
	rootCmd.AddCommand(approveCmd)
	rootCmd.AddCommand(mergeCmd)
}
