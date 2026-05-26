package main

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"

	"golang.org/x/mod/semver"
)

var versionExp = regexp.MustCompile(`ruby (\d+\.\d+\.\d+)`)

func checkRubyVersion(minVersion string) error {
	if !semver.IsValid(minVersion) {
		return errors.New("Invalid version: " + minVersion)
	}

	out, err := exec.Command("ruby", "--version").Output()

	if err != nil {
		return err
	}

	matches := versionExp.FindStringSubmatch(string(out))

	if len(matches) < 2 {
		return errors.New("Could not parse Ruby version from output: " + string(out))
	}

	version := "v" + matches[1]

	if !semver.IsValid(version) {
		return errors.New("Invalid version: " + version)
	}

	if semver.Compare(version, minVersion) < 0 {
		return errors.New("Ruby version (" + version + ") is less than expected version (" + minVersion + ")")
	}

	fmt.Printf("Ruby version is acceptable: %s\n", version)

	return nil
}
