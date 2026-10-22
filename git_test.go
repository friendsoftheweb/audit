package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitCurrentBranchReturnsCurrentBranch(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "user.email", "test@example.com")

	readmePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmePath, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}

	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "initial")
	runGit(t, dir, "checkout", "-b", "feature")

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to repo: %v", err)
	}

	got, err := gitCurrentBranch()
	if err != nil {
		t.Fatalf("gitCurrentBranch() returned error: %v", err)
	}
	if got != "feature" {
		t.Fatalf("gitCurrentBranch() = %q, want %q", got, "feature")
	}
}

func TestGitCurrentBranchHandlesDetachedHead(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "user.email", "test@example.com")

	readmePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmePath, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}

	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "initial")

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to repo: %v", err)
	}

	runGit(t, dir, "checkout", "--detach", "HEAD")

	got, err := gitCurrentBranch()
	if err != nil {
		t.Fatalf("gitCurrentBranch() returned error: %v", err)
	}
	if got != "" {
		t.Fatalf("gitCurrentBranch() = %q, want empty string for detached HEAD", got)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v\n%s", args, err, out)
	}
}

func TestGitCommitsBehindCountsUnpulledCommits(t *testing.T) {
	local := setupRepoBehindRemote(t)

	upstream := gitUpstreamBranch()
	if upstream != "origin/main" {
		t.Fatalf("gitUpstreamBranch() = %q, want %q", upstream, "origin/main")
	}

	runGit(t, local, "fetch")

	got, err := gitCommitsBehind(upstream)
	if err != nil {
		t.Fatalf("gitCommitsBehind() returned error: %v", err)
	}
	if got != 2 {
		t.Fatalf("gitCommitsBehind() = %d, want 2", got)
	}
}

func TestGitPullFastForwardsToUpstream(t *testing.T) {
	setupRepoBehindRemote(t)

	if err := gitPull(); err != nil {
		t.Fatalf("gitPull() returned error: %v", err)
	}

	got, err := gitCommitsBehind("origin/main")
	if err != nil {
		t.Fatalf("gitCommitsBehind() returned error: %v", err)
	}
	if got != 0 {
		t.Fatalf("gitCommitsBehind() = %d after pull, want 0", got)
	}
}

// setupRepoBehindRemote creates a clone whose main branch is two commits behind
// its remote, changes into it for the duration of the test, and returns its path.
func setupRepoBehindRemote(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	local := filepath.Join(root, "local")
	other := filepath.Join(root, "other")

	runGit(t, root, "init", "--bare", "--initial-branch=main", remote)
	runGit(t, root, "clone", remote, local)
	runGit(t, local, "config", "user.name", "Test User")
	runGit(t, local, "config", "user.email", "test@example.com")
	runGit(t, local, "checkout", "-b", "main")
	runGit(t, local, "commit", "--allow-empty", "-m", "initial")
	runGit(t, local, "push", "--set-upstream", "origin", "main")

	runGit(t, root, "clone", remote, other)
	runGit(t, other, "config", "user.name", "Test User")
	runGit(t, other, "config", "user.email", "test@example.com")
	runGit(t, other, "commit", "--allow-empty", "-m", "second")
	runGit(t, other, "commit", "--allow-empty", "-m", "third")
	runGit(t, other, "push")

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	if err := os.Chdir(local); err != nil {
		t.Fatalf("chdir to repo: %v", err)
	}

	return local
}
