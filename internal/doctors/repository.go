package doctors

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

func (r *Repository) GetAll(page, limit int, search string) ([]Doctor, int, error) {
	offset := (page - 1) * limit
	args := []interface{}{}
	where := ""

	if search != "" {
		where = "WHERE (d.full_name ILIKE $1 OR d.doctor_code ILIKE $1 OR d.specialization ILIKE $1)"
		args = append(args, "%"+search+"%")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM doctors d %s", where)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung dokter: %w", err)
	}

	args = append(args, limit, offset)
	idx := len(args)
	query := fmt.Sprintf(`
		SELECT d.id, d.user_id, d.doctor_code, d.full_name, d.specialization,
			COALESCE(d.license_number,''), COALESCE(d.phone,''), COALESCE(d.email,''),
			d.consultation_fee, d.is_active, d.created_at, d.updated_at
		FROM doctors d %s
		ORDER BY d.id DESC
		LIMIT $%d OFFSET $%d
	`, where, idx-1, idx)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil data dokter: %w", err)
	}
	defer rows.Close()

	var doctors []Doctor
	for rows.Next() {
		var d Doctor
		if err := rows.Scan(&d.ID, &d.UserID, &d.DoctorCode, &d.FullName, &d.Specialization,
			&d.LicenseNumber, &d.Phone, &d.Email, &d.ConsultationFee, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("gagal scan dokter: %w", err)
		}
		doctors = append(doctors, d)
	}
	return doctors, total, nil
}

func (r *Repository) GetByID(id int) (*Doctor, error) {
	d := &Doctor{}
	query := `SELECT id, user_id, doctor_code, full_name, specialization,
		COALESCE(license_number,''), COALESCE(phone,''), COALESCE(email,''),
		consultation_fee, is_active, created_at, updated_at
		FROM doctors WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&d.ID, &d.UserID, &d.DoctorCode, &d.FullName, &d.Specialization,
		&d.LicenseNumber, &d.Phone, &d.Email, &d.ConsultationFee, &d.IsActive, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("dokter tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil data dokter: %w", err)
	}
	return d, nil
}

func (r *Repository) Create(d *Doctor) error {
	query := `INSERT INTO doctors (doctor_code, full_name, specialization, license_number, phone, email, consultation_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, d.DoctorCode, d.FullName, d.Specialization,
		d.LicenseNumber, d.Phone, d.Email, d.ConsultationFee).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
}

func (r *Repository) Update(id int, d *Doctor) error {
	query := `UPDATE doctors SET doctor_code=$1, full_name=$2, specialization=$3,
		license_number=$4, phone=$5, email=$6, consultation_fee=$7, is_active=$8
		WHERE id=$9`
	_, err := r.db.Exec(query, d.DoctorCode, d.FullName, d.Specialization,
		d.LicenseNumber, d.Phone, d.Email, d.ConsultationFee, d.IsActive, id)
	return err
}

func (r *Repository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM doctors WHERE id=$1", id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("dokter tidak ditemukan")
	}
	return nil
}

func (r *Repository) CodeExists(code string, excludeID int) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM doctors WHERE doctor_code = $1 AND id != $2"
	err := r.db.QueryRow(query, code, excludeID).Scan(&count)
	return count > 0, err
}

func (r *Repository) GenerateCode() (string, error) {
	var lastCode sql.NullString
	err := r.db.QueryRow("SELECT doctor_code FROM doctors ORDER BY id DESC LIMIT 1").Scan(&lastCode)
	if err != nil || !lastCode.Valid {
		return "D001", nil
	}
	code := strings.TrimPrefix(lastCode.String, "D")
	num := 0
	fmt.Sscanf(code, "%d", &num)
	return fmt.Sprintf("D%03d", num+1), nil
}

func (r *Repository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM doctors WHERE is_active = true").Scan(&count)
	return count, err
}
