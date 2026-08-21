package finance

import "time"

type Invoice struct {
	ID               int       `json:"id"`
	InvoiceNumber    string    `json:"invoice_number"`
	PatientID        int       `json:"patient_id"`
	PatientName      string    `json:"patient_name,omitempty"`
	PatientMRN       string    `json:"patient_mrn,omitempty"`
	RegistrationID   *int      `json:"registration_id,omitempty"`
	MedicalRecordID  *int      `json:"medical_record_id,omitempty"`
	InvoiceDate      string    `json:"invoice_date"`
	Subtotal         float64   `json:"subtotal"`
	Discount         float64   `json:"discount"`
	Total            float64   `json:"total"`
	Notes            string    `json:"notes"`
	Status           string    `json:"status"`
	CreatedBy        *int      `json:"created_by,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Items            []InvoiceItem
	Payments         []Payment
}

type InvoiceItem struct {
	ID          int     `json:"id"`
	InvoiceID   int     `json:"invoice_id"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
	ItemType    string  `json:"item_type"`
	ReferenceID *int    `json:"reference_id,omitempty"`
}

type Payment struct {
	ID              int       `json:"id"`
	PaymentNumber   string    `json:"payment_number"`
	InvoiceID       int       `json:"invoice_id"`
	PatientID       int       `json:"patient_id"`
	PatientName     string    `json:"patient_name,omitempty"`
	PaymentDate     string    `json:"payment_date"`
	Amount          float64   `json:"amount"`
	PaymentMethod   string    `json:"payment_method"`
	ReferenceNumber string    `json:"reference_number"`
	Notes           string    `json:"notes"`
	Status          string    `json:"status"`
	ReceivedBy      *int     `json:"received_by,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}
