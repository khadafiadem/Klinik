package pain

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

// Save melakukan validasi lalu menyimpan (insert/update) penilaian nyeri.
func (s *Service) Save(p *PainAssessment) error {
	if p.MedicalRecordID == 0 {
		return fmt.Errorf("rekam medis wajib diisi")
	}
	if p.Intensity < 0 || p.Intensity > 10 {
		return fmt.Errorf("skala nyeri harus antara 0 sampai 10")
	}
	return s.repo.Upsert(p)
}

func (s *Service) GetByMedicalRecordID(mrID int) (*PainAssessment, error) {
	return s.repo.GetByMedicalRecordID(mrID)
}
