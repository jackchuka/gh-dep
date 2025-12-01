package cmd

import (
	"strings"

	"github.com/jackchuka/gh-dep/internal/config"
	"github.com/spf13/cobra"
)

// resolveScope determines owner and repo targets, falling back to @me when nothing is provided.
func resolveScope(cmd *cobra.Command, repoValue, ownerValue string, cfg *config.Config) (string, []string) {
	var repos []string

	if cmd.Flags().Changed("repo") {
		repos = cleanRepos(repoValue)
	} else if cfg != nil && len(cfg.GetRepos()) > 0 {
		repos = append(repos, cfg.GetRepos()...)
	}

	owner := ownerValue
	if owner == "" && len(repos) == 0 {
		owner = "@me"
	}

	return owner, repos
}

// cleanRepos splits comma-separated repos, trimming blanks.
func cleanRepos(repoValue string) []string {
	var repos []string
	for r := range strings.SplitSeq(repoValue, ",") {
		r = strings.TrimSpace(r)
		if r != "" {
			repos = append(repos, r)
		}
	}
	return repos
}
