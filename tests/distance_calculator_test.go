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

// Test data structures for JSON loading
type DistanceTestData struct {
	JaroWinklerTests []DistanceTest        `json:"jaro_winkler_tests"`
	LevenshteinTests []DistanceTest        `json:"levenshtein_tests"`
	DefaultTests     []DistanceTestDefault `json:"default_tests"`
}

type DistanceTest struct {
	Name     string  `json:"name"`
	S1       string  `json:"s1"`
	S2       string  `json:"s2"`
	Expected float64 `json:"expected"`
	Delta    float64 `json:"delta"`
	Note     string  `json:"note,omitempty"`
}

type DistanceTestDefault struct {
	Name     string  `json:"name"`
	S1       string  `json:"s1"`
	S2       string  `json:"s2"`
	Expected float64 `json:"expected"`
	Note     string  `json:"note,omitempty"`
}

type BasicTestData struct {
	BasicTestMembers []BasicTestMember `json:"basic_test_members"`
	TestQueries      TestQueries       `json:"test_queries"`
	ValidationData   ValidationData    `json:"validation_data"`
}

type BasicTestMember struct {
	ID          int    `json:"id"`
	Firstname   string `json:"firstname"`
	Surname     string `json:"surname"`
	Birthdate   string `json:"birthdate"`
	EventEndUtc string `json:"event_end_utc"`
	Note        string `json:"note,omitempty"`
}

type TestQueries struct {
	ExactMatch  TestQuery `json:"exact_match"`
	FuzzyMatch  TestQuery `json:"fuzzy_match"`
	EmptySearch TestQuery `json:"empty_search"`
}

type TestQuery struct {
	ID            int    `json:"id"`
	Firstname     string `json:"firstname"`
	Surname       string `json:"surname"`
	Birthdate     string `json:"birthdate"`
	EventStartUtc string `json:"event_start_utc,omitempty"`
	Note          string `json:"note,omitempty"`
}

type ValidationData struct {
	FuzzyEntryKeys     map[string]string `json:"fuzzy_entry_keys"`
	FuzzyEntryMetadata FuzzyEntryMeta    `json:"fuzzy_entry_metadata"`
	Parameters         TestParameters    `json:"parameters"`
}

type FuzzyEntryMeta struct {
	ID     int    `json:"id"`
	Expiry string `json:"expiry"`
}

type TestParameters struct {
	MaxDepth           map[string]int     `json:"max_depth"`
	MaxEdits           map[string]int     `json:"max_edits"`
	Weights            map[string]float64 `json:"weights"`
	CalculationMethods map[string]string  `json:"calculation_methods"`
	MinDistances       map[string]float64 `json:"min_distances"`
}

// Helper functions for loading test data
func loadDistanceTestData(t *testing.T) DistanceTestData {
	data, err := os.ReadFile("test_data/distance_tests.json")
	require.NoError(t, err, "Failed to read distance test data")

	var testData DistanceTestData
	err = json.Unmarshal(data, &testData)
	require.NoError(t, err, "Failed to unmarshal distance test data")

	return testData
}

func loadBasicTestData(t *testing.T) BasicTestData {
	data, err := os.ReadFile("test_data/basic_tests.json")
	require.NoError(t, err, "Failed to read basic test data")

	var testData BasicTestData
	err = json.Unmarshal(data, &testData)
	require.NoError(t, err, "Failed to unmarshal basic test data")

	return testData
}

func convertToWaveMember(member BasicTestMember) fc.WaveMembershipSource {
	birthdate, _ := time.Parse("2006-01-02", member.Birthdate)
	eventEnd, _ := time.Parse(time.RFC3339, member.EventEndUtc)

	return fc.WaveMembershipSource{
		ID:          member.ID,
		Firstname:   member.Firstname,
		Surname:     member.Surname,
		Birthdate:   birthdate,
		EventEndUtc: eventEnd,
	}
}

