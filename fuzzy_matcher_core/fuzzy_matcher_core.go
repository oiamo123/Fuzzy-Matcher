package fuzzymatchercore

import (
	"container/heap"
	"fmt"
	"sort"
	"strings"
	"sync"

	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

// Key to identify visited nodes
type VisitKey struct {
	index int
	node  *FuzzyMatcherNode
	edits int
	depth int
}

type FieldResult struct {
	key     ft.Field
	matches []ft.MatchCandidate
	err     error
}

// FuzzyMatcherCore represents the core structure of the fuzzy matcher
type FuzzyMatcherCore[T ft.FuzzyMatcherDataSource] struct {
	Root               *FuzzyMatcherNode
	CoreParams         ft.FuzzyMatcherCoreParameters[T]
	ExpiryHeap         ExpiryHeap
	Entries            map[int]T
}

const (
	MaxDepth          int                  = 5
	MaxEdits          int                  = 2
	MinDistance       float32              = 0.8
	CalculationMethod ft.CalculationMethod = ft.JaroWinkler
)

// FuzzyMatcherNode represents a node in the FuzzyMatcher structure
type FuzzyMatcherNode struct {
	Char          rune
	Children      map[rune]*FuzzyMatcherNode
	IsEndofString bool
	ID            map[int]bool
	Parent        *FuzzyMatcherNode
	Count         int
}

// Inserts a word into the fuzzy matcher
func (fmc *FuzzyMatcherCore[T]) Insert(word string, ID int) *FuzzyMatcherNode {
	node := fmc.Root

	for _, char := range word {
		c := rune(char)

		if node.Children[c] == nil {
			node.Children[c] = &FuzzyMatcherNode{
				Children: make(map[rune]*FuzzyMatcherNode),
				Char:     c,
				Parent:   node,
				Count:    0,
			}
		}

		node = node.Children[c]
		node.Count++
	}

	// Mark the last node with the entry ID
	if node.ID == nil {
		node.ID = make(map[int]bool)
	}

	node.ID[ID] = true

	return node
}

// Builds the fuzzy matcher with a list of fuzzy entries
func (fmc *FuzzyMatcherCore[T]) Build(entries []T) error {
	// Init the expiry heap if it is nil
	if fmc.ExpiryHeap == nil && fmc.CoreParams.UseExpiration {
		heap.Init(&fmc.ExpiryHeap)
	}

	// Init the root node if it is nil
	if fmc.Root == nil {
		fmc.Root = &FuzzyMatcherNode{
			Children: make(map[rune]*FuzzyMatcherNode),
		}
	}

	// Insert each word into the fuzzy matcher
	for _, entry := range entries {
		fuzzyEntry := entry.CreateFuzzyEntry()
		for key, field := range fuzzyEntry.Key {
			// Prefix the string with the field name ie 'firstname:'
			normalized := fmc.NormalizeField(field)
			searchString := string(key) + ":" + normalized

			node := fmc.Insert(searchString, fuzzyEntry.ID)

			node.IsEndofString = true

			// Create an expiry for the entry
			if fmc.CoreParams.UseExpiration {
				if fuzzyEntry.Expiry.IsZero() {
					return fmt.Errorf("UseExpiration set to true. Cannot insert entry with no expiry: %v", entry)
				}

				heap.Push(&fmc.ExpiryHeap, ExpiryEntry{
					Node:   node,
					Expiry: fuzzyEntry.Expiry,
					ID:     fuzzyEntry.ID,
				})
			}
		}

		if fmc.Entries == nil {
			fmc.Entries = make(map[int]T)
		}

		fmc.Entries[fuzzyEntry.ID] = entry
	}

	return nil
}

// Searches the fuzzy matcher for the given entry
func (fmc *FuzzyMatcherCore[T]) SearchFuzzy(entry ft.FuzzyMatcherDataSource) (bool, []ft.FuzzyMatch[T]) {
	if fmc.CoreParams.UseExpiration {
		fmc.Clean()
	}

	fuzzyEntry := entry.CreateFuzzyEntry()
	parameters := entry.GetSearchParameters()

	var wg sync.WaitGroup
	results := make(chan FieldResult, len(fuzzyEntry.Key))

	// Per-field goroutines
	for key, field := range fuzzyEntry.Key {
		wg.Add(1)
		go func(key ft.Field, field string) {
			defer wg.Done()

			normalized := fmc.NormalizeField(field)
			searchString := string(key) + ":" + normalized

			valueStart := len(key) + 1
			editableFields := make([]bool, len(searchString))
			numEdits, numEditsOk := parameters.MaxEdits[key]

			// Initialize editableFields based on the search parameters
			for i := valueStart; i < len(editableFields); i++ {
				if numEditsOk && numEdits > 0 {
					editableFields[i] = true
				} else {
					editableFields[i] = false
				}
			}

			recurseParameters := RecurseParameters{
				word: []rune(searchString),
				index: 0,
				node: fmc.Root,
				path: make([]rune, 0),
				maxDepth: parameters.MaxDepth[key],
				depth: 0,
				depthIncrement: 0,
				numEdits: 0,
				maxEdits: parameters.MaxEdits[key],
				numEditsIncrement: 0,
				editableFields: editableFields,
				visited: make(map[VisitKey]struct{}),
			}

			matches := fmc.Recurse(recurseParameters)

			results <- FieldResult{key: key, matches: matches}
		}(key, field)
	}

	// Close results channel after all workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect all results first (thread-safe)
	allResults := make(map[ft.Field][]ft.MatchCandidate)
	for res := range results {
		if res.err != nil {
			// Handle error if needed
			continue
		}
		allResults[res.key] = res.matches
	}

	// Now merge results sequentially (no race conditions)
	matchedEntries := make(map[int]map[ft.Field]string)
	matchedEntriesCount := make(map[int]map[ft.Field]int)

	for key, matches := range allResults {
		for _, match := range matches {
			for _, id := range match.ID {
				if match.EditCount > parameters.MaxEdits[key] {
					continue
				}

				if matchedEntries[id] == nil {
					matchedEntries[id] = make(map[ft.Field]string)
				}
				matchedEntries[id][key] = strings.Replace(match.Text, string(key)+":", "", 1)

				if matchedEntriesCount[id] == nil {
					matchedEntriesCount[id] = make(map[ft.Field]int)
				}

				if currentCount, exists := matchedEntriesCount[id][key]; !exists || currentCount > match.EditCount {
					matchedEntriesCount[id][key] = match.EditCount
				}
			}
		}
	}

	// Remove all incomplete entries or entries that exceed max edits
	// An entry is incomplete if it has any empty fields
	matchedEntriesCleaned := fmc.CleanMatches(matchedEntries, matchedEntriesCount, fuzzyEntry)

	if len(matchedEntriesCleaned) == 0 {
		return false, nil
	}

	// track valid entries
	finalMatchedEntries := []ft.FuzzyMatch[T]{}

	for id, match := range matchedEntriesCleaned {
		similarities := make(map[ft.Field]float64)
		reject := false

		// iterate through the keys
		for key := range fuzzyEntry.Key {
			matchVal, exists := match[key]
			origVal := fuzzyEntry.Key[key]
			min := parameters.MinDistances[key]

			// Missing required field
			if (!exists || matchVal == "") && min > 0 {
				reject = true
				break
			}

			matchNormalized := fmc.NormalizeField(matchVal)
			originalNormalized := fmc.NormalizeField(origVal)

			similarity := fmc.CalculateSimilarity(originalNormalized, matchNormalized, parameters.CalculationMethods[key], parameters.MinDistances[key])

			// if the min distance is not 0 and the distance == 0
			if min == 0 && similarity == 0 {
				continue
			}

			if min > 0 && similarity < min {
				reject = true
				break
			}

			similarities[key] = similarity
		}

		// skip entry
		if reject {
			continue
		}

		var score float64
		for key, weight := range parameters.Weights {
			if distance, exists := similarities[key]; exists {
				score += weight * distance
			}
		}

		// add to list
		finalMatchedEntries = append(finalMatchedEntries, ft.FuzzyMatch[T]{
			Score: score,
			Entry: fmc.Entries[id],
		})
	}

	if len(finalMatchedEntries) == 0 {
		return false, nil
	}

	// Return top n best matches
	sort.Slice(finalMatchedEntries, func(i, j int) bool {
		return finalMatchedEntries[i].Score > finalMatchedEntries[j].Score
	})

	if len(finalMatchedEntries) > 5 {
		finalMatchedEntries = finalMatchedEntries[:5]
	}

	// return true, matchedEntries
	return true, finalMatchedEntries
}
