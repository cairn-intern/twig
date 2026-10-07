package client

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFakeHelper(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatalf("writing fake helper: %v", err)
	}
}

// TestGitURLSchemeMatchesInstalledHelper verifies generated URLs use the
// transport of the helper actually on PATH.
func TestGitURLSchemeMatchesInstalledHelper(t *testing.T) {
	dir := t.TempDir()
	writeFakeHelper(t, dir, "git-remote-gitlawb")
	t.Setenv("PATH", dir)

	if got := GitURLScheme(); got != "gitlawb" {
		t.Fatalf("GitURLScheme() = %q, want gitlawb", got)
	}
	if got := FormatGitURL("owner", "repo"); got != "gitlawb://owner/repo" {
		t.Fatalf("FormatGitURL() = %q, want gitlawb://owner/repo", got)
	}
}

// TestGitURLSchemePrefersTwigpine verifies the new helper wins when both are
// installed.
func TestGitURLSchemePrefersTwigpine(t *testing.T) {
	dir := t.TempDir()
	writeFakeHelper(t, dir, "git-remote-gitlawb")
	writeFakeHelper(t, dir, "git-remote-twigpine")
	t.Setenv("PATH", dir)

	if got := GitURLScheme(); got != "twigpine" {
		t.Fatalf("GitURLScheme() = %q, want twigpine", got)
	}
	if got := FormatGitURL("owner", "repo"); got != "twigpine://owner/repo" {
		t.Fatalf("FormatGitURL() = %q, want twigpine://owner/repo", got)
	}
}