func convertQueryToWaveMember(query TestQuery) fc.WaveMembershipSource {
	birthdate, _ := time.Parse("2006-01-02", query.Birthdate)

	// Parse EventStartUtc if provided, otherwise default to future date
	var eventStartUtc time.Time
	if query.EventStartUtc != "" {
		eventStartUtc, _ = time.Parse("2006-01-02", query.EventStartUtc)
	} else {
		// Default to a future date if not provided
		eventStartUtc = time.Date(2999, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	return fc.WaveMembershipSource{
		ID:            query.ID,
		Firstname:     query.Firstname,
		Surname:       query.Surname,
		Birthdate:     birthdate,
		EventStartUtc: eventStartUtc,
	}
}

func TestFuzzyMatcherCore_CalculateSimilarity_JaroWinkler(t *testing.T) {
	fuzzyMatcherCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{}

	// Load test data from JSON
	distanceTests := loadDistanceTestData(t)

	for _, tt := range distanceTests.JaroWinklerTests {
		t.Run(tt.Name, func(t *testing.T) {
			result := fuzzyMatcherCore.CalculateSimilarity(tt.S1, tt.S2, ft.JaroWinkler, 0.0)
			assert.InDelta(t, tt.Expected, result, tt.Delta,
				"JaroWinkler(%q, %q) = %f, expected ~%f", tt.S1, tt.S2, result, tt.Expected)
		})
	}
}

func TestFuzzyMatcherCore_CalculateSimilarity_Levenshtein(t *testing.T) {
	fuzzyMatcherCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{}

	// Load test data from JSON
	distanceTests := loadDistanceTestData(t)

	for _, tt := range distanceTests.LevenshteinTests {
		t.Run(tt.Name, func(t *testing.T) {
			result := fuzzyMatcherCore.CalculateSimilarity(tt.S1, tt.S2, ft.Levenshtein, 0.0)
			assert.InDelta(t, tt.Expected, result, tt.Delta,
				"Levenshtein(%q, %q) = %f, expected ~%f", tt.S1, tt.S2, result, tt.Expected)
		})
	}
}

// Can't really test default because number of edits is considered before it gets to the distance metric
func TestFuzzyMatcherCore_CalculateSimilarity_Default(t *testing.T) {
	fuzzyMatcherCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{}

	// Load test data from JSON
	distanceTests := loadDistanceTestData(t)

	for _, tt := range distanceTests.DefaultTests {
		t.Run(tt.Name, func(t *testing.T) {
			result := fuzzyMatcherCore.CalculateSimilarity(tt.S1, tt.S2, ft.Default, 0.0)
			assert.Equal(t, tt.Expected, result,
				"Default(%q, %q) = %f, expected %f", tt.S1, tt.S2, result, tt.Expected)
		})
	}
}

func TestFuzzyMatcherCore_Insert_And_Search(t *testing.T) {
	// Load test data from JSON
	basicTests := loadBasicTestData(t)

	// Convert JSON members to WaveMembershipSource
	members := make([]fc.WaveMembershipSource, len(basicTests.BasicTestMembers))
	for i, member := range basicTests.BasicTestMembers {
		members[i] = convertToWaveMember(member)
	}

	
	params := ft.FuzzyMatcherCoreParameters[fc.WaveMembershipSource]{
		CorrectOcrMisreads: false,
		UseExpiration:      false,
		MaxEdits: 6,
	}

	fuzzyMatcherCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{
		CoreParams: params,
	}

	fuzzyMatcherCore.Build(members)

	// Test search for existing member using JSON query
	query := convertQueryToWaveMember(basicTests.TestQueries.ExactMatch)

	found, matches := fuzzyMatcherCore.SearchFuzzy(query)
	assert.True(t, found, "Should find exact match")
	assert.NotEmpty(t, matches, "Should return matches")

	// Verify the correct member was found
	foundJohn := false
	for _, match := range matches {
		if match.Entry.ID == 1 && match.Entry.Firstname == "John" {
			foundJohn = true
			assert.Greater(t, match.Score, 0.8, "Exact match should have high score")
		}
	}
	assert.True(t, foundJohn, "Should find John Smith")
}

func TestFuzzyMatcherCore_FuzzySearch_Comprehensive(t *testing.T) {
	// Load test data from JSON
	testData := loadFuzzySearchTestCases(t)
	members := loadWaveMembersTestData(t)

	// Create fuzzyMatcherCore with all test data
	params := ft.FuzzyMatcherCoreParameters[fc.WaveMembershipSource]{
		CorrectOcrMisreads: false,
		UseExpiration:      false,
		MaxEdits: 6,
	}

	fuzzyMatcherCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{
		CoreParams: params,
	}

	fuzzyMatcherCore.Build(members)

	// Run each test case
	for _, testCase := range testData.TestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// Convert test case query to WaveMembershipSource
			birthdate, err := time.Parse("2006-01-02", testCase.Query.Birthdate)
			require.NoError(t, err, "Failed to parse birthdate for test case %s", testCase.Name)

			// Parse EventStartUtc if provided, otherwise default to future date
			var eventStartUtc time.Time
			if testCase.Query.EventStartUtc != "" {
				eventStartUtc, _ = time.Parse("2006-01-02", testCase.Query.EventStartUtc)
			} else {
				// Default to a future date if not provided
				eventStartUtc = time.Date(2999, 1, 1, 0, 0, 0, 0, time.UTC)
			}

			query := fc.WaveMembershipSource{
				ID:            999, // Use different ID to avoid self-match
				Firstname:     testCase.Query.Firstname,
				Surname:       testCase.Query.Surname,
				Birthdate:     birthdate,
				EventStartUtc: eventStartUtc,
			}

			// Execute search
			found, matches := fuzzyMatcherCore.SearchFuzzy(query)

			// Validate basic expectations
			assert.Equal(t, testCase.Expected.ShouldFind, found,
				"Test case %s: expected should_find=%t, got %t",
				testCase.Name, testCase.Expected.ShouldFind, found)

			if testCase.Expected.MinMatches > 0 {
				assert.GreaterOrEqual(t, len(matches), testCase.Expected.MinMatches,
					"Test case %s: expected at least %d matches, got %d",
					testCase.Name, testCase.Expected.MinMatches, len(matches))
			} else {
				assert.Empty(t, matches, "Test case %s: expected no matches", testCase.Name)
			}

			// Validate specific expected matches
			for _, expectedMatch := range testCase.Expected.ExpectedMatches {
				foundMatch := false
				for _, actualMatch := range matches {
					if actualMatch.Entry.ID == expectedMatch.MemberID {
						foundMatch = true
						assert.GreaterOrEqual(t, actualMatch.Score, expectedMatch.MinScore,
							"Test case %s: member %d score %f should be >= %f",
							testCase.Name, expectedMatch.MemberID, actualMatch.Score, expectedMatch.MinScore)
						assert.LessOrEqual(t, actualMatch.Score, expectedMatch.MaxScore,
							"Test case %s: member %d score %f should be <= %f",
							testCase.Name, expectedMatch.MemberID, actualMatch.Score, expectedMatch.MaxScore)
						break
					}
				}
				assert.True(t, foundMatch,
					"Test case %s: expected to find member %d in results",
					testCase.Name, expectedMatch.MemberID)
			}

			t.Logf("Test case %s: found=%t, matches=%d", testCase.Name, found, len(matches))
			for i, match := range matches {
				t.Logf("  Match %d: ID=%d, Name=%s %s, Score=%.6f",
					i, match.Entry.ID, match.Entry.Firstname, match.Entry.Surname, match.Score)
			}
		})
	}
}

