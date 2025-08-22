package fuzzymatchercore

import (
	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

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
	key := fmc.MakeKey(params.Index, params.NumEdits, params.Depth, int(params.Node.Char))

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
COMPUTES SCORE
- Uses next character prediction + distance to calculate similarity
*/
func (fmc *FuzzyMatcherCore[T]) ComputeScore(
	path, word, key []rune, 
	parent, child *ft.FuzzyMatcherNode, 
	method ft.CalculationMethod, 
) float64 {
	// Next character prediction
	predictedChar := float64(child.Count) / float64(parent.Count)

	s1 := path[len(key)+1:]
	s2 := word[len(key)+1:]

	distance := fmc.CalculateSimilarity(string(s1), string(s2), method)

    return float64(predictedChar*0.4) + float64(distance*0.6)
}

func (fmc *FuzzyMatcherCore[T]) MakeKey(index, edits, depth, nodeID int) ft.VisitKey {
	return ft.VisitKey(
		(uint64(index) << 48) |
        (uint64(edits) << 32) |
        (uint64(depth) << 16) |
        uint64(nodeID & 0xFFFF),
	)
}