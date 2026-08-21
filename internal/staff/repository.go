package staff

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

func (r *Repository) GetAll(page, limit int, search string) ([]Staff, int, error) {
	offset := (page - 1) * limit
	args := []interface{}{}
	where := ""

	if search != "" {
		where = "WHERE (s.full_name ILIKE $1 OR s.staff_code ILIKE $1 OR s.position ILIKE $1)"
		args = append(args, "%"+search+"%")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM staff s %s", where)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	idx := len(args)
	query := fmt.Sprintf(`
		SELECT s.id, s.user_id, s.staff_code, s.full_name, s.position,
			COALESCE(s.phone,''), COALESCE(s.email,''),
			s.is_active, s.created_at, s.updated_at
		FROM staff s %s
		ORDER BY s.id DESC
		LIMIT $%d OFFSET $%d
	`, where, idx-1, idx)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var staffList []Staff
	for rows.Next() {
		var s Staff
		if err := rows.Scan(&s.ID, &s.UserID, &s.StaffCode, &s.FullName, &s.Position,
			&s.Phone, &s.Email, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		staffList = append(staffList, s)
	}
	return staffList, total, nil
}

func (r *Repository) GetByID(id int) (*Staff, error) {
	s := &Staff{}
	query := `SELECT id, user_id, staff_code, full_name, position,
		COALESCE(phone,''), COALESCE(email,''),
		is_active, created_at, updated_at
		FROM staff WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&s.ID, &s.UserID, &s.StaffCode, &s.FullName, &s.Position,
		&s.Phone, &s.Email, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("staff tidak ditemukan")
		}
		return nil, err
	}
	return s, nil
}

func (r *Repository) Create(s *Staff) error {
	query := `INSERT INTO staff (staff_code, full_name, position, phone, email)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, s.StaffCode, s.FullName, s.Position, s.Phone, s.Email).
		Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *Repository) Update(id int, s *Staff) error {
	query := `UPDATE staff SET staff_code=$1, full_name=$2, position=$3, phone=$4, email=$5, is_active=$6 WHERE id=$7`
	_, err := r.db.Exec(query, s.StaffCode, s.FullName, s.Position, s.Phone, s.Email, s.IsActive, id)
	return err
}

func (r *Repository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM staff WHERE id=$1", id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("staff tidak ditemukan")
	}
	return nil
}

func (r *Repository) CodeExists(code string, excludeID int) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM staff WHERE staff_code = $1 AND id != $2", code, excludeID).Scan(&count)
	return count > 0, err
}

func (r *Repository) GenerateCode() (string, error) {
	var lastCode sql.NullString
	err := r.db.QueryRow("SELECT staff_code FROM staff ORDER BY id DESC LIMIT 1").Scan(&lastCode)
	if err != nil || !lastCode.Valid {
		return "S001", nil
	}
	code := strings.TrimPrefix(lastCode.String, "S")
	num := 0
	fmt.Sscanf(code, "%d", &num)
	return fmt.Sprintf("S%03d", num+1), nil
}

func (r *Repository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM staff WHERE is_active = true").Scan(&count)
	return count, err
}
