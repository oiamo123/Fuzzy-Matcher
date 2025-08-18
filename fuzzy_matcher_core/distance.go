package fuzzymatchercore

import (
	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"

	"github.com/antzucaro/matchr"
)

func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// Calculate the distance between 2 strings based on the specified method
// Returns a similarity score between 0 and 1 where 1 is a 100% match
func (fmc *FuzzyMatcherCore[T]) CalculateSimilarity(s1, s2 string, distanceMethod ft.CalculationMethod, minSimilarity float64) float64 {
	switch distanceMethod {
	case ft.JaroWinkler:
		sim := matchr.JaroWinkler(s1, s2, false)

		if sim >= minSimilarity {
			return sim
		}

		return 0

	case ft.Levenshtein:
		maxLen := maxInt(len(s1), len(s2))
		if maxLen == 0 {
			return 1
		}

		dist := matchr.Levenshtein(s1, s2)
		sim := 1.0 - float64(dist)/float64(maxLen)

		if sim >= minSimilarity {
			return sim
		}
		return 0

	default:
		return 1
	}
}
