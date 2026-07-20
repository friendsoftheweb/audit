package main

import (
	"fmt"
	"log"
	"strconv"

	"charm.land/huh/v2"
)

type UpgradeResult struct {
	Id                 string
	PackageName        string
	Versions           []string
	VulnerableVersions []string
	PatchedVersions    []string
	Url                string
}

func auditNodePackages(upgradeAll bool) ([]UpgradeResult, error) {
	var results = []UpgradeResult{}

	issuesBefore, err := yarnAudit()

	if err != nil {
		log.Fatal(err)
	}

	if len(issuesBefore) == 0 {
		fmt.Println("No packages with CVEs found")

		return results, nil
	}

	var selectedPackages []string

	if upgradeAll {
		for _, issue := range issuesBefore {
			selectedPackages = append(selectedPackages, issue.PackageName)
		}
	} else {
		var options = make([]huh.Option[string], 0)

		for _, issue := range issuesBefore {
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

		err = huh.NewMultiSelect[string]().
			Options(
				options...,
			).
			Title("Select packages to upgrade").
			Value(&selectedPackages).Run()

		if err != nil {
			log.Fatal(err)
		}
	}

	selectedPackages = unique(selectedPackages)

	if len(selectedPackages) == 0 {
		fmt.Println("No packages selected. Aborting...")

		return results, nil
	}

	for _, packageName := range selectedPackages {
		err := yarnUpgrade(packageName)

		if err != nil {
			log.Printf("Error upgrading package %s: %s\n", packageName, err)
		}
	}

	issuesAfter, err := yarnAudit()

	for _, issue := range issuesAfter {
		results = append(results, UpgradeResult{
			Id:                 strconv.Itoa(issue.Data.Id),
			PackageName:        issue.PackageName,
			VulnerableVersions: []string{issue.Data.VulnerableVersions},
			Versions:           issue.Data.TreeVersions,
			Url:                issue.Data.Url,
		})
	}

	return results, nil
}

func auditRubyGems(upgradeAll bool) ([]UpgradeResult, error) {
	var results = []UpgradeResult{}

	issuesBefore, err := bundlerAuditCheck()

	if err != nil {
		return results, err
	}

	var gemNames []string

	for _, issue := range issuesBefore {
		gemNames = append(gemNames, issue.Gem.Name)
	}

	gemNames = unique(gemNames)

	if len(gemNames) > 0 {
		var selectedGems []string

		if upgradeAll {
			selectedGems = gemNames
		} else {
			var options = make([]huh.Option[string], 0)

			for _, gemName := range gemNames {
				options = append(options, huh.NewOption(gemName, gemName).Selected(true))
			}

			err = huh.NewMultiSelect[string]().
				Options(
					options...,
				).
				Title("Select gems to upgrade").
				Value(&selectedGems).Run()

			if err != nil {
				return results, err
			}
		}

		command("Upgrading gems...", append([]string{"bundle", "update"}, selectedGems...))
	}

	issuesAfter, err := bundlerAuditCheck()

	if err != nil {
		return results, err
	}

	for _, issue := range issuesAfter {
		results = append(results, UpgradeResult{
			PackageName:     issue.Gem.Name,
			Versions:        []string{issue.Gem.Version},
			PatchedVersions: issue.Advisory.PatchedVersions,
			Url:             issue.Advisory.Url,
		})
	}

	return results, nil
}
