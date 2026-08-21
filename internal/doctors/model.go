package doctors

import "time"

type Doctor struct {
	ID               int       `json:"id"`
	UserID           *int      `json:"user_id,omitempty"`
	DoctorCode       string    `json:"doctor_code"`
	FullName         string    `json:"full_name"`
	Specialization   string    `json:"specialization"`
	LicenseNumber    string    `json:"license_number"`
	Phone            string    `json:"phone"`
	Email            string    `json:"email"`
	ConsultationFee  float64   `json:"consultation_fee"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
