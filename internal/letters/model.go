package letters

import "time"

type Template struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	Subject      string    `json:"subject"`
	BodyTemplate string    `json:"body_template"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Issuance struct {
	ID           int       `json:"id"`
	TemplateID   int       `json:"template_id"`
	LetterNumber string    `json:"letter_number"`
	PatientID    int       `json:"patient_id"`
	DoctorID     *int      `json:"doctor_id"`
	Notes        string    `json:"notes"`
	IssuedBy     *int      `json:"issued_by"`
	CreatedAt    time.Time `json:"created_at"`

	PatientName  string `json:"patient_name,omitempty"`
	PatientMRN   string `json:"patient_mrn,omitempty"`
	DoctorName   string `json:"doctor_name,omitempty"`
	TemplateName string `json:"template_name,omitempty"`
	TemplateCode string `json:"template_code,omitempty"`
	Subject      string `json:"subject,omitempty"`
	BodyTemplate string `json:"body_template,omitempty"`
}
