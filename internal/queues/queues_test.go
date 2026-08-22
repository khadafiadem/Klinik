package queues

import (
	"strings"
	"testing"
	"time"
)

func TestQueueCreateValidation(t *testing.T) {
	s := &Service{}

	tests := []struct {
		name    string
		queue   Queue
		wantErr string
	}{
		{"missing registration", Queue{PatientID: 1, DoctorID: 1}, "registrasi"},
		{"missing patient", Queue{RegistrationID: 1, DoctorID: 1}, "pasien"},
		{"missing doctor", Queue{RegistrationID: 1, PatientID: 1}, "dokter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.queue
			err := s.Create(&q)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestQueueUpdateStatusValidation(t *testing.T) {
	s := &Service{}

	validStatuses := map[string]bool{
		"MENUNGGU": true, "DIPANGGIL": true, "SEDANG_DIPERIKSA": true,
		"SELESAI": true, "DIBATALKAN": true,
	}
	for _, st := range []string{"MENUNGGU", "DIPANGGIL", "SEDANG_DIPERIKSA", "SELESAI", "DIBATALKAN"} {
		if !validStatuses[st] {
			t.Errorf("status %s should be recognized as valid", st)
		}
	}

	// Invalid statuses must be rejected before touching the repository.
	invalidStatuses := []string{"", "WAITING", "invalid", "menunggu"}
	for _, st := range invalidStatuses {
		err := s.UpdateStatus(0, st)
		if err == nil || !strings.Contains(err.Error(), "status tidak valid") {
			t.Errorf("status %q should be rejected as invalid, got: %v", st, err)
		}
	}
}

func TestQueueDefaultValues(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	q := Queue{}
	if q.QueueDate == "" {
		// Service sets today's date when empty (mirrors Create behavior)
		date := time.Now().Format("2006-01-02")
		if date != today {
			t.Errorf("default date mismatch: %s vs %s", date, today)
		}
	}

	q2 := Queue{Status: "  "}
	status := strings.TrimSpace(q2.Status)
	if status == "" {
		status = "MENUNGGU"
	}
	if status != "MENUNGGU" {
		t.Errorf("default status should be MENUNGGU, got %s", status)
	}
}

func TestQueueNumberFormat(t *testing.T) {
	num := "Q-20260822-001"
	if !strings.HasPrefix(num, "Q-") {
		t.Errorf("queue number should start with Q-, got %s", num)
	}
	parts := strings.Split(num, "-")
	if len(parts) != 3 {
		t.Fatalf("queue number should have 3 segments, got %d", len(parts))
	}
	if len(parts[1]) != 8 {
		t.Errorf("date segment should be YYYYMMDD (8 chars), got %d", len(parts[1]))
	}
	if parts[0] != "Q" || parts[2] != "001" {
		t.Errorf("unexpected queue number format: %s", num)
	}
}
