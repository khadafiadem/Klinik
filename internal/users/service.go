package users

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"klinik-app/internal/auth"
)

type Service struct {
	repo *Repository
}

func NewService(db *sql.DB) *Service {
	return &Service{
		repo: NewRepository(db),
	}
}

func (s *Service) GetAll(page, limit int, search string) ([]User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.GetAll(page, limit, search)
}

func (s *Service) GetByID(id int) (*User, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(req CreateUserRequest) (*User, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" || req.FullName == "" {
		return nil, fmt.Errorf("semua field wajib diisi")
	}

	if len(req.Password) < 6 {
		return nil, fmt.Errorf("password minimal 6 karakter")
	}

	exists, err := s.repo.UsernameExists(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("username sudah digunakan")
	}

	exists, err = s.repo.EmailExists(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("email sudah digunakan")
	}

	if req.RoleID > 0 {
		_, err := s.repo.GetRoleByID(req.RoleID)
		if err != nil {
			return nil, fmt.Errorf("role tidak valid: %w", err)
		}
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
		IsActive: true,
	}

	if err := s.repo.Create(user, passwordHash, req.RoleID); err != nil {
		return nil, err
	}

	return s.repo.GetByID(user.ID)
}

func (s *Service) Update(id int, req UpdateUserRequest) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if req.Email != "" {
		exists, err := s.repo.EmailExists(req.Email)
		if err != nil {
			return err
		}
		if exists {
			user, _ := s.repo.GetByID(id)
			if user != nil && user.Email != req.Email {
				return fmt.Errorf("email sudah digunakan")
			}
		}
	}

	return s.repo.Update(id, req)
}

func (s *Service) Delete(id int) error {
	return s.repo.Delete(id)
}

func (s *Service) AssignRole(userID, roleID int) error {
	_, err := s.repo.GetByID(userID)
	if err != nil {
		return err
	}

	_, err = s.repo.GetRoleByID(roleID)
	if err != nil {
		return err
	}

	return s.repo.AssignRole(userID, roleID)
}

func (s *Service) GetAllRoles() ([]RoleInfo, error) {
	return s.repo.GetAllRoles()
}

func (s *Service) AssignPrimaryRole(userID, roleID int) error {
	if _, err := s.repo.GetByID(userID); err != nil {
		return fmt.Errorf("user tidak ditemukan")
	}
	if _, err := s.repo.GetRoleByID(roleID); err != nil {
		return fmt.Errorf("role tidak valid")
	}
	return s.repo.ReplaceUserRole(userID, roleID)
}

func (s *Service) SetPassword(id int, password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password minimal 6 karakter")
	}
	if _, err := s.repo.GetByID(id); err != nil {
		return fmt.Errorf("user tidak ditemukan")
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(id, hash)
}

// ChangeOwnPassword mengubah password oleh user itu sendiri.
// Password saat ini wajib diverifikasi terlebih dahulu.
func (s *Service) ChangeOwnPassword(id int, currentPassword, newPassword string) error {
	if err := ValidatePasswordChange(currentPassword, newPassword); err != nil {
		return err
	}

	hash, err := s.repo.GetPasswordHash(id)
	if err != nil {
		return fmt.Errorf("user tidak ditemukan")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(currentPassword)); err != nil {
		return fmt.Errorf("password lama tidak sesuai")
	}

	newHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.repo.UpdatePassword(id, newHash)
}

// ValidatePasswordChange memvalidasi aturan penggantian password.
func ValidatePasswordChange(currentPassword, newPassword string) error {
	if currentPassword == "" || newPassword == "" {
		return fmt.Errorf("password lama dan password baru wajib diisi")
	}
	if len(newPassword) < 6 {
		return fmt.Errorf("password baru minimal 6 karakter")
	}
	if newPassword == currentPassword {
		return fmt.Errorf("password baru tidak boleh sama dengan password lama")
	}
	return nil
}
