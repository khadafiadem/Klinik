package appointments

import (
	"database/sql"
	"fmt"
	"strings"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(page, limit int, search, date string) ([]Appointment, int, error) {
	offset := (page - 1) * limit
	args := []interface{}{}
	where := ""
	ph := 1

	conds := []string{}
	if search != "" {
		conds = append(conds, fmt.Sprintf("(p.full_name ILIKE $%d OR p.medical_record_number ILIKE $%d OR a.appointment_number ILIKE $%d)", ph, ph+1, ph+2))
		args = append(args, "%"+search+"%", "%"+search+"%", "%"+search+"%")
		ph += 3
	}
	if date != "" {
		conds = append(conds, fmt.Sprintf("a.appointment_date = $%d", ph))
		args = append(args, date)
		ph++
	}
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM appointments a
		JOIN patients p ON a.patient_id = p.id
		JOIN doctors d ON a.doctor_id = d.id %s`, where)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung janji temu: %w", err)
	}

	query := fmt.Sprintf(`SELECT a.id, a.appointment_number, a.patient_id, p.full_name, p.medical_record_number,
		a.doctor_id, d.full_name, a.appointment_date, a.start_time,
		COALESCE(a.purpose,''), a.status, COALESCE(a.notes,''),
		a.created_by, a.created_at, a.updated_at
		FROM appointments a
		JOIN patients p ON a.patient_id = p.id
		JOIN doctors d ON a.doctor_id = d.id
		%s
		ORDER BY a.appointment_date DESC, a.start_time ASC
		LIMIT $%d OFFSET $%d
		`, where, ph, ph+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil data janji temu: %w", err)
	}
	defer rows.Close()

	var list []Appointment
	for rows.Next() {
		var a Appointment
		if err := rows.Scan(&a.ID, &a.AppointmentNumber, &a.PatientID, &a.PatientName, &a.PatientMRN,
			&a.DoctorID, &a.DoctorName, &a.AppointmentDate, &a.StartTime,
			&a.Purpose, &a.Status, &a.Notes,
			&a.CreatedBy, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("gagal scan janji temu: %w", err)
		}
		list = append(list, a)
	}
	return list, total, nil
}

// GetUpcoming mengembalikan janji temu dengan tanggal mulai dari hari ini ke depan.
func (r *Repository) GetUpcoming(today string) ([]Appointment, error) {
	query := `SELECT a.id, a.appointment_number, a.patient_id, p.full_name, p.medical_record_number,
		a.doctor_id, d.full_name, a.appointment_date, a.start_time,
		COALESCE(a.purpose,''), a.status, COALESCE(a.notes,''),
		a.created_by, a.created_at, a.updated_at
		FROM appointments a
		JOIN patients p ON a.patient_id = p.id
		JOIN doctors d ON a.doctor_id = d.id
		WHERE a.appointment_date >= $1 AND a.status NOT IN ('SELESAI', 'DIBATALKAN')
		ORDER BY a.appointment_date ASC, a.start_time ASC
		LIMIT 100`
	rows, err := r.db.Query(query, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Appointment
	for rows.Next() {
		var a Appointment
		if err := rows.Scan(&a.ID, &a.AppointmentNumber, &a.PatientID, &a.PatientName, &a.PatientMRN,
			&a.DoctorID, &a.DoctorName, &a.AppointmentDate, &a.StartTime,
			&a.Purpose, &a.Status, &a.Notes,
			&a.CreatedBy, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (r *Repository) GetByID(id int) (*Appointment, error) {
	a := &Appointment{}
	query := `SELECT a.id, a.appointment_number, a.patient_id, p.full_name, p.medical_record_number,
		a.doctor_id, d.full_name, a.appointment_date, a.start_time,
		COALESCE(a.purpose,''), a.status, COALESCE(a.notes,''),
		a.created_by, a.created_at, a.updated_at
		FROM appointments a
		JOIN patients p ON a.patient_id = p.id
		JOIN doctors d ON a.doctor_id = d.id
		WHERE a.id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&a.ID, &a.AppointmentNumber, &a.PatientID, &a.PatientName, &a.PatientMRN,
		&a.DoctorID, &a.DoctorName, &a.AppointmentDate, &a.StartTime,
		&a.Purpose, &a.Status, &a.Notes,
		&a.CreatedBy, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("janji temu tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil data janji temu: %w", err)
	}
	return a, nil
}

func (r *Repository) Create(a *Appointment) error {
	query := `INSERT INTO appointments (appointment_number, patient_id, doctor_id, appointment_date, start_time, purpose, status, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, a.AppointmentNumber, a.PatientID, a.DoctorID, a.AppointmentDate,
		a.StartTime, a.Purpose, a.Status, a.Notes, a.CreatedBy).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

func (r *Repository) Update(id int, a *Appointment) error {
	query := `UPDATE appointments SET patient_id=$1, doctor_id=$2, appointment_date=$3, start_time=$4,
		purpose=$5, notes=$6, updated_at=NOW() WHERE id=$7`
	_, err := r.db.Exec(query, a.PatientID, a.DoctorID, a.AppointmentDate, a.StartTime, a.Purpose, a.Notes, id)
	return err
}

func (r *Repository) UpdateStatus(id int, status string) error {
	query := `UPDATE appointments SET status=$1, updated_at=NOW() WHERE id=$2`
	result, err := r.db.Exec(query, status, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("janji temu tidak ditemukan")
	}
	return nil
}

func (r *Repository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM appointments WHERE id=$1", id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("janji temu tidak ditemukan")
	}
	return nil
}

func (r *Repository) GenerateNumber() (string, error) {
	var last string
	err := r.db.QueryRow("SELECT appointment_number FROM appointments ORDER BY id DESC LIMIT 1").Scan(&last)
	if err != nil || last == "" {
		return "APT0001", nil
	}
	num := 0
	fmt.Sscanf(strings.TrimPrefix(last, "APT"), "%d", &num)
	return fmt.Sprintf("APT%04d", num+1), nil
}

func (r *Repository) CountUpcoming(today string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM appointments
		WHERE appointment_date >= $1 AND status NOT IN ('SELESAI', 'DIBATALKAN')`, today).Scan(&count)
	return count, err
}