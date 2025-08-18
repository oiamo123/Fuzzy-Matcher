package fuzzymatchertypes

import (
	"time"
)

type CalculationMethod string

type Field string

const (
	JaroWinkler CalculationMethod = "jaro"
	Levenshtein CalculationMethod = "levenshtein"
	Default     CalculationMethod = ""
)

const (
	Firstname  Field = "firstname"
	Middlename Field = "middlename"
	Surname    Field = "surname"
	Birthdate  Field = "birthdate"
	CustomerId Field = "customer_id"
)

type FuzzyEntry struct {
	Key    map[Field]string // Key for the entry, e.g. {"firstname": "John", "surname": "Doe"}
	ID     int              // Unique identifier for the entry
	Expiry time.Time        // Expiry time for the entry
}

type FuzzyMatch[T FuzzyMatcherDataSource] struct {
	Entry T
	Score float64 // Score of the match
}

// Defines the search parameters for fuzzy matching
type FuzzyMatcherParameters struct {
	MaxDepth           map[Field]int               // Maximum search depth for each field
	MaxEdits           map[Field]int               // Maximum number of edits allowed for each field
	Weights            map[Field]float64           // Minimum distance for each field
	CalculationMethods map[Field]CalculationMethod // Calculation method for each field, e.g. "jaro", "levenshtein"
	MinDistances       map[Field]float64           // Minimum distance for each field
}

type FuzzyMatcherCoreParameters[T FuzzyMatcherDataSource] struct {
	CorrectOcrMisreads bool
	MaxEdits           int
	UseExpiration      bool
}

type MatchCandidate struct {
	Text        string
	EditCount   int
	SearchDepth int
	ID          []int
}

// ApiResponse is a generic structure for API responses that returns the FuzzyMatcherDataSource type
type ApiResponse[T FuzzyMatcherDataSource] struct {
	Success bool `json:"success"`
	Data    []T  `json:"data"`
}

// Defines required methods for data sources that can be used with FuzzyMatcher
type FuzzyMatcherDataSource interface {
	CreateFuzzyEntry() *FuzzyEntry               // Converts the data source to a FuzzyEntry
	GetSearchParameters() FuzzyMatcherParameters // Returns search restrictions for the entry
}
