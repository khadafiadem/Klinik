package prescriptions

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

func (s *Service) GetAll(page, limit int, search, status string) ([]Prescription, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.GetAll(page, limit, search, status)
}

func (s *Service) GetByID(id int) (*Prescription, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(pr *Prescription) error {
	if pr.PatientID == 0 {
		return fmt.Errorf("pasien wajib dipilih")
	}
	if pr.DoctorID == 0 {
		return fmt.Errorf("dokter wajib dipilih")
	}
	if pr.MedicalRecordID == 0 {
		return fmt.Errorf("rekam medis wajib dipilih")
	}

	num, err := s.repo.GenerateNumber()
	if err != nil {
		return err
	}
	pr.PrescriptionNumber = num

	if strings.TrimSpace(pr.Status) == "" {
		pr.Status = "PENDING"
	}

	return s.repo.Create(pr)
}

// Process: PENDING -> PROCESSING (apotek mulai menyiapkan obat).
func (s *Service) Process(id int) error {
	return s.repo.UpdateStatusGuarded(id, "PROCESSING", []string{"PENDING"})
}

// Complete: PROCESSING -> COMPLETED sekaligus memotong stok obat (dispensasi).
func (s *Service) Complete(id int, userID int) error {
	return s.repo.CompleteWithDispense(id, userID)
}

// Cancel: PENDING/PROCESSING -> CANCELLED.
func (s *Service) Cancel(id int) error {
	return s.repo.UpdateStatusGuarded(id, "CANCELLED", []string{"PENDING", "PROCESSING"})
}

func (s *Service) AddItem(pi *PrescriptionItem) error {
	if pi.PrescriptionID == 0 {
		return fmt.Errorf("resep tidak valid")
	}
	if pi.Quantity < 1 {
		return fmt.Errorf("jumlah obat minimal 1")
	}

	pr, err := s.repo.GetByID(pi.PrescriptionID)
	if err != nil {
		return err
	}
	if pr.Status != "PENDING" {
		return fmt.Errorf("obat hanya dapat ditambahkan pada resep berstatus PENDING")
	}

	if pi.MedicineID != nil && *pi.MedicineID > 0 {
		name, code, err := s.repo.MedicineInfo(*pi.MedicineID)
		if err != nil {
			return err
		}
		pi.MedicineName = name
		pi.MedicineCode = code
	} else {
		pi.MedicineID = nil
	}

	if strings.TrimSpace(pi.MedicineName) == "" {
		return fmt.Errorf("obat wajib dipilih")
	}

	return s.repo.AddItem(pi)
}

func (s *Service) RemoveItem(id int) error {
	return s.repo.RemoveItem(id)
}

func (s *Service) Count() (int, error) {
	return s.repo.Count()
}
