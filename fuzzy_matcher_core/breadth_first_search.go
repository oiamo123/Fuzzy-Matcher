package fuzzymatchercore

import (
	"container/heap"
	"strings"

	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

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

    return float64(predictedChar*0.4) + float64(distance*0.6)
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
   4.4. Compute the current nodes score using prefix prediction / similarity
   4.5. Add the new branch to the priority queue
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

			// 4.4
			score := fmc.ComputeScore(
				branch.Path,
				branch.Word,
				branch.Node.Parent,
				branch.Node,
				branch.CalculationMethod,
			)

			if len(strings.Split(string(branch.Path), ":")[1]) >= 4 && score < float64(params.MinDistance) {
				continue
			}

			// 4.5
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