// Helper functions for loading test data
type FuzzySearchTestData struct {
	TestCases []FuzzySearchTestCase `json:"test_cases"`
}

type FuzzySearchTestCase struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Query       FuzzySearchQuery           `json:"query"`
	Expected    FuzzySearchExpectedResults `json:"expected"`
}

type FuzzySearchQuery struct {
	Firstname     string `json:"firstname"`
	Surname       string `json:"surname"`
	Birthdate     string `json:"birthdate"`
	EventStartUtc string `json:"event_start_utc,omitempty"`
}

type FuzzySearchExpectedResults struct {
	ShouldFind      bool                       `json:"should_find"`
	MinMatches      int                        `json:"min_matches"`
	ExpectedMatches []FuzzySearchExpectedMatch `json:"expected_matches"`
	Note            string                     `json:"note,omitempty"`
}

type FuzzySearchExpectedMatch struct {
	MemberID int     `json:"member_id"`
	MinScore float64 `json:"min_score"`
	MaxScore float64 `json:"max_score"`
}

func loadFuzzySearchTestCases(t *testing.T) FuzzySearchTestData {
	data, err := os.ReadFile("test_data/fuzzy_search_cases.json")
	require.NoError(t, err, "Failed to read fuzzy search test cases")

	var testData FuzzySearchTestData
	err = json.Unmarshal(data, &testData)
	require.NoError(t, err, "Failed to unmarshal fuzzy search test cases")

	return testData
}

