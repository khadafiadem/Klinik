package auth

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

func (r *Repository) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, email, password_hash, full_name, is_active, last_login, created_at, updated_at 
		FROM users WHERE username = $1`
	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FullName, &user.IsActive, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil data user: %w", err)
	}
	return user, nil
}

func (r *Repository) GetUserByID(id int) (*User, error) {
	user := &User{}
	query := `SELECT id, username, email, password_hash, full_name, is_active, last_login, created_at, updated_at 
		FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FullName, &user.IsActive, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil data user: %w", err)
	}
	roles, _ := r.GetUserRoles(id)
	user.Roles = roles
	return user, nil
}

func (r *Repository) CreateUser(user *User) error {
	query := `INSERT INTO users (username, email, password_hash, full_name) 
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, user.Username, user.Email, user.PasswordHash, user.FullName).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *Repository) UpdateLastLogin(userID int) error {
	query := `UPDATE users SET last_login = $1 WHERE id = $2`
	_, err := r.db.Exec(query, time.Now(), userID)
	return err
}

func (r *Repository) GetUserRoles(userID int) ([]Role, error) {
	query := `SELECT r.id, r.name, r.description, r.created_at, r.updated_at 
		FROM roles r 
		JOIN user_roles ur ON r.id = ur.role_id 
		WHERE ur.user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil role user: %w", err)
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, fmt.Errorf("gagal scan role: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *Repository) GetUserPermissions(userID int) ([]Permission, error) {
	query := `SELECT DISTINCT p.id, p.name, p.description, p.created_at 
		FROM permissions p 
		JOIN role_permissions rp ON p.id = rp.permission_id 
		JOIN user_roles ur ON rp.role_id = ur.role_id 
		WHERE ur.user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil permission user: %w", err)
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var perm Permission
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.Description, &perm.CreatedAt); err != nil {
			return nil, fmt.Errorf("gagal scan permission: %w", err)
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}

func (r *Repository) AssignRole(userID, roleID int) error {
	query := `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(query, userID, roleID)
	return err
}

func (r *Repository) UserExists(username, email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2`
	err := r.db.QueryRow(query, username, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("gagal mengecek keberadaan user: %w", err)
	}
	return count > 0, nil
}
