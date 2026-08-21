package medical_records

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

func (r *Repository) GetAll(page, limit int, search string) ([]MedicalRecord, int, error) {
	offset := (page - 1) * limit
	args := []interface{}{}
	where := ""

	if search != "" {
		where = `WHERE (mr.medical_record_number ILIKE $1 OR p.full_name ILIKE $1 OR p.medical_record_number ILIKE $1 OR d.full_name ILIKE $1)`
		args = append(args, "%"+search+"%")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM medical_records mr
		JOIN patients p ON mr.patient_id = p.id
		JOIN doctors d ON mr.doctor_id = d.id %s`, where)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	idx := len(args)
	query := fmt.Sprintf(`SELECT mr.id, mr.medical_record_number, mr.patient_id, p.full_name, p.medical_record_number,
		mr.doctor_id, d.full_name, mr.registration_id,
		COALESCE((SELECT r2.registration_number FROM registrations r2 WHERE r2.id = mr.registration_id),''),
		mr.examination_date, COALESCE(mr.chief_complaint,''),
		COALESCE(mr.vital_signs::text,'{}'), COALESCE(mr.anamnesis,''),
		COALESCE(mr.physical_examination,''), COALESCE(mr.notes,''), mr.status,
		mr.created_by, mr.created_at, mr.updated_at
		FROM medical_records mr
		JOIN patients p ON mr.patient_id = p.id
		JOIN doctors d ON mr.doctor_id = d.id
		%s ORDER BY mr.id DESC LIMIT $%d OFFSET $%d`, where, idx-1, idx)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []MedicalRecord
	for rows.Next() {
		var mr MedicalRecord
		if err := rows.Scan(&mr.ID, &mr.MedicalRecordNumber, &mr.PatientID, &mr.PatientName, &mr.PatientMRN,
			&mr.DoctorID, &mr.DoctorName, &mr.RegistrationID, &mr.RegistrationNumber,
			&mr.ExaminationDate, &mr.ChiefComplaint, &mr.VitalSigns, &mr.Anamnesis,
			&mr.PhysicalExamination, &mr.Notes, &mr.Status, &mr.CreatedBy, &mr.CreatedAt, &mr.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, mr)
	}
	return list, total, nil
}

func (r *Repository) GetByID(id int) (*MedicalRecord, error) {
	mr := &MedicalRecord{}
	query := `SELECT mr.id, mr.medical_record_number, mr.patient_id, p.full_name, p.medical_record_number,
		mr.doctor_id, d.full_name, mr.registration_id,
		COALESCE((SELECT r2.registration_number FROM registrations r2 WHERE r2.id = mr.registration_id),''),
		mr.examination_date, COALESCE(mr.chief_complaint,''),
		COALESCE(mr.vital_signs::text,'{}'), COALESCE(mr.anamnesis,''),
		COALESCE(mr.physical_examination,''), COALESCE(mr.notes,''), mr.status,
		mr.created_by, mr.created_at, mr.updated_at
		FROM medical_records mr
		JOIN patients p ON mr.patient_id = p.id
		JOIN doctors d ON mr.doctor_id = d.id
		WHERE mr.id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&mr.ID, &mr.MedicalRecordNumber, &mr.PatientID, &mr.PatientName, &mr.PatientMRN,
		&mr.DoctorID, &mr.DoctorName, &mr.RegistrationID, &mr.RegistrationNumber,
		&mr.ExaminationDate, &mr.ChiefComplaint, &mr.VitalSigns, &mr.Anamnesis,
		&mr.PhysicalExamination, &mr.Notes, &mr.Status, &mr.CreatedBy, &mr.CreatedAt, &mr.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("rekam medis tidak ditemukan")
		}
		return nil, err
	}

	mr.Diagnoses, _ = r.GetDiagnoses(id)
	mr.Treatments, _ = r.GetTreatments(id)
	return mr, nil
}

