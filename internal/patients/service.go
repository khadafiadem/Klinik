package patients

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

func (s *Service) GetAll(page, limit int, search string) ([]Patient, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.GetAll(page, limit, search)
}

func (s *Service) GetByID(id int) (*Patient, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(p *Patient) error {
	if strings.TrimSpace(p.FullName) == "" {
		return fmt.Errorf("nama lengkap wajib diisi")
	}
	if p.Gender != "LAKI_LAKI" && p.Gender != "PEREMPUAN" {
		return fmt.Errorf("jenis kelamin tidak valid")
	}
	if p.DateOfBirth == "" {
		return fmt.Errorf("tanggal lahir wajib diisi")
	}

	mrn, err := s.repo.GenerateMRN()
	if err != nil {
		return err
	}
	p.MedicalRecordNumber = mrn

	return s.repo.Create(p)
}

func (s *Service) Update(id int, p *Patient) error {
	if strings.TrimSpace(p.FullName) == "" {
		return fmt.Errorf("nama lengkap wajib diisi")
	}
	if p.Gender != "LAKI_LAKI" && p.Gender != "PEREMPUAN" {
		return fmt.Errorf("jenis kelamin tidak valid")
	}

	return s.repo.Update(id, p)
}

func (s *Service) Delete(id int) error {
	return s.repo.Delete(id)
}

func (s *Service) Count() (int, error) {
	return s.repo.Count()
}
