package fuzzymatcherclasses

import (
	"strings"
	"time"

	ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"
)

type WaveMembershipSource struct {
	ID             int        `json:"id"`
	TicketQuantity int        `json:"ticket_quantity"`
	Firstname      string     `json:"firstname"`
	Surname        string     `json:"surname"`
	Birthdate      time.Time  `json:"birthdate"`
	Tag            string     `json:"tag"`
	DeletedAt      *time.Time `json:"deleted_at"`
	EventStartUtc  time.Time  `json:"event_start_utc"`
	EventEndUtc    time.Time  `json:"event_end_utc"`
}

// Defines the number of edits and search depth allowed for each field
func (w WaveMembershipSource) GetSearchParameters() ft.FuzzyMatcherParameters {
	isValid := w.ValidateEntry()

	var maxDepth map[ft.Field]int
	var maxEdits map[ft.Field]int

	// If the entry isn't valid (name isn't long enough)
	// Look for exact match
	if !isValid {
		maxDepth = map[ft.Field]int{
			ft.Firstname: 0,
			ft.Surname:   0,
			ft.Birthdate: 0,
		}

		maxEdits = map[ft.Field]int{
			ft.Firstname: 0,
			ft.Surname:   0,
			ft.Birthdate: 0,
		}
		// Else fuzzy match
	} else {
		maxDepth = map[ft.Field]int{
			ft.Firstname: 6,
			ft.Surname:   2,
			ft.Birthdate: 2,
		}

		maxEdits = map[ft.Field]int{
			ft.Firstname: 6,
			ft.Surname:   2,
			ft.Birthdate: 2,
		}
	}

	// Has to add up to 1.0
	weights := map[ft.Field]float64{
		ft.Firstname: 0.2,
		ft.Surname:   0.4,
		ft.Birthdate: 0.4,
	}

	calculationMethods := map[ft.Field]ft.CalculationMethod{
		ft.Firstname: ft.JaroWinkler,
		ft.Surname:   ft.JaroWinkler,
		ft.Birthdate: ft.Default,
	}

	minDistances := map[ft.Field]float64{
		ft.Firstname: 0.7,
		ft.Surname:   0.9,
		ft.Birthdate: 1,
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
func (w WaveMembershipSource) ValidateEntry() bool {
	firstName := strings.ToLower(strings.TrimSpace(w.Firstname))
	surname := strings.ToLower(strings.TrimSpace(w.Surname))
	birthdate := w.Birthdate.Format("20060102")

	if firstName == "" || surname == "" || birthdate == "" {
		return false
	}

	averageLength := float64(len(firstName)+len(surname)) / 2.0

	// Check if the average length of the first name and surname is greater than 3
	if averageLength <= 3.5 {
		return false
	}

	return true
}

// converts the WaveMembershipSource to a FuzzyEntry
func (w WaveMembershipSource) CreateFuzzyEntry() *ft.FuzzyEntry {
	// Formats the string to be used for fuzzy matching
	if w.DeletedAt != nil {
		return nil
	}

	firstName := strings.ToLower(strings.TrimSpace(w.Firstname))
	surname := strings.ToLower(strings.TrimSpace(w.Surname))
	birthdate := w.Birthdate.Format("20060102")

	key := make(map[ft.Field]string)
	key[ft.Firstname] = firstName
	key[ft.Surname] = surname
	key[ft.Birthdate] = birthdate

	return &ft.FuzzyEntry{
		Key:    key,
		ID:     w.ID,
		Expiry: w.EventEndUtc.Add(12 * time.Hour), // Set expiry to 12 hours after the event end time
	}
}
