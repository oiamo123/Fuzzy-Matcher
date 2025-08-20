package fuzzyclasses

import ft "github.com/oiamo123/fuzzy_matcher/fuzzy_types"

const (
	AdmissionType ft.Field = "admission_type"
	Age           ft.Field = "age"
	BillingAmount ft.Field = "billing_amount"
	BloodType     ft.Field = "blood_type"
	DateOfAdmission ft.Field = "date_of_admission"
	DischargeDate ft.Field = "discharge_date"
	Doctor       ft.Field = "doctor"
	Gender       ft.Field = "gender"
	Hospital    ft.Field = "hospital"
	InsuranceProvider ft.Field = "insurance_provider"
	MedicalCondition ft.Field = "medical_condition"
	Medication   ft.Field = "medication"
	Name         ft.Field = "name"
	RoomNumber   ft.Field = "room_number"
	TestResults  ft.Field = "test_results"
)

type BenchmarkSource struct {
	ID int `json:"el"`
	AdmissionType string `json:"Admission Type"`
	Age int `json:"Age"`
	BillingAmount float64 `json:"Billing Amount"`
	BloodType string `json:"Blood Type"`
	DateOfAdmission string `json:"Date of Admission"`
	DischargeDate string `json:"Discharge Date"`
	Doctor string `json:"Doctor"`
	Gender string `json:"Gender"`
	Hospital string `json:"Hospital"`
	InsuranceProvider string `json:"Insurance Provider"`
	MedicalCondition string `json:"Medical Condition"`
	Medication string `json:"Medication"`
	Name string `json:"Name"`
	RoomNumber int `json:"Room Number"`
	TestResults string `json:"Test Results"`
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