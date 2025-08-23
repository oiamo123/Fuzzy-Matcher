package fuzzymatchercore

import (
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
RECURSE FLOW:
1. Perform BFS if the index has reached the end of the search word
	- IE: Searching for "Mike", "Michael" is a match

2. Early exit
	- IE: We're at MaxEdits - 1 and the current nodes children doesn't contain the current character

3. Perform Normal match
	- IE: Current char is 'm' and current node has child node 'm'

4. If the current character is editable, we can also:
	4.1. Skip the current character
	- IE: Searching for "Mike" and on 'i', we can skip to 'k'
	4.2. Perform BFS
	- IE: Searching for "Mike" and on 'i', we can perform BFS to find "Michael"

5. If the fuzzy matcher is correcting OCR misreads, we can also:
	5.1. Check for single-character ocr misreads
	- "Searching for "M1ke", switch the '1' out with an 'i'"
	5.2. Check for multi-character ocr misreads
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

	// 2.
	if params.NumEdits == params.MaxEdits-1 && params.Node.Children[char] == nil {
		return matches
	}

	// 3.
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

	// 4.
	if params.EditableFields[params.Index] {
		// 4.1.
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

		// 4.2. 
		matches = append(matches, fmc.BreadthFirstSearch(params.Clone())...)

		// 5.
		if fmc.CoreParams.CorrectOcrMisreads {
			// 5.1
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

			// 5.2
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
