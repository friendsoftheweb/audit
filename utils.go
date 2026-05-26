package main

import (
	"log"
	"os"

	"charm.land/huh/v2"
)

func confirm(message string, initial bool) bool {
	var confirm = initial

	err := huh.NewConfirm().
		Title(message).
		Affirmative("Yes").
		Negative("No").
		Value(&confirm).Run()

	if err != nil {
		log.Fatal(err)
	}

	return confirm
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)

	return !os.IsNotExist(err)
}

func unique(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	result := make([]string, 0, len(input))

	for _, s := range input {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}

	return result
}
