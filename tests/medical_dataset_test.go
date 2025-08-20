package fuzzymatchertests

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	fm "github.com/oiamo123/fuzzy_matcher"
	fc "github.com/oiamo123/fuzzy_matcher/fuzzy_classes"
	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

// MedicalDataset represents the structure of the JSON dataset
type MedicalDataset struct {
	Elements []MedicalDataElement `json:"elements"`
}

type MedicalDataElement struct {
	ID       int                  `json:"el"`
	FromData map[string]interface{} `json:"fr"`
	ToData   []ToElement          `json:"to"`
	Weight   float64              `json:"w"` // Match confidence
}

type ToElement struct {
	ID     int                  `json:"el"`
	ToData map[string]interface{} `json:"to"`
}

// convertToBenchmarkSource converts a raw data map to BenchmarkSource
func convertToBenchmarkSource(data map[string]interface{}, id int) *fc.BenchmarkSource {
	source := &fc.BenchmarkSource{
		ID: id,
	}

	// Extract fields from the map, handling different types
	if name, ok := data["Name"].(string); ok {
		source.Name = name
	}
	if doctor, ok := data["Doctor"].(string); ok {
		source.Doctor = doctor
	}
	if gender, ok := data["Gender"].(string); ok {
		source.Gender = gender
	}
	if hospital, ok := data["Hospital"].(string); ok {
		source.Hospital = hospital
	}
	if dateOfAdm, ok := data["Date of Admission"].(string); ok {
		source.DateOfAdmission = dateOfAdm
	}
	if medCondition, ok := data["Medical Condition"].(string); ok {
		source.MedicalCondition = medCondition
	}
	if admissionType, ok := data["Admission Type"].(string); ok {
		source.AdmissionType = admissionType
	}
	
	// Handle numeric fields
	if age, ok := data["Age"].(float64); ok {
		source.Age = int(age)
	}
	if billingAmount, ok := data["Billing Amount"].(float64); ok {
		source.BillingAmount = billingAmount
	}
	if roomNumber, ok := data["Room Number"].(float64); ok {
		source.RoomNumber = int(roomNumber)
	}
	
	// Handle remaining string fields
	if bloodType, ok := data["Blood Type"].(string); ok {
		source.BloodType = bloodType
	}
	if dischargeDate, ok := data["Discharge Date"].(string); ok {
		source.DischargeDate = dischargeDate
	}
	if insProvider, ok := data["Insurance Provider"].(string); ok {
		source.InsuranceProvider = insProvider
	}
	if medication, ok := data["Medication"].(string); ok {
		source.Medication = medication
	}
	if testResults, ok := data["Test Results"].(string); ok {
		source.TestResults = testResults
	}
	
	return source
}

