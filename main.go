package main

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/fatih/color"
)

var MIN_NODE_VERSION = "v24.0.0"
var MIN_RUBY_VERSION = "v3.3.0"

func main() {
	var upgradeAll = false

	args := os.Args

	if len(args) > 1 {
		for _, arg := range args[1:] {
			if arg == "--upgrade-all" || arg == "-a" {
				upgradeAll = true
			} else {
				fmt.Printf(color.RedString("Unknown argument: %s\n"), arg)

				os.Exit(1)
			}
		}
	}

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
		fmt.Printf("\nBranch %s already exists. Checking it out...\n", branchName)

		err := exec.Command("git", "checkout", branchName).Run()

		if err != nil {
			fmt.Print(color.RedString(err.Error()))

			os.Exit(1)
		}
	} else {
		fmt.Printf("\nCreating and checking out branch %s...\n", branchName)

		gitCreateBranch(branchName)
	}

	var nodeResults []UpgradeResult
	var rubyResults []UpgradeResult

	if nodeProject {
		nodeResults, _ = auditNodePackages(upgradeAll)
	}

	if rubyProject {
		rubyResults, _ = auditRubyGems(upgradeAll)
	}

	if len(nodeResults) > 0 {
		fmt.Println("\nRemaining Node CVEs:")

		printResultsTable(nodeResults)
	}

	if len(rubyResults) > 0 {
		fmt.Println("\nRemaining Ruby CVEs:")

		printResultsTable(rubyResults)
	}

	modifiedFiles = gitStatus()

	nodeLockFileModified := slices.Contains(modifiedFiles, "yarn.lock")
	nodePackageFileModified := slices.Contains(modifiedFiles, "package.json")

	rubyLockFileModified := slices.Contains(modifiedFiles, "Gemfile.lock")

	if !(nodeLockFileModified || rubyLockFileModified) {
		fmt.Println("\nNo changes to commit")
	} else {
		fmt.Println()

		if confirm("Commit and push changes?", true) {
			if nodeLockFileModified {
				gitAdd([]string{"yarn.lock"})
			}

			if nodePackageFileModified {
				gitAdd([]string{"package.json"})
			}

			if rubyLockFileModified {
				gitAdd([]string{"Gemfile.lock"})
			}

			gitCommit("Upgrade packages with CVEs")

			gitPush(branchName)
		}
	}
}

func printResultsTable(results []UpgradeResult) {
	if len(results) == 0 {
		return
	}

	showPatchedVersions := len(results[0].PatchedVersions) > 0

	var headers []string

	if showPatchedVersions {
		headers = []string{"Package", "Patched", "Installed", "URL"}
	} else {
		headers = []string{"Package", "Vulnerable", "Installed", "URL"}
	}

	rows := make([][]string, len(results))

	for i, result := range results {
		if showPatchedVersions {
			rows[i] = []string{result.PackageName, strings.Join(result.PatchedVersions, ", "), strings.Join(result.Versions, ", "), result.Url}
		} else {
			rows[i] = []string{result.PackageName, strings.Join(result.VulnerableVersions, ", "), strings.Join(result.Versions, ", "), result.Url}
		}
	}

	purple := lipgloss.Color("99")
	gray := lipgloss.Color("245")

	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := cellStyle.Foreground(purple).Bold(true)

	t := table.New().
		Headers(headers...).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == -1 {
				return headerStyle
			}

			return cellStyle.Foreground(gray)
		})

	fmt.Println(t)
}
