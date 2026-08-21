package doctors

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

func (s *Service) GetAll(page, limit int, search string) ([]Doctor, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.GetAll(page, limit, search)
}

func (s *Service) GetByID(id int) (*Doctor, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(d *Doctor) error {
	if strings.TrimSpace(d.DoctorCode) == "" {
		code, err := s.repo.GenerateCode()
		if err != nil {
			return err
		}
		d.DoctorCode = code
	}

	if strings.TrimSpace(d.FullName) == "" || strings.TrimSpace(d.Specialization) == "" {
		return fmt.Errorf("nama lengkap dan spesialisasi wajib diisi")
	}

	exists, err := s.repo.CodeExists(d.DoctorCode, 0)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("kode dokter sudah digunakan")
	}

	return s.repo.Create(d)
}

func (s *Service) Update(id int, d *Doctor) error {
	if strings.TrimSpace(d.FullName) == "" || strings.TrimSpace(d.Specialization) == "" {
		return fmt.Errorf("nama lengkap dan spesialisasi wajib diisi")
	}

	exists, err := s.repo.CodeExists(d.DoctorCode, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("kode dokter sudah digunakan")
	}

	return s.repo.Update(id, d)
}

func (s *Service) Delete(id int) error {
	return s.repo.Delete(id)
}

func (s *Service) Count() (int, error) {
	return s.repo.Count()
}
