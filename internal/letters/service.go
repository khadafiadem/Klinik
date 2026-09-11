package letters

import (
	"database/sql"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(db *sql.DB) *Service {
	return &Service{repo: NewRepository(db)}
}

func (s *Service) ListActiveTemplates() ([]Template, error) {
	return s.repo.ListActiveTemplates()
}

func (s *Service) GetTemplateByCode(code string) (*Template, error) {
	return s.repo.GetTemplateByCode(code)
}

func (s *Service) CreateIssuance(iss *Issuance) error {
	if iss.TemplateID == 0 {
		return fmt.Errorf("template wajib dipilih")
	}
	if iss.PatientID == 0 {
		return fmt.Errorf("pasien wajib dipilih")
	}
	count, err := s.repo.CountIssuances(iss.TemplateID)
	if err != nil {
		return err
	}
	iss.LetterNumber = fmt.Sprintf("%04d/%d", count+1, iss.TemplateID)
	return s.repo.CreateIssuance(iss)
}

func (s *Service) GetIssuance(id int) (*Issuance, error) {
	return s.repo.GetIssuance(id)
}

func (s *Service) ListRecent(limit int) ([]Issuance, error) {
	if limit < 1 {
		limit = 20
	}
	return s.repo.ListRecent(limit)
}
