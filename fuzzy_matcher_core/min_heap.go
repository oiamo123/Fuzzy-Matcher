package fuzzymatchercore

import (
	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

type ExpiryHeap []ft.ExpiryEntry

// Heap interface implementation for min heap (earliest expiry first)
func (h ExpiryHeap) Len() int           { return len(h) }
func (h ExpiryHeap) Less(i, j int) bool { return h[i].Expiry.Before(h[j].Expiry) }
func (h ExpiryHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

// Adds an element to the heap, maintaining the heap property
func (h *ExpiryHeap) Push(x interface{}) {
	*h = append(*h, x.(ft.ExpiryEntry))
}

// Removes the element with the earliest expiry time from the heap
func (h *ExpiryHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
