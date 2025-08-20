# Fuzzy Matcher

A high-performance, field-based fuzzy string matcher for multi-criteria text matching with generic support. Achieve superior matching accuracy with customizable field parameters.

## Performance Highlights

Benchmarked against a [healthcare dataset with 55,500 records](https://www.kaggle.com/datasets/aahz78/dublicate-in-healthcare-dateset):

| Metric                  | Value  |
| ----------------------- | ------ |
| **Precision**           | 99.67% |
| **Recall**              | 100%   |
| **F1 Score**            | 99.84% |
| **Accuracy**            | 99.94% |
| **Average Search Time** | 1.89ms |

These results were achieved using only three fields: patient name, hospital name, and date of admission. The implementation outperforms traditional database LIKE queries while offering intelligent fuzzy matching capabilities.

## Why Use Fuzzy Matcher?

- **Superior to Database LIKE Queries**: Replace expensive and inflexible database searches with high-performance fuzzy matching
- **Field-Specific Customization**: Set different matching rules for each field
- **Intelligent Scoring**: Results include confidence scores rather than binary matches
- **Fast Performance**: Sub-2ms search times even on large datasets
- **Generic Support**: Works with any data structure using Go generics

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
            Fields.Name:  m.Name,  // Case sensitivity is not preserved
            Fields.Email: m.Email,
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
            Fields.Name:  3,  // Maximum edit distance for fuzzy matching
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
        CalculationMethods: map[ft.Field]ft.CalculationMethod{
            Fields.Name:  ft.JaroWinkler,  // Better for names
            Fields.Email: ft.Levenshtein,  // Better for structured data
            Fields.Phone: ft.Default,      // Default is exact matching
        },
    }
}
```

## Usage

````go
```go
// Initialize the matcher with your data type
matcher := &fuzzymatcher.FuzzyMatcher[MyData]{}
_ = matcher.Init(&ft.FuzzyMatcherCoreParameters[MyData]{
    CorrectOcrMisreads: true,  // Optional: Enable OCR correction
    UseExpiration:      false,
    MaxEdits:           3,      // Global maximum edit distance
})

// Build the matcher with your data
var data []MyData = LoadData()
matcher.InsertEntries(data)

// For search with expiry support
expiryMatcher := &fuzzymatcher.FuzzyMatcher[MyDataWithExpiry]{}
expiryMatcher.Init(&ft.FuzzyMatcherCoreParameters[MyDataWithExpiry]{
    UseExpiration: true,
})
expiryMatcher.InsertEntries(expiryData)

// Search for matches
query := MyData{Name: "John Smth", Email: "john@example.com"}
found, matches := matcher.Search(query)
if found {
    for _, match := range matches {
        fmt.Printf("Match: %+v, Score: %.3f\n", match.Entry, match.Score)
    }
}
````

## Features

- Generic support for any data source
- Multi-field fuzzy matching with configurable parameters
- Field weighting and scoring
- Optional expiration support
- OCR error correction support
- Case-sensitive matching
- High performance with large datasets
- Short name validation

## Running Tests

```bash
go test ./tests/... -v
```

---

## Real-World Applications

This fuzzy matcher is ideal for:

- **Healthcare Record Linkage**: Find duplicate patient records despite data entry variations
- **Customer Data Integration**: Match customer records across disparate systems
- **Fraud Detection**: Identify accounts with suspiciously similar details
- **Data Cleansing**: Find and merge duplicate entities in databases
- **Search Engines**: Power "did you mean" and suggestion features

## Benchmark Details

The fuzzy matcher was tested against a healthcare dataset containing 55,500 patient records with fields including names, admission dates, hospitals, medical conditions, and more.

### Test Configuration

For the benchmark test, we used these fields with the following parameters:

```go
func (s BenchmarkSource) CreateFuzzyEntry() *ft.FuzzyEntry {
    return &ft.FuzzyEntry{
        ID: s.ID,
        Key: map[ft.Field]string{
            Name: s.Name,
            DateOfAdmission: s.DateOfAdmission,
            Hospital: s.Hospital,
        },
    }
}

func (s BenchmarkSource) GetSearchParameters() ft.FuzzyMatcherParameters {
    maxDepth := map[ft.Field]int{
        Name: 2,
        DateOfAdmission: 0,
        Hospital: 0,
    }

    maxEdits := map[ft.Field]int{
        Name: 2,
        DateOfAdmission: 0,
        Hospital: 0,
    }

    weights := map[ft.Field]float64{
        Name: 0.2,
        DateOfAdmission: 0.5,
        Hospital: 0.3,
    }

    minDistances := map[ft.Field]float64{
        Name: 0.7,
        DateOfAdmission: 1,
        Hospital: 1,
    }

    calculationMethods := map[ft.Field]ft.CalculationMethod{
        Name: ft.JaroWinkler,
        DateOfAdmission: ft.Default,
        Hospital: ft.Default,
    }

    return ft.FuzzyMatcherParameters{
        Weights: weights,
        MinDistances: minDistances,
        CalculationMethods: calculationMethods,
        MaxDepth: maxDepth,
        MaxEdits: maxEdits,
    }
}
```

### Benchmark Results

```
Total Elements: 55,500
True Positives: 11,000
False Positives: 36
True Negatives: 44,464
False Negatives: 0
Average Search Time: 1.89ms
Precision: 99.67%
Recall: 100%
F1 Score: 99.84%
Accuracy: 99.94%
```

These results demonstrate the fuzzy matcher's exceptional performance in both accuracy and speed.

## Example Data Sources

The package includes several example data source implementations:

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

### BenchmarkSource

```go
type BenchmarkSource struct {
    ID             int
    Name           string
    Hospital       string
    DateOfAdmission string
    // Additional fields omitted for brevity
}
```

Use these as reference implementations or extend them for your own data sources.
