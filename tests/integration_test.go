package fuzzymatchertests

import (
	"testing"
	"time"

	fm "github.com/oiamo123/fuzzy_matcher"
	fc "github.com/oiamo123/fuzzy_matcher/fuzzy_classes"
	fmc "github.com/oiamo123/fuzzy_matcher/fuzzy_matcher_core"
	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

func TestFuzzyMatcher_Integration(t *testing.T) {
	// Create fuzzy matcher (Client not needed for local testing)
	matcher := &fm.FuzzyMatcher[fc.WaveMembershipSource]{}
	matcher.Init(&ft.FuzzyMatcherCoreParameters[fc.WaveMembershipSource]{
		CorrectOcrMisreads: false,
		UseExpiration:      false,
		MaxEdits:           6,
	})

	// Create test data
	testMembers := []fc.WaveMembershipSource{
		{
			ID:             1,
			TicketQuantity: 2,
			Firstname:      "John",
			Surname:        "Smith",
			Birthdate:      time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
			Tag:            "VIP",
			DeletedAt:      nil,
			EventStartUtc:  time.Date(2025, 8, 20, 18, 0, 0, 0, time.UTC),
			EventEndUtc:    time.Date(2025, 8, 20, 23, 0, 0, 0, time.UTC),
		},
		{
			ID:             2,
			TicketQuantity: 1,
			Firstname:      "Sarah",
			Surname:        "Johnson",
			Birthdate:      time.Date(1985, 12, 3, 0, 0, 0, 0, time.UTC),
			Tag:            "GENERAL",
			DeletedAt:      nil,
			EventStartUtc:  time.Date(2025, 8, 20, 18, 0, 0, 0, time.UTC),
			EventEndUtc:    time.Date(2025, 8, 20, 23, 0, 0, 0, time.UTC),
		},
	}

	// Build fuzzyMatcherCore manually for testing (bypassing Init/Sync which requires network)
	fuzzyMatcherCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{
		Root: &ft.FuzzyMatcherNode{
			Children: make(map[rune]*ft.FuzzyMatcherNode),
		},
	}
	fuzzyMatcherCore.Build(testMembers)
	matcher.FuzzyMatcherCore = fuzzyMatcherCore

	// Test exact match using the fuzzyMatcherCore directly
	t.Run("ExactMatch", func(t *testing.T) {
		query := fc.WaveMembershipSource{
			ID:        999, // Different ID to avoid self-match
			Firstname: "John",
			Surname:   "Smith",
			Birthdate: time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
		}

		// Use fuzzyMatcherCore directly to avoid network calls
		found, matches := fuzzyMatcherCore.SearchFuzzy(query)
		if !found {
			t.Error("Expected to find exact match")
		}
		if len(matches) == 0 {
			t.Error("Expected to return matches")
		}

		// Check that we found the correct member
		foundJohn := false
		for _, match := range matches {
			if match.Entry.ID == 1 && match.Entry.Firstname == "John" {
				foundJohn = true
				if match.Score <= 0.8 {
					t.Errorf("Expected high score for exact match, got %f", match.Score)
				}
			}
		}
		if !foundJohn {
			t.Error("Expected to find John Smith")
		}
	})

	// Test fuzzy match
	t.Run("FuzzyMatch", func(t *testing.T) {
		query := fc.WaveMembershipSource{
			ID:        999,
			Firstname: "Jon",   // Missing 'h'
			Surname:   "Smyth", // 'y' instead of 'i'
			Birthdate: time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
		}

		// Use fuzzyMatcherCore directly to avoid network calls
		found, matches := fuzzyMatcherCore.SearchFuzzy(query)
		if found {
			if len(matches) == 0 {
				t.Error("Expected to return matches if found")
			}

			// Check scores are reasonable for fuzzy matches
			for _, match := range matches {
				if match.Entry.ID == 1 {
					if match.Score >= 1.0 {
						t.Errorf("Fuzzy match should not be perfect, got %f", match.Score)
					}
					if match.Score <= 0.3 {
						t.Errorf("Fuzzy match should have reasonable score, got %f", match.Score)
					}
				}
			}
		}
	})

	// Test no match
	t.Run("NoMatch", func(t *testing.T) {
		query := fc.WaveMembershipSource{
			ID:        999,
			Firstname: "Nonexistent",
			Surname:   "Person",
			Birthdate: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		// Use fuzzyMatcherCore directly to avoid network calls
		found, matches := fuzzyMatcherCore.SearchFuzzy(query)
		if found {
			t.Error("Expected no match for non-existent person")
		}
		if len(matches) != 0 {
			t.Error("Expected no matches for non-existent person")
		}
	})
}

func TestTypedFieldsSafety(t *testing.T) {
	// Test that typed fields prevent common errors
	member := fc.WaveMembershipSource{
		Firstname: "Test",
		Surname:   "User",
		Birthdate: time.Now(),
	}

	entry := member.CreateFuzzyEntry()
	if entry == nil {
		t.Fatal("CreateFuzzyEntry returned nil")
	}

	// Test typed field access
	firstname, exists := entry.Key[ft.Firstname]
	if !exists {
		t.Error("Firstname field should exist")
	}
	if firstname != "test" {
		t.Errorf("Expected 'test', got '%s'", firstname)
	}

	surname, exists := entry.Key[ft.Surname]
	if !exists {
		t.Error("Surname field should exist")
	}
	if surname != "user" {
		t.Errorf("Expected 'user', got '%s'", surname)
	}

	// This would be a compile error if attempted:
	// entry.Key["firstnam"] = "typo" // ← Field type prevents this
}

func TestFieldParameterConsistency(t *testing.T) {
	member := fc.WaveMembershipSource{}
	params := member.GetSearchParameters()

	// Test that weights sum to 1.0
	totalWeight := 0.0
	for _, weight := range params.Weights {
		totalWeight += weight
	}
	if totalWeight < 0.99 || totalWeight > 1.01 {
		t.Errorf("Weights should sum to 1.0, got %f", totalWeight)
	}

	// Test that all fields are configured consistently
	expectedFields := []ft.Field{
		ft.Firstname,
		ft.Surname,
		ft.Birthdate,
	}

	for _, field := range expectedFields {
		if _, exists := params.MaxDepth[field]; !exists {
			t.Errorf("MaxDepth missing for field %s", field)
		}
		if _, exists := params.MaxEdits[field]; !exists {
			t.Errorf("MaxEdits missing for field %s", field)
		}
		if _, exists := params.Weights[field]; !exists {
			t.Errorf("Weights missing for field %s", field)
		}
		if _, exists := params.CalculationMethods[field]; !exists {
			t.Errorf("CalculationMethods missing for field %s", field)
		}
		if _, exists := params.MinDistances[field]; !exists {
			t.Errorf("MinDistances missing for field %s", field)
		}
	}
}
