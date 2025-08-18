package fuzzymatchercore

import (
	"container/heap"
	"fmt"
	"strings"

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

/*
HELPERS
*/

func (fmc *FuzzyMatcherCore[T]) ComputeScore(
	path, word []rune, 
	parent, child *ft.FuzzyMatcherNode, 
	method ft.CalculationMethod, 
) float64 {
	// Next character prediction
	predictedChar := float64(child.Count) / float64(parent.Count)

	s1 := strings.Split(string(path), ":")[1]
	s2 := strings.Split(string(word), ":")[1]
	distance := fmc.CalculateSimilarity(s1, s2, method)

    return (predictedChar * 0.6) + (distance * 0.4)
}

/*
CHECKS FLOW
1. Increment depth and num edits
2. Check if node has been visited and update as needed
3. Check if current node is end of string
4. Check if we've exceeded limits
*/

func (fmc *FuzzyMatcherCore[T]) ProcessNode(params *ft.RecurseParameters) ([]ft.MatchCandidate, bool) {
    // 1. Apply depth and edit costs
    params.Depth += params.DepthIncrement
    params.NumEdits += params.NumEditsIncrement

    // 2. Check if already visited
	key := ft.VisitKey{
		Index: params.Index,
		Node:  params.Node,
		Edits: params.NumEdits,
		Depth:   params.Depth,
	}

    params.Visited[key] = struct{}{}

    matches := []ft.MatchCandidate{}

    // 3. If this node is an end-of-string, add match
    if params.Node.IsEndofString {
        ids := make([]int, 0, len(params.Node.ID))
        for id := range params.Node.ID {
            ids = append(ids, id)
        }

        matches = append(matches, ft.MatchCandidate{
            Text:        string(params.Path),
            EditCount:   params.NumEdits,
            SearchDepth: params.Depth,
            ID:          ids,
        })
    }

    // 4. Early exit if over limits
    if params.NumEdits > params.MaxEdits || params.Depth > params.MaxDepth {
        return matches, false // stop further recursion/BFS
    }

    return matches, true // continue exploring
}

/*
BREADTH-FIRST-SEARCH FLOW
1. Initialize a priority queue (max heap) for exploring nodes
2. Add the initial node to the priority queue
3. Remove the node from the visited array
4. Loop over all nodes in the priority queue
   4.1. Get the highest priority node
   4.2. Process the node
   4.3. Expand the node's children
   4.4. Add the new branch to the priority queue
5. Add the node back to the visited array
*/

func (fmc *FuzzyMatcherCore[T]) BreadthFirstSearch(params ft.RecurseParameters) []ft.MatchCandidate {
	// 1.
	maxHeap := &MaxHeap{}
	heap.Init(maxHeap)
	matches := []ft.MatchCandidate{}

	// 2.
	heap.Push(maxHeap, ft.NodePriority{
		Params: params,
		Score:  0,
	})

	// 3.
	key := ft.VisitKey{
		Index:   params.Index,
		Node:    params.Node,
		Edits:   params.NumEdits,
		Depth:   params.Depth,
	}

	delete(params.Visited, key)

	// 4.
	for maxHeap.Len() > 0 {
		// 4.1
		nodePriority := heap.Pop(maxHeap).(ft.NodePriority)
		node := nodePriority.Params.Node

		// 4.2
		match, ok := fmc.ProcessNode(&nodePriority.Params)
		// fmt.Printf("BFS Path: %s, Num Edits: %d, Depth: %d, Char: %c Word: %s, Ok: %v\n", 
		// 	string(nodePriority.Params.Path), 
		// 	nodePriority.Params.NumEdits, 
		// 	nodePriority.Params.Depth, 
		// 	node.Char, 
		// 	string(nodePriority.Params.Word), 
		// 	ok,
		// )

		matches = append(matches, match...)

		if !ok {
			continue
		} 

		// 4.3
		for ch, child := range node.Children {
			branch := nodePriority.Params.Clone()
			branch.Path = append(branch.Path, ch)
			branch.Node = child
			branch.Index++
			branch.DepthIncrement = 0
			branch.NumEditsIncrement = 0
			
			if branch.Index-1 < len(branch.Word) && ch != branch.Word[branch.Index-1] {
    			branch.NumEditsIncrement = 1
    			branch.DepthIncrement = 1
			}

			score := fmc.ComputeScore(
				branch.Path,
				branch.Word,
				branch.Node.Parent,
				branch.Node,
				branch.CalculationMethod,
			)

			// 4.4
			heap.Push(maxHeap, ft.NodePriority{
				Params: branch,
				Score:  score,
			})
		}
	}

	// 5.
	params.Visited[key] = struct{}{}

	return matches
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

func (fmc *FuzzyMatcherCore[T]) Recurse(params ft.RecurseParameters) []ft.MatchCandidate {
	// 1.
	if params.Index >= len(params.Word) {
		return fmc.BreadthFirstSearch(params.Clone())
	}

	matches, ok := fmc.ProcessNode(&params)
	if !ok {
    	return matches // stop recursion
	}

	char := params.Word[params.Index]

	fmt.Printf("Recurse Path: %s, Num Edits: %d, Depth: %d\n", string(params.Path), params.NumEdits, params.Depth)

	// 2.
	if params.Node.Children[char] != nil {
		branch := params.Clone()
		branch.Index++
		branch.Node = branch.Node.Children[char]
		branch.Path = append(branch.Path, char)
		branch.DepthIncrement = 0
		branch.NumEditsIncrement = 0

		result := fmc.Recurse(branch)

		if result != nil {
			matches = append(matches, result...)
		}
	}

	// 3.
	if params.EditableFields[params.Index] {
		// 3.1.
		if params.Index+1 <= len(params.Word) {
			branch := params.Clone()
			branch.Index++
			branch.DepthIncrement = 1
			branch.NumEditsIncrement = 1

			result := fmc.Recurse(branch)

			if result != nil {
				matches = append(matches, result...)
			}
		}

		// 3.2. 
		matches = append(matches, fmc.BreadthFirstSearch(params.Clone())...)

		// If the fuzzy matcher is correcting OCR misreads
		if fmc.CoreParams.CorrectOcrMisreads {
			// 5.
			for _, sub := range ocrMisreads[char] {
				if params.Node.Children[sub] != nil {
					branch := params.Clone()	
					branch.Index++
					branch.Node = branch.Node.Children[sub]
					branch.Path = append(branch.Path, sub)
					branch.DepthIncrement = 1
					branch.NumEditsIncrement = 1

					result := fmc.Recurse(branch)

					if result != nil {
						matches = append(matches, result...)
					}
				}
			}

			// 6. 
			if params.Index+1 < len(params.Word) {
				twoChars := string(params.Word[params.Index : params.Index+2])

				if multiCharMisreads[twoChars] != nil {
					for _, subRunes := range multiCharMisreads[twoChars] {
						child := params.Node
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
							branch := params.Clone()
							branch.Index += len(twoChars)
							branch.Node = child
							branch.Path = append(branch.Path, subRunes...)
							branch.DepthIncrement = 1
							branch.NumEditsIncrement = 1

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