func loadWaveMembersTestData(t *testing.T) []fc.WaveMembershipSource {
	data, err := os.ReadFile("test_data/wave_members.json")
	require.NoError(t, err, "Failed to read wave members test data")

	var testData struct {
		Members []struct {
			ID             string  `json:"id"`
			TicketQuantity string  `json:"ticket_quantity"`
			Firstname      string  `json:"firstname"`
			Surname        string  `json:"surname"`
			Birthdate      string  `json:"birthdate"`
			Tag            string  `json:"tag"`
			DeletedAt      *string `json:"deleted_at"`
			EventStartUtc  string  `json:"event_start_utc"`
			EventEndUtc    string  `json:"event_end_utc"`
		} `json:"members"`
	}
	err = json.Unmarshal(data, &testData)
	require.NoError(t, err, "Failed to unmarshal wave members test data")

	members := make([]fc.WaveMembershipSource, len(testData.Members))
	for i, tm := range testData.Members {
		id, _ := strconv.Atoi(tm.ID)
		ticketQty, _ := strconv.Atoi(tm.TicketQuantity)
		birthdate, _ := time.Parse("2006-01-02", tm.Birthdate)
		eventStart, _ := time.Parse(time.RFC3339, tm.EventStartUtc)
		eventEnd, _ := time.Parse(time.RFC3339, tm.EventEndUtc)

		var deletedAt *time.Time
		if tm.DeletedAt != nil && *tm.DeletedAt != "" {
			parsed, err := time.Parse(time.RFC3339, *tm.DeletedAt)
			if err == nil {
				deletedAt = &parsed
			}
		}

		members[i] = fc.WaveMembershipSource{
			ID:             id,
			TicketQuantity: ticketQty,
			Firstname:      tm.Firstname,
			Surname:        tm.Surname,
			Birthdate:      birthdate,
			Tag:            tm.Tag,
			DeletedAt:      deletedAt,
			EventStartUtc:  eventStart,
			EventEndUtc:    eventEnd,
		}
	}

	return members
}

