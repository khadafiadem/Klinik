package prescriptions

import (
	"database/sql"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(page, limit int, search string) ([]Prescription, int, error) {
	offset := (page - 1) * limit
	args := []interface{}{}
	where := ""

	if search != "" {
		where = `WHERE (pr.prescription_number ILIKE $1 OR p.full_name ILIKE $1 OR pr.medicine_name ILIKE $1)`
		args = append(args, "%"+search+"%")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM prescriptions pr
		JOIN patients p ON pr.patient_id = p.id %s`, where)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	idx := len(args)
	query := fmt.Sprintf(`SELECT pr.id, pr.prescription_number, pr.medical_record_id,
		COALESCE(mr.medical_record_number,''), pr.patient_id, p.full_name, p.medical_record_number,
		pr.doctor_id, d.full_name, pr.prescription_date,
		COALESCE(pr.notes,''), pr.status,
		COALESCE((SELECT COUNT(*) FROM prescription_items pi WHERE pi.prescription_id = pr.id),0),
		pr.created_at, pr.updated_at
		FROM prescriptions pr
		JOIN patients p ON pr.patient_id = p.id
		JOIN doctors d ON pr.doctor_id = d.id
		LEFT JOIN medical_records mr ON pr.medical_record_id = mr.id
		%s ORDER BY pr.id DESC LIMIT $%d OFFSET $%d`, where, idx-1, idx)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Prescription
	for rows.Next() {
		var pr Prescription
		if err := rows.Scan(&pr.ID, &pr.PrescriptionNumber, &pr.MedicalRecordID, &pr.MedicalRecordNum,
			&pr.PatientID, &pr.PatientName, &pr.PatientMRN,
			&pr.DoctorID, &pr.DoctorName, &pr.PrescriptionDate,
			&pr.Notes, &pr.Status, &pr.ItemCount, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, pr)
	}
	return list, total, nil
}

func (r *Repository) GetByID(id int) (*Prescription, error) {
	pr := &Prescription{}
	query := `SELECT pr.id, pr.prescription_number, pr.medical_record_id,
		COALESCE(mr.medical_record_number,''), pr.patient_id, p.full_name, p.medical_record_number,
		pr.doctor_id, d.full_name, pr.prescription_date,
		COALESCE(pr.notes,''), pr.status, pr.created_at, pr.updated_at
		FROM prescriptions pr
		JOIN patients p ON pr.patient_id = p.id
		JOIN doctors d ON pr.doctor_id = d.id
		LEFT JOIN medical_records mr ON pr.medical_record_id = mr.id
		WHERE pr.id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&pr.ID, &pr.PrescriptionNumber, &pr.MedicalRecordID, &pr.MedicalRecordNum,
		&pr.PatientID, &pr.PatientName, &pr.PatientMRN,
		&pr.DoctorID, &pr.DoctorName, &pr.PrescriptionDate,
		&pr.Notes, &pr.Status, &pr.CreatedAt, &pr.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("resep tidak ditemukan")
		}
		return nil, err
	}

	pr.Items, _ = r.GetItems(id)
	return pr, nil
}

