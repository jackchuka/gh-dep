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
	listLabel           string
	listAuthor          string
	listGroup           bool
	listLimit           int
	listRepo            string
	listOwner           string
	listJSON            bool
	listReviewRequested string
	listArchived        bool
)

func init() {
	listCmd.Flags().BoolVar(&listGroup, "group", false, "Group PRs by package@version")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")

	listCmd.Flags().IntVar(&listLimit, "limit", 200, "Max PRs to fetch per repo")

	// additional filters
	listCmd.Flags().StringVarP(&listRepo, "repo", "R", "", "Target repo(s), comma-separated")
	listCmd.Flags().StringVar(&listLabel, "label", "", "PR label to filter")
	listCmd.Flags().StringVar(&listAuthor, "author", "dependabot[bot]", "PR author to filter (use 'any' for all)")
	listCmd.Flags().StringVar(&listOwner, "owner", "", "Target owner (user or org)")
	listCmd.Flags().StringVar(&listReviewRequested, "review-requested", "", "Filter PRs by review requested from user or team (e.g., '@me' or 'username')")
	listCmd.Flags().BoolVar(&listArchived, "archived", false, "Include PRs from archived repositories")
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

	searchParams := github.SearchParams{
		Owner:           rootOwner,
		Repos:           strings.Split(repo, ","),
		Label:           label,
		Author:          author,
		Limit:           rootLimit,
		ReviewRequested: listReviewRequested,
		Archived:        listArchived,
	}

	allPRs, err := github.SearchPRs(searchParams)
	if err != nil {
		return fmt.Errorf("failed to search PRs: %w", err)
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
