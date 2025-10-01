package cmd

import (
	"fmt"

	"github.com/jackchuka/gh-dep/internal/cache"
	"github.com/jackchuka/gh-dep/internal/github"
	"github.com/jackchuka/gh-dep/internal/ui"
	"github.com/spf13/cobra"
)

var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Bulk approve all PRs in a group",
	RunE:  runApprove,
}

var (
	approveGroup  string
	approveDryRun bool
	approveRepo   string
	approveOrg    string
)

func init() {
	approveCmd.Flags().StringVar(&approveGroup, "group", "", "Group key (package@version)")
	_ = approveCmd.MarkFlagRequired("group")

	approveCmd.Flags().BoolVar(&approveDryRun, "dry-run", false, "Print actions without executing")
	approveCmd.Flags().StringVarP(&approveRepo, "repo", "R", "", "Target repo(s), comma-separated")
	approveCmd.Flags().StringVarP(&approveOrg, "org", "O", "", "Target organization")
}

func runApprove(cmd *cobra.Command, args []string) error {
	c, err := cache.Load()
	if err != nil {
		return fmt.Errorf("failed to load cache: %w", err)
	}

	if c == nil || len(c.Groups) == 0 {
		return fmt.Errorf("no cached groups found. Run 'gh dep list --group' first")
	}

	prs, ok := c.Groups[approveGroup]
	if !ok {
		return fmt.Errorf("group '%s' not found in cache", approveGroup)
	}

	display := ui.New(prs, false)

	for _, pr := range prs {
		if approveDryRun {
			display.PrintAction("approve", pr)
			continue
		}

		if err := github.ApprovePR(pr.Repo, pr.Number); err != nil {
			display.PrintError("approve", pr, err)
			continue
		}

		display.PrintAction("approve", pr)
	}

	return nil
}
