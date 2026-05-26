package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
)

type BundlerResult struct {
	Results []BundlerIssue `json:"results"`
}

type BundlerIssue struct {
	Gem      BundlerIssueGem      `json:"gem"`
	Advisory BundlerIssueAdvisory `json:"advisory"`
}

type BundlerIssueGem struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type BundlerIssueAdvisory struct {
	Title string `json:"title"`
}

func bundlerAuditCheck() ([]BundlerIssue, error) {
	var result = BundlerResult{}

	out, err := command("Auditing gems...", []string{"bundle", "exec", "bundle-audit", "check", "--update", "--format=json"})

	if err != nil {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) {
			// We're expecting an exit code of 1 if there are vulnerabilities, so we
			// can ignore that code and just return the results
			if exitErr.ExitCode() != 1 {
				return result.Results, err
			}
		} else {
			return result.Results, err
		}
	}

	// The output from bundle-audit may contain some text before the JSON, so we
	// need to find the first '{' character and trim everything before it
	if i := bytes.IndexByte(out, '{'); i != -1 {
		out = out[i:]
	}

	err = json.Unmarshal(out, &result)

	return result.Results, nil
}
