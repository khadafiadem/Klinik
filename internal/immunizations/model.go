package immunizations

import "time"

type Schedule struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	AgeLabel  string    `json:"age_label"`
	MonthFrom int       `json:"month_from"`
	MonthTo   *int      `json:"month_to"`
	Category  string    `json:"category"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Immunization struct {
	ID              int       `json:"id"`
	PatientID       int       `json:"patient_id"`
	PatientName     string    `json:"patient_name,omitempty"`
	PatientMRN      string    `json:"patient_mrn,omitempty"`
	ImmunizationID  *int      `json:"immunization_id"`
	VaccinationDate string    `json:"vaccination_date"`
	VaccineName     string    `json:"vaccine_name"`
	BatchNumber     string    `json:"batch_number"`
	ProviderName    string    `json:"provider_name"`
	Notes           string    `json:"notes"`
	CreatedBy       *int      `json:"created_by,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
