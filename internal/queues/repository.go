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
	query := `SELECT q.id, q.queue_number, q.registration_id, r.registration_number,
		q.patient_id, p.full_name, p.medical_record_number,
		q.doctor_id, d.full_name, q.queue_date, q.status,
		q.called_at, q.started_at, q.completed_at, q.created_at, q.updated_at
		FROM queues q
		JOIN registrations r ON q.registration_id = r.id
		JOIN patients p ON q.patient_id = p.id
		JOIN doctors d ON q.doctor_id = d.id
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
			&q.DoctorID, &q.DoctorName, &q.QueueDate, &q.Status,
			&q.CalledAt, &q.StartedAt, &q.CompletedAt, &q.CreatedAt, &q.UpdatedAt); err != nil {
			return nil, fmt.Errorf("gagal scan antrian: %w", err)
		}
		list = append(list, q)
	}
	return list, nil
}

func (r *Repository) GetByID(id int) (*Queue, error) {
	q := &Queue{}
	query := `SELECT q.id, q.queue_number, q.registration_id, r.registration_number,
		q.patient_id, p.full_name, p.medical_record_number,
		q.doctor_id, d.full_name, q.queue_date, q.status,
		q.called_at, q.started_at, q.completed_at, q.created_at, q.updated_at
		FROM queues q
		JOIN registrations r ON q.registration_id = r.id
		JOIN patients p ON q.patient_id = p.id
		JOIN doctors d ON q.doctor_id = d.id
		WHERE q.id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&q.ID, &q.QueueNumber, &q.RegistrationID, &q.RegistrationNum,
		&q.PatientID, &q.PatientName, &q.PatientMRN,
		&q.DoctorID, &q.DoctorName, &q.QueueDate, &q.Status,
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
	query := `INSERT INTO queues (queue_number, registration_id, patient_id, doctor_id, queue_date, status)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, q.QueueNumber, q.RegistrationID, q.PatientID,
		q.DoctorID, q.QueueDate, q.Status).Scan(&q.ID, &q.CreatedAt, &q.UpdatedAt)
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

func (r *Repository) GenerateNumber(date string) (string, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM queues WHERE queue_date = $1", date).Scan(&count)
	if err != nil {
		return "A-001", nil
	}
	return fmt.Sprintf("A-%03d", count+1), nil
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
