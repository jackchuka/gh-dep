package cmd

import (
	"fmt"

	"github.com/jackchuka/gh-dep/internal/cache"
	"github.com/jackchuka/gh-dep/internal/github"
	"github.com/jackchuka/gh-dep/internal/ui"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Bulk merge all PRs in a group",
	RunE:  runMerge,
}

var (
	mergeGroup         string
	mergeDryRun        bool
	mergeMethod        string
	mergeRequireChecks bool
)

func init() {
	mergeCmd.Flags().StringVar(&mergeGroup, "group", "", "Group key (package@version)")
	_ = mergeCmd.MarkFlagRequired("group")

	mergeCmd.Flags().BoolVar(&mergeDryRun, "dry-run", false, "Print actions without executing")
	mergeCmd.Flags().StringVar(&mergeMethod, "method", "squash", "Merge method: merge, squash, or rebase")
	mergeCmd.Flags().BoolVar(&mergeRequireChecks, "require-checks", true, "Require CI checks to pass")
}

func runMerge(cmd *cobra.Command, args []string) error {
	if mergeMethod != "merge" && mergeMethod != "squash" && mergeMethod != "rebase" {
		return fmt.Errorf("invalid merge method: %s (must be 'merge', 'squash', or 'rebase')", mergeMethod)
	}

	c, err := cache.Load()
	if err != nil {
		return fmt.Errorf("failed to load cache: %w", err)
	}

	if c == nil || len(c.Groups) == 0 {
		return fmt.Errorf("no cached groups found. Run 'gh dep list --group' first")
	}

	prs, ok := c.Groups[mergeGroup]
	if !ok {
		return fmt.Errorf("group '%s' not found in cache", mergeGroup)
	}

	display := ui.New(prs, false)

	for _, pr := range prs {
		if mergeRequireChecks {
			headSHA := pr.HeadSHA
			if headSHA == "" {
				sha, err := github.GetPRHead(pr.Repo, pr.Number)
				if err != nil {
					display.PrintAction("skipped", pr, fmt.Sprintf("failed to fetch PR head: %v", err))
					continue
				}
				headSHA = sha
			}

			status, err := github.GetCIStatus(pr.Repo, headSHA)
			if err != nil {
				display.PrintAction("skipped", pr, fmt.Sprintf("failed to check CI status: %v", err))
				continue
			}

			if !status.AllPassed {
				display.PrintAction("skipped", pr, fmt.Sprintf("CI checks not passing (state: %s)", status.State))
				continue
			}
		}

		if mergeDryRun {
			display.PrintAction("[dry-run] merge", pr)
			continue
		}

		mergeErr := github.MergeViaPR(pr.Repo, pr.Number, mergeMethod)
		if mergeErr != nil {
			display.PrintError("merge", pr, mergeErr)
			continue
		}

		display.PrintAction("merge", pr, "via API")
	}

	return nil
}
