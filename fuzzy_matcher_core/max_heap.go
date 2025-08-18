package fuzzymatchercore

type MaxHeap []ExpiryEntry

// Heap interface implementation for min heap (earliest expiry first)
func (h MaxHeap) Len() int           { return len(h) }
func (h MaxHeap) Less(i, j int) bool { return h[i].Expiry.Before(h[j].Expiry) }
func (h MaxHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

// Adds an element to the heap, maintaining the heap property
func (h *MaxHeap) Push(x interface{}) {
	*h = append(*h, x.(ExpiryEntry))
}

// Removes the element with the earliest expiry time from the heap
func (h *MaxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
