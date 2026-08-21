package users

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

func (r *Repository) GetAll(page, limit int, search string) ([]User, int, error) {
	offset := (page - 1) * limit

	searchPattern := ""
	args := []interface{}{}
	if search != "" {
		searchPattern = "WHERE (u.username ILIKE $1 OR u.email ILIKE $1 OR u.full_name ILIKE $1)"
		args = append(args, "%"+search+"%")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users u %s", searchPattern)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung user: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT u.id, u.username, u.email, u.full_name, u.is_active, u.last_login, u.created_at, u.updated_at
		FROM users u %s
		ORDER BY u.id DESC
		LIMIT $%d OFFSET $%d
	`, searchPattern, len(args)+1, len(args)+2)

	args = append(args, limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil data user: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.IsActive, &u.LastLogin, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("gagal scan user: %w", err)
		}
		users = append(users, u)
	}

	return users, total, nil
}

func (r *Repository) GetByID(id int) (*User, error) {
	user := &User{}
	query := `SELECT id, username, email, full_name, is_active, last_login, created_at, updated_at 
		FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName,
		&user.IsActive, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil data user: %w", err)
	}

	roles, err := r.GetUserRoles(id)
	if err == nil {
		user.Roles = roles
	}

	return user, nil
}

func (r *Repository) GetUserRoles(userID int) ([]RoleInfo, error) {
	query := `SELECT r.id, r.name FROM roles r JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil role user: %w", err)
	}
	defer rows.Close()

	var roles []RoleInfo
	for rows.Next() {
		var role RoleInfo
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, fmt.Errorf("gagal scan role: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *Repository) Create(user *User, passwordHash string, roleID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO users (username, email, password_hash, full_name) 
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	err = tx.QueryRow(query, user.Username, user.Email, passwordHash, user.FullName).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("gagal membuat user: %w", err)
	}

	if roleID > 0 {
		roleQuery := `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
		if _, err := tx.Exec(roleQuery, user.ID, roleID); err != nil {
			return fmt.Errorf("gagal assign role: %w", err)
		}
	}

	return tx.Commit()
}

func (r *Repository) Update(id int, req UpdateUserRequest) error {
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Email != "" {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, req.Email)
		argIndex++
	}
	if req.FullName != "" {
		setClauses = append(setClauses, fmt.Sprintf("full_name = $%d", argIndex))
		args = append(args, req.FullName)
		argIndex++
	}
	if req.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("tidak ada data yang diupdate")
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIndex)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("gagal update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal mendapatkan hasil update: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user tidak ditemukan")
	}

	return nil
}

func (r *Repository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal mendapatkan hasil hapus: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user tidak ditemukan")
	}

	return nil
}

func (r *Repository) UsernameExists(username string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE username = $1`
	err := r.db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("gagal mengecek username: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) EmailExists(email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = $1`
	err := r.db.QueryRow(query, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("gagal mengecek email: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) GetRoleByID(id int) (*RoleInfo, error) {
	role := &RoleInfo{}
	query := `SELECT id, name FROM roles WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&role.ID, &role.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil data role: %w", err)
	}
	return role, nil
}

func (r *Repository) AssignRole(userID, roleID int) error {
	query := `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(query, userID, roleID)
	return err
}
