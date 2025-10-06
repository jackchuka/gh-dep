package types

// PR represents a pull request
type PR struct {
	Number   int    `json:"number"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Repo     string `json:"repo"` // OWNER/REPO format
	URL      string `json:"url"`
	HeadSHA  string `json:"-"`         // For CI status checks
	CIStatus string `json:"ci_status"` // CI status: success, pending, failure, or empty
}

// Group represents a collection of PRs for the same package@version
type Group struct {
	Key string // package@version
	PRs []PR
}

// Cache represents the cached groups from list --group
type Cache struct {
	Groups map[string][]PR `json:"groups"` // key: package@version, value: list of PRs
}