func TestFuzzyMatcherCore_FuzzySearch_EdgeCases(t *testing.T) {
	// Load edge case test data
	edgeCaseData := loadEdgeCaseTestData(t)
	members := loadWaveMembersTestData(t)

	// Create fuzzyMatcherCore with all test data
	params := ft.FuzzyMatcherCoreParameters[fc.WaveMembershipSource]{
		CorrectOcrMisreads: false,
		UseExpiration:      false,
		MaxEdits: 6,
	}

	fuzzyMatcherCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{
		CoreParams: params,
	}

	fuzzyMatcherCore.Build(members)

	// Run each edge case test
	for _, testCase := range edgeCaseData.TestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// Convert test case query to WaveMembershipSource
			birthdate, err := time.Parse("2006-01-02", testCase.Query.Birthdate)
			require.NoError(t, err, "Failed to parse birthdate for test case %s", testCase.Name)

			// Parse EventStartUtc if provided, otherwise default to future date
			var eventStartUtc time.Time
			if testCase.Query.EventStartUtc != "" {
				eventStartUtc, _ = time.Parse("2006-01-02", testCase.Query.EventStartUtc)
			} else {
				// Default to a future date if not provided
				eventStartUtc = time.Date(2999, 1, 1, 0, 0, 0, 0, time.UTC)
			}

			query := fc.WaveMembershipSource{
				ID:            999,
				Firstname:     testCase.Query.Firstname,
				Surname:       testCase.Query.Surname,
				Birthdate:     birthdate,
				EventStartUtc: eventStartUtc,
			}

			// Execute search
			found, matches := fuzzyMatcherCore.SearchFuzzy(query)

			// Validate expectations
			assert.Equal(t, testCase.Expected.ShouldFind, found,
				"Test case %s: expected should_find=%t, got %t",
				testCase.Name, testCase.Expected.ShouldFind, found)

			if testCase.Expected.MinMatches > 0 {
				assert.GreaterOrEqual(t, len(matches), testCase.Expected.MinMatches,
					"Test case %s: expected at least %d matches, got %d",
					testCase.Name, testCase.Expected.MinMatches, len(matches))
			} else {
				assert.Empty(t, matches, "Test case %s: expected no matches", testCase.Name)
			}

			// Log results for analysis
			t.Logf("Edge case %s: found=%t, matches=%d", testCase.Name, found, len(matches))
			if testCase.Expected.Note != "" {
				t.Logf("  Note: %s", testCase.Expected.Note)
			}
			for i, match := range matches {
				t.Logf("  Match %d: ID=%d, Name=%s %s, Score=%.6f",
					i, match.Entry.ID, match.Entry.Firstname, match.Entry.Surname, match.Score)
			}
		})
	}
}

func loadEdgeCaseTestData(t *testing.T) FuzzySearchTestData {
	data, err := os.ReadFile("test_data/edge_case_tests.json")
	require.NoError(t, err, "Failed to read edge case test data")

	var testData FuzzySearchTestData
	err = json.Unmarshal(data, &testData)
	require.NoError(t, err, "Failed to unmarshal edge case test data")

	return testData
}

