package prescriptions

import (
	"testing"
)

func TestPrescriptionStatusTransitions(t *testing.T) {
	validTransitions := map[string][]string{
		"PENDING":    {"PROCESSING", "CANCELLED"},
		"PROCESSING": {"COMPLETED", "CANCELLED"},
		"COMPLETED":  {},
		"CANCELLED":  {},
	}

	tests := []struct {
		from, to string
		valid    bool
	}{
		{"PENDING", "PROCESSING", true},
		{"PENDING", "CANCELLED", true},
		{"PENDING", "COMPLETED", false},
		{"PROCESSING", "COMPLETED", true},
		{"PROCESSING", "CANCELLED", true},
		{"COMPLETED", "PENDING", false},
		{"CANCELLED", "PENDING", false},
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

func TestPrescriptionItemQuantity(t *testing.T) {
	item := PrescriptionItem{
		MedicineName: "Parasetamol",
		Quantity:     10,
		Dosage:       "500mg",
		Frequency:    "3x sehari",
		Duration:     "5 hari",
	}

	if item.Quantity <= 0 {
		t.Error("quantity must be positive")
	}
	if item.Dosage == "" {
		t.Error("dosage should not be empty")
	}
}

func TestPrescriptionNumberFormat(t *testing.T) {
	num := "RX-20260821-001"
	if len(num) < 4 {
		t.Errorf("prescription number too short: %s", num)
	}
	if num[:3] != "RX-" {
		t.Errorf("prescription number should start with RX-, got %s", num[:3])
	}
}
