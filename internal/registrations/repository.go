package registrations

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

func (r *Repository) GetAll(page, limit int, search string) ([]Registration, int, error) {
	offset := (page - 1) * limit
	args := []interface{}{}
	where := ""

	if search != "" {
		where = `WHERE (r.registration_number ILIKE $1 OR p.full_name ILIKE $1 OR p.medical_record_number ILIKE $1 OR d.full_name ILIKE $1)`
		args = append(args, "%"+search+"%")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM registrations r
		JOIN patients p ON r.patient_id = p.id
		JOIN doctors d ON r.doctor_id = d.id %s`, where)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung registrasi: %w", err)
	}

	args = append(args, limit, offset)
	idx := len(args)
	query := fmt.Sprintf(`SELECT r.id, r.registration_number, r.patient_id, p.full_name, p.medical_record_number,
		r.doctor_id, d.full_name, r.registration_date, r.registration_type,
		COALESCE(r.complaint,''), r.status, COALESCE(r.notes,''),
		r.created_by, r.created_at, r.updated_at
		FROM registrations r
		JOIN patients p ON r.patient_id = p.id
		JOIN doctors d ON r.doctor_id = d.id
		%s
		ORDER BY r.id DESC
		LIMIT $%d OFFSET $%d
	`, where, idx-1, idx)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil data registrasi: %w", err)
	}
	defer rows.Close()

	var list []Registration
	for rows.Next() {
		var reg Registration
		if err := rows.Scan(&reg.ID, &reg.RegistrationNumber, &reg.PatientID, &reg.PatientName, &reg.PatientMRN,
			&reg.DoctorID, &reg.DoctorName, &reg.RegistrationDate, &reg.RegistrationType,
			&reg.Complaint, &reg.Status, &reg.Notes,
			&reg.CreatedBy, &reg.CreatedAt, &reg.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("gagal scan registrasi: %w", err)
		}
		list = append(list, reg)
	}
	return list, total, nil
}

func (r *Repository) GetByID(id int) (*Registration, error) {
	reg := &Registration{}
	query := `SELECT r.id, r.registration_number, r.patient_id, p.full_name, p.medical_record_number,
		r.doctor_id, d.full_name, r.registration_date, r.registration_type,
		COALESCE(r.complaint,''), r.status, COALESCE(r.notes,''),
		r.created_by, r.created_at, r.updated_at
		FROM registrations r
		JOIN patients p ON r.patient_id = p.id
		JOIN doctors d ON r.doctor_id = d.id
		WHERE r.id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&reg.ID, &reg.RegistrationNumber, &reg.PatientID, &reg.PatientName, &reg.PatientMRN,
		&reg.DoctorID, &reg.DoctorName, &reg.RegistrationDate, &reg.RegistrationType,
		&reg.Complaint, &reg.Status, &reg.Notes,
		&reg.CreatedBy, &reg.CreatedAt, &reg.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("registrasi tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil data registrasi: %w", err)
	}
	return reg, nil
}

func (r *Repository) Create(reg *Registration) error {
	query := `INSERT INTO registrations (registration_number, patient_id, doctor_id, registration_date,
		registration_type, complaint, status, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, reg.RegistrationNumber, reg.PatientID, reg.DoctorID,
		reg.RegistrationDate, reg.RegistrationType, reg.Complaint, reg.Status, reg.Notes, reg.CreatedBy).
		Scan(&reg.ID, &reg.CreatedAt, &reg.UpdatedAt)
}

func (r *Repository) UpdateStatus(id int, status string) error {
	_, err := r.db.Exec("UPDATE registrations SET status=$1 WHERE id=$2", status, id)
	return err
}

func (r *Repository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM registrations WHERE registration_date = CURRENT_DATE").Scan(&count)
	return count, err
}

func (r *Repository) GenerateNumber() (string, error) {
	var next int
	err := r.db.QueryRow("SELECT nextval('reg_seq')").Scan(&next)
	if err != nil {
		return "REG-20260821-001", nil
	}
	var dateStr string
	_ = r.db.QueryRow("SELECT TO_CHAR(CURRENT_DATE, 'YYYYMMDD')").Scan(&dateStr)
	if dateStr == "" {
		dateStr = "20260821"
	}
	return fmt.Sprintf("REG-%s-%03d", dateStr, next), nil
}
