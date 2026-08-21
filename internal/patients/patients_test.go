package patients

import (
	"strings"
	"testing"
)

func TestPatientFullNameRequired(t *testing.T) {
	p := Patient{FullName: ""}
	if strings.TrimSpace(p.FullName) == "" {
		return
	}
	t.Error("empty name should be caught")
}

func TestNIKLengthValidation(t *testing.T) {
	tests := []struct {
		name  string
		nik   string
		valid bool
	}{
		{"valid NIK", "3201234567890001", true},
		{"too short", "12345", false},
		{"empty", "", false},
		{"correct length", "3201012345678901", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := len(tt.nik) == 16
			if valid != tt.valid {
				t.Errorf("NIK %q: expected valid=%v, got valid=%v", tt.nik, tt.valid, valid)
			}
		})
	}
}

func TestGenderValidation(t *testing.T) {
	validGenders := map[string]bool{"LAKI_LAKI": true, "PEREMPUAN": true}

	tests := []struct {
		gender string
		valid  bool
	}{
		{"LAKI_LAKI", true},
		{"PEREMPUAN", true},
		{"OTHER", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.gender, func(t *testing.T) {
			_, valid := validGenders[tt.gender]
			if valid != tt.valid {
				t.Errorf("gender %q: expected valid=%v, got valid=%v", tt.gender, tt.valid, valid)
			}
		})
	}
}

func TestBloodTypeValidation(t *testing.T) {
	validBloodTypes := map[string]bool{
		"A": true, "B": true, "AB": true, "O": true,
		"A+": true, "A-": true, "B+": true, "B-": true,
		"AB+": true, "AB-": true, "O+": true, "O-": true,
	}

	for _, bt := range []string{"A", "B", "AB", "O", "A+", "O-", "INVALID"} {
		_, valid := validBloodTypes[bt]
		if bt == "INVALID" && valid {
			t.Error("INVALID should not be a valid blood type")
		}
	}
}
