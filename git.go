package main

import (
	"fmt"
	"log"
	"os/exec"
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
