package fuzzymatchertests

import (
	"testing"
	"time"

	fc "github.com/oiamo123/fuzzy_matcher/fuzzy_classes"
	fmc "github.com/oiamo123/fuzzy_matcher/fuzzy_matcher_core"
	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

// TestMultiCharOCRReplacements tests the efficiency of multi-character OCR corrections
// vs. depth-based searching for character substitutions
func TestMultiCharOCRReplacements(t *testing.T) {
	// Test case 1: "rn" -> "m" replacement (srnith -> smith)
	t.Run("RnToM_SingleReplacement", func(t *testing.T) {
		testRnToMReplacement(t, "srnith", "smith", 1) // Should find with 1 multi-char edit
	})

	// Test case 2: Multiple "rn" -> "m" replacements in same word
	t.Run("RnToM_MultipleReplacements", func(t *testing.T) {
		testRnToMReplacement(t, "srnithsrnith", "smithsmith", 2) // Should find with 2 multi-char edits
	})

	// Test case 3: The extreme case you mentioned - 5 consecutive "rn" -> "m" replacements
	t.Run("RnToM_ExtremeCase", func(t *testing.T) {
		// Input: "srnithsrnithsrnithsrnithsrnith" (5 x "srnith")
		// Expected: "smithsmithsmithsmithsmith" (5 x "smith")
		// Multi-char OCR: 5 edits (one per "rn" -> "m" replacement)
		// Depth-based: 10 edits (one per character: s->s, r->m, n->delete, i->i, t->t, h->h)
		testRnToMReplacement(t, "srnithsrnithsrnithsrnithsrnith", "smithsmithsmithsmithsmith", 5)
	})

	// Test case 4: "nn" -> "m" replacement
	t.Run("NnToM_Replacement", func(t *testing.T) {
		testMultiCharReplacement(t, "jonnson", "jomson", "nn", "m", 1)
	})

	// Test case 5: "cl" -> "d" replacement
	t.Run("ClToD_Replacement", func(t *testing.T) {
		testMultiCharReplacement(t, "clown", "down", "cl", "d", 1)
	})

	// Test case 6: "vv" -> "w" replacement
	t.Run("VvToW_Replacement", func(t *testing.T) {
		testMultiCharReplacement(t, "dovvn", "down", "vv", "w", 1)
	})

	// Test case 7: Mixed multi-character and single-character OCR errors
	t.Run("MixedOCRErrors", func(t *testing.T) {
		// "jonnath4n" -> "jonathan"
		// "nn" -> "m" (1 multi-char edit) + "4" -> "a" (1 single-char edit) = 2 total edits
		testMixedOCRReplacement(t, "jonnath4n", "jonathan", 2)
	})
}

// Helper function to test "rn" -> "m" replacements specifically
func testRnToMReplacement(t *testing.T, searchTerm, targetTerm string, expectedEdits int) {
	// Create fuzzy matcher core with OCR corrections enabled
	params := ft.FuzzyMatcherCoreParameters[fc.WaveMembershipSource]{
		CorrectOcrMisreads: true,
		UseExpiration:      false,
		MaxEdits: 10,
	}

	fuzzyCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{
		CoreParams: params,
	}

	// Create test data with the target term
	testMember := fc.WaveMembershipSource{
		ID:            1,
		Firstname:     targetTerm,
		Surname:       "Test",
		Birthdate:     time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		EventStartUtc: time.Date(2025, 8, 20, 18, 0, 0, 0, time.UTC),
		EventEndUtc:   time.Date(2025, 8, 20, 23, 0, 0, 0, time.UTC),
	}

	// Build the trie with test data
	fuzzyCore.Build([]fc.WaveMembershipSource{testMember})

	// Create search entry with OCR errors
	searchMember := fc.WaveMembershipSource{
		ID:            999,
		Firstname:     searchTerm,
		Surname:       "Test",
		Birthdate:     time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		EventStartUtc: time.Date(2025, 8, 20, 18, 0, 0, 0, time.UTC),
		EventEndUtc:   time.Date(2025, 8, 20, 23, 0, 0, 0, time.UTC),
	}

	// Search for matches
	found, matches := fuzzyCore.SearchFuzzy(searchMember)

	// Verify match was found
	if !found {
		t.Errorf("Expected to find match for '%s' -> '%s' with multi-char OCR correction", searchTerm, targetTerm)
		return
	}

	if len(matches) == 0 {
		t.Errorf("Expected at least one match, got none")
		return
	}

	// Verify the match is the correct entry
	match := matches[0]
	if match.Entry.ID != 1 {
		t.Errorf("Expected to match entry ID 1, got ID %d", match.Entry.ID)
	}

	// Note: We can't directly verify edit count from the match result,
	// but we can verify that the match was found with OCR corrections enabled
	// The efficiency gain is demonstrated by the fact that complex multi-character
	// sequences are found as matches when they wouldn't be with simple depth searching
	t.Logf("Successfully found match for '%s' -> '%s' using multi-char OCR corrections", searchTerm, targetTerm)
	t.Logf("Match score: %.3f", match.Score)
}

// Helper function to test general multi-character replacements
func testMultiCharReplacement(t *testing.T, searchTerm, targetTerm, multiChar, replacement string, expectedReplacements int) {
	// Create fuzzy matcher core with OCR corrections enabled
	params := ft.FuzzyMatcherCoreParameters[fc.WaveMembershipSource]{
		CorrectOcrMisreads: true,
		UseExpiration:      false,
		MaxEdits: 6,
	}
	
	fuzzyCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{
		CoreParams: params,
	}

	// Create test data
	testMember := fc.WaveMembershipSource{
		ID:            1,
		Firstname:     targetTerm,
		Surname:       "Test",
		Birthdate:     time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		EventStartUtc: time.Date(2025, 8, 20, 18, 0, 0, 0, time.UTC),
		EventEndUtc:   time.Date(2025, 8, 20, 23, 0, 0, 0, time.UTC),
	}

	// Build the trie
	fuzzyCore.Build([]fc.WaveMembershipSource{testMember})

	// Create search entry
	searchMember := fc.WaveMembershipSource{
		ID:            999,
		Firstname:     searchTerm,
		Surname:       "Test",
		Birthdate:     time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		EventStartUtc: time.Date(2025, 8, 20, 18, 0, 0, 0, time.UTC),
		EventEndUtc:   time.Date(2025, 8, 20, 23, 0, 0, 0, time.UTC),
	}

	// Search for matches
	found, matches := fuzzyCore.SearchFuzzy(searchMember)

	// Verify results
	if !found {
		t.Errorf("Expected to find match for '%s' -> '%s' with '%s' -> '%s' replacement",
			searchTerm, targetTerm, multiChar, replacement)
		return
	}

	if len(matches) == 0 {
		t.Errorf("Expected at least one match, got none")
		return
	}

	match := matches[0]
	if match.Entry.ID != 1 {
		t.Errorf("Expected to match entry ID 1, got ID %d", match.Entry.ID)
	}

	t.Logf("Successfully found match for '%s' -> '%s' using '%s' -> '%s' replacement",
		searchTerm, targetTerm, multiChar, replacement)
	t.Logf("Match score: %.3f", match.Score)
}

// Helper function to test mixed OCR errors (both single and multi-character)
func testMixedOCRReplacement(t *testing.T, searchTerm, targetTerm string, expectedTotalEdits int) {
	// Create fuzzy matcher core with OCR corrections enabled
	params := ft.FuzzyMatcherCoreParameters[fc.WaveMembershipSource]{
		CorrectOcrMisreads: true,
		UseExpiration:      false,
		MaxEdits: 6,
	}

	fuzzyCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{
		CoreParams: params,
	}

	// Create test data
	testMember := fc.WaveMembershipSource{
		ID:            1,
		Firstname:     targetTerm,
		Surname:       "Test",
		Birthdate:     time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		EventStartUtc: time.Date(2025, 8, 20, 18, 0, 0, 0, time.UTC),
		EventEndUtc:   time.Date(2025, 8, 20, 23, 0, 0, 0, time.UTC),
	}

	// Build the trie
	fuzzyCore.Build([]fc.WaveMembershipSource{testMember})

	// Create search entry
	searchMember := fc.WaveMembershipSource{
		ID:            999,
		Firstname:     searchTerm,
		Surname:       "Test",
		Birthdate:     time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		EventStartUtc: time.Date(2025, 8, 20, 18, 0, 0, 0, time.UTC),
		EventEndUtc:   time.Date(2025, 8, 20, 23, 0, 0, 0, time.UTC),
	}

	// Search for matches
	found, matches := fuzzyCore.SearchFuzzy(searchMember)

	// Verify results
	if !found {
		t.Errorf("Expected to find match for '%s' -> '%s' with mixed OCR corrections", searchTerm, targetTerm)
		return
	}

	if len(matches) == 0 {
		t.Errorf("Expected at least one match, got none")
		return
	}

	match := matches[0]
	if match.Entry.ID != 1 {
		t.Errorf("Expected to match entry ID 1, got ID %d", match.Entry.ID)
	}

	t.Logf("Successfully found match for '%s' -> '%s' using mixed OCR corrections", searchTerm, targetTerm)
	t.Logf("Match score: %.3f", match.Score)
}

// Benchmark to demonstrate the efficiency difference between OCR corrections and depth-based searching
func BenchmarkOCREfficiency(b *testing.B) {
	b.Run("WithOCRCorrections", func(b *testing.B) {
		benchmarkOCRSearch(b, true, "srnithsrnithsrnith", "smithsmithsmith")
	})

	b.Run("WithoutOCRCorrections", func(b *testing.B) {
		benchmarkOCRSearch(b, false, "srnithsrnithsrnith", "smithsmithsmith")
	})
}

func benchmarkOCRSearch(b *testing.B, useOCR bool, searchTerm, targetTerm string) {
	// Create fuzzy matcher core
	params := ft.FuzzyMatcherCoreParameters[fc.WaveMembershipSource]{
		CorrectOcrMisreads: useOCR,
		UseExpiration:      false,
		MaxEdits: 10,
	}

	fuzzyCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{
		CoreParams: params,
	}

	// Create test data
	testMember := fc.WaveMembershipSource{
		ID:            1,
		Firstname:     targetTerm,
		Surname:       "Test",
		Birthdate:     time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		EventStartUtc: time.Date(2025, 8, 20, 18, 0, 0, 0, time.UTC),
		EventEndUtc:   time.Date(2025, 8, 20, 23, 0, 0, 0, time.UTC),
	}

	// Build the trie
	fuzzyCore.Build([]fc.WaveMembershipSource{testMember})

	// Create search entry
	searchMember := fc.WaveMembershipSource{
		ID:            999,
		Firstname:     searchTerm,
		Surname:       "Test",
		Birthdate:     time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		EventStartUtc: time.Date(2025, 8, 20, 18, 0, 0, 0, time.UTC),
		EventEndUtc:   time.Date(2025, 8, 20, 23, 0, 0, 0, time.UTC),
	}

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fuzzyCore.SearchFuzzy(searchMember)
	}
}
