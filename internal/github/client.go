package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/jackchuka/gh-dep/internal/parser"
	"github.com/jackchuka/gh-dep/internal/types"
)

// GetClient returns a GitHub REST API client
func GetClient() (*api.RESTClient, error) {
	return api.DefaultRESTClient()
}

// SearchPRs searches for PRs based on org/repo, label, author, and limit
func SearchPRs(
	owner string,
	repos []string,
	label string,
	author string,
	limit int,
) ([]types.PR, error) {
	args := []string{"search", "prs", "is:open"}

	if owner != "" {
		args = append(args, "--owner", owner)
	}
	if len(repos) > 0 {
		for _, repo := range repos {
			args = append(args, "--repo", repo)
		}
	}

	if label != "" {
		args = append(args, "--label", label)
	}
	if author != "" && author != "any" {
		args = append(args, "--author", author)
	}
	args = append(args, "--json", "number,title,author,url,repository")
	if limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", limit))
	}

	var rawPRs []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		URL        string `json:"url"`
		Repository struct {
			NameWithOwner string `json:"nameWithOwner"`
		} `json:"repository"`
	}

	stdOut, stdErr, err := gh.Exec(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search PRs: %w\n%s", err, stdErr.String())
	}

	if err := parseJSON(stdOut.String(), &rawPRs); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	// Convert raw PRs to types.PR first
	prs := make([]types.PR, len(rawPRs))
	for i, raw := range rawPRs {
		prs[i] = types.PR{
			Number: raw.Number,
			Title:  raw.Title,
			Author: raw.Author.Login,
			Repo:   raw.Repository.NameWithOwner,
			URL:    raw.URL,
		}
	}

	// Fetch CI status concurrently with worker pool
	const maxWorkers = 10
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxWorkers)

	for i := range prs {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			// Fetch HEAD SHA and CI status
			headSHA, err := GetPRHead(prs[idx].Repo, prs[idx].Number)
			if err == nil {
				prs[idx].HeadSHA = headSHA
				ciStatus, err := GetCIStatus(prs[idx].Repo, headSHA)
				if err == nil && ciStatus != nil {
					prs[idx].CIStatus = ciStatus.State
				}
			}
		}(i)
	}

	wg.Wait()
	return prs, nil
}

// ListRepos fetches all repositories for an organization
func ListRepos(org string) ([]string, error) {
	args := []string{"repo", "list", org, "--json", "nameWithOwner", "--limit", "1000"}

	var repos []struct {
		NameWithOwner string `json:"nameWithOwner"`
	}

	stdOut, _, err := gh.Exec(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list repos: %w", err)
	}

	if err := parseJSON(stdOut.String(), &repos); err != nil {
		return nil, fmt.Errorf("failed to parse repos: %w", err)
	}

	var repoNames []string
	for _, repo := range repos {
		repoNames = append(repoNames, repo.NameWithOwner)
	}

	return repoNames, nil
}

// GetCurrentRepo gets the current repository from cwd
func GetCurrentRepo() (string, error) {
	args := []string{"repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner"}

	stdOut, _, err := gh.Exec(args...)
	if err != nil {
		return "", fmt.Errorf("failed to get current repo (are you in a git repository?): %w", err)
	}

	return strings.TrimSpace(stdOut.String()), nil
}

// GroupPRs groups PRs by package@version
func GroupPRs(prs []types.PR, customPatterns []string) map[string][]types.PR {
	groups := make(map[string][]types.PR)

	for _, pr := range prs {
		update := parser.ParseTitle(pr.Title, customPatterns)
		key := update.GroupKey()
		groups[key] = append(groups[key], pr)
	}

	return groups
}

// ApprovePR approves a pull request
func ApprovePR(repo string, number int) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	body := map[string]string{
		"event": "APPROVE",
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("repos/%s/pulls/%d/reviews", repo, number)
	if err := client.Post(path, bytes.NewReader(bodyBytes), nil); err != nil {
		return fmt.Errorf("failed to approve PR #%d: %w", number, err)
	}

	return nil
}

