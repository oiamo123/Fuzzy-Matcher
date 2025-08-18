package fuzzymatcherclasses

// THIS CLASS IS USED TO TEST REPEAT FAKES, IGNORE FOR NOW

import (
	"strconv"
	"strings"
	"time"

	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

type FakePatronSource struct {
	Birthdate      string `json:"birthdate"`
	Firstname      string `json:"firstname"`
	Middlename     string `json:"middlename"`
	ID             int    `json:"pk_patron_id"`
	Surname        string `json:"surname"`
	InsertDatetime string `json:"insert_datetime_utc"`
	CustomerId     int    `json:"fk_company_id"`
}

// Defines the number of edits and search depth allowed for each field
func (w FakePatronSource) GetSearchParameters() ft.FuzzyMatcherParameters {
	maxDepth := map[ft.Field]int{
		ft.Firstname:  2,
		ft.Middlename: 2,
		ft.Surname:    2,
		ft.Birthdate:  5,
		ft.CustomerId: 0,
	}

	maxEdits := map[ft.Field]int{
		ft.Firstname:  2,
		ft.Middlename: 2,
		ft.Surname:    2,
		ft.Birthdate:  5,
		ft.CustomerId: 0,
	}

	// Has to add up to 1.0
	weights := map[ft.Field]float64{
		ft.Firstname:  0.30,
		ft.Middlename: 0.00,
		ft.Surname:    0.30,
		ft.Birthdate:  0.4,
		ft.CustomerId: 0.0,
	}

	calculationMethods := map[ft.Field]ft.CalculationMethod{
		ft.Firstname:  ft.JaroWinkler,
		ft.Middlename: ft.JaroWinkler,
		ft.Surname:    ft.JaroWinkler,
		ft.Birthdate:  ft.Levenshtein,
		ft.CustomerId: ft.Default,
	}

	minDistances := map[ft.Field]float64{
		ft.Firstname:  0.85, // slightly looser than 0.8
		ft.Middlename: 0.0,
		ft.Surname:    0.85,
		ft.Birthdate:  0.6,
		ft.CustomerId: 1.0,
	}

	return ft.FuzzyMatcherParameters{
		MaxDepth:           maxDepth,
		MaxEdits:           maxEdits,
		Weights:            weights,
		CalculationMethods: calculationMethods,
		MinDistances:       minDistances,
	}
}

// Validates the entry by checking if the required fields are present and have valid values
func (w FakePatronSource) ValidateEntry() bool {
	firstName := strings.ToLower(strings.TrimSpace(w.Firstname))
	middlename := strings.ToLower(strings.TrimSpace(w.Middlename))
	surname := strings.ToLower(strings.TrimSpace(w.Surname))
	birthdate := strings.ToLower(strings.TrimSpace(w.Birthdate))

	if firstName == "" || surname == "" || birthdate == "" {
		return false
	}

	averageLength := float32(len(firstName)+len(surname)+len(middlename)) / 3

	// Check if the average length of the first name and surname is greater than 3
	if averageLength <= 3.5 {
		return false
	}

	return true
}

// converts the FakePatronSource to a FuzzyEntry
func (p FakePatronSource) CreateFuzzyEntry() *ft.FuzzyEntry {
	firstName := strings.ToLower(strings.TrimSpace(strings.SplitN(p.Firstname, " ", 2)[0]))
	middlename := strings.ToLower(strings.TrimSpace(strings.SplitN(p.Middlename, " ", 2)[0]))
	surname := strings.ToLower(strings.TrimSpace(strings.SplitN(p.Surname, " ", 2)[0]))
	birthdate := strings.ToLower(strings.TrimSpace(p.Birthdate))
	customerId := strings.ToLower(strings.TrimSpace(strconv.Itoa(p.CustomerId)))

	key := make(map[ft.Field]string)
	key[ft.Firstname] = firstName
	key[ft.Middlename] = middlename
	key[ft.Surname] = surname
	key[ft.Birthdate] = birthdate
	key[ft.CustomerId] = customerId

	return &ft.FuzzyEntry{
		Key:    key,
		ID:     p.ID,
		Expiry: time.Now().Add(24 * time.Hour),
	}
}
