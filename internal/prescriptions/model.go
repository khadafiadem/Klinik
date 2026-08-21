package prescriptions

import "time"

type Prescription struct {
	ID                 int              `json:"id"`
	PrescriptionNumber string           `json:"prescription_number"`
	MedicalRecordID    int              `json:"medical_record_id"`
	MedicalRecordNum   string           `json:"medical_record_number,omitempty"`
	PatientID          int              `json:"patient_id"`
	PatientName        string           `json:"patient_name,omitempty"`
	PatientMRN         string           `json:"patient_mrn,omitempty"`
	DoctorID           int              `json:"doctor_id"`
	DoctorName         string           `json:"doctor_name,omitempty"`
	PrescriptionDate   string           `json:"prescription_date"`
	Notes              string           `json:"notes"`
	Status             string           `json:"status"`
	ItemCount          int              `json:"item_count"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	Items              []PrescriptionItem
}

type PrescriptionItem struct {
	ID             int    `json:"id"`
	PrescriptionID int    `json:"prescription_id"`
	MedicineID     *int   `json:"medicine_id,omitempty"`
	MedicineName   string `json:"medicine_name"`
	MedicineCode   string `json:"medicine_code"`
	Quantity       int    `json:"quantity"`
	Dosage         string `json:"dosage"`
	Frequency      string `json:"frequency"`
	Duration       string `json:"duration"`
	Instructions   string `json:"instructions"`
	Stock          int    `json:"stock,omitempty"`
}
