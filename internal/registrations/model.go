package registrations

import "time"

type Registration struct {
	ID                 int       `json:"id"`
	RegistrationNumber string    `json:"registration_number"`
	PatientID          int       `json:"patient_id"`
	PatientName        string    `json:"patient_name,omitempty"`
	PatientMRN         string    `json:"patient_mrn,omitempty"`
	DoctorID           int       `json:"doctor_id"`
	DoctorName         string    `json:"doctor_name,omitempty"`
	RegistrationDate   string    `json:"registration_date"`
	RegistrationType   string    `json:"registration_type"`
	Complaint          string    `json:"complaint"`
	Status             string    `json:"status"`
	Notes              string    `json:"notes"`
	CreatedBy          *int      `json:"created_by,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
