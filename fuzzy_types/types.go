package fuzzymatchertypes

import (
	"time"
)

type Field string
type CalculationMethod string

// Calculation methods
const (
    JaroWinkler CalculationMethod = "jaro"
    Levenshtein CalculationMethod = "levenshtein"
    Default     CalculationMethod = ""
)

// Common field types
const (
    Firstname  Field = "firstname"
    Middlename Field = "middlename"
    Surname    Field = "surname"
    Birthdate  Field = "birthdate"
    CustomerId Field = "customer_id"
)

// FuzzyEntry represents a single entry in the fuzzy matcher
type FuzzyEntry struct {
    Key    map[Field]string // Key for the entry, e.g. {"firstname": "John", "surname": "Doe"}
    ID     int              // Unique identifier for the entry
    Expiry time.Time        // Expiry time for the entry
}

// FuzzyMatcherNode represents a node in the FuzzyMatcher trie structure
type FuzzyMatcherNode struct {
    Char          rune
    Children      map[rune]*FuzzyMatcherNode
    IsEndofString bool
    ID            map[int]bool
    Parent        *FuzzyMatcherNode
    Count         int
}

// FuzzyMatch represents a match result with score
type FuzzyMatch[T FuzzyMatcherDataSource] struct {
    Entry T
    Score float64
}

// FuzzyMatcherParameters defines the search parameters for fuzzy matching
type FuzzyMatcherParameters struct {
    MaxDepth           map[Field]int               // Maximum search depth for each field
    MaxEdits           map[Field]int               // Maximum number of edits allowed for each field
    Weights            map[Field]float64           // Weights for each field
    CalculationMethods map[Field]CalculationMethod // Calculation method for each field
    MinDistances       map[Field]float64           // Minimum distance for each field
}

// FuzzyMatcherCoreParameters defines core behavior of the fuzzy matcher
type FuzzyMatcherCoreParameters[T FuzzyMatcherDataSource] struct {
    CorrectOcrMisreads bool
    MaxEdits           int
    UseExpiration      bool
}

// VisitKey is a key to identify visited nodes during recursion
type VisitKey uint64

// MatchCandidate represents a potential match during search
type MatchCandidate struct {
    Text        string
    EditCount   int
    SearchDepth int
    ID          []int
}

type ExpiryEntry struct {
	Expiry time.Time
	Node   *FuzzyMatcherNode
	ID     int
}

// BreadthFirstSearchNode represents a node in the BFS queue
type NodePriority struct {
    Params RecurseParameters
    Score  float64
}

// FieldResult represents the result of searching a specific field
type FieldResult struct {
    Key     Field
    Matches []MatchCandidate
    Err     error
}

// ApiResponse is a generic structure for API responses
type ApiResponse[T FuzzyMatcherDataSource] struct {
    Success bool `json:"success"`
    Data    []T  `json:"data"`
}

// FuzzyMatcherDataSource defines required methods for data sources
type FuzzyMatcherDataSource interface {
    CreateFuzzyEntry() *FuzzyEntry               // Converts the data source to a FuzzyEntry
    GetSearchParameters() FuzzyMatcherParameters // Returns search restrictions for the entry
}