func (r *Repository) Create(mr *MedicalRecord) error {
	query := `INSERT INTO medical_records (medical_record_number, patient_id, doctor_id, registration_id,
		examination_date, chief_complaint, vital_signs, anamnesis, physical_examination, notes, status, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, mr.MedicalRecordNumber, mr.PatientID, mr.DoctorID, mr.RegistrationID,
		mr.ExaminationDate, mr.ChiefComplaint, mr.VitalSigns, mr.Anamnesis,
		mr.PhysicalExamination, mr.Notes, mr.Status, mr.CreatedBy).Scan(&mr.ID, &mr.CreatedAt, &mr.UpdatedAt)
}

func (r *Repository) Update(mr *MedicalRecord) error {
	_, err := r.db.Exec(`UPDATE medical_records SET
		chief_complaint=$1, vital_signs=$2, anamnesis=$3, physical_examination=$4,
		notes=$5, status=$6
		WHERE id=$7`,
		mr.ChiefComplaint, mr.VitalSigns, mr.Anamnesis, mr.PhysicalExamination,
		mr.Notes, mr.Status, mr.ID)
	return err
}

func (r *Repository) UpdateStatus(id int, status string) error {
	_, err := r.db.Exec("UPDATE medical_records SET status=$1 WHERE id=$2", status, id)
	return err
}

func (r *Repository) GetDiagnoses(medicalRecordID int) ([]DiagnosisEntry, error) {
	query := `SELECT mrd.id, mrd.medical_record_id, mrd.diagnosis_id,
		COALESCE(d.diagnosis_code,''), COALESCE(d.name,''), mrd.diagnosis_type, COALESCE(mrd.notes,'')
		FROM medical_record_diagnoses mrd
		LEFT JOIN diagnoses d ON mrd.diagnosis_id = d.id
		WHERE mrd.medical_record_id = $1`
	rows, err := r.db.Query(query, medicalRecordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DiagnosisEntry
	for rows.Next() {
		var de DiagnosisEntry
		if err := rows.Scan(&de.ID, &de.MedicalRecordID, &de.DiagnosisID,
			&de.DiagnosisCode, &de.DiagnosisName, &de.DiagnosisType, &de.Notes); err != nil {
			return nil, err
		}
		list = append(list, de)
	}
	return list, nil
}

func (r *Repository) AddDiagnosis(medicalRecordID, diagnosisID int, diagnosisType, notes string) error {
	_, err := r.db.Exec(`INSERT INTO medical_record_diagnoses (medical_record_id, diagnosis_id, diagnosis_type, notes)
		VALUES ($1,$2,$3,$4)`, medicalRecordID, diagnosisID, diagnosisType, notes)
	return err
}

func (r *Repository) RemoveDiagnosis(id int) error {
	_, err := r.db.Exec("DELETE FROM medical_record_diagnoses WHERE id=$1", id)
	return err
}

func (r *Repository) GetTreatments(medicalRecordID int) ([]TreatmentEntry, error) {
	query := `SELECT mrt.id, mrt.medical_record_id, mrt.treatment_id,
		COALESCE(t.treatment_code,''), COALESCE(t.name,''), mrt.cost, COALESCE(mrt.notes,'')
		FROM medical_record_treatments mrt
		LEFT JOIN treatments t ON mrt.treatment_id = t.id
		WHERE mrt.medical_record_id = $1`
	rows, err := r.db.Query(query, medicalRecordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []TreatmentEntry
	for rows.Next() {
		var te TreatmentEntry
		if err := rows.Scan(&te.ID, &te.MedicalRecordID, &te.TreatmentID,
			&te.TreatmentCode, &te.TreatmentName, &te.Cost, &te.Notes); err != nil {
			return nil, err
		}
		list = append(list, te)
	}
	return list, nil
}

func (r *Repository) AddTreatment(medicalRecordID, treatmentID int, cost float64, notes string) error {
	_, err := r.db.Exec(`INSERT INTO medical_record_treatments (medical_record_id, treatment_id, cost, notes)
		VALUES ($1,$2,$3,$4)`, medicalRecordID, treatmentID, cost, notes)
	return err
}

func (r *Repository) RemoveTreatment(id int) error {
	_, err := r.db.Exec("DELETE FROM medical_record_treatments WHERE id=$1", id)
	return err
}

func (r *Repository) GenerateNumber() (string, error) {
	var next int
	err := r.db.QueryRow("SELECT nextval('mr_seq')").Scan(&next)
	if err != nil {
		return "MR-20260821-001", nil
	}
	var dateStr string
	_ = r.db.QueryRow("SELECT TO_CHAR(CURRENT_DATE, 'YYYYMMDD')").Scan(&dateStr)
	if dateStr == "" {
		dateStr = "20260821"
	}
	return fmt.Sprintf("RM-%s-%03d", dateStr, next), nil
}

func (r *Repository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM medical_records WHERE examination_date = CURRENT_DATE").Scan(&count)
	return count, err
}

// Diagnosis master data
func (r *Repository) GetAllDiagnoses() ([]Diagnosis, error) {
	rows, err := r.db.Query("SELECT id, diagnosis_code, name, COALESCE(description,''), is_active FROM diagnoses WHERE is_active=true ORDER BY diagnosis_code")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Diagnosis
	for rows.Next() {
		var d Diagnosis
		if err := rows.Scan(&d.ID, &d.DiagnosisCode, &d.Name, &d.Description, &d.IsActive); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, nil
}

func (r *Repository) GetDiagnosisByID(id int) (*Diagnosis, error) {
	d := &Diagnosis{}
	err := r.db.QueryRow("SELECT id, diagnosis_code, name, COALESCE(description,''), is_active FROM diagnoses WHERE id=$1", id).
		Scan(&d.ID, &d.DiagnosisCode, &d.Name, &d.Description, &d.IsActive)
	return d, err
}

// Treatment master data
func (r *Repository) GetAllTreatments() ([]Treatment, error) {
	rows, err := r.db.Query("SELECT id, treatment_code, name, COALESCE(description,''), default_cost, is_active FROM treatments WHERE is_active=true ORDER BY treatment_code")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Treatment
	for rows.Next() {
		var t Treatment
		if err := rows.Scan(&t.ID, &t.TreatmentCode, &t.Name, &t.Description, &t.DefaultCost, &t.IsActive); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *Repository) GetTreatmentByID(id int) (*Treatment, error) {
	t := &Treatment{}
	err := r.db.QueryRow("SELECT id, treatment_code, name, COALESCE(description,''), default_cost, is_active FROM treatments WHERE id=$1", id).
		Scan(&t.ID, &t.TreatmentCode, &t.Name, &t.Description, &t.DefaultCost, &t.IsActive)
	return t, err
}