func TestMedicalDatasetBenchmark(t *testing.T) {
	// Configure the match threshold
	matchThreshold := 0.7 // If weight >= 0.7, it should match

	// Load dataset - adjust the path to your file
	jsonData, err := os.ReadFile("test_data/pretty_dd_hcds.json") 
	if err != nil {
		t.Fatalf("Failed to load medical dataset: %v", err)
	}

	var dataset []MedicalDataElement
	err = json.Unmarshal(jsonData, &dataset)
	if err != nil {
		t.Fatalf("Failed to parse dataset: %v", err)
	}
	
	t.Logf("Loaded dataset with %d elements", len(dataset))

	// Initialize the fuzzy matcher
	matcher := &fm.FuzzyMatcher[*fc.BenchmarkSource]{}
	matcher.Init(&ft.FuzzyMatcherCoreParameters[*fc.BenchmarkSource]{
		CorrectOcrMisreads: true,
		UseExpiration:      false,
		MaxEdits:           2, // Adjusted based on your parameters
	})

	// We need to avoid inserting the "from" entries, only the "to" entries
	// so we don't get false positives when searching for exact matches
	var toEntries []*fc.BenchmarkSource
	
	// Track IDs we insert to avoid duplicates
	insertedIDs := make(map[int]bool)
	
	// Insert only the "to" entries (destination records)
	for _, element := range dataset {
		for _, toElement := range element.ToData {
			// Skip if we've already inserted this ID
			if insertedIDs[toElement.ID] {
				continue
			}
			
			benchmarkSource := convertToBenchmarkSource(toElement.ToData, toElement.ID)
			toEntries = append(toEntries, benchmarkSource)
			insertedIDs[toElement.ID] = true
		}
	}
	
	// Insert entries into the fuzzy matcher
	err = matcher.InsertEntries(toEntries)
	if err != nil {
		t.Fatalf("Failed to insert entries: %v", err)
	}
	t.Logf("Inserted %d unique 'to' entries into fuzzy matcher", len(toEntries))

	// Test results tracking
	results := struct {
		TotalTests      int
		TruePositives   int
		FalsePositives  int
		TrueNegatives   int
		FalseNegatives  int
		TotalSearchTime time.Duration
	}{
		TotalTests: len(dataset),
	}

	// Process each "from" entry and compare with expected matches
	for _, element := range dataset {
		// Convert from entry
		fromEntry := convertToBenchmarkSource(element.FromData, element.ID)
		
		// Determine expected match status
		expectedMatch := element.Weight >= matchThreshold
		
		// Search
		startTime := time.Now()
		found, matches := matcher.Search(fromEntry)
		searchTime := time.Since(startTime)
		results.TotalSearchTime += searchTime

		// Check if any of the matches have the same ID as the original entry
		// (this would be a trivial match)
		hasNonTrivialMatch := false
		
		if found {
			for _, match := range matches {
				// We only care if there's a non-trivial match (different ID)
				if match.Entry.ID != element.ID {
					hasNonTrivialMatch = true
					break
				}
			}
		}
		
		// Analyze results - only count as false positive if it found a non-trivial match
		// that should not have matched based on weight
		if expectedMatch && hasNonTrivialMatch {
			results.TruePositives++
			if testing.Verbose() {
				t.Logf("True Positive: ID %d (w=%.2f) matched as expected", element.ID, element.Weight)
				t.Logf("  From: Name=%s, Hospital=%s, DateOfAdm=%s", 
					fromEntry.Name, fromEntry.Hospital, fromEntry.DateOfAdmission)
				for i, match := range matches {
					if i < 3 && match.Entry.ID != element.ID { // Show up to 3 non-trivial matches
						t.Logf("  Match %d: ID=%d, Name=%s, Hospital=%s, DateOfAdm=%s (Score: %.4f)", 
							i+1, match.Entry.ID, match.Entry.Name, match.Entry.Hospital, 
							match.Entry.DateOfAdmission, match.Score)
					}
				}
			}
		} else if !expectedMatch && hasNonTrivialMatch {
			// Only count as false positive if a non-trivial match is found
			results.FalsePositives++
			t.Logf("False Positive: ID %d (w=%.2f) should not match but did", element.ID, element.Weight)
			t.Logf("  From: Name=%s, Hospital=%s, DateOfAdm=%s", 
				fromEntry.Name, fromEntry.Hospital, fromEntry.DateOfAdmission)
			for i, match := range matches {
				if i < 3 && match.Entry.ID != element.ID { // Show up to 3 non-trivial matches
					t.Logf("  Match %d: ID=%d, Name=%s, Hospital=%s, DateOfAdm=%s (Score: %.4f)", 
						i+1, match.Entry.ID, match.Entry.Name, match.Entry.Hospital, 
						match.Entry.DateOfAdmission, match.Score)
				}
			}
		} else if !expectedMatch && !hasNonTrivialMatch {
			results.TrueNegatives++
			if testing.Verbose() {
				t.Logf("True Negative: ID %d (w=%.2f) correctly did not match", element.ID, element.Weight)
			}
		} else if expectedMatch && !hasNonTrivialMatch {
			results.FalseNegatives++
			t.Logf("False Negative: ID %d (w=%.2f) should match but didn't", element.ID, element.Weight)
			t.Logf("  From: Name=%s, Hospital=%s, DateOfAdm=%s", 
				fromEntry.Name, fromEntry.Hospital, fromEntry.DateOfAdmission)
		}
	}

	// Calculate metrics
	precision := float64(0)
	if results.TruePositives+results.FalsePositives > 0 {
		precision = float64(results.TruePositives) / float64(results.TruePositives+results.FalsePositives)
	}
	
	recall := float64(0)
	if results.TruePositives+results.FalseNegatives > 0 {
		recall = float64(results.TruePositives) / float64(results.TruePositives+results.FalseNegatives)
	}
	
	f1 := float64(0)
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}
	
	accuracy := float64(0)
	if results.TotalTests > 0 {
		accuracy = float64(results.TruePositives+results.TrueNegatives) / float64(results.TotalTests)
	}
	
	avgSearchTime := results.TotalSearchTime / time.Duration(results.TotalTests)
	
	// Log results
	t.Logf("Medical Dataset Benchmark Results:")
	t.Logf("Total Elements: %d", results.TotalTests)
	t.Logf("True Positives: %d", results.TruePositives)
	t.Logf("False Positives: %d", results.FalsePositives)
	t.Logf("True Negatives: %d", results.TrueNegatives)
	t.Logf("False Negatives: %d", results.FalseNegatives)
	t.Logf("Average Search Time: %v", avgSearchTime)
	t.Logf("Precision: %.4f", precision)
	t.Logf("Recall: %.4f", recall)
	t.Logf("F1 Score: %.4f", f1)
	t.Logf("Accuracy: %.4f", accuracy)
}

