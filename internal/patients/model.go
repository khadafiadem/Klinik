package patients

import "time"

type Patient struct {
	ID                     int       `json:"id"`
	MedicalRecordNumber    string    `json:"medical_record_number"`
	FullName               string    `json:"full_name"`
	NIK                    string    `json:"nik"`
	Gender                 string    `json:"gender"`
	DateOfBirth            string    `json:"date_of_birth"`
	BloodType              string    `json:"blood_type"`
	Phone                  string    `json:"phone"`
	Email                  string    `json:"email"`
	Address                string    `json:"address"`
	City                   string    `json:"city"`
	Province               string    `json:"province"`
	PostalCode             string    `json:"postal_code"`
	EmergencyContactName   string    `json:"emergency_contact_name"`
	EmergencyContactPhone  string    `json:"emergency_contact_phone"`
	EmergencyContactRelation string `json:"emergency_contact_relation"`
	InsuranceName          string    `json:"insurance_name"`
	InsuranceNumber        string    `json:"insurance_number"`
	Allergies              string    `json:"allergies"`
	Notes                  string    `json:"notes"`
	IsActive               bool      `json:"is_active"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}
