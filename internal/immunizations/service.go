package immunizations

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

func (s *Service) ListSchedules() ([]Schedule, error) {
	return s.repo.ListSchedules()
}

func (s *Service) ListImmunizations(page, limit int, search string) ([]Immunization, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	return s.repo.ListImmunizations(page, limit, search)
}

func (s *Service) Create(im *Immunization) error {
	if im.PatientID == 0 {
		return fmt.Errorf("pasien wajib dipilih")
	}
	if strings.TrimSpace(im.VaccinationDate) == "" {
		return fmt.Errorf("tanggal pemberian wajib diisi")
	}
	if strings.TrimSpace(im.VaccineName) == "" {
		return fmt.Errorf("vaksin wajib dipilih")
	}
	return s.repo.Create(im)
}

func (s *Service) Delete(id int) error {
	return s.repo.Delete(id)
}
