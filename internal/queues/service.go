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
	if q.PatientID == nil || *q.PatientID == 0 {
		return fmt.Errorf("pasien wajib dipilih")
	}
	if q.DoctorID == nil || *q.DoctorID == 0 {
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
	if strings.TrimSpace(q.QueueSource) == "" {
		q.QueueSource = "ADMIN"
	}

	return s.repo.Create(q)
}

func (s *Service) CreateKiosk() (*Queue, error) {
	q := &Queue{
		QueueDate:  time.Now().Format("2006-01-02"),
		Status:     "MENUNGGU",
		QueueSource: "KIOSK",
	}

	num, err := s.repo.GenerateKioskNumber()
	if err != nil {
		return nil, err
	}
	q.QueueNumber = num

	if err := s.repo.Create(q); err != nil {
		return nil, err
	}

	return q, nil
}

func (s *Service) GetKioskQueuesToday() ([]Queue, error) {
	return s.repo.GetKioskQueuesToday()
}

func (s *Service) LinkToRegistration(queueID int, registrationID, patientID, doctorID int) error {
	return s.repo.LinkToRegistration(queueID, registrationID, patientID, doctorID)
}

func (s *Service) GetMonitorData(date string) ([]Queue, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	return s.repo.GetMonitorData(date)
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

func (s *Service) UpdateStatusCalledBy(id int, status string, calledBy int) error {
	validStatuses := map[string]bool{
		"MENUNGGU": true, "DIPANGGIL": true, "SEDANG_DIPERIKSA": true,
		"SELESAI": true, "DIBATALKAN": true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("status tidak valid: %s", status)
	}
	return s.repo.UpdateStatusCalledBy(id, status, calledBy)
}

func (s *Service) GetTodayStats() (waiting, inProgress, completed int, err error) {
	return s.repo.GetTodayStats()
}
