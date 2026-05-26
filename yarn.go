package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"
)

type YarnIssue struct {
	PackageName string        `json:"value"`
	Data        YarnIssueData `json:"children"`
}

type YarnIssueData struct {
	Severity           string   `json:"severity"`
	Issue              string   `json:"Issue"`
	VulnerableVersions string   `json:"Vulnerable Versions"`
	TreeVersions       []string `json:"Tree Versions"`
}

type YarnInfo struct {
	Data YarnInfoData `json:"children"`
}

type YarnInfoData struct {
	Version string `json:"Version"`
}

func yarnAudit() ([]YarnIssue, error) {
	var issues = []YarnIssue{}

	out, err := command("Auditing packages...", []string{"yarn", "npm", "audit", "--recursive", "--no-deprecations", "--json"})

	if err != nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.TrimSpace(line) != "" {
				issue := YarnIssue{}
				err := json.Unmarshal([]byte(line), &issue)

				if err != nil {
					return nil, err
				}

				issues = append(issues, issue)

				// fmt.Println("----------------------------------------")
				// fmt.Printf("Package: %s\n", issue.PackageName)
				// fmt.Printf("Severity: %s\n", issue.Data.Severity)
				// fmt.Printf("Issue: %s\n", issue.Data.Issue)
				// fmt.Printf("Vulnerable Versions: %s\n", issue.Data.VulnerableVersions)
				// fmt.Printf("Tree Versions: %s\n", strings.Join(issue.Data.TreeVersions, ", "))
				// fmt.Println("----------------------------------------")
				// fmt.Println()
			}
		}
	}

	return issues, nil
}

func yarnInfo(packageName string) ([]YarnInfo, error) {
	out, err := command("Getting package info...", []string{"yarn", "info", "--recursive", "--json", packageName})

	if err != nil {
		fmt.Printf("Error getting info for package %s: %s\n", packageName, string(out))

		return nil, err
	}

	infos := []YarnInfo{}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		var info YarnInfo
		err := json.Unmarshal([]byte(line), &info)
		if err != nil {
			log.Fatal("Failed to parse JSON output from `yarn info`:", err)
		}

		infos = append(infos, info)
	}

	return infos, nil
}

func yarnInfoVersions(packageName string) ([]string, error) {
	var versions []string

	infos, err := yarnInfo(packageName)

	if err != nil {
		return versions, err
	}

	for _, info := range infos {
		if info.Data.Version != "" {
			versions = append(versions, info.Data.Version)
		}
	}

	if len(versions) > 0 {
		return versions, nil
	}

	return versions, fmt.Errorf("no version found for package %s", packageName)
}

func yarnUpgrade(packageName string) error {
	previousVersions, err := yarnInfoVersions(packageName)

	if err != nil {
		return err
	}

	out, err := command(fmt.Sprintf("Upgrading package %s...", packageName), []string{"yarn", "up", "--recursive", packageName})

	if err != nil {
		fmt.Printf("Error upgrading package %s: %s\n", packageName, string(out))

		return err
	}

	versions, err := yarnInfoVersions(packageName)

	if err != nil {
		return err
	}

	var upgraded = false

	for index, previousVersion := range previousVersions {
		if versions[index] != previousVersion {
			upgraded = true

			break
		}
	}

	if upgraded {
		fmt.Printf(color.GreenString("Package %s upgraded successfully\n"), packageName)
	} else {
		fmt.Printf(color.YellowString("Package %s could not be upgraded\n"), packageName)
	}

	return nil
}
