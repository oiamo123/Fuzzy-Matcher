package fuzzymatcher

import (
	"fmt"

	fmcore "github.com/oiamo123/fuzzy_matcher/fuzzy_matcher_core"
	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

// FuzzyMatcher is a generic structure that holds the fuzzy matcher core and other configurations for fuzzy matching
type FuzzyMatcher[T ft.FuzzyMatcherDataSource] struct {
	FuzzyMatcherCore   *fmcore.FuzzyMatcherCore[T]
}

func (fuzzyMatcher *FuzzyMatcher[T]) Init(params *ft.FuzzyMatcherCoreParameters[T]) {
	if fuzzyMatcher.FuzzyMatcherCore == nil {
		fuzzyMatcher.FuzzyMatcherCore = &fmcore.FuzzyMatcherCore[T]{
			CoreParams: ft.FuzzyMatcherCoreParameters[T]{
				CorrectOcrMisreads: false,
				UseExpiration:      false,
				MaxEdits:           9999,
			},
		}
	}

	if params != nil {
		fuzzyMatcher.FuzzyMatcherCore.CoreParams = *params
	}
}

// Inserts an array of entries into the fuzzy matcher
func (fuzzyMatcher *FuzzyMatcher[T]) InsertEntries(entries []T) error {
	if fuzzyMatcher.FuzzyMatcherCore == nil {
		return fmt.Errorf("FuzzyMatcherCore is not initialized")
	}

	fuzzyMatcher.FuzzyMatcherCore.Build(entries)

	return nil
}

func (fuzzyMatcher *FuzzyMatcher[T]) Search(entry T) (bool, []ft.FuzzyMatch[T]) {
	// Verify fuzzyMatcherCore is initialized
	if fuzzyMatcher.FuzzyMatcherCore == nil {
		return false, nil
	}

	fuzzyMatcher.FuzzyMatcherCore.Clean()

	// Even if sync fails, we can still search with the existing data
	return fuzzyMatcher.FuzzyMatcherCore.SearchFuzzy(entry)
}
