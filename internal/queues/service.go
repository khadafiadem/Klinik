package queues

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(db *sql.DB) *Service {
	return &Service{repo: NewRepository(db)}
}

func (s *Service) GetAllByDate(date string) ([]Queue, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	return s.repo.GetAllByDate(date)
}

func (s *Service) GetByID(id int) (*Queue, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(q *Queue) error {
	if q.RegistrationID == 0 {
		return fmt.Errorf("registrasi wajib dipilih")
	}
	if q.PatientID == 0 {
		return fmt.Errorf("pasien wajib dipilih")
	}
	if q.DoctorID == 0 {
		return fmt.Errorf("dokter wajib dipilih")
	}

	if q.QueueDate == "" {
		q.QueueDate = time.Now().Format("2006-01-02")
	}

	num, err := s.repo.GenerateNumber(q.QueueDate)
	if err != nil {
		return err
	}
	q.QueueNumber = num

	if strings.TrimSpace(q.Status) == "" {
		q.Status = "MENUNGGU"
	}

	return s.repo.Create(q)
}

func (s *Service) UpdateStatus(id int, status string) error {
	validStatuses := map[string]bool{
		"MENUNGGU": true, "DIPANGGIL": true, "SEDANG_DIPERIKSA": true,
		"SELESAI": true, "DIBATALKAN": true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %s", status)
	}
	return s.repo.UpdateStatus(id, status)
}

func (s *Service) GetTodayStats() (waiting, inProgress, completed int, err error) {
	return s.repo.GetTodayStats()
}
