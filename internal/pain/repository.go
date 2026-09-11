package pain

import (
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Upsert menyimpan penilaian nyeri; jika sudah ada untuk rekam medis yang sama,
// data diperbarui (satu penilaian per rekam medis).
func (r *Repository) Upsert(p *PainAssessment) error {
	query := `INSERT INTO pain_assessments (medical_record_id, location, quality, intensity,
		onset_duration, radiation, aggravating_factors, relieving_factors, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (medical_record_id) DO UPDATE SET
			location = EXCLUDED.location,
			quality = EXCLUDED.quality,
			intensity = EXCLUDED.intensity,
			onset_duration = EXCLUDED.onset_duration,
			radiation = EXCLUDED.radiation,
			aggravating_factors = EXCLUDED.aggravating_factors,
			relieving_factors = EXCLUDED.relieving_factors,
			notes = EXCLUDED.notes,
			updated_at = NOW()
		RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, p.MedicalRecordID, p.Location, p.Quality, p.Intensity,
		p.OnsetDuration, p.Radiation, p.AggravatingFactors, p.RelievingFactors, p.Notes, p.CreatedBy).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *Repository) GetByMedicalRecordID(mrID int) (*PainAssessment, error) {
	p := &PainAssessment{}
	query := `SELECT id, medical_record_id, COALESCE(location,''), COALESCE(quality,''),
		intensity, COALESCE(onset_duration,''), COALESCE(radiation,''),
		COALESCE(aggravating_factors,''), COALESCE(relieving_factors,''), COALESCE(notes,''),
		created_by, created_at, updated_at
		FROM pain_assessments WHERE medical_record_id = $1`
	err := r.db.QueryRow(query, mrID).Scan(
		&p.ID, &p.MedicalRecordID, &p.Location, &p.Quality,
		&p.Intensity, &p.OnsetDuration, &p.Radiation,
		&p.AggravatingFactors, &p.RelievingFactors, &p.Notes,
		&p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mengambil pengkajian nyeri: %w", err)
	}
	return p, nil
}
