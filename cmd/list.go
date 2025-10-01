package cmd

import (
	"fmt"
	"strings"

	"github.com/jackchuka/gh-dep/internal/cache"
	"github.com/jackchuka/gh-dep/internal/config"
	"github.com/jackchuka/gh-dep/internal/github"
	"github.com/jackchuka/gh-dep/internal/types"
	"github.com/jackchuka/gh-dep/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List dependency PRs, optionally grouped by package@version",
	RunE:  runList,
}

var (
	listLabel  string
	listAuthor string
	listGroup  bool
	listLimit  int
	listRepo   string
	listOwner  string
	listJSON   bool
)

func init() {
	listCmd.Flags().BoolVar(&listGroup, "group", false, "Group PRs by package@version")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")

	listCmd.Flags().IntVar(&listLimit, "limit", 200, "Max PRs to fetch per repo")

	// additional filters
	listCmd.Flags().StringVar(&listLabel, "label", "", "PR label to filter")
	listCmd.Flags().StringVar(&listAuthor, "author", "dependabot[bot]", "PR author to filter (use 'any' for all)")
	listCmd.Flags().StringVarP(&listRepo, "repo", "R", "", "Target repo(s), comma-separated")
	listCmd.Flags().StringVarP(&listOwner, "owner", "O", "", "Target owner (user or org)")
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	label := listLabel
	author := listAuthor
	repo := listRepo

	if !cmd.Flags().Changed("repo") && len(cfg.GetRepos()) > 0 {
		repo = strings.Join(cfg.GetRepos(), ",")
	}

	var allPRs []types.PR

	if listOwner != "" {
		allPRs, err = github.SearchPRs(listOwner, nil, label, author, listLimit)
		if err != nil {
			return fmt.Errorf("failed to search PRs in owner %s: %w", listOwner, err)
		}
	} else {
		if repo == "" {
			return fmt.Errorf("either --owner or --repo must be specified")
		}
		repos := strings.Split(repo, ",")
		allPRs, err = github.SearchPRs("", repos, label, author, listLimit)
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

	if listGroup {
		groups := github.GroupPRs(allPRs, cfg.GetPatterns())

		// Cache the groups
		c := &types.Cache{
			Groups: groups,
		}
		if err := cache.Save(c); err != nil {
			return fmt.Errorf("failed to save cache: %w", err)
		}

		// Display groups
		display := ui.NewFromGroups(groups, listJSON)
		return display.DisplayGroups(groups)
	}

	display := ui.New(allPRs, listJSON)
	return display.DisplayList(allPRs)
}
