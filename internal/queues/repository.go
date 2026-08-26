package queues

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

func (r *Repository) GetAllByDate(date string) ([]Queue, error) {
	query := `SELECT q.id, q.queue_number, q.registration_id,
		COALESCE(r.registration_number,''),
		q.patient_id, COALESCE(p.full_name,''), COALESCE(p.medical_record_number,''),
		q.doctor_id, COALESCE(d.full_name,''), COALESCE(q.doctor_name_snapshot,''),
		q.queue_date, q.status, q.queue_source, q.called_by,
		COALESCE(u.full_name,''),
		q.called_at, q.started_at, q.completed_at, q.created_at, q.updated_at
		FROM queues q
		LEFT JOIN registrations r ON q.registration_id = r.id
		LEFT JOIN patients p ON q.patient_id = p.id
		LEFT JOIN doctors d ON q.doctor_id = d.id
		LEFT JOIN users u ON q.called_by = u.id
		WHERE q.queue_date = $1
		ORDER BY q.queue_number ASC`

	rows, err := r.db.Query(query, date)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data antrian: %w", err)
	}
	defer rows.Close()

	var list []Queue
	for rows.Next() {
		var q Queue
		if err := rows.Scan(&q.ID, &q.QueueNumber, &q.RegistrationID, &q.RegistrationNum,
			&q.PatientID, &q.PatientName, &q.PatientMRN,
			&q.DoctorID, &q.DoctorName, &q.DoctorNameSnapshot,
			&q.QueueDate, &q.Status, &q.QueueSource, &q.CalledBy,
			&q.CalledByName,
			&q.CalledAt, &q.StartedAt, &q.CompletedAt, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return nil, fmt.Errorf("gagal scan antrian: %w", err)
		}
		list = append(list, q)
	}
	return list, nil
}

func (r *Repository) GetByID(id int) (*Queue, error) {
	q := &Queue{}
	query := `SELECT q.id, q.queue_number, q.registration_id,
		COALESCE(r.registration_number,''),
		q.patient_id, COALESCE(p.full_name,''), COALESCE(p.medical_record_number,''),
		q.doctor_id, COALESCE(d.full_name,''), COALESCE(q.doctor_name_snapshot,''),
		q.queue_date, q.status, q.queue_source, q.called_by,
		COALESCE(u.full_name,''),
		q.called_at, q.started_at, q.completed_at, q.created_at, q.updated_at
		FROM queues q
		LEFT JOIN registrations r ON q.registration_id = r.id
		LEFT JOIN patients p ON q.patient_id = p.id
		LEFT JOIN doctors d ON q.doctor_id = d.id
		LEFT JOIN users u ON q.called_by = u.id
		WHERE q.id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&q.ID, &q.QueueNumber, &q.RegistrationID, &q.RegistrationNum,
		&q.PatientID, &q.PatientName, &q.PatientMRN,
		&q.DoctorID, &q.DoctorName, &q.DoctorNameSnapshot,
		&q.QueueDate, &q.Status, &q.QueueSource, &q.CalledBy,
		&q.CalledByName,
		&q.CalledAt, &q.StartedAt, &q.CompletedAt, &q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("antrian tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil data antrian: %w", err)
	}
	return q, nil
}

func (r *Repository) Create(q *Queue) error {
	var regID, patientID, doctorID interface{}
	if q.RegistrationID != nil && *q.RegistrationID > 0 {
		regID = *q.RegistrationID
	}
	if q.PatientID != nil && *q.PatientID > 0 {
		patientID = *q.PatientID
	}
	if q.DoctorID != nil && *q.DoctorID > 0 {
		doctorID = *q.DoctorID
	}
	query := `INSERT INTO queues (queue_number, registration_id, patient_id, doctor_id, queue_date, status, queue_source, doctor_name_snapshot)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, q.QueueNumber, regID, patientID, doctorID,
		q.QueueDate, q.Status, q.QueueSource, q.DoctorNameSnapshot).Scan(&q.ID, &q.CreatedAt, &q.UpdatedAt)
}

func (r *Repository) UpdateStatus(id int, status string) error {
	var extra string
	switch status {
	case "DIPANGGIL":
		extra = ", called_at = NOW()"
	case "SEDANG_DIPERIKSA":
		extra = ", started_at = NOW()"
	case "SELESAI":
		extra = ", completed_at = NOW()"
	}
	query := fmt.Sprintf("UPDATE queues SET status=$1%s WHERE id=$2", extra)
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *Repository) UpdateStatusCalledBy(id int, status string, calledBy int) error {
	var extra string
	switch status {
	case "DIPANGGIL":
		extra = ", called_at = NOW(), called_by = $3"
	case "SEDANG_DIPERIKSA":
		extra = ", started_at = NOW()"
	case "SELESAI":
		extra = ", completed_at = NOW()"
	}
	query := fmt.Sprintf("UPDATE queues SET status=$1%s WHERE id=$2", extra)
	_, err := r.db.Exec(query, status, id, calledBy)
	return err
}

