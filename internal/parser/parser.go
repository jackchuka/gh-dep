package parser

import (
	"regexp"
)

// PackageUpdate represents a parsed dependency update
type PackageUpdate struct {
	Package   string
	ToVersion string
}

// Patterns for matching PR titles
var patterns = []*regexp.Regexp{
	// Pattern 1: "bump/update PACKAGE from X to VERSION"
	// Matches most Dependabot/Renovate formats with "from...to"
	// Examples:
	// - "Bump lodash from 4.17.15 to 4.17.21"
	// - "Update dependency typescript from 4.5.2 to 5.6.0"
	// - "Bump package-name from 1.2.3 to 2.0.0"
	regexp.MustCompile(`(?i)(?:bump|update)[:\s]+([^\s]+)\s+from\s+[^\s]+\s+to\s+v?(\d+\.\d+(?:\.\d+)?)`),

	// Pattern 2: "[Uu]pdate PACKAGE to VERSION"
	// Matches: "Update typescript to 5.6.0", "update dependency eslint to 8.57.0"
	regexp.MustCompile(`(?i)update\s+(?:dependency\s+)?([^\s]+)\s+to\s+v?(\d+\.\d+(?:\.\d+)?)`),

	// Pattern 3: Catch-all semver pattern
	// Extracts package name and version from any title with "X to Y" format
	// This is very permissive - just finds the last word before "to" and a semver after
	regexp.MustCompile(`(?i)([^\s:]+)\s+to\s+v?(\d+\.\d+(?:\.\d+)?)`),
}

// ParseTitle attempts to extract package and version from a PR title
// Custom patterns (if provided) are tried first, then default patterns
// Returns (package, version) or ("unknown", "unknown") if parsing fails
func ParseTitle(title string, customPatterns []string) PackageUpdate {
	for _, patternStr := range customPatterns {
		if re, err := regexp.Compile(patternStr); err == nil {
			matches := re.FindStringSubmatch(title)
			if len(matches) == 3 {
				return PackageUpdate{
					Package:   matches[1],
					ToVersion: matches[2],
				}
			}
		}
	}

	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(title)
		if len(matches) == 3 {
			return PackageUpdate{
				Package:   matches[1],
				ToVersion: matches[2],
			}
		}
	}

	return PackageUpdate{
		Package:   "unknown",
		ToVersion: "unknown",
	}
}

// GroupKey returns the group key for this update
func (u PackageUpdate) GroupKey() string {
	return u.Package + "@" + u.ToVersion
}
