package fuzzymatchertests

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	fc "github.com/oiamo123/fuzzy_matcher/fuzzy_classes"
	fmc "github.com/oiamo123/fuzzy_matcher/fuzzy_matcher_core"
	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestData represents the structure of our test JSON file
type TestData struct {
	Members []TestMember `json:"members"`
}

type TestMember struct {
	ID             string  `json:"id"`
	Firstname      string  `json:"firstname"`
	Surname        string  `json:"surname"`
	Birthdate      string  `json:"birthdate"`
	EventStartUtc  string  `json:"event_start_utc"`
	EventEndUtc    string  `json:"event_end_utc"`
}

// Convert test member to ExampleSource
func (tm TestMember) ToExampleSource() fc.ExampleSource {
	id, _ := strconv.Atoi(tm.ID)
	birthdate, _ := time.Parse("2006-01-02", tm.Birthdate)
	eventStart, _ := time.Parse(time.RFC3339, tm.EventStartUtc)
	eventEnd, _ := time.Parse(time.RFC3339, tm.EventEndUtc)

	return fc.ExampleSource{
		ID:             id,
		Firstname:      tm.Firstname,
		Surname:        tm.Surname,
		Birthdate:      birthdate,
		EventStartUtc:  eventStart,
		EventEndUtc:    eventEnd,
	}
}

// loadTestData loads wave members from the JSON test file
func loadTestData(t *testing.T) []fc.ExampleSource {
	data, err := os.ReadFile("test_data/example_members.json")
	require.NoError(t, err, "Failed to read test data file")

	var testData TestData
	err = json.Unmarshal(data, &testData)
	require.NoError(t, err, "Failed to unmarshal test data")

	members := make([]fc.ExampleSource, len(testData.Members))
	for i, tm := range testData.Members {
		members[i] = tm.ToExampleSource()
	}

	return members
}

// createMockFuzzyMatcherCore creates a fuzzyMatcherCore populated with test data
func createMockFuzzyMatcherCore(t *testing.T, members []fc.ExampleSource) *fmc.FuzzyMatcherCore[fc.ExampleSource] {
	params := ft.FuzzyMatcherCoreParameters[fc.ExampleSource]{
		CorrectOcrMisreads: false,
		UseExpiration:      false,
		MaxEdits: 6,
	}

	fuzzyMatcherCore := &fmc.FuzzyMatcherCore[fc.ExampleSource]{
		CoreParams: params,
	}

	fuzzyMatcherCore.Build(members)
	return fuzzyMatcherCore
}

func TestExampleSource_CreateFuzzyEntry(t *testing.T) {
	member := fc.ExampleSource{
		ID:          123,
		Firstname:   "John",
		Surname:     "Smith",
		Birthdate:   time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
		EventEndUtc: time.Date(2025, 8, 20, 23, 0, 0, 0, time.UTC),
	}

	entry := member.CreateFuzzyEntry()
	require.NotNil(t, entry, "CreateFuzzyEntry should not return nil")

	// Check basic properties
	assert.Equal(t, 123, entry.ID)
	assert.Equal(t, time.Date(2025, 8, 21, 11, 0, 0, 0, time.UTC), entry.Expiry)

	// Check field mapping
	expectedFields := map[ft.Field]string{
		ft.Firstname: "john",
		ft.Surname:   "smith",
		ft.Birthdate: "19900515",
	}

	for field, expectedValue := range expectedFields {
		actualValue, exists := entry.Key[field]
		assert.True(t, exists, "Field %s should exist in entry", field)
		assert.Equal(t, expectedValue, actualValue, "Field %s should have correct value", field)
	}
}

