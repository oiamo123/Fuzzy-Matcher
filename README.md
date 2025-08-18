# Fuzzy Matcher

A simple, field-based fuzzy string matcher for multi-criteria text matching.

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

### 2. Implement the Data Interface

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
        Expiry: time.Now().Add(24 * time.Hour),
    }
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
        MaxEdits: map[ft.Field]int{
            Fields.Name:  6,
            Fields.Email: 2,
            Fields.Phone: 3,
        },
        Weights: map[ft.Field]float64{
            Fields.Name:  0.3,
            Fields.Email: 0.5,
            Fields.Phone: 0.2,
        },
        MinDistances: map[ft.Field]float64{
            Fields.Name:  0.7,
            Fields.Email: 0.9,
            Fields.Phone: 1.0,
        },
    }
}
```

## Usage

```go
matcher := &fuzzymatcher.FuzzyMatcher[MyData]{}
_ = matcher.Init()
matcher.FuzzyMatcherCore.Build(data, true)

found, matches := matcher.Search(searchEntry)
if found {
    for _, match := range matches {
        fmt.Println(match.Entry, match.Score)
    }
}
```

## Running Tests

```bash
cd tests
go test -v
```

---

Supports multi-field fuzzy matching, scoring, and expiry out of the
