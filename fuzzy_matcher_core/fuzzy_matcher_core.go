package fuzzymatchercore

import (
	"container/heap"
	"fmt"
	"sort"
	"sync"

	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

// FuzzyMatcherCore represents the core structure of the fuzzy matcher
type FuzzyMatcherCore[T ft.FuzzyMatcherDataSource] struct {
	Root       *ft.FuzzyMatcherNode
	CoreParams ft.FuzzyMatcherCoreParameters[T]
	ExpiryHeap ExpiryHeap
	Entries    map[int]T
}

const (
	MaxDepth          int                  = 5
	MaxEdits          int                  = 2
	MinDistance       float32              = 0.8
	CalculationMethod ft.CalculationMethod = ft.JaroWinkler
)

// Inserts a word into the fuzzy matcher
func (fmc *FuzzyMatcherCore[T]) Insert(word string, ID int) *ft.FuzzyMatcherNode {
	node := fmc.Root

	for _, char := range word {
		c := rune(char)

		if node.Children[c] == nil {
			node.Children[c] = &ft.FuzzyMatcherNode{
				Children: make(map[rune]*ft.FuzzyMatcherNode),
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
		fmc.Root = &ft.FuzzyMatcherNode{
			Children: make(map[rune]*ft.FuzzyMatcherNode),
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

				heap.Push(&fmc.ExpiryHeap, ft.ExpiryEntry{
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

	// Shared maps for results - no race conditions since each goroutine writes to different fields
	matchedEntries := make(map[int]map[ft.Field]string)
	matchedEntriesCount := make(map[int]map[ft.Field]int)
	matchedSimilarities := make(map[int]map[ft.Field]float64)
	var mapMutex sync.Mutex

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

			recurseParameters := &ft.RecurseParameters{
				Word:              []rune(searchString),
				Key:               []rune(key),
				Index:             0,
				Node:              fmc.Root,
				Path:              make([]rune, 0),
				MaxDepth:          parameters.MaxDepth[key],
				Depth:             0,
				DepthIncrement:    0,
				NumEdits:          0,
				MaxEdits:          parameters.MaxEdits[key],
				NumEditsIncrement: 0,
				EditableFields:    editableFields,
				Visited:           make(map[*ft.FuzzyMatcherNode]int),
				CalculationMethod: parameters.CalculationMethods[key],
				MinDistance:       parameters.MinDistances[key],
			}

			matches := fmc.Recurse(recurseParameters)

			// Filter by edit count only in the goroutine, similarity filtering happens later
			maxEdits := parameters.MaxEdits[key]

			mapMutex.Lock()
			for _, match := range matches {
				// Filter by edit count only
				if match.EditCount > maxEdits {
					continue
				}

				for _, id := range match.ID {
					if matchedEntries[id] == nil {
						matchedEntries[id] = make(map[ft.Field]string)
						matchedEntriesCount[id] = make(map[ft.Field]int)
						matchedSimilarities[id] = make(map[ft.Field]float64)
					}

					// Take the best match (lowest edit count or highest similarity)
					if currentCount, exists := matchedEntriesCount[id][key]; !exists || currentCount > match.EditCount {
						matchedEntries[id][key] = match.Text[len(string(key))+1:]
						matchedEntriesCount[id][key] = match.EditCount
						matchedSimilarities[id][key] = match.Similarity
					}
				}
			}
			mapMutex.Unlock()
		}(key, field)
	}

	wg.Wait()

	// Remove all incomplete entries or entries that exceed max edits
	// An entry is incomplete if it has any empty fields
	matchedEntriesCleaned := fmc.CleanMatches(matchedEntries, matchedEntriesCount, fuzzyEntry)

	if len(matchedEntriesCleaned) == 0 {
		return false, nil
	}

	// Build final entries with weighted scores
	finalMatchedEntries := []ft.FuzzyMatch[T]{}

	for id := range matchedEntriesCleaned {
		reject := false

		// Check if all required fields meet minimum similarity thresholds
		for key := range fuzzyEntry.Key {
			min := parameters.MinDistances[key]

			similarity, exists := matchedSimilarities[id][key]

			// Missing required field
			if !exists && min > 0 {
				reject = true
				break
			}

			// Check minimum similarity threshold for individual fields
			if min > 0 && similarity < min {
				reject = true
				break
			}
		}

		// skip entry if individual field requirements not met
		if reject {
			continue
		}

		// Calculate weighted score using pre-calculated similarities
		var score float64
		for key, weight := range parameters.Weights {
			if similarity, exists := matchedSimilarities[id][key]; exists {
				score += weight * similarity
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
