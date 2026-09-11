package immunizations

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

func (r *Repository) ListSchedules() ([]Schedule, error) {
	rows, err := r.db.Query(`SELECT id, name, age_label, month_from, month_to, category, sort_order, created_at, updated_at
		FROM immunization_schedules ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Schedule
	for rows.Next() {
		var s Schedule
		if err := rows.Scan(&s.ID, &s.Name, &s.AgeLabel, &s.MonthFrom, &s.MonthTo, &s.Category, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *Repository) ListImmunizations(page, limit int, search string) ([]Immunization, int, error) {
	offset := (page - 1) * limit
	args := []interface{}{}
	where := ""
	if search != "" {
		where = "WHERE (p.full_name ILIKE $1 OR p.medical_record_number ILIKE $1)"
		args = append(args, "%"+search+"%")
	}

	var total int
	if err := r.db.QueryRow(fmt.Sprintf(`SELECT COUNT(*) FROM patient_immunizations pi
		JOIN patients p ON pi.patient_id = p.id %s`, where), args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung imunisasi: %w", err)
	}

	args = append(args, limit, offset)
	query := fmt.Sprintf(`SELECT pi.id, pi.patient_id, p.full_name, p.medical_record_number,
		pi.immunization_id, pi.vaccination_date, pi.vaccine_name,
		COALESCE(pi.batch_number,''), COALESCE(pi.provider_name,''), COALESCE(pi.notes,''),
		pi.created_by, pi.created_at, pi.updated_at
		FROM patient_immunizations pi
		JOIN patients p ON pi.patient_id = p.id
		%s
		ORDER BY pi.vaccination_date DESC, pi.id DESC
		LIMIT $%d OFFSET $%d`, where, len(args)-1, len(args))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil data imunisasi: %w", err)
	}
	defer rows.Close()

	var list []Immunization
	for rows.Next() {
		var im Immunization
		if err := rows.Scan(&im.ID, &im.PatientID, &im.PatientName, &im.PatientMRN,
			&im.ImmunizationID, &im.VaccinationDate, &im.VaccineName,
			&im.BatchNumber, &im.ProviderName, &im.Notes,
			&im.CreatedBy, &im.CreatedAt, &im.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("gagal scan imunisasi: %w", err)
		}
		list = append(list, im)
	}
	return list, total, nil
}

func (r *Repository) Create(im *Immunization) error {
	query := `INSERT INTO patient_immunizations (patient_id, immunization_id, vaccination_date, vaccine_name, batch_number, provider_name, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, im.PatientID, im.ImmunizationID, im.VaccinationDate, im.VaccineName,
		im.BatchNumber, im.ProviderName, im.Notes, im.CreatedBy).Scan(&im.ID, &im.CreatedAt, &im.UpdatedAt)
}

func (r *Repository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM patient_immunizations WHERE id=$1", id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("imunisasi tidak ditemukan")
	}
	return nil
}
