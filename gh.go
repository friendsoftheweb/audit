package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type GhPullRequest struct {
	Url   string `json:"url"`
	State string `json:"state"`
}

func ghInstalled() bool {
	_, err := exec.LookPath("gh")

	return err == nil
}

// ghOpenPullRequestUrl returns the URL of the open pull request for a branch,
// or an empty string if there isn't one.
func ghOpenPullRequestUrl(branchName string) string {
	out, err := exec.Command("gh", "pr", "view", branchName, "--json", "url,state").Output()

	if err != nil {
		return ""
	}

	var pr GhPullRequest

	if err := json.Unmarshal(out, &pr); err != nil || pr.State != "OPEN" {
		return ""
	}

	return pr.Url
}

// ghCreatePullRequest opens a pull request for a branch against the repo's
// default branch, using the branch's commits for the title and body. It returns
// the new pull request's URL.
func ghCreatePullRequest(branchName string) (string, error) {
	out, err := command("Creating pull request...", []string{"gh", "pr", "create", "--head", branchName, "--fill"})

	if err != nil {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return "", fmt.Errorf("%s", strings.TrimSpace(string(exitErr.Stderr)))
		}

		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
