package staff

import (
	"database/sql"
	"fmt"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(db *sql.DB) *Service {
	return &Service{repo: NewRepository(db)}
}

func (s *Service) GetAll(page, limit int, search string) ([]Staff, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.GetAll(page, limit, search)
}

func (s *Service) GetByID(id int) (*Staff, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(staff *Staff) error {
	if strings.TrimSpace(staff.StaffCode) == "" {
		code, err := s.repo.GenerateCode()
		if err != nil {
			return err
		}
		staff.StaffCode = code
	}

	if strings.TrimSpace(staff.FullName) == "" || strings.TrimSpace(staff.Position) == "" {
		return fmt.Errorf("nama lengkap dan posisi wajib diisi")
	}

	exists, err := s.repo.CodeExists(staff.StaffCode, 0)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("kode staff sudah digunakan")
	}

	return s.repo.Create(staff)
}

func (s *Service) Update(id int, staff *Staff) error {
	if strings.TrimSpace(staff.FullName) == "" || strings.TrimSpace(staff.Position) == "" {
		return fmt.Errorf("nama lengkap dan posisi wajib diisi")
	}

	exists, err := s.repo.CodeExists(staff.StaffCode, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("kode staff sudah digunakan")
	}

	return s.repo.Update(id, staff)
}

func (s *Service) Delete(id int) error {
	return s.repo.Delete(id)
}

func (s *Service) Count() (int, error) {
	return s.repo.Count()
}
