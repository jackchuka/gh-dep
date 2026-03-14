package cmd

import (
	"fmt"
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

// resolveAuthors picks the effective author filters based on flags.
// --author wins; otherwise --bot is mapped to known logins.
func resolveAuthors(cmd *cobra.Command, authorValue, botValue string) ([]string, error) {
	if cmd.Flags().Changed("author") {
		return []string{authorValue}, nil
	}

	switch normalized := strings.TrimSpace(strings.ToLower(strings.TrimPrefix(botValue, "@"))); normalized {
	case "all", "both", "":
		return []string{"dependabot[bot]", "renovate[bot]"}, nil
	case "dependabot", "dependabot[bot]":
		return []string{"dependabot[bot]"}, nil
	case "renovate", "renovate[bot]":
		return []string{"renovate[bot]"}, nil
	default:
		return nil, fmt.Errorf("invalid value for --bot: %q (expected all, dependabot, or renovate)", botValue)
	}
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
