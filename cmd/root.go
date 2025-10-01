package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-dep",
	Short: "Streamline dependency PR review and merge workflow",
	Long: `gh-dep is a GitHub CLI extension that helps you manage automated
dependency update PRs by grouping, bulk approving, and bulk merging them.`,
	SilenceUsage: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(groupsCmd)
	rootCmd.AddCommand(approveCmd)
	rootCmd.AddCommand(mergeCmd)
}
