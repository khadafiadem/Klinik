package medical_records

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

func (s *Service) GetAll(page, limit int, search string) ([]MedicalRecord, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.GetAll(page, limit, search)
}

func (s *Service) GetByID(id int) (*MedicalRecord, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(mr *MedicalRecord) error {
	if mr.PatientID == 0 {
		return fmt.Errorf("pasien wajib dipilih")
	}
	if mr.DoctorID == 0 {
		return fmt.Errorf("dokter wajib dipilih")
	}

	num, err := s.repo.GenerateNumber()
	if err != nil {
		return err
	}
	mr.MedicalRecordNumber = num

	if strings.TrimSpace(mr.Status) == "" {
		mr.Status = "DRAFT"
	}

	return s.repo.Create(mr)
}

func (s *Service) Update(mr *MedicalRecord) error {
	return s.repo.Update(mr)
}

func (s *Service) AddDiagnosis(medicalRecordID, diagnosisID int, diagnosisType, notes string) error {
	return s.repo.AddDiagnosis(medicalRecordID, diagnosisID, diagnosisType, notes)
}

func (s *Service) RemoveDiagnosis(id int) error {
	return s.repo.RemoveDiagnosis(id)
}

func (s *Service) AddTreatment(medicalRecordID, treatmentID int, cost float64, notes string) error {
	return s.repo.AddTreatment(medicalRecordID, treatmentID, cost, notes)
}

func (s *Service) RemoveTreatment(id int) error {
	return s.repo.RemoveTreatment(id)
}

func (s *Service) Count() (int, error) {
	return s.repo.Count()
}

type MRInput struct {
	PatientID       int
	DoctorID        int
	RegistrationID  *int
	ExaminationDate string
	ChiefComplaint  string
	VitalSigns      string
	Anamnesis       string
	PhysicalExam    string
	Notes           string
}

func (s *Service) CreateFromInput(input *MRInput) error {
	mr := &MedicalRecord{
		PatientID:            input.PatientID,
		DoctorID:             input.DoctorID,
		RegistrationID:       input.RegistrationID,
		ExaminationDate:      input.ExaminationDate,
		ChiefComplaint:       input.ChiefComplaint,
		VitalSigns:           input.VitalSigns,
		Anamnesis:            input.Anamnesis,
		PhysicalExamination:  input.PhysicalExam,
		Notes:                input.Notes,
	}
	return s.Create(mr)
}

func (s *Service) UpdateStatus(id int, status string) error {
	return s.repo.UpdateStatus(id, status)
}

func (s *Service) GetAllDiagnoses() ([]Diagnosis, error) {
	return s.repo.GetAllDiagnoses()
}

func (s *Service) GetDiagnosisByID(id int) (*Diagnosis, error) {
	return s.repo.GetDiagnosisByID(id)
}

func (s *Service) GetAllTreatments() ([]Treatment, error) {
	return s.repo.GetAllTreatments()
}

func (s *Service) GetTreatmentByID(id int) (*Treatment, error) {
	return s.repo.GetTreatmentByID(id)
}
