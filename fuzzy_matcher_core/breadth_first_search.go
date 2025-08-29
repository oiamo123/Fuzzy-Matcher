package fuzzymatchercore

import (
	"container/heap"

	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

/*
BREADTH-FIRST-SEARCH FLOW
1. Initialize a priority queue (max heap) for exploring nodes
2. Add the initial node to the priority queue
3. Remove the node from the visited array
4. Loop over all nodes in the priority queue
   4.1. Get the highest priority node
   4.2. Process the node
   4.3. Expand the node's children
   4.4. Prune node
      4.4.1. Skip node if we're at maxEdits-1 and the current node's children doesn't contain the current character
	  4.4.2. Only search the character if the current edits is less than the number of edits in the visited key
   4.5. Compute the current nodes score using prefix prediction / similarity
   4.6. Add the new branch to the priority queue
5. Add the node back to the visited array
*/

func (fmc *FuzzyMatcherCore[T]) BreadthFirstSearch(params *ft.RecurseParameters) []*ft.MatchCandidate {
	// 1.
	maxHeap := &MaxHeap{}
	heap.Init(maxHeap)
	matches := []*ft.MatchCandidate{}

	// 2.
	heap.Push(maxHeap, ft.NodePriority{
		Params: params,
		Score:  0,
	})

	// 4.
	for maxHeap.Len() > 0 {
		// 4.1
		nodePriority := heap.Pop(maxHeap).(ft.NodePriority)
		node := nodePriority.Params.Node

		// 4.2
		match, ok := fmc.ProcessNode(nodePriority.Params)

		matches = append(matches, match...)

		if !ok {
			continue
		}

		// 4.3
		for ch, child := range node.Children {
			// 4.4.1
			if params.NumEdits == params.MaxEdits-1 && params.Node.Children[child.Char] == nil {
				continue
			}

			// 4.4.2
			if params.NumEdits > params.Visited[child] {
				continue
			}

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

			// 4.5
			score := fmc.ComputeScore(branch)

			if len(branch.Path[len(branch.Key)+1:]) >= 4 && score < float64(branch.MinDistance) {
				continue
			}

			// 4.6
			heap.Push(maxHeap, ft.NodePriority{
				Params: branch,
				Score:  score,
			})
		}
	}

	return matches
}