func (r *Repository) GenerateNumber(date string) (string, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM queues WHERE queue_date = $1", date).Scan(&count)
	if err != nil {
		return "A-001", nil
	}
	return fmt.Sprintf("A-%03d", count+1), nil
}

func (r *Repository) GenerateKioskNumber() (string, error) {
	var num string
	err := r.db.QueryRow("SELECT generate_kiosk_queue_number(CURRENT_DATE)").Scan(&num)
	if err != nil {
		return "", fmt.Errorf("gagal generate nomor antrian: %w", err)
	}
	return num, nil
}

func (r *Repository) GetTodayStats() (waiting, inProgress, completed int, err error) {
	err = r.db.QueryRow("SELECT COUNT(*) FROM queues WHERE queue_date = CURRENT_DATE AND status = 'MENUNGGU'").Scan(&waiting)
	if err != nil {
		return
	}
	_ = r.db.QueryRow("SELECT COUNT(*) FROM queues WHERE queue_date = CURRENT_DATE AND status IN ('DIPANGGIL', 'SEDANG_DIPERIKSA')").Scan(&inProgress)
	_ = r.db.QueryRow("SELECT COUNT(*) FROM queues WHERE queue_date = CURRENT_DATE AND status = 'SELESAI'").Scan(&completed)
	return
}

func (r *Repository) GetKioskQueuesToday() ([]Queue, error) {
	query := `SELECT q.id, q.queue_number, q.registration_id,
		COALESCE(r.registration_number,''),
		q.patient_id, COALESCE(p.full_name,''), COALESCE(p.medical_record_number,''),
		q.doctor_id, COALESCE(d.full_name,''), COALESCE(q.doctor_name_snapshot,''),
		q.queue_date, q.status, q.queue_source, q.called_by, COALESCE(u.full_name,''),
		q.called_at, q.started_at, q.completed_at, q.created_at, q.updated_at
		FROM queues q
		LEFT JOIN registrations r ON q.registration_id = r.id
		LEFT JOIN patients p ON q.patient_id = p.id
		LEFT JOIN doctors d ON q.doctor_id = d.id
		LEFT JOIN users u ON q.called_by = u.id
		WHERE q.queue_date = CURRENT_DATE AND q.queue_source = 'KIOSK' AND q.registration_id IS NULL AND q.status = 'MENUNGGU'
		ORDER BY q.created_at ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Queue
	for rows.Next() {
		var q Queue
		if err := rows.Scan(&q.ID, &q.QueueNumber, &q.RegistrationID, &q.RegistrationNum,
			&q.PatientID, &q.PatientName, &q.PatientMRN,
			&q.DoctorID, &q.DoctorName, &q.DoctorNameSnapshot,
			&q.QueueDate, &q.Status, &q.QueueSource, &q.CalledBy, &q.CalledByName,
			&q.CalledAt, &q.StartedAt, &q.CompletedAt, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, q)
	}
	return list, nil
}

func (r *Repository) LinkToRegistration(queueID int, registrationID, patientID, doctorID int) error {
	_, err := r.db.Exec(
		`UPDATE queues SET registration_id = $1, patient_id = $2, doctor_id = $3 WHERE id = $4`,
		registrationID, patientID, doctorID, queueID)
	return err
}

func (r *Repository) GetMonitorData(date string) ([]Queue, error) {
	query := `SELECT q.id, q.queue_number, q.registration_id,
		COALESCE(r.registration_number,''),
		q.patient_id, COALESCE(p.full_name,''), COALESCE(p.medical_record_number,''),
		q.doctor_id, COALESCE(d.full_name,''), COALESCE(q.doctor_name_snapshot,''),
		q.queue_date, q.status, q.queue_source, q.called_by, COALESCE(u.full_name,''),
		q.called_at, q.started_at, q.completed_at, q.created_at, q.updated_at
		FROM queues q
		LEFT JOIN registrations r ON q.registration_id = r.id
		LEFT JOIN patients p ON q.patient_id = p.id
		LEFT JOIN doctors d ON q.doctor_id = d.id
		LEFT JOIN users u ON q.called_by = u.id
		WHERE q.queue_date = $1 AND q.status != 'DIBATALKAN'
		ORDER BY
			CASE q.status
				WHEN 'DIPANGGIL' THEN 1
				WHEN 'SEDANG_DIPERIKSA' THEN 2
				WHEN 'MENUNGGU' THEN 3
				WHEN 'SELESAI' THEN 4
			END,
			q.queue_number ASC`

	rows, err := r.db.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Queue
	for rows.Next() {
		var q Queue
		if err := rows.Scan(&q.ID, &q.QueueNumber, &q.RegistrationID, &q.RegistrationNum,
			&q.PatientID, &q.PatientName, &q.PatientMRN,
			&q.DoctorID, &q.DoctorName, &q.DoctorNameSnapshot,
			&q.QueueDate, &q.Status, &q.QueueSource, &q.CalledBy, &q.CalledByName,
			&q.CalledAt, &q.StartedAt, &q.CompletedAt, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, q)
	}
	return list, nil
}
