package clinic

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

func (s *Service) Get() (*ClinicSettings, error) {
	return s.repo.Get()
}

func (s *Service) Update(settings *ClinicSettings) error {
	return s.repo.Update(settings)
}

func (s *Service) ListInsuranceProviders() ([]InsuranceProvider, error) {
	return s.repo.ListInsuranceProviders()
}

func (s *Service) AddInsuranceProvider(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("nama asuransi wajib diisi")
	}
	return s.repo.AddInsuranceProvider(name)
}

func (s *Service) DeleteInsuranceProvider(id int) error {
	return s.repo.DeleteInsuranceProvider(id)
}