// MergeViaPR merges a PR via GitHub API
func MergeViaPR(repo string, number int, method string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	body := map[string]string{
		"merge_method": method,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("repos/%s/pulls/%d/merge", repo, number)
	if err := client.Put(path, bytes.NewReader(bodyBytes), nil); err != nil {
		return fmt.Errorf("failed to merge PR #%d: %w", number, err)
	}

	return nil
}

// MergeViaDependabot posts a dependabot merge comment
func MergeViaDependabot(repo string, number int, method string) error {
	client, err := GetClient()
	if err != nil {
		return err
	}

	var comment string
	switch method {
	case "squash":
		comment = "@dependabot squash and merge"
	case "rebase":
		comment = "@dependabot rebase and merge"
	default: // merge
		comment = "@dependabot merge"
	}

	body := map[string]string{
		"body": comment,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("repos/%s/issues/%d/comments", repo, number)
	if err := client.Post(path, bytes.NewReader(bodyBytes), nil); err != nil {
		return fmt.Errorf("failed to comment on PR #%d: %w", number, err)
	}

	return nil
}

// CheckStatus represents CI status
type CheckStatus struct {
	State     string // success, pending, failure, error
	AllPassed bool
}

type statusResponse struct {
	State    string          `json:"state"`
	Statuses []statusContext `json:"statuses"`
}

type statusContext struct {
	State string `json:"state"`
}

type checkSuiteResponse struct {
	CheckSuites []checkSuite `json:"check_suites"`
}

type checkSuite struct {
	Status     string  `json:"status"`
	Conclusion *string `json:"conclusion"`
}

// GetPRHead fetches the HEAD SHA for a PR (useful when SearchPRs doesn't return it)
func GetPRHead(repo string, number int) (string, error) {
	client, err := GetClient()
	if err != nil {
		return "", err
	}

	var pr struct {
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}

	path := fmt.Sprintf("repos/%s/pulls/%d", repo, number)
	if err := client.Get(path, &pr); err != nil {
		return "", fmt.Errorf("failed to get PR #%d: %w", number, err)
	}

	return pr.Head.SHA, nil
}

// GetCIStatus checks the CI status for a PR
func GetCIStatus(repo string, sha string) (*CheckStatus, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	var suites checkSuiteResponse

	suitePath := fmt.Sprintf("repos/%s/commits/%s/check-suites", repo, sha)
	suitesErr := client.Get(suitePath, &suites)

	var status statusResponse

	statusPath := fmt.Sprintf("repos/%s/commits/%s/status", repo, sha)
	statusErr := client.Get(statusPath, &status)

	if statusErr != nil && suitesErr != nil {
		return nil, fmt.Errorf("failed to get status for %s@%s: status error: %v; check suites error: %v",
			repo, sha, statusErr, suitesErr)
	}

	state := deriveCIState(suites, status)

	return &CheckStatus{
		State:     state,
		AllPassed: state == "success",
	}, nil
}

// parseJSON is a helper to parse JSON strings (gh.Exec returns Bytes that have String() method)
func parseJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

func deriveCIState(
	suites checkSuiteResponse,
	status statusResponse,
) string {
	allCompleted := true
	allSuccess := true

	if len(suites.CheckSuites) > 0 {
		for _, suite := range suites.CheckSuites {
			// ignore queued suites
			if suite.Status == "queued" {
				continue
			}
			if suite.Status != "completed" {
				allCompleted = false
				break
			}
			// check conclusion
			if suite.Conclusion != nil &&
				slices.Contains([]string{"neutral", "skipped", "success"}, *suite.Conclusion) {
				continue
			}
			allSuccess = false
		}
	}

	if len(status.Statuses) > 0 {
		for _, s := range status.Statuses {
			if s.State != "success" {
				allSuccess = false
			}
			if s.State == "pending" {
				allCompleted = false
			}
		}
	}

	if allCompleted {
		if allSuccess {
			return "success"
		} else {
			return "failure"
		}
	}
	return "pending"
}
