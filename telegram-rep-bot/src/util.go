package main

import (
	"errors"
	"regexp"
)

// Attempts to match the target string with any of the regex expressions in the conditions slice
//
// Will complain if it cannot compile the regex
func regexMatchArray(conditions *[]string, target *string) (bool, error) {
	for _, condition := range *conditions {
		// Attempt to match
		result, err := regexp.MatchString(condition, *target)

		// Issues?
		if err != nil {
			return false, errors.New("Regex compile failed for " + condition)
		}

		// Found?
		if result {
			return true, nil
		}

	}

	return false, nil
}
