package fuzzymatchercore

import (
	"regexp"
	"strings"
)

// Normalizes an entry by converting it to lowercase and removing non-alphanumeric characters
func (fmc *FuzzyMatcherCore[T]) NormalizeField(entry string) string {
	normalizeRegex := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	lower := strings.ToLower(entry)
	normalized := normalizeRegex.ReplaceAllString(lower, "")

	return normalized
}
