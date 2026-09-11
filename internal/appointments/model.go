package appointments

import "time"

type Appointment struct {
	ID                int       `json:"id"`
	AppointmentNumber string    `json:"appointment_number"`
	PatientID         int       `json:"patient_id"`
	PatientName       string    `json:"patient_name,omitempty"`
	PatientMRN        string    `json:"patient_mrn,omitempty"`
	DoctorID          int       `json:"doctor_id"`
	DoctorName        string    `json:"doctor_name,omitempty"`
	AppointmentDate   string    `json:"appointment_date"`
	StartTime         string    `json:"start_time"`
	Purpose           string    `json:"purpose"`
	Status            string    `json:"status"`
	Notes             string    `json:"notes"`
	CreatedBy         *int      `json:"created_by,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}