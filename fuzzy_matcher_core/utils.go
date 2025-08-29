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

func (fmc *FuzzyMatcherCore[T]) ProcessNode(params *ft.RecurseParameters) ([]*ft.MatchCandidate, bool) {
	// 1. Apply depth and edit costs
	params.Depth += params.DepthIncrement
	params.NumEdits += params.NumEditsIncrement

	// 2. Check if already visited
	if params.Visited[params.Node] != 0 && params.NumEdits >= params.Visited[params.Node] {
		return nil, false // stop further recursion/BFS
	}

	params.Visited[params.Node] = params.NumEdits

	matches := []*ft.MatchCandidate{}

	// 3. If this node is an end-of-string, add match
	if params.Node.IsEndofString {
		ids := make([]int, 0, len(params.Node.ID))
		for id := range params.Node.ID {
			ids = append(ids, id)
		}

		matches = append(matches, &ft.MatchCandidate{
			Text:        string(params.Path),
			EditCount:   params.NumEdits,
			SearchDepth: params.Depth,
			ID:          ids,
			Similarity: fmc.CalculateSimilarity(
				string(params.Path[len(params.Key)+1:]),
				string(params.Word[len(params.Key)+1:]),
				params.CalculationMethod,
			),
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
func (fmc *FuzzyMatcherCore[T]) ComputeScore(params *ft.RecurseParameters) float64 {
	// Next character prediction
	predictedChar := float64(params.Node.Count) / float64(params.Node.Parent.Count)

	s1 := params.Path[len(params.Key)+1:]
	s2 := params.Word[len(params.Key)+1:]

	distance := fmc.CalculateSimilarity(string(s1), string(s2), params.CalculationMethod)

	return float64(predictedChar*0.4) + float64(distance*0.6)
}
