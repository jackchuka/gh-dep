package cmd

import (
	"fmt"

	"github.com/jackchuka/gh-dep/internal/cache"
	"github.com/jackchuka/gh-dep/internal/ui"
	"github.com/spf13/cobra"
)

var groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Show cached groups from the last list --group",
	RunE:  runGroups,
}

var groupsJSON bool

func init() {
	groupsCmd.Flags().BoolVar(&groupsJSON, "json", false, "Output as JSON")
}

func runGroups(cmd *cobra.Command, args []string) error {
	c, err := cache.Load()
	if err != nil {
		return fmt.Errorf("failed to load cache: %w", err)
	}

	if c == nil || len(c.Groups) == 0 {
		return fmt.Errorf("no cached groups found. Run 'gh dep list --group' first")
	}

	display := ui.NewFromGroups(c.Groups, groupsJSON)
	return display.DisplayGroups(c.Groups)
}
