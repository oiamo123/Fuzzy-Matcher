# Fuzzy Matcher

A field-based fuzzy string matching system for multi-criteria text matching.

## Running Tests

```bash
cd tests
go test -v
```

## Implementing Your Own Fuzzy Matcher

### 1. Define Your Field Types

```go
// Define typed fields for compile-time safety
type MyDataFields struct {
    Name      ft.Field
    Email     ft.Field
    Phone     ft.Field
}

var Fields = MyDataFields{
    Name:  "name",
    Email: "email", 
    Phone: "phone",
}
```

### 2. Implement the Interface

Your data type must implement `FuzzyMatcherDataSource`:

```go
type MyDataType struct {
    ID    int
    Name  string
    Email string
    Phone string
}

// Convert your data to fuzzy entry format
func (m MyDataType) CreateFuzzyEntry() *ft.FuzzyEntry {
    key := make(map[ft.Field]string)
    key[Fields.Name] = strings.ToLower(strings.TrimSpace(m.Name))
    key[Fields.Email] = strings.ToLower(strings.TrimSpace(m.Email))
    key[Fields.Phone] = m.Phone
    
    return &ft.FuzzyEntry{
        Key:    key,
        ID:     m.ID,
        Expiry: time.Now().Add(24 * time.Hour), // Expires in 24 hours
    }
}

// Validate entry has required fields
func (m MyDataType) ValidateEntry() bool {
    return m.Name != "" && m.Email != ""
}

// Configure matching parameters per field
func (m MyDataType) GetSearchParameters() ft.FuzzyMatcherParameters {
    return ft.FuzzyMatcherParameters{
        MaxEdits: map[ft.Field]int{
            Fields.Name:  6,  // Allow more errors in names
            Fields.Email: 2,  // Be strict with emails
            Fields.Phone: 3,  // Moderate errors for phone
        },
        Weights: map[ft.Field]float64{
            Fields.Name:  0.3,  // 30% contribution to final score
            Fields.Email: 0.5,  // 50% contribution  
            Fields.Phone: 0.2,  // 20% contribution
        },
        CalculationMethods: map[ft.Field]string{
            Fields.Name:  string(ft.JaroWinkler),  // Good for names
            Fields.Email: string(ft.Levenshtein),  // General purpose
            Fields.Phone: string(ft.Default),      // Threshold-based
        },
        MinDistances: map[ft.Field]float64{
            Fields.Name:  0.7,  // 70% similarity required
            Fields.Email: 0.9,  // 90% similarity required
            Fields.Phone: 1.0,  // Exact or very close match
        },
    }
}
```

## Using the Fuzzy Matcher

### Initialize and Build

```go
// Create matcher
matcher := &fuzzymatcher.FuzzyMatcher[MyDataType]{
    MaxEdits:           6,
    CorrectOcrMisreads: true,
}

// Initialize
err := matcher.Init()
if err != nil {
    log.Fatal(err)
}

// Build from data (usually from API/database)
data := []MyDataType{ /* your data */ }
matcher.FuzzyMatcherCore.Build(data, true)
```

### Search for Matches

```go
// Create search entry
searchEntry := MyDataType{
    Name:  "John Smith",
    Email: "john@example.com", 
    Phone: "555-1234",
}

// Search for matches
found, matches := matcher.Search(searchEntry)
if found {
    for _, match := range matches {
        fmt.Printf("Match: %+v, Score: %.2f\n", match.Entry, match.Score)
    }
}
```

That's it! The system handles multi-field matching, scoring, and expiry automatically.
