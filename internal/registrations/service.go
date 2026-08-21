package registrations

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

func (s *Service) GetAll(page, limit int, search string) ([]Registration, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.GetAll(page, limit, search)
}

func (s *Service) GetByID(id int) (*Registration, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(reg *Registration) error {
	if reg.PatientID == 0 {
		return fmt.Errorf("pasien wajib dipilih")
	}
	if reg.DoctorID == 0 {
		return fmt.Errorf("dokter wajib dipilih")
	}

	num, err := s.repo.GenerateNumber()
	if err != nil {
		return err
	}
	reg.RegistrationNumber = num

	if strings.TrimSpace(reg.RegistrationType) == "" {
		reg.RegistrationType = "UMUM"
	}
	if strings.TrimSpace(reg.Status) == "" {
		reg.Status = "TERDAFTAR"
	}

	return s.repo.Create(reg)
}

func (s *Service) UpdateStatus(id int, status string) error {
	return s.repo.UpdateStatus(id, status)
}

func (s *Service) Count() (int, error) {
	return s.repo.Count()
}
