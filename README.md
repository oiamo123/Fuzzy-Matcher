# Fuzzy Matcher

A simple, field-based fuzzy string matcher for multi-criteria text matching with generic support.

## Quick Start

### 1. Define Your Fields

```go
type MyFields struct {
    Name  ft.Field
    Email ft.Field
    Phone ft.Field
}

var Fields = MyFields{
    Name:  "name",
    Email: "email",
    Phone: "phone",
}
```

### 2. Implement the Generic Data Source Interface

```go
type MyData struct {
    ID    int
    Name  string
    Email string
    Phone string
}

func (m MyData) CreateFuzzyEntry() *ft.FuzzyEntry {
    return &ft.FuzzyEntry{
        Key: map[ft.Field]string{
            Fields.Name:  strings.ToLower(m.Name),
            Fields.Email: strings.ToLower(m.Email),
            Fields.Phone: m.Phone,
        },
        ID: m.ID,
    }
}

// Use ExampleWithExpirySource if you need expiration support
type MyDataWithExpiry struct {
    MyData
    ExpiryTime time.Time
}

func (m MyDataWithExpiry) CreateFuzzyEntry() *ft.FuzzyEntry {
    entry := m.MyData.CreateFuzzyEntry()
    entry.Expiry = m.ExpiryTime
    return entry
}

func (m MyData) ValidateEntry() bool {
    return m.Name != "" && m.Email != ""
}
```

### 3. (Optional) Custom Matching Parameters

You can customize matching per field:

```go
func (m MyData) GetSearchParameters() ft.FuzzyMatcherParameters {
    return ft.FuzzyMatcherParameters{
        MaxDepth: map[ft.Field]int{
            Fields.Name:  6,  // Maximum recursion depth for fuzzy matching
            Fields.Email: 2,
            Fields.Phone: 0,  // 0 means exact matching only
        },
        MaxEdits: map[ft.Field]int{
            Fields.Name:  3,  // Maximum number of edits for fuzzy matching
            Fields.Email: 2,
            Fields.Phone: 0,  // 0 means exact matching only
        },
        Weights: map[ft.Field]float64{
            Fields.Name:  0.3,  // Field weights for scoring (must sum to 1.0)
            Fields.Email: 0.5,
            Fields.Phone: 0.2,
        },
        MinDistances: map[ft.Field]float64{
            Fields.Name:  0.7,  // Minimum similarity threshold (0.0-1.0)
            Fields.Email: 0.9,
            Fields.Phone: 1.0,  // 1.0 means exact match required
        },
    }
}
```

## Usage

```go
// Initialize the matcher with your data type
matcher := &fuzzymatcher.FuzzyMatcher[MyData]{}
_ = matcher.Init()

// Build the matcher with your data
var data []MyData = LoadData()
matcher.FuzzyMatcherCore.Build(data)

// For search with expiry support
expiryMatcher := &fuzzymatcher.FuzzyMatcher[MyDataWithExpiry]{}
expiryMatcher.Init()
expiryMatcher.CoreParams.UseExpiration = true
expiryMatcher.FuzzyMatcherCore.Build(expiryData)

// Search for matches
query := MyData{Name: "John Smth", Email: "john@example.com"}
found, matches := matcher.Search(query)
if found {
    for _, match := range matches {
        fmt.Printf("Match: %+v, Score: %.3f\n", match.Entry, match.Score)
    }
}
```

## Features

- Generic support for any data source
- Multi-field fuzzy matching with configurable parameters
- Field weighting and scoring
- Optional expiration support
- OCR error correction support
- Short name validation

## Running Tests

```bash
go test ./tests/... -v
```

---

## Example Data Sources

The package comes with two example data source implementations:

### ExampleSource

```go
type ExampleSource struct {
    ID        int
    Firstname string
    Surname   string
    Birthdate time.Time
}
```

### ExampleWithExpirySource

```go
type ExampleWithExpirySource struct {
    ID            int
    Firstname     string
    Surname       string
    Birthdate     time.Time
    EventStartUtc time.Time
    EventEndUtc   time.Time
}
```

Use these as reference implementations or extend them for your own data sources.
