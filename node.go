package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/mod/semver"
)

func checkNodeVersion(minVersion string) error {
	if !semver.IsValid(minVersion) {
		return errors.New("Invalid version: " + minVersion)
	}

	out, err := exec.Command("node", "--version").Output()

	if err != nil {
		return err
	}

	version := strings.TrimSpace(string(out))

	if !semver.IsValid(version) {
		return errors.New("Invalid version: " + version)
	}

	if semver.Compare(version, minVersion) < 0 {
		return errors.New("Node version (" + version + ") is less than expected version (" + minVersion + ")")
	}

	fmt.Printf("Node version is acceptable: %s\n", version)

	return nil
}