func TestFuzzyMatcherCore_FuzzySearch_Nicknames(t *testing.T) {
	// Load nickname test data
	nicknameData := loadNicknameTestData(t)
	members := loadWaveMembersTestData(t)

	// Create fuzzyMatcherCore with all test data
	params := ft.FuzzyMatcherCoreParameters[fc.WaveMembershipSource]{
		CorrectOcrMisreads: false,
		UseExpiration:      false,
		MaxEdits: 6,
	}

	fuzzyMatcherCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{
		CoreParams: params,
	}

	fuzzyMatcherCore.Build(members)

	// Run each nickname test
	for _, testCase := range nicknameData.TestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// Convert test case query to WaveMembershipSource
			birthdate, err := time.Parse("2006-01-02", testCase.Query.Birthdate)
			require.NoError(t, err, "Failed to parse birthdate for test case %s", testCase.Name)

			query := fc.WaveMembershipSource{
				ID:            999,
				Firstname:     testCase.Query.Firstname,
				Surname:       testCase.Query.Surname,
				Birthdate:     birthdate,
				EventStartUtc: time.Date(2999, 1, 1, 0, 0, 0, 0, time.UTC),
			}

			// Execute search
			found, matches := fuzzyMatcherCore.SearchFuzzy(query)

			// Validate expectations
			assert.Equal(t, testCase.Expected.ShouldFind, found,
				"Test case %s: expected should_find=%t, got %t",
				testCase.Name, testCase.Expected.ShouldFind, found)

			if testCase.Expected.MinMatches > 0 {
				assert.GreaterOrEqual(t, len(matches), testCase.Expected.MinMatches,
					"Test case %s: expected at least %d matches, got %d",
					testCase.Name, testCase.Expected.MinMatches, len(matches))
			} else {
				assert.Empty(t, matches, "Test case %s: expected no matches", testCase.Name)
			}

			// Validate specific expected matches
			for _, expectedMatch := range testCase.Expected.ExpectedMatches {
				foundMatch := false
				for _, actualMatch := range matches {
					if actualMatch.Entry.ID == expectedMatch.MemberID {
						foundMatch = true
						assert.GreaterOrEqual(t, actualMatch.Score, expectedMatch.MinScore,
							"Test case %s: member %d score %f should be >= %f",
							testCase.Name, expectedMatch.MemberID, actualMatch.Score, expectedMatch.MinScore)
						assert.LessOrEqual(t, actualMatch.Score, expectedMatch.MaxScore,
							"Test case %s: member %d score %f should be <= %f",
							testCase.Name, expectedMatch.MemberID, actualMatch.Score, expectedMatch.MaxScore)
						break
					}
				}
				if len(testCase.Expected.ExpectedMatches) > 0 {
					assert.True(t, foundMatch,
						"Test case %s: expected to find member %d in results",
						testCase.Name, expectedMatch.MemberID)
				}
			}

			// Log results for analysis
			t.Logf("Nickname test %s: found=%t, matches=%d", testCase.Name, found, len(matches))
			if testCase.Expected.Note != "" {
				t.Logf("  Note: %s", testCase.Expected.Note)
			}
			for i, match := range matches {
				t.Logf("  Match %d: ID=%d, Name=%s %s, Score=%.6f",
					i, match.Entry.ID, match.Entry.Firstname, match.Entry.Surname, match.Score)
			}
		})
	}
}

func loadNicknameTestData(t *testing.T) FuzzySearchTestData {
	data, err := os.ReadFile("test_data/nickname_tests.json")
	require.NoError(t, err, "Failed to read nickname test data")

	var testData FuzzySearchTestData
	err = json.Unmarshal(data, &testData)
	require.NoError(t, err, "Failed to unmarshal nickname test data")

	return testData
}

func TestFuzzyMatcherCore_FuzzySearch_Legacy(t *testing.T) {
	// Load test data from JSON
	basicTests := loadBasicTestData(t)

	// Use only the first member for legacy test
	members := []fc.WaveMembershipSource{
		convertToWaveMember(basicTests.BasicTestMembers[0]),
	}

	
	params := ft.FuzzyMatcherCoreParameters[fc.WaveMembershipSource]{
		CorrectOcrMisreads: false,
		UseExpiration:      false,
		MaxEdits: 6,
	}

	fuzzyMatcherCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{
		CoreParams: params,
	}
	
	fuzzyMatcherCore.Build(members)

	// Test fuzzy search with typos using JSON query
	query := convertQueryToWaveMember(basicTests.TestQueries.FuzzyMatch)

	found, matches := fuzzyMatcherCore.SearchFuzzy(query)

	if found {
		assert.NotEmpty(t, matches, "Should return matches if found")
		// Check that scores are reasonable for fuzzy matches
		for _, match := range matches {
			if match.Entry.ID == 1 {
				assert.Greater(t, match.Score, 0.3, "Fuzzy match should have some score")
				assert.Less(t, match.Score, 1.0, "Fuzzy match should not be perfect")
			}
		}
	}
}

func TestFuzzyMatcherCore_EmptySearch(t *testing.T) {
	// Load test data from JSON
	basicTests := loadBasicTestData(t)

	// Test search on empty fuzzyMatcherCore
	params := ft.FuzzyMatcherCoreParameters[fc.WaveMembershipSource]{
		CorrectOcrMisreads: false,
		UseExpiration:      false,
		MaxEdits: 6,
	}

	fuzzyMatcherCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{
		Root: &ft.FuzzyMatcherNode{
			Children: make(map[rune]*ft.FuzzyMatcherNode),
		},
		CoreParams: params,
	}

	query := convertQueryToWaveMember(basicTests.TestQueries.EmptySearch)

	found, matches := fuzzyMatcherCore.SearchFuzzy(query)
	assert.False(t, found, "Should not find anything in empty fuzzyMatcherCore")
	assert.Empty(t, matches, "Should return no matches for empty fuzzyMatcherCore")
}

