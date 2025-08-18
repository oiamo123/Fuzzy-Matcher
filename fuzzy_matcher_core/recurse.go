package fuzzymatchercore

import (
	"fmt"

	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

var ocrMisreads = map[rune][]rune{
	'0': {'o', 'd', 'q'},
	'1': {'l', 'i'},
	'2': {'z', 's'},
	'3': {'e', '8', 'b'},
	'4': {'a', 'h'},
	'5': {'s'},
	'6': {'b', 'g', 'G'},
	'7': {'t', 'y'},
	'8': {'b', '3', 'B'},
	'9': {'g', 'q'},
	'o': {'0', 'a'},
	'i': {'1', 'l'},
	'l': {'1', 'i'},
	'b': {'8', '3', '6'},
	'g': {'6', '9'},
	'z': {'2'},
	'c': {'e', 'o'},
	's': {'5'},
	'n': {'m', 'r'},
	'a': {'o'},
	'e': {'c'},
	'r': {'n'},
	'v': {'u'},
	'u': {'v'},
}

var multiCharMisreads = map[string][][]rune{
	"m":  {{'r', 'n'}, {'n', 'n'}},
	"cl": {{'d'}},
	"rn": {{'m'}},
	"nn": {{'m'}},
	"w":  {{'v', 'v'}},
	"d":  {{'c', 'l'}},
}

type RecurseParameters struct {
	word []rune
	index int
	node *FuzzyMatcherNode
	path []rune
	maxDepth int
	depth int
	depthIncrement int
	numEdits int
	maxEdits int
	numEditsIncrement int
	editableFields []bool
	visited map[VisitKey]struct{}
}

/*
CHECKS FLOW
1. Increment depth and num edits
2. Check if node has been visited and update as needed
3. Check if current node is end of string
4. Check if we've exceeded limits
*/

func (fmc *FuzzyMatcherCore[T]) ProcessNode(params *RecurseParameters) ([]ft.MatchCandidate, bool) {
    // 1. Apply depth and edit costs
    params.depth += params.depthIncrement
    params.numEdits += params.numEditsIncrement

    // 2. Check if already visited
    key := VisitKey{params.index, params.node, params.numEdits, params.depth}
    if _, seen := params.visited[key]; seen {
        return nil, false // already visited, skip
    }
    params.visited[key] = struct{}{}

    matches := []ft.MatchCandidate{}

    // 3. If this node is an end-of-string, add match
    if params.node.IsEndofString {
        ids := make([]int, 0, len(params.node.ID))
        for id := range params.node.ID {
            ids = append(ids, id)
        }
        matches = append(matches, ft.MatchCandidate{
            Text:        string(params.path),
            EditCount:   params.numEdits,
            SearchDepth: params.depth,
            ID:          ids,
        })
    }

    // 4. Early exit if over limits
    if params.numEdits > params.maxEdits || params.depth > params.maxDepth {
        return matches, false // stop further recursion/BFS
    }

    return matches, true // continue exploring
}

/*
BREADTH-FIRST-SEARCH FLOW
1. 
*/

func (fuzzyMatcherCore *FuzzyMatcherCore[T]) BreadthFirstSearch(params RecurseParameters) []ft.MatchCandidate {
}

/*
RECURSE FLOW:
1. Perform BFS if the index has reached the end of the search word
	- IE: Searching for "Mike", "Michael" is a match

2. Perform Normal match
	- IE: Current char is 'm' and current node has child node 'm'

3. If the current character is editable, we can also:
	3.1. Skip the current character
	- IE: Searching for "Mike" and on 'i', we can skip to 'k'
	3.2. Perform BFS
	- IE: Searching for "Mike" and on 'i', we can perform BFS to find "Michael"

4. If the fuzzy matcher is correcting OCR misreads, we can also:
	4.1. Check for single-character ocr misreads
	- "Searching for "M1ke", switch the '1' out with an 'i'"
	4.2. Check for multi-character ocr misreads
	- "Searching for "Srnith", switch the 'rn' out with an 'm'"
*/

func (fmc *FuzzyMatcherCore[T]) Recurse(params RecurseParameters) []ft.MatchCandidate {
	matches, ok := fmc.ProcessNode(&params)
	if !ok {
    	return matches // stop recursion
	}

	// Exploration
	// 1.
	if params.index >= len(params.word) {
		for ch, child := range params.node.Children {
			branch := params
			branch.path = append(branch.path, ch)
			branch.depthIncrement = 1
			branch.numEditsIncrement = 1
			branch.node = child

			res := fmc.Recurse(branch)

			if res != nil {
				matches = append(matches, res...)
			}
		}
		return matches
	}

	char := params.word[params.index]

	fmt.Printf("Path: %s, Current Char: %s, Node count: %d\n", string(params.path), string(char), params.node.Count)

	// 2.
	if params.node.Children[char] != nil {
		branch := params
		branch.index++
		branch.node = branch.node.Children[char]
		branch.path = append(branch.path, char)
		branch.depthIncrement = 0
		branch.numEditsIncrement = 0

		result := fmc.Recurse(branch)

		if result != nil {
			matches = append(matches, result...)
		}
	}

	// 3.
	if params.editableFields[params.index] {
		// 3.1.
		if params.index+1 <= len(params.word) {
			branch := params
			branch.index++
			branch.depthIncrement = 1
			branch.numEditsIncrement = 1

			result := fmc.Recurse(branch)

			if result != nil {
				matches = append(matches, result...)
			}
		}

		// 3.2. 
		for ch, child := range params.node.Children {
			branch := params
			branch.index++
			branch.node = child
			branch.path = append(branch.path, ch)
			branch.depthIncrement = 1
			branch.numEditsIncrement = 1

			result := fmc.Recurse(branch)

			if result != nil {
				matches = append(matches, result...)
			}
		}

		// If the fuzzy matcher is correcting OCR misreads
		if fmc.CoreParams.CorrectOcrMisreads {
			// 5.
			for _, sub := range ocrMisreads[char] {
				if params.node.Children[sub] != nil {
					branch := params
					branch.index++
					branch.node = branch.node.Children[sub]
					branch.path = append(branch.path, sub)
					branch.depthIncrement = 1
					branch.numEditsIncrement = 1

					result := fmc.Recurse(branch)

					if result != nil {
						matches = append(matches, result...)
					}
				}
			}

			// 6. 
			if params.index+1 < len(params.word) {
				twoChars := string(params.word[params.index : params.index+2])

				if multiCharMisreads[twoChars] != nil {
					for _, subRunes := range multiCharMisreads[twoChars] {
						child := params.node
						valid := true
						for _, r := range subRunes {
							next := child.Children[r]
							if next == nil {
								valid = false
								break
							}
							child = next
						}
						if valid {
							branch := params
							branch.index += len(twoChars)
							branch.node = child
							branch.path = append(branch.path, subRunes...)
							branch.depthIncrement = 1
							branch.numEditsIncrement = 1

							result := fmc.Recurse(branch)

							matches = append(matches, result...)
						}
					}
				}
			}
		}
	}

	return matches
}
