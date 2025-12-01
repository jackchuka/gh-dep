package cmd

import (
	"slices"
	"testing"

	"github.com/jackchuka/gh-dep/internal/config"
	"github.com/spf13/cobra"
)

func TestResolveScopeDefaultsToAuthenticatedUser(t *testing.T) {
	c := newTestCommand()
	cfg := &config.Config{}

	owner, repos := resolveScope(c, "", "", cfg)

	if owner != "@me" {
		t.Fatalf("expected owner to default to @me, got %q", owner)
	}
	if len(repos) != 0 {
		t.Fatalf("expected no repos, got %v", repos)
	}
}

func TestResolveScopePrefersFlaggedRepos(t *testing.T) {
	c := newTestCommand()
	if err := c.Flags().Set("repo", "cli/cli, tailwindlabs/tailwindcss"); err != nil {
		t.Fatalf("failed to set repo flag: %v", err)
	}

	owner, repos := resolveScope(c, "cli/cli, tailwindlabs/tailwindcss", "", &config.Config{})

	if owner != "" {
		t.Fatalf("expected owner to remain empty, got %q", owner)
	}

	expected := []string{"cli/cli", "tailwindlabs/tailwindcss"}
	if !slices.Equal(repos, expected) {
		t.Fatalf("expected repos %v, got %v", expected, repos)
	}
}

func TestResolveScopeUsesConfigRepos(t *testing.T) {
	c := newTestCommand()
	cfg := &config.Config{
		Repos: []string{"jackchuka/gh-dep"},
	}

	owner, repos := resolveScope(c, "", "", cfg)

	if owner != "" {
		t.Fatalf("expected owner to remain empty when repos are configured, got %q", owner)
	}
	if !slices.Equal(repos, cfg.Repos) {
		t.Fatalf("expected repos %v, got %v", cfg.Repos, repos)
	}
}

func TestResolveScopeKeepsExplicitOwner(t *testing.T) {
	c := newTestCommand()
	owner, repos := resolveScope(c, "", "myorg", &config.Config{})

	if owner != "myorg" {
		t.Fatalf("expected owner to be %q, got %q", "myorg", owner)
	}
	if len(repos) != 0 {
		t.Fatalf("expected no repos, got %v", repos)
	}
}

func newTestCommand() *cobra.Command {
	c := &cobra.Command{}
	c.Flags().String("repo", "", "")
	return c
}