// BenchmarkMedicalDataset performs benchmark tests on the medical dataset
func BenchmarkMedicalDataset(b *testing.B) {
	// Load dataset - adjust the path to your file
	jsonData, err := os.ReadFile("test_data/medical_dataset.json")
	if err != nil {
		b.Fatalf("Failed to load medical dataset: %v", err)
	}

	var dataset []MedicalDataElement
	err = json.Unmarshal(jsonData, &dataset)
	if err != nil {
		b.Fatalf("Failed to parse dataset: %v", err)
	}
	
	// Limit dataset size for benchmarking if needed
	maxDatasetSize := 1000
	if len(dataset) > maxDatasetSize {
		dataset = dataset[:maxDatasetSize]
	}
	
	// Initialize the fuzzy matcher
	matcher := &fm.FuzzyMatcher[*fc.BenchmarkSource]{}
	matcher.Init(&ft.FuzzyMatcherCoreParameters[*fc.BenchmarkSource]{
		CorrectOcrMisreads: true,
		UseExpiration:      false,
		MaxEdits:           2,
	})

	// Insert only the "to" entries
	var toEntries []*fc.BenchmarkSource
	insertedIDs := make(map[int]bool)
	
	for _, element := range dataset {
		for _, toElement := range element.ToData {
			// Skip if we've already inserted this ID
			if insertedIDs[toElement.ID] {
				continue
			}
			
			benchmarkSource := convertToBenchmarkSource(toElement.ToData, toElement.ID)
			toEntries = append(toEntries, benchmarkSource)
			insertedIDs[toElement.ID] = true
		}
	}
	
	err = matcher.InsertEntries(toEntries)
	if err != nil {
		b.Fatalf("Failed to insert entries: %v", err)
	}
	b.Logf("Inserted %d unique 'to' entries into fuzzy matcher", len(toEntries))

	// Prepare query entries (use a subset for benchmarking)
	var queryEntries []*fc.BenchmarkSource
	for i, element := range dataset {
		if i >= 100 { // Only use first 100 entries for queries
			break
		}
		// Convert the "from" entry for querying
		fromEntry := convertToBenchmarkSource(element.FromData, element.ID)
		queryEntries = append(queryEntries, fromEntry)
	}
	
	b.ResetTimer()
	
	// Run benchmark
	for i := 0; i < b.N; i++ {
		// Use modulo to cycle through test entries
		queryEntry := queryEntries[i%len(queryEntries)]
		_, _ = matcher.Search(queryEntry)
	}
}