func TestExampleSource_ValidateEntry(t *testing.T) {
	tests := []struct {
		name     string
		member   fc.ExampleSource
		expected bool
	}{
		{
			name: "Valid member",
			member: fc.ExampleSource{
				Firstname:     "John",
				Surname:       "Smith",
				Birthdate:     time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
				EventStartUtc: time.Now().Add(24 * time.Hour), // Future event
			},
			expected: true,
		},
		{
			name: "Empty firstname",
			member: fc.ExampleSource{
				Firstname:     "",
				Surname:       "Smith",
				Birthdate:     time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
				EventStartUtc: time.Now().Add(24 * time.Hour),
			},
			expected: false,
		},
		{
			name: "Empty surname",
			member: fc.ExampleSource{
				Firstname:     "John",
				Surname:       "",
				Birthdate:     time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
				EventStartUtc: time.Now().Add(24 * time.Hour),
			},
			expected: false,
		},
		{
			name: "Short names (average length <= 3)",
			member: fc.ExampleSource{
				Firstname:     "Jo",
				Surname:       "Li",
				Birthdate:     time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
				EventStartUtc: time.Now().Add(24 * time.Hour),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.member.ValidateEntry()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExampleSource_GetSearchParameters(t *testing.T) {
	member := fc.ExampleSource{}
	params := member.GetSearchParameters()

	// Test field configuration existence
	expectedFields := []ft.Field{
		ft.Firstname,
		ft.Surname,
		ft.Birthdate,
	}

	for _, field := range expectedFields {
		assert.Contains(t, params.MaxDepth, field, "MaxDepth should contain field %s", field)
		assert.Contains(t, params.MaxEdits, field, "MaxEdits should contain field %s", field)
		assert.Contains(t, params.Weights, field, "Weights should contain field %s", field)
		assert.Contains(t, params.CalculationMethods, field, "CalculationMethods should contain field %s", field)
	}

	// Test weight distribution sums to 1.0
	totalWeight := 0.0
	for _, weight := range params.Weights {
		totalWeight += weight
	}
	assert.InDelta(t, 1.0, totalWeight, 0.001, "Weights should sum to 1.0")

	// Test specific configurations
	assert.Equal(t, ft.JaroWinkler, params.CalculationMethods[ft.Firstname])
	assert.Equal(t, ft.JaroWinkler, params.CalculationMethods[ft.Surname])
	assert.Equal(t, ft.Default, params.CalculationMethods[ft.Birthdate])
}

// ShortNameValidationTestCase represents a test case for short name validation
type ShortNameValidationTestCase struct {
	Name   string `json:"name"`
	Member struct {
		Firstname     string `json:"firstname"`
		Surname       string `json:"surname"`
		Birthdate     string `json:"birthdate"`
		EventStartUtc string `json:"event_start_utc"`
	} `json:"member"`
	ExpectedMaxDepth struct {
		Firstname int `json:"firstname"`
		Surname   int `json:"surname"`
		Birthdate int `json:"birthdate"`
	} `json:"expected_max_depth"`
	ExpectedMaxEdits struct {
		Firstname int `json:"firstname"`
		Surname   int `json:"surname"`
		Birthdate int `json:"birthdate"`
	} `json:"expected_max_edits"`
	Description string `json:"description"`
}

// ShortNameValidationTestData represents the structure of the test JSON file
type ShortNameValidationTestData struct {
	TestCases []ShortNameValidationTestCase `json:"test_cases"`
}

func TestExampleSource_GetSearchParameters_ShortNameValidation(t *testing.T) {
	// Load test data from JSON
	data, err := os.ReadFile("test_data/short_name_validation_tests.json")
	require.NoError(t, err, "Failed to read test data file")

	var testData ShortNameValidationTestData
	err = json.Unmarshal(data, &testData)
	require.NoError(t, err, "Failed to unmarshal test data")

	for _, tt := range testData.TestCases {
		t.Run(tt.Name, func(t *testing.T) {
			// Parse dates
			birthdate, err := time.Parse("2006-01-02", tt.Member.Birthdate)
			require.NoError(t, err, "Failed to parse birthdate")

			eventStartUtc, err := time.Parse(time.RFC3339, tt.Member.EventStartUtc)
			require.NoError(t, err, "Failed to parse event start UTC")

			// Create member from test data
			member := fc.ExampleSource{
				Firstname:     tt.Member.Firstname,
				Surname:       tt.Member.Surname,
				Birthdate:     birthdate,
				EventStartUtc: eventStartUtc,
			}

			params := member.GetSearchParameters()

			// Test MaxDepth values
			assert.Equal(t, tt.ExpectedMaxDepth.Firstname, params.MaxDepth[ft.Firstname],
				"MaxDepth for firstname: %s", tt.Description)
			assert.Equal(t, tt.ExpectedMaxDepth.Surname, params.MaxDepth[ft.Surname],
				"MaxDepth for surname: %s", tt.Description)
			assert.Equal(t, tt.ExpectedMaxDepth.Birthdate, params.MaxDepth[ft.Birthdate],
				"MaxDepth for birthdate: %s", tt.Description)

			// Test MaxEdits values
			assert.Equal(t, tt.ExpectedMaxEdits.Firstname, params.MaxEdits[ft.Firstname],
				"MaxEdits for firstname: %s", tt.Description)
			assert.Equal(t, tt.ExpectedMaxEdits.Surname, params.MaxEdits[ft.Surname],
				"MaxEdits for surname: %s", tt.Description)
			assert.Equal(t, tt.ExpectedMaxEdits.Birthdate, params.MaxEdits[ft.Birthdate],
				"MaxEdits for birthdate: %s", tt.Description)

			// Verify that validation matches our expectation
			isValidEntry := member.ValidateEntry()
			if tt.ExpectedMaxDepth.Firstname == 0 && tt.ExpectedMaxDepth.Surname == 0 {
				// Short names should fail validation
				assert.False(t, isValidEntry, "Short names should fail validation: %s", tt.Description)
			} else {
				// Normal names should pass validation (assuming other conditions are met)
				assert.True(t, isValidEntry, "Normal names should pass validation: %s", tt.Description)
			}

			// Test that all other parameters are consistent regardless of name length
			expectedWeights := map[ft.Field]float64{
				ft.Firstname: 0.2,
				ft.Surname:   0.4,
				ft.Birthdate: 0.4,
			}

			expectedCalculationMethods := map[ft.Field]ft.CalculationMethod{
				ft.Firstname: ft.JaroWinkler,
				ft.Surname:   ft.JaroWinkler,
				ft.Birthdate: ft.Default,
			}

			expectedMinDistances := map[ft.Field]float64{
				ft.Firstname: 0.7,
				ft.Surname:   0.9,
				ft.Birthdate: 1.0,
			}

			assert.Equal(t, expectedWeights, params.Weights, "Weights should be consistent")
			assert.Equal(t, expectedCalculationMethods, params.CalculationMethods, "Calculation methods should be consistent")
			assert.Equal(t, expectedMinDistances, params.MinDistances, "Min distances should be consistent")
		})
	}
}

func TestFuzzyMatcher_ExactMatch(t *testing.T) {
	members := loadTestData(t)
	fuzzyMatcherCore := createMockFuzzyMatcherCore(t, members)

	// Test exact match for John Smith
	query := fc.ExampleSource{
		ID:        999, // Different ID to avoid self-match
		Firstname: "John",
		Surname:   "Smith",
		Birthdate: time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
	}

	found, matches := fuzzyMatcherCore.SearchFuzzy(query)
	assert.True(t, found, "Should find exact match")
	assert.NotEmpty(t, matches, "Should return matches")

	// Check that we found the correct member
	foundCorrectMember := false
	for _, match := range matches {
		if match.Entry.ID == 1 && match.Entry.Firstname == "John" && match.Entry.Surname == "Smith" {
			foundCorrectMember = true
			assert.Greater(t, match.Score, 0.8, "Exact match should have high score")
		}
	}
	assert.True(t, foundCorrectMember, "Should find the correct John Smith")
}

func TestFuzzyMatcher_FuzzyMatch(t *testing.T) {
	members := loadTestData(t)
	t.Logf("Loaded %d members", len(members))
	for i, m := range members[:3] { // Show first 3 members
		t.Logf("Member %d: ID=%d, Name=%s %s, Birthdate=%s, Valid=%t",
			i, m.ID, m.Firstname, m.Surname, m.Birthdate.Format("2006-01-02"), m.ValidateEntry())
	}

	fuzzyMatcherCore := createMockFuzzyMatcherCore(t, members)

	// Test fuzzy match with typos
	query := fc.ExampleSource{
		ID:        999,
		Firstname: "Jon",   // Typo: missing 'h' (JaroWinkler: 0.933 > 0.7 threshold)
		Surname:   "Smith", // Exact match to ensure it passes surname threshold
		Birthdate: time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
	}

	t.Logf("Query: Name=%s %s, Birthdate=%s, Valid=%t",
		query.Firstname, query.Surname, query.Birthdate.Format("2006-01-02"), query.ValidateEntry())

	found, matches := fuzzyMatcherCore.SearchFuzzy(query)
	t.Logf("SearchFuzzy result: found=%t, matches=%d", found, len(matches))
	for i, match := range matches {
		t.Logf("Match %d: ID=%d, Name=%s %s, Score=%f",
			i, match.Entry.ID, match.Entry.Firstname, match.Entry.Surname, match.Score)
	}

	assert.True(t, found, "Should find fuzzy match")
	assert.NotEmpty(t, matches, "Should return matches")

	// Check for reasonable scores
	for _, match := range matches {
		if match.Entry.ID == 1 {
			assert.Greater(t, match.Score, 0.5, "Fuzzy match should have reasonable score")
			assert.Less(t, match.Score, 1.0, "Fuzzy match should not be perfect")
		}
	}
}

func TestFuzzyMatcher_NoMatch(t *testing.T) {
	members := loadTestData(t)
	fuzzyMatcherCore := createMockFuzzyMatcherCore(t, members)

	// Test with completely different name
	query := fc.ExampleSource{
		ID:        999,
		Firstname: "Xyz",
		Surname:   "Nonexistent",
		Birthdate: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	found, matches := fuzzyMatcherCore.SearchFuzzy(query)
	assert.False(t, found, "Should not find match for non-existent member")
	assert.Empty(t, matches, "Should return no matches")
}

func TestFuzzyMatcher_PartialMatch(t *testing.T) {
	members := loadTestData(t)
	fuzzyMatcherCore := createMockFuzzyMatcherCore(t, members)

	// Test with matching name but different birthdate
	query := fc.ExampleSource{
		ID:        999,
		Firstname: "John",
		Surname:   "Smith",
		Birthdate: time.Date(1985, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	found, matches := fuzzyMatcherCore.SearchFuzzy(query)

	// Depending on your configuration, this might or might not match
	// The birthdate has high weight (0.4) so wrong birthdate might eliminate matches
	if found {
		// If matches are found, they should have lower scores due to birthdate mismatch
		for _, match := range matches {
			if match.Entry.Firstname == "John" && match.Entry.Surname == "Smith" {
				// Score should be lower due to birthdate mismatch
				assert.Less(t, match.Score, 0.7, "Partial match should have lower score")
			}
		}
	}
}

func TestTypedFields_CompileSafety(t *testing.T) {
	// This test ensures typed fields are working correctly
	member := fc.ExampleSource{
		Firstname: "Test",
		Surname:   "User",
		Birthdate: time.Now(),
	}

	entry := member.CreateFuzzyEntry()
	require.NotNil(t, entry)

	// Test that we can access fields using typed constants
	firstname, exists := entry.Key[ft.Firstname]
	assert.True(t, exists)
	assert.Equal(t, "test", firstname)

	surname, exists := entry.Key[ft.Surname]
	assert.True(t, exists)
	assert.Equal(t, "user", surname)

	// Test that the field type prevents typos (this would be a compile error)
	// entry.Key["firstnam"] = "typo" // ← This would not compile
}

func TestFuzzyEntry_Expiry(t *testing.T) {
	eventEnd := time.Date(2025, 8, 20, 23, 0, 0, 0, time.UTC)
	member := fc.ExampleSource{
		ID:          123,
		Firstname:   "John",
		Surname:     "Smith",
		Birthdate:   time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
		EventEndUtc: eventEnd,
	}

	entry := member.CreateFuzzyEntry()
	require.NotNil(t, entry)

	expectedExpiry := eventEnd.Add(12 * time.Hour)
	assert.Equal(t, expectedExpiry, entry.Expiry, "Expiry should be 12 hours after event end")
}

func BenchmarkExampleSource_CreateFuzzyEntry(b *testing.B) {
	member := fc.ExampleSource{
		ID:        123,
		Firstname: "John",
		Surname:   "Smith",
		Birthdate: time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry := member.CreateFuzzyEntry()
		_ = entry
	}
}

func BenchmarkFuzzyMatcher_Search(b *testing.B) {
	// Load test data once
	data, _ := os.ReadFile("test_data/example_members.json")
	var testData TestData
	json.Unmarshal(data, &testData)

	members := make([]fc.ExampleSource, len(testData.Members))
	for i, tm := range testData.Members {
		members[i] = tm.ToExampleSource()
	}

	fuzzyMatcherCore := createMockFuzzyMatcherCore(nil, members)

	query := fc.ExampleSource{
		ID:        999,
		Firstname: "John",
		Surname:   "Smith",
		Birthdate: time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fuzzyMatcherCore.SearchFuzzy(query)
	}
}
