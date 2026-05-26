package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"time"

	"charm.land/huh/v2"
	"github.com/fatih/color"
)

var MIN_NODE_VERSION = "v24.0.0"
var MIN_RUBY_VERSION = "v3.3.0"

func main() {
	modifiedFiles := gitStatus()

	// Make sure the user is aware of any modified files before proceeding
	if len(modifiedFiles) > 0 {
		fmt.Println("You have modified files:")

		for _, file := range modifiedFiles {
			fmt.Printf("- %s\n", file)
		}

		fmt.Println()

		if !confirm("Are you sure you want to continue?", false) {
			fmt.Println("Aborting...")

			return
		}
	}

	nodeProject := fileExists("package.json")
	rubyProject := fileExists("Gemfile")

	if nodeProject {
		err := checkNodeVersion(MIN_NODE_VERSION)

		if err != nil {
			fmt.Print(color.RedString(err.Error()))

			os.Exit(1)
		}
	}

	if rubyProject {
		err := checkRubyVersion(MIN_RUBY_VERSION)

		if err != nil {
			fmt.Print(color.RedString(err.Error()))

			os.Exit(1)
		}
	}

	date := time.Now().Format("2006-01-02")
	branchName := fmt.Sprintf("audit-%s", date)

	if gitBranchExists(branchName) {
		fmt.Printf("\nBranch %s already exists. Checking it out...\n\n", branchName)

		err := exec.Command("git", "checkout", branchName).Run()

		if err != nil {
			fmt.Print(color.RedString(err.Error()))

			os.Exit(1)
		}
	} else {
		fmt.Printf("Creating and checking out branch %s...\n\n", branchName)

		gitCreateBranch(branchName)
	}

	if nodeProject {
		auditPackages()
	}

	if rubyProject {
		auditGems()
	}

	modifiedFiles = gitStatus()

	yarnLockfileModified := slices.Contains(modifiedFiles, "yarn.lock")
	bundlerLockfileModified := slices.Contains(modifiedFiles, "Gemfile.lock")

	if !(yarnLockfileModified || bundlerLockfileModified) {
		fmt.Println("\nNo changes to commit")

		return
	}

	fmt.Println("")

	if confirm("Commit and push changes?", true) {
		if yarnLockfileModified {
			gitAdd([]string{"yarn.lock"})
		}

		if bundlerLockfileModified {
			gitAdd([]string{"Gemfile.lock"})
		}

		gitCommit("Upgrade packages with CVEs")

		gitPush(branchName)
	}

	fmt.Println("\n\nAll done!")
}

func auditPackages() {
	issues, err := yarnAudit()

	if err != nil {
		log.Fatal(err)
	}

	if len(issues) == 0 {
		fmt.Println("No packages with CVEs found")
		return
	}

	// fmt.Printf("Found %d packages with CVEs", len(issues))

	var options = make([]huh.Option[string], 0)

	for _, issue := range issues {
		var found *YarnIssue

		for _, option := range options {
			if option.Value == issue.PackageName {
				found = &issue
				break
			}
		}

		if found == nil {
			options = append(options, huh.NewOption(issue.PackageName, issue.PackageName).Selected(true))
		}
	}

	var selectedPackages []string

	err = huh.NewMultiSelect[string]().
		Options(
			options...,
		).
		Title("Select packages to upgrade").
		Value(&selectedPackages).Run()

	if err != nil {
		log.Fatal(err)
	}

	if len(selectedPackages) == 0 {
		fmt.Println("No packages selected. Aborting...")

		return
	}

	for _, packageName := range selectedPackages {
		err := yarnUpgrade(packageName)

		if err != nil {
			log.Printf("Error upgrading package %s: %s\n", packageName, err)
		}
	}

	fmt.Println()
}

func auditGems() {
	issues, err := bundlerAuditCheck()

	if err != nil {
		log.Fatal(err)
	}

	// if len(issues) > 0 {
	// 	for _, issue := range issues {
	// 		fmt.Println("----------------------------------------")
	// 		fmt.Printf("Gem: %s\n", issue.Gem.Name)
	// 		fmt.Printf("Version: %s\n", issue.Gem.Version)
	// 		fmt.Printf("Advisory: %s\n", issue.Advisory.Title)
	// 		fmt.Println("----------------------------------------")
	// 		fmt.Println()
	// 	}
	// }

	var gemNames []string

	for _, issue := range issues {
		gemNames = append(gemNames, issue.Gem.Name)
	}

	gemNames = unique(gemNames)

	if len(gemNames) > 0 {
		var options = make([]huh.Option[string], 0)

		for _, gemName := range gemNames {
			options = append(options, huh.NewOption(gemName, gemName).Selected(true))
		}

		var selectedGems []string

		err = huh.NewMultiSelect[string]().
			Options(
				options...,
			).
			Title("Select gems to upgrade").
			Value(&selectedGems).Run()

		if err != nil {
			log.Fatal(err)
		}

		command("Upgrading gems...", append([]string{"bundle", "update"}, selectedGems...))
	}
}
