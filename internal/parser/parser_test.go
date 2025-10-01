package parser

import (
	"testing"
)

func TestParseTitle(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantPkg string
		wantVer string
	}{
		// Dependabot patterns
		{
			name:    "Dependabot bump pattern",
			title:   "Bump lodash from 4.17.20 to 4.17.21",
			wantPkg: "lodash",
			wantVer: "4.17.21",
		},
		{
			name:    "Dependabot bump pattern with v prefix",
			title:   "Bump axios from 1.6.0 to v1.7.3",
			wantPkg: "axios",
			wantVer: "1.7.3",
		},
		{
			name:    "Dependabot update pattern",
			title:   "Update typescript to 5.6.0",
			wantPkg: "typescript",
			wantVer: "5.6.0",
		},
		{
			name:    "Dependabot update pattern with v prefix",
			title:   "Update commander to v12.1.0",
			wantPkg: "commander",
			wantVer: "12.1.0",
		},

		// Renovate patterns
		{
			name:    "Renovate Update dependency pattern",
			title:   "Update dependency eslint to 8.57.0",
			wantPkg: "eslint",
			wantVer: "8.57.0",
		},
		{
			name:    "Renovate Update dependency pattern with v prefix",
			title:   "Update dependency prettier to v3.0.0",
			wantPkg: "prettier",
			wantVer: "3.0.0",
		},
		{
			name:    "Renovate chore(deps) update pattern",
			title:   "chore(deps): update vitest to 1.2.0",
			wantPkg: "vitest",
			wantVer: "1.2.0",
		},
		{
			name:    "Renovate chore(deps) update pattern with v prefix",
			title:   "chore(deps): update jest to v29.7.0",
			wantPkg: "jest",
			wantVer: "29.7.0",
		},
		{
			name:    "chore(deps) bump pattern",
			title:   "chore(deps): bump axios from 1.11.0 to 1.12.0",
			wantPkg: "axios",
			wantVer: "1.12.0",
		},

		// Generic pattern
		{
			name:    "Generic colon pattern",
			title:   "deps: upgrade webpack to 5.89.0",
			wantPkg: "webpack",
			wantVer: "5.89.0",
		},

		// Unknown/parse failures
		{
			name:    "Unparseable title",
			title:   "Some random PR title",
			wantPkg: "unknown",
			wantVer: "unknown",
		},
		{
			name:    "Empty title",
			title:   "",
			wantPkg: "unknown",
			wantVer: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTitle(tt.title, nil)
			if got.Package != tt.wantPkg {
				t.Errorf("ParseTitle(%q).Package = %q, want %q", tt.title, got.Package, tt.wantPkg)
			}
			if got.ToVersion != tt.wantVer {
				t.Errorf("ParseTitle(%q).ToVersion = %q, want %q", tt.title, got.ToVersion, tt.wantVer)
			}
		})
	}
}

func TestGroupKey(t *testing.T) {
	tests := []struct {
		name   string
		update PackageUpdate
		want   string
	}{
		{
			name: "Normal package",
			update: PackageUpdate{
				Package:   "lodash",
				ToVersion: "4.17.21",
			},
			want: "lodash@4.17.21",
		},
		{
			name: "Scoped package",
			update: PackageUpdate{
				Package:   "@types/node",
				ToVersion: "20.10.0",
			},
			want: "@types/node@20.10.0",
		},
		{
			name: "Unknown",
			update: PackageUpdate{
				Package:   "unknown",
				ToVersion: "unknown",
			},
			want: "unknown@unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.update.GroupKey()
			if got != tt.want {
				t.Errorf("GroupKey() = %q, want %q", got, tt.want)
			}
		})
	}
}