func (r *Repository) Create(pr *Prescription) error {
	query := `INSERT INTO prescriptions (prescription_number, medical_record_id, patient_id, doctor_id,
		prescription_date, notes, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, pr.PrescriptionNumber, pr.MedicalRecordID, pr.PatientID,
		pr.DoctorID, pr.PrescriptionDate, pr.Notes, pr.Status).Scan(&pr.ID, &pr.CreatedAt, &pr.UpdatedAt)
}

func (r *Repository) GetItems(prescriptionID int) ([]PrescriptionItem, error) {
	rows, err := r.db.Query(`SELECT pi.id, pi.prescription_id, pi.medicine_id, pi.medicine_name, COALESCE(pi.medicine_code,''),
		pi.quantity, COALESCE(pi.dosage,''), COALESCE(pi.frequency,''), COALESCE(pi.duration,''), COALESCE(pi.instructions,''),
		COALESCE(m.stock, 0)
		FROM prescription_items pi
		LEFT JOIN medicines m ON m.id = pi.medicine_id
		WHERE pi.prescription_id = $1 ORDER BY pi.id`, prescriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []PrescriptionItem
	for rows.Next() {
		var pi PrescriptionItem
		if err := rows.Scan(&pi.ID, &pi.PrescriptionID, &pi.MedicineID, &pi.MedicineName, &pi.MedicineCode,
			&pi.Quantity, &pi.Dosage, &pi.Frequency, &pi.Duration, &pi.Instructions,
			&pi.Stock); err != nil {
			return nil, err
		}
		list = append(list, pi)
	}
	return list, nil
}

func (r *Repository) AddItem(pi *PrescriptionItem) error {
	query := `INSERT INTO prescription_items (prescription_id, medicine_id, medicine_name, medicine_code, quantity, dosage, frequency, duration, instructions)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`
	return r.db.QueryRow(query, pi.PrescriptionID, pi.MedicineID, pi.MedicineName, pi.MedicineCode,
		pi.Quantity, pi.Dosage, pi.Frequency, pi.Duration, pi.Instructions).Scan(&pi.ID)
}

// MedicineInfo mengambil snapshot nama/kode obat aktif dari master obat.
func (r *Repository) MedicineInfo(medicineID int) (name, code string, err error) {
	err = r.db.QueryRow(`SELECT name, medicine_code FROM medicines WHERE id=$1 AND is_active`, medicineID).Scan(&name, &code)
	if err == sql.ErrNoRows {
		return "", "", fmt.Errorf("obat tidak ditemukan atau tidak aktif")
	}
	return name, code, err
}

func (r *Repository) RemoveItem(id int) error {
	_, err := r.db.Exec("DELETE FROM prescription_items WHERE id=$1", id)
	return err
}

// UpdateStatusGuarded mengubah status hanya jika status saat ini termasuk salah satu fromStatus.
func (r *Repository) UpdateStatusGuarded(id int, status string, fromStatus []string) error {
	query := `UPDATE prescriptions SET status=$1 WHERE id=$2 AND status IN (`
	args := []interface{}{status, id}
	for i, fs := range fromStatus {
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("$%d", i+3)
		args = append(args, fs)
	}
	query += ")"
	res, err := r.db.Exec(query, args...)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("resep tidak dapat diubah dari status saat ini")
	}
	return nil
}

// CompleteWithDispense menyelesaikan resep dan memotong stok obat dalam satu transaksi DB.
// Item tanpa medicine_id (resep lama) dilewati tanpa memotong stok.
func (r *Repository) CompleteWithDispense(id int, userID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	err = tx.QueryRow("SELECT status FROM prescriptions WHERE id=$1 FOR UPDATE", id).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("resep tidak ditemukan")
		}
		return err
	}
	if status != "PROCESSING" {
		return fmt.Errorf("hanya resep berstatus PROCESSING yang dapat diselesaikan")
	}

	rows, err := tx.Query(`SELECT pi.medicine_id, pi.quantity, m.name, m.stock
		FROM prescription_items pi
		JOIN medicines m ON m.id = pi.medicine_id
		WHERE pi.prescription_id=$1 AND pi.medicine_id IS NOT NULL`, id)
	if err != nil {
		return err
	}

	type dispense struct {
		medicineID int
		quantity   int
		name       string
		stock      int
	}
	var items []dispense
	for rows.Next() {
		var d dispense
		if err := rows.Scan(&d.medicineID, &d.quantity, &d.name, &d.stock); err != nil {
			rows.Close()
			return err
		}
		items = append(items, d)
	}
	rows.Close()

	for _, d := range items {
		if d.stock < d.quantity {
			return fmt.Errorf("stok %s tidak mencukupi (tersedia %d, dibutuhkan %d)", d.name, d.stock, d.quantity)
		}
		res, err := tx.Exec(`UPDATE medicines SET stock = stock - $1 WHERE id=$2 AND stock >= $1`, d.quantity, d.medicineID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return fmt.Errorf("stok %s tidak mencukupi (tersedia %d, dibutuhkan %d)", d.name, d.stock, d.quantity)
		}
	}

	var rxNum string
	if err := tx.QueryRow("SELECT prescription_number FROM prescriptions WHERE id=$1", id).Scan(&rxNum); err != nil {
		return err
	}

	for _, d := range items {
		if _, err := tx.Exec(`INSERT INTO medicine_stock_transactions
			(medicine_id, transaction_type, quantity, notes, reference_type, reference_id, created_by)
			VALUES ($1,'KELUAR',$2,$3,'prescription',$4,$5)`,
			d.medicineID, d.quantity, "Dispensasi resep "+rxNum, id, userID); err != nil {
			return err
		}
	}

	if _, err := tx.Exec("UPDATE prescriptions SET status='COMPLETED' WHERE id=$1", id); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) GenerateNumber() (string, error) {
	var next int
	err := r.db.QueryRow("SELECT nextval('rx_seq')").Scan(&next)
	if err != nil {
		return "", fmt.Errorf("gagal generate nomor resep: %w", err)
	}
	var dateStr string
	if err := r.db.QueryRow("SELECT TO_CHAR(CURRENT_DATE, 'YYYYMMDD')").Scan(&dateStr); err != nil || dateStr == "" {
		dateStr = time.Now().Format("20060102")
	}
	return fmt.Sprintf("RX-%s-%03d", dateStr, next), nil
}

func (r *Repository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM prescriptions").Scan(&count)
	return count, err
}
