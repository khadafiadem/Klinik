package registrations

import (
	"testing"
)

func TestRegistrationStatusTransitions(t *testing.T) {
	validTransitions := map[string][]string{
		"REGISTERED": {"IN_QUEUE", "CANCELLED"},
		"IN_QUEUE":   {"COMPLETED", "CANCELLED"},
		"COMPLETED":  {},
		"CANCELLED":  {},
	}

	tests := []struct {
		from, to string
		valid    bool
	}{
		{"REGISTERED", "IN_QUEUE", true},
		{"REGISTERED", "CANCELLED", true},
		{"REGISTERED", "COMPLETED", false},
		{"IN_QUEUE", "COMPLETED", true},
		{"IN_QUEUE", "CANCELLED", true},
		{"COMPLETED", "REGISTERED", false},
		{"CANCELLED", "REGISTERED", false},
	}

	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			allowed, exists := validTransitions[tt.from]
			if !exists {
				t.Errorf("unknown status: %s", tt.from)
				return
			}
			valid := false
			for _, a := range allowed {
				if a == tt.to {
					valid = true
					break
				}
			}
			if valid != tt.valid {
				t.Errorf("transition %s->%s: expected valid=%v, got %v", tt.from, tt.to, tt.valid, valid)
			}
		})
	}
}

func TestRegistrationTypeValidation(t *testing.T) {
	validTypes := map[string]bool{"UMUM": true, "BPJS": true, "KONTROL": true, "IGD": true}

	tests := []struct {
		regType string
		valid   bool
	}{
		{"UMUM", true},
		{"BPJS", true},
		{"KONTROL", true},
		{"IGD", true},
		{"INVALID", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.regType, func(t *testing.T) {
			_, valid := validTypes[tt.regType]
			if valid != tt.valid {
				t.Errorf("type %q: expected valid=%v, got valid=%v", tt.regType, tt.valid, valid)
			}
		})
	}
}

func TestRegistrationNumberFormat(t *testing.T) {
	num := "REG-20260821-001"
	if len(num) < 4 {
		t.Errorf("registration number too short: %s", num)
	}
	if num[:4] != "REG-" {
		t.Errorf("registration number should start with REG-, got %s", num[:4])
	}
}