func TestFuzzyEntry_Creation(t *testing.T) {
	// Load test data from JSON
	basicTests := loadBasicTestData(t)

	key := make(map[ft.Field]string)
	for k, v := range basicTests.ValidationData.FuzzyEntryKeys {
		key[ft.Field(k)] = v
	}

	expiry, _ := time.Parse(time.RFC3339, basicTests.ValidationData.FuzzyEntryMetadata.Expiry)

	entry := &ft.FuzzyEntry{
		Key:    key,
		ID:     basicTests.ValidationData.FuzzyEntryMetadata.ID,
		Expiry: expiry,
	}

	assert.Equal(t, basicTests.ValidationData.FuzzyEntryMetadata.ID, entry.ID)
	assert.Equal(t, "john", entry.Key[ft.Field("firstname")])
	assert.Equal(t, "smith", entry.Key[ft.Field("surname")])
	assert.Equal(t, expiry, entry.Expiry)
}

func TestFuzzyMatcherParameters_Validation(t *testing.T) {
	// Load test data from JSON
	basicTests := loadBasicTestData(t)
	params := basicTests.ValidationData.Parameters

	// Test valid parameters
	validParams := ft.FuzzyMatcherParameters{
		MaxDepth:           make(map[ft.Field]int),
		MaxEdits:           make(map[ft.Field]int),
		Weights:            make(map[ft.Field]float64),
		CalculationMethods: make(map[ft.Field]ft.CalculationMethod),
		MinDistances:       make(map[ft.Field]float64),
	}

	// Convert JSON data to proper types
	for k, v := range params.MaxDepth {
		validParams.MaxDepth[ft.Field(k)] = v
	}
	for k, v := range params.MaxEdits {
		validParams.MaxEdits[ft.Field(k)] = v
	}
	for k, v := range params.Weights {
		validParams.Weights[ft.Field(k)] = v
	}
	for k, v := range params.CalculationMethods {
		// Convert string to CalculationMethod enum
		validParams.CalculationMethods[ft.Field(k)] = ft.CalculationMethod(v)
	}
	for k, v := range params.MinDistances {
		validParams.MinDistances[ft.Field(k)] = v
	}

	// Test weight sum
	totalWeight := 0.0
	for _, weight := range validParams.Weights {
		totalWeight += weight
	}
	assert.InDelta(t, 1.0, totalWeight, 0.001, "Weights should sum to 1.0")

	// Test that all fields are present in all maps
	fields := []ft.Field{ft.Field("firstname"), ft.Field("surname")}
	for _, field := range fields {
		assert.Contains(t, validParams.MaxDepth, field)
		assert.Contains(t, validParams.MaxEdits, field)
		assert.Contains(t, validParams.Weights, field)
		assert.Contains(t, validParams.CalculationMethods, field)
		assert.Contains(t, validParams.MinDistances, field)
	}
}

func BenchmarkFuzzyMatcherCore_CalculateSimilarity_JaroWinkler(b *testing.B) {
	fuzzyMatcherCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{}
	s1 := "john"
	s2 := "jon"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fuzzyMatcherCore.CalculateSimilarity(s1, s2, ft.JaroWinkler, 0.0)
	}
}

func BenchmarkFuzzyMatcherCore_CalculateSimilarity_Levenshtein(b *testing.B) {
	fuzzyMatcherCore := &fmc.FuzzyMatcherCore[fc.WaveMembershipSource]{}
	s1 := "hello"
	s2 := "hallo"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fuzzyMatcherCore.CalculateSimilarity(s1, s2, ft.Levenshtein, 0.0)
	}
}
