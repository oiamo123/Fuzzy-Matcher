package fuzzymatchercore

import (
	"container/heap"
	"time"

	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

// Propogate backwards to prune the fuzzy matcher
func (fmc *FuzzyMatcherCore[T]) Prune(node *FuzzyMatcherNode) {
	if node == nil {
		return
	}

	// If the node is an end of string or has 1 or more children or has an ID, it cannot be pruned
	if node.IsEndofString || len(node.Children) >= 1 || len(node.ID) > 0 {
		return
	}

	// Else remove the node from its parent's children and continue pruning
	if node.Parent != nil {
		delete(node.Parent.Children, node.Char)
		fmc.Prune(node.Parent)
	}
}

// Cleans up the fuzzy matcher by removing expired entries
func (fmc *FuzzyMatcherCore[T]) Clean() {
	if !fmc.CoreParams.UseExpiration {
		return
	}

	if fmc.Root == nil {
		return
	}

	now := time.Now()
	for fmc.ExpiryHeap.Len() > 0 && fmc.ExpiryHeap[0].Expiry.Before(now) {
		entry := heap.Pop(&fmc.ExpiryHeap).(ExpiryEntry)

		// Remove the ID from the node
		delete(entry.Node.ID, entry.ID)

		if len(entry.Node.ID) == 0 {
			// If the node has no IDs left, prune it
			entry.Node.IsEndofString = false
			fmc.Prune(entry.Node)
		}
	}
}

// Cleans up the matched entries by removing those that exceed max edits or have empty fields
func (fmc *FuzzyMatcherCore[T]) CleanMatches(
	matchedEntries map[int]map[ft.Field]string,
	matchedEntriesCount map[int]map[ft.Field]int,
	fuzzyEntry *ft.FuzzyEntry,
) map[int]map[ft.Field]string {
	if len(matchedEntries) == 0 {
		return matchedEntries
	}

	var matchedEntriesCleaned = make(map[int]map[ft.Field]string)

	for id, match := range matchedEntries {
		shouldDelete := false

		// Check total edits for the entire entry first
		totalNumEdits := 0
		for _, count := range matchedEntriesCount[id] {
			totalNumEdits += count
		}

		totalEdits := totalNumEdits

		if totalEdits > fmc.CoreParams.MaxEdits {
			shouldDelete = true
		}

		if shouldDelete {
			delete(matchedEntries, id)
			continue
		}

		// If all checks pass, add to the cleaned entries
		matchedEntriesCleaned[id] = match
	}

	return matchedEntriesCleaned
}
