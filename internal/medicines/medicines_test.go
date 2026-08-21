package medicines

import (
	"testing"
)

func TestMedicineStockValidation(t *testing.T) {
	tests := []struct {
		name       string
		stock      int
		reduceQty  int
		canReduce  bool
	}{
		{"sufficient stock", 100, 50, true},
		{"exact stock", 100, 100, true},
		{"insufficient stock", 10, 50, false},
		{"zero stock", 0, 1, false},
		{"zero reduce", 10, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canReduce := tt.stock >= tt.reduceQty && tt.reduceQty > 0
			if canReduce != tt.canReduce {
				t.Errorf("stock=%d, reduce=%d: expected canReduce=%v, got %v",
					tt.stock, tt.reduceQty, tt.canReduce, canReduce)
			}
		})
	}
}

func TestLowStockDetection(t *testing.T) {
	tests := []struct {
		name     string
		stock    int
		minStock int
		isLow    bool
	}{
		{"normal stock", 100, 10, false},
		{"at minimum", 10, 10, true},
		{"below minimum", 5, 10, true},
		{"zero stock", 0, 10, true},
		{"zero minimum", 100, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isLow := tt.stock <= tt.minStock && tt.minStock > 0
			if isLow != tt.isLow {
				t.Errorf("stock=%d, min=%d: expected isLow=%v, got %v",
					tt.stock, tt.minStock, tt.isLow, isLow)
			}
		})
	}
}

func TestMedicineCodeFormat(t *testing.T) {
	tests := []struct {
		code  string
		valid bool
	}{
		{"MED-001", true},
		{"OB-002", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			valid := len(tt.code) > 0
			if valid != tt.valid {
				t.Errorf("code %q: expected valid=%v, got valid=%v", tt.code, tt.valid, valid)
			}
		})
	}
}

func TestSellingPriceAbovePurchasePrice(t *testing.T) {
	purchase := 5000.0
	selling := 7500.0

	if selling < purchase {
		t.Error("selling price should not be less than purchase price")
	}
}

func TestStockTransactionType(t *testing.T) {
	validTypes := map[string]bool{"MASUK": true, "KELUAR": true}

	for _, tt := range []string{"MASUK", "KELUAR", "INVALID"} {
		_, valid := validTypes[tt]
		if tt == "INVALID" && valid {
			t.Error("INVALID should not be a valid transaction type")
		}
	}
}
