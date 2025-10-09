package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jackchuka/gh-dep/internal/config"
	"github.com/jackchuka/gh-dep/internal/github"
	"github.com/jackchuka/gh-dep/internal/tui"
	"github.com/jackchuka/gh-dep/internal/types"
	"github.com/spf13/cobra"
)

var (
	rootLabel        string
	rootAuthor       string
	rootLimit        int
	rootRepo         string
	rootOwner        string
	rootMergeMethod  string
	rootMergeMode    string
	rootRequireCheck bool
	rootMode         string
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
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	label := rootLabel
	author := rootAuthor
	repo := rootRepo

	if !cmd.Flags().Changed("repo") && len(cfg.GetRepos()) > 0 {
		repo = strings.Join(cfg.GetRepos(), ",")
	}

	var allPRs []types.PR

	if rootOwner != "" {
		allPRs, err = github.SearchPRs(rootOwner, nil, label, author, rootLimit)
		if err != nil {
			return fmt.Errorf("failed to search PRs in owner %s: %w", rootOwner, err)
		}
	} else {
		if repo == "" {
			return fmt.Errorf("either --owner or --repo must be specified")
		}
		repos := strings.Split(repo, ",")
		allPRs, err = github.SearchPRs("", repos, label, author, rootLimit)
		if err != nil {
			if len(repos) > 1 {
				return fmt.Errorf("failed to search PRs across repos: %w", err)
			}
			return fmt.Errorf("failed to list PRs for %s: %w", repos[0], err)
		}
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
	model := tui.NewModelWithExecutor(allPRs, rootMergeMethod, rootMergeMode, rootRequireCheck, mode)

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

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(groupsCmd)
	rootCmd.AddCommand(approveCmd)
	rootCmd.AddCommand(mergeCmd)
}
