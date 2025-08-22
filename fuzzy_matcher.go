package fuzzymatcher

import (
	fmcore "github.com/oiamo123/fuzzy_matcher/fuzzy_matcher_core"
	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

// FuzzyMatcher is a generic structure that holds the fuzzy matcher core and other configurations for fuzzy matching
type FuzzyMatcher[T ft.FuzzyMatcherDataSource] struct {
	FuzzyMatcherCore fmcore.FuzzyMatcherCore[T]
}

// Initializes the fuzzy matcher
// If params == ft.FuzzyMatcherCoreParameters[T]{}, it will default to:
// CorrectOcrMisreads: false, UseExpiration: false, maxEdits: 0
func (fuzzyMatcher *FuzzyMatcher[T]) Init(params ft.FuzzyMatcherCoreParameters[T]) {
	fuzzyMatcher.FuzzyMatcherCore.CoreParams = params
}

// Inserts an array of entries into the fuzzy matcher
func (fuzzyMatcher *FuzzyMatcher[T]) InsertEntries(entries []T) {
	if len(entries) == 0 {
		return
	}

	fuzzyMatcher.FuzzyMatcherCore.Build(entries)
}

func (fuzzyMatcher *FuzzyMatcher[T]) Search(entry T) (bool, []ft.FuzzyMatch[T]) {
	fuzzyMatcher.FuzzyMatcherCore.Clean()
	return fuzzyMatcher.FuzzyMatcherCore.SearchFuzzy(entry)
}

func (fuzzyMatcher *FuzzyMatcher[T]) RemoveEntries(entries []T) {
	fuzzyMatcher.FuzzyMatcherCore.RemoveEntries(entries)
}
