package medical_records

import "time"

type MedicalRecord struct {
	ID                     int       `json:"id"`
	MedicalRecordNumber    string    `json:"medical_record_number"`
	PatientID              int       `json:"patient_id"`
	PatientName            string    `json:"patient_name,omitempty"`
	PatientMRN             string    `json:"patient_mrn,omitempty"`
	DoctorID               int       `json:"doctor_id"`
	DoctorName             string    `json:"doctor_name,omitempty"`
	RegistrationID         *int      `json:"registration_id,omitempty"`
	RegistrationNumber     string    `json:"registration_number,omitempty"`
	ExaminationDate        string    `json:"examination_date"`
	ChiefComplaint         string    `json:"chief_complaint"`
	VitalSigns             string    `json:"vital_signs"`
	Anamnesis              string    `json:"anamnesis"`
	PhysicalExamination    string    `json:"physical_examination"`
	Notes                  string    `json:"notes"`
	Status                 string    `json:"status"`
	CreatedBy              *int      `json:"created_by,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
	Diagnoses              []DiagnosisEntry
	Treatments             []TreatmentEntry
}

type DiagnosisEntry struct {
	ID             int    `json:"id"`
	MedicalRecordID int   `json:"medical_record_id"`
	DiagnosisID    int    `json:"diagnosis_id"`
	DiagnosisCode  string `json:"diagnosis_code"`
	DiagnosisName  string `json:"diagnosis_name"`
	DiagnosisType  string `json:"diagnosis_type"`
	Notes          string `json:"notes"`
}

type TreatmentEntry struct {
	ID              int     `json:"id"`
	MedicalRecordID int     `json:"medical_record_id"`
	TreatmentID     int     `json:"treatment_id"`
	TreatmentCode   string  `json:"treatment_code"`
	TreatmentName   string  `json:"treatment_name"`
	Cost            float64 `json:"cost"`
	Notes           string  `json:"notes"`
}

type Diagnosis struct {
	ID             int    `json:"id"`
	DiagnosisCode  string `json:"diagnosis_code"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	IsActive       bool   `json:"is_active"`
}

type Treatment struct {
	ID              int     `json:"id"`
	TreatmentCode   string  `json:"treatment_code"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	DefaultCost     float64 `json:"default_cost"`
	IsActive        bool    `json:"is_active"`
}
