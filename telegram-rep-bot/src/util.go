package main

import (
	"errors"
	"regexp"
	"strings"
)

// Attempts to match the target string with any of the regex expressions in the conditions slice
//
// Will complain if it cannot compile the regex
func regexMatchArray(conditions *[]string, target *string) (bool, error) {
	for _, condition := range *conditions {
		// Attempt to match
		result, err := regexp.MatchString(condition, *target)

		// Issues
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

// Check if slice contains item
func arrayContains(item string, array []string) bool {
	for _, val := range array {
		if strings.Compare(val, item) == 0 {
			return true
		}
	}

	return false
}
