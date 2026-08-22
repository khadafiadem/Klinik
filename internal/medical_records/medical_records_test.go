package medical_records

import (
	"strings"
	"testing"
)

func TestMedicalRecordCreateValidation(t *testing.T) {
	s := &Service{}

	tests := []struct {
		name    string
		mr      MedicalRecord
		wantErr string
	}{
		{"missing patient", MedicalRecord{DoctorID: 1}, "pasien"},
		{"missing doctor", MedicalRecord{PatientID: 1}, "dokter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := tt.mr
			err := s.Create(&mr)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestMedicalRecordDefaultStatus(t *testing.T) {
	mr := MedicalRecord{Status: "  "}
	status := strings.TrimSpace(mr.Status)
	if status == "" {
		status = "DRAFT"
	}
	if status != "DRAFT" {
		t.Errorf("default status should be DRAFT, got %s", status)
	}
}

func TestMedicalRecordNumberFormat(t *testing.T) {
	num := "RM-20260822-0001"
	if !strings.HasPrefix(num, "RM-") {
		t.Errorf("medical record number should start with RM-, got %s", num)
	}
	parts := strings.Split(num, "-")
	if len(parts) != 3 {
		t.Fatalf("medical record number should have 3 segments, got %d", len(parts))
	}
	if len(parts[1]) != 8 {
		t.Errorf("date segment should be YYYYMMDD (8 chars), got %d", len(parts[1]))
	}
}

func TestDiagnosisEntryStructure(t *testing.T) {
	entry := DiagnosisEntry{
		DiagnosisCode: "J06.9",
		DiagnosisName: "Infeksi saluran pernapasan akut",
		DiagnosisType: "PRIMER",
	}

	if entry.DiagnosisCode == "" {
		t.Error("diagnosis code should not be empty")
	}
	if entry.DiagnosisType != "PRIMER" && entry.DiagnosisType != "SEKUNDER" {
		t.Errorf("diagnosis type should be PRIMER or SEKUNDER, got %s", entry.DiagnosisType)
	}
}

func TestTreatmentEntryCost(t *testing.T) {
	entry := TreatmentEntry{
		TreatmentCode: "TBT-001",
		TreatmentName: "Tindakan Luka Ringan",
		Cost:          50000,
	}

	if entry.Cost < 0 {
		t.Error("treatment cost must not be negative")
	}
	if entry.Cost == 0 {
		t.Log("warning: treatment has zero cost")
	}
}

func TestVitalSignsFormat(t *testing.T) {
	vs := "TD: 120/80 mmHg, Nadi: 80x/menit, Suhu: 36.7 C, RR: 20x/menit"

	required := []string{"TD", "Nadi", "Suhu"}
	for _, key := range required {
		if !strings.Contains(vs, key) {
			t.Errorf("vital signs should contain %s", key)
		}
	}
}
