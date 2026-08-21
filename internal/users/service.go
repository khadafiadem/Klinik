package users

import (
	"database/sql"
	"fmt"

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
