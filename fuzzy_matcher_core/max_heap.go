package fuzzymatchercore

import (
	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

type MaxHeap []ft.NodePriority

// Heap interface implementation for max heap (highest score first)
func (m MaxHeap) Len() int           { return len(m) }
func (m MaxHeap) Less(i, j int) bool { return m[i].Score > m[j].Score }
func (m MaxHeap) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

// Adds an element to the heap, maintaining the heap property
func (m *MaxHeap) Push(x interface{}) {
	*m = append(*m, x.(ft.NodePriority))
}

// Removes the element with the earliest expiry time from the heap
func (m *MaxHeap) Pop() interface{} {
	old := *m
	n := len(old)
	x := old[n-1]
	*m = old[0 : n-1]
	return x
}
