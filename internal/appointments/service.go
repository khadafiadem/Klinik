package appointments

import (
	"database/sql"
	"fmt"
	"strings"
)

type Service struct {
	repo *Repository
}

var validStatuses = map[string]bool{
	"TERJADWAL":   true,
	"DIKONFIRMASI": true,
	"HADIR":       true,
	"SELESAI":     true,
	"DIBATALKAN":  true,
}

func NewService(db *sql.DB) *Service {
	return &Service{repo: NewRepository(db)}
}

func (s *Service) GetAll(page, limit int, search, date string) ([]Appointment, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	return s.repo.GetAll(page, limit, search, date)
}

func (s *Service) GetUpcoming(today string) ([]Appointment, error) {
	return s.repo.GetUpcoming(today)
}

func (s *Service) GetByID(id int) (*Appointment, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(a *Appointment) error {
	if a.PatientID == 0 {
		return fmt.Errorf("pasien wajib dipilih")
	}
	if a.DoctorID == 0 {
		return fmt.Errorf("dokter wajib dipilih")
	}
	if strings.TrimSpace(a.AppointmentDate) == "" {
		return fmt.Errorf("tanggal janji temu wajib diisi")
	}
	if strings.TrimSpace(a.StartTime) == "" {
		a.StartTime = "09:00"
	}

	num, err := s.repo.GenerateNumber()
	if err != nil {
		return err
	}
	a.AppointmentNumber = num

	if strings.TrimSpace(a.Status) == "" {
		a.Status = "TERJADWAL"
	}
	return s.repo.Create(a)
}

func (s *Service) Update(id int, a *Appointment) error {
	if a.PatientID == 0 {
		return fmt.Errorf("pasien wajib dipilih")
	}
	if a.DoctorID == 0 {
		return fmt.Errorf("dokter wajib dipilih")
	}
	if strings.TrimSpace(a.AppointmentDate) == "" {
		return fmt.Errorf("tanggal janji temu wajib diisi")
	}
	return s.repo.Update(id, a)
}

func (s *Service) UpdateStatus(id int, status string) error {
	if !validStatuses[status] {
		return fmt.Errorf("status janji temu tidak valid")
	}
	return s.repo.UpdateStatus(id, status)
}

func (s *Service) Delete(id int) error {
	return s.repo.Delete(id)
}

func (s *Service) CountUpcoming(today string) (int, error) {
	return s.repo.CountUpcoming(today)
}