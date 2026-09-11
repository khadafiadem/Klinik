package pain

import "time"

type PainAssessment struct {
	ID                 int       `json:"id"`
	MedicalRecordID    int       `json:"medical_record_id"`
	Location           string    `json:"location"`
	Quality            string    `json:"quality"`
	Intensity          int       `json:"intensity"`
	OnsetDuration      string    `json:"onset_duration"`
	Radiation          string    `json:"radiation"`
	AggravatingFactors string    `json:"aggravating_factors"`
	RelievingFactors   string    `json:"relieving_factors"`
	Notes              string    `json:"notes"`
	CreatedBy          *int      `json:"created_by,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
