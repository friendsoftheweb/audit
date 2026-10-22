package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func gitStatus() []string {
	out, err := exec.Command("git", "status", "--porcelain").Output()

	if err != nil {
		log.Fatal(err)
	}

	var modifiedFiles []string

	for _, line := range strings.Split(string(out), "\n") {
		if line != "" {
			modifiedFiles = append(modifiedFiles, strings.TrimSpace(line[3:]))
		}
	}

	return modifiedFiles
}

func gitCurrentBranch() (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--quiet", "--short", "HEAD")
	out, err := cmd.CombinedOutput()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			return "", nil // Detached HEAD state
		}

		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

// gitUpstreamBranch returns the upstream of the current branch (e.g.
// "origin/main"), or an empty string if the branch has no upstream.
func gitUpstreamBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}").Output()

	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

func gitFetch() error {
	_, err := command("Fetching from remote...", []string{"git", "fetch", "--quiet"})

	return err
}

// gitPull fast-forwards the current branch to its upstream. It never creates a
// merge commit, so it fails if the branch has diverged from the upstream.
func gitPull() error {
	_, err := command("Pulling from remote...", []string{"git", "pull", "--ff-only", "--quiet"})

	return err
}

// gitCommitsBehind returns the number of commits on upstream that aren't in HEAD.
func gitCommitsBehind(upstream string) (int, error) {
	out, err := exec.Command("git", "rev-list", "--count", "HEAD.."+upstream).Output()

	if err != nil {
		return 0, err
	}

	return strconv.Atoi(strings.TrimSpace(string(out)))
}

func gitBranchExists(name string) bool {
	out, err := exec.Command("git", "branch", "--list", name).Output()

	if err != nil {
		log.Fatal(err)
	}

	return strings.TrimSpace(string(out)) != ""
}

func gitCreateBranch(name string) {
	err := exec.Command("git", "checkout", "-b", name).Run()

	if err != nil {
		log.Fatal(err)
	}
}

func gitAdd(files []string) {
	err := exec.Command("git", "add", strings.Join(files, " ")).Run()

	if err != nil {
		log.Fatal(err)
	}
}

func gitAddAll() {
	err := exec.Command("git", "add", ".").Run()

	if err != nil {
		log.Fatal(err)
	}
}

func gitCommit(message string) {
	err := exec.Command("git", "commit", "-m", message).Run()

	if err != nil {
		log.Fatal(err)
	}
}

func gitPush(branchName string) {
	err := exec.Command("git", "push", "--set-upstream", "origin", branchName).Run()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Changes pushed to origin/%s", branchName)
}
