package fuzzymatchercore

// Resets the fuzzy matcher core
func (fmc *FuzzyMatcherCore[T]) Reset() {
	// We don't want to reset the params since they stay the same
	fmc.Root = nil
	fmc.Entries = nil
	fmc.ExpiryHeap = nil
}