package queues

import "time"

type Queue struct {
	ID               int        `json:"id"`
	QueueNumber      string     `json:"queue_number"`
	RegistrationID   int        `json:"registration_id"`
	RegistrationNum  string     `json:"registration_number,omitempty"`
	PatientID        int        `json:"patient_id"`
	PatientName      string     `json:"patient_name,omitempty"`
	PatientMRN       string     `json:"patient_mrn,omitempty"`
	DoctorID         int        `json:"doctor_id"`
	DoctorName       string     `json:"doctor_name,omitempty"`
	QueueDate        string     `json:"queue_date"`
	Status           string     `json:"status"`
	CalledAt         *time.Time `json:"called_at,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
