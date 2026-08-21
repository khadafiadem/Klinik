package patients

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

func (r *Repository) GetAll(page, limit int, search string) ([]Patient, int, error) {
	offset := (page - 1) * limit
	args := []interface{}{}
	where := ""

	if search != "" {
		where = `WHERE (p.full_name ILIKE $1 OR p.medical_record_number ILIKE $1 
			OR p.nik ILIKE $1 OR p.phone ILIKE $1)`
		args = append(args, "%"+search+"%")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM patients p %s", where)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	idx := len(args)
	query := fmt.Sprintf(`
		SELECT p.id, p.medical_record_number, p.full_name, COALESCE(p.nik,''), p.gender,
			TO_CHAR(p.date_of_birth, 'YYYY-MM-DD'), COALESCE(p.blood_type,''),
			COALESCE(p.phone,''), COALESCE(p.email,''), COALESCE(p.address,''),
			COALESCE(p.city,''), COALESCE(p.province,''), COALESCE(p.postal_code,''),
			COALESCE(p.emergency_contact_name,''), COALESCE(p.emergency_contact_phone,''),
			COALESCE(p.emergency_contact_relation,''),
			COALESCE(p.insurance_name,''), COALESCE(p.insurance_number,''),
			COALESCE(p.allergies,''), COALESCE(p.notes,''),
			p.is_active, p.created_at, p.updated_at
		FROM patients p %s
		ORDER BY p.id DESC
		LIMIT $%d OFFSET $%d
	`, where, idx-1, idx)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var patients []Patient
	for rows.Next() {
		var p Patient
		if err := rows.Scan(
			&p.ID, &p.MedicalRecordNumber, &p.FullName, &p.NIK, &p.Gender,
			&p.DateOfBirth, &p.BloodType, &p.Phone, &p.Email, &p.Address,
			&p.City, &p.Province, &p.PostalCode,
			&p.EmergencyContactName, &p.EmergencyContactPhone, &p.EmergencyContactRelation,
			&p.InsuranceName, &p.InsuranceNumber, &p.Allergies, &p.Notes,
			&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		patients = append(patients, p)
	}
	return patients, total, nil
}

func (r *Repository) GetByID(id int) (*Patient, error) {
	p := &Patient{}
	query := `SELECT id, medical_record_number, full_name, COALESCE(nik,''), gender,
		TO_CHAR(date_of_birth, 'YYYY-MM-DD'), COALESCE(blood_type,''),
		COALESCE(phone,''), COALESCE(email,''), COALESCE(address,''),
		COALESCE(city,''), COALESCE(province,''), COALESCE(postal_code,''),
		COALESCE(emergency_contact_name,''), COALESCE(emergency_contact_phone,''),
		COALESCE(emergency_contact_relation,''),
		COALESCE(insurance_name,''), COALESCE(insurance_number,''),
		COALESCE(allergies,''), COALESCE(notes,''),
		is_active, created_at, updated_at
		FROM patients WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.MedicalRecordNumber, &p.FullName, &p.NIK, &p.Gender,
		&p.DateOfBirth, &p.BloodType, &p.Phone, &p.Email, &p.Address,
		&p.City, &p.Province, &p.PostalCode,
		&p.EmergencyContactName, &p.EmergencyContactPhone, &p.EmergencyContactRelation,
		&p.InsuranceName, &p.InsuranceNumber, &p.Allergies, &p.Notes,
		&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pasien tidak ditemukan")
		}
		return nil, err
	}
	return p, nil
}

func (r *Repository) Create(p *Patient) error {
	query := `INSERT INTO patients (medical_record_number, full_name, nik, gender, date_of_birth,
		blood_type, phone, email, address, city, province, postal_code,
		emergency_contact_name, emergency_contact_phone, emergency_contact_relation,
		insurance_name, insurance_number, allergies, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query,
		p.MedicalRecordNumber, p.FullName, p.NIK, p.Gender, p.DateOfBirth,
		p.BloodType, p.Phone, p.Email, p.Address, p.City, p.Province, p.PostalCode,
		p.EmergencyContactName, p.EmergencyContactPhone, p.EmergencyContactRelation,
		p.InsuranceName, p.InsuranceNumber, p.Allergies, p.Notes,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *Repository) Update(id int, p *Patient) error {
	query := `UPDATE patients SET full_name=$1, nik=$2, gender=$3, date_of_birth=$4,
		blood_type=$5, phone=$6, email=$7, address=$8, city=$9, province=$10, postal_code=$11,
		emergency_contact_name=$12, emergency_contact_phone=$13, emergency_contact_relation=$14,
		insurance_name=$15, insurance_number=$16, allergies=$17, notes=$18, is_active=$19
		WHERE id=$20`
	_, err := r.db.Exec(query,
		p.FullName, p.NIK, p.Gender, p.DateOfBirth, p.BloodType,
		p.Phone, p.Email, p.Address, p.City, p.Province, p.PostalCode,
		p.EmergencyContactName, p.EmergencyContactPhone, p.EmergencyContactRelation,
		p.InsuranceName, p.InsuranceNumber, p.Allergies, p.Notes, p.IsActive, id,
	)
	return err
}

func (r *Repository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM patients WHERE id=$1", id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("pasien tidak ditemukan")
	}
	return nil
}

func (r *Repository) GenerateMRN() (string, error) {
	var mrn sql.NullString
	err := r.db.QueryRow("SELECT medical_record_number FROM patients ORDER BY id DESC LIMIT 1").Scan(&mrn)
	if err != nil || !mrn.Valid {
		return "MR-000001", nil
	}
	var num int
	fmt.Sscanf(mrn.String, "MR-%06d", &num)
	return fmt.Sprintf("MR-%06d", num+1), nil
}

func (r *Repository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM patients WHERE is_active = true").Scan(&count)
	return count, err
}
