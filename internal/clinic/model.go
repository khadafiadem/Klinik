package clinic

import "time"

type ClinicSettings struct {
	ID                   int       `json:"id"`
	ClinicName           string    `json:"clinic_name"`
	ClinicAddress        string    `json:"clinic_address"`
	ClinicPhone          string    `json:"clinic_phone"`
	ClinicEmail          string    `json:"clinic_email"`
	ClinicLogo           string    `json:"clinic_logo"`
	OpeningTime          string    `json:"opening_time"`
	ClosingTime          string    `json:"closing_time"`
	MaxPatientsPerDay    int       `json:"max_patients_per_day"`
	RegistrationFee      float64   `json:"registration_fee"`
	ConsultationFee      float64   `json:"consultation_fee"`
	TaxPercentage        float64   `json:"tax_percentage"`
	Currency             string    `json:"currency"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
