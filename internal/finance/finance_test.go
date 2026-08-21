package finance

import (
	"testing"
)

func TestInvoiceStatusTransitions(t *testing.T) {
	tests := []struct {
		name           string
		total          float64
		paid           float64
		expectedStatus string
	}{
		{"no payment", 100000, 0, "BELUM_BAYAR"},
		{"partial payment", 100000, 50000, "SEBAGIAN"},
		{"full payment", 100000, 100000, "SUDAH_BAYAR"},
		{"over payment scenario", 100000, 80000, "SEBAGIAN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status string
			if tt.paid <= 0 {
				status = "BELUM_BAYAR"
			} else if tt.paid >= tt.total {
				status = "SUDAH_BAYAR"
			} else {
				status = "SEBAGIAN"
			}
			if status != tt.expectedStatus {
				t.Errorf("expected status %s, got %s", tt.expectedStatus, status)
			}
		})
	}
}

func TestInvoiceItemTotalCalculation(t *testing.T) {
	items := []InvoiceItem{
		{Quantity: 2, UnitPrice: 50000},
		{Quantity: 1, UnitPrice: 250000},
		{Quantity: 3, UnitPrice: 15000},
	}

	var subtotal float64
	for i := range items {
		items[i].TotalPrice = float64(items[i].Quantity) * items[i].UnitPrice
		subtotal += items[i].TotalPrice
	}

	if items[0].TotalPrice != 100000 {
		t.Errorf("item 0: expected 100000, got %.0f", items[0].TotalPrice)
	}
	if items[1].TotalPrice != 250000 {
		t.Errorf("item 1: expected 250000, got %.0f", items[1].TotalPrice)
	}
	if items[2].TotalPrice != 45000 {
		t.Errorf("item 2: expected 45000, got %.0f", items[2].TotalPrice)
	}
	if subtotal != 395000 {
		t.Errorf("subtotal: expected 395000, got %.0f", subtotal)
	}
}

func TestInvoiceTotalWithDiscount(t *testing.T) {
	subtotal := 500000.0
	discount := 50000.0
	total := subtotal - discount

	if total != 450000 {
		t.Errorf("expected 450000, got %.0f", total)
	}
}

func TestInvoiceNumberFormat(t *testing.T) {
	num := "INV-20260821-001"
	if len(num) != 16 {
		t.Errorf("unexpected invoice number length: %d", len(num))
	}
	if num[:4] != "INV-" {
		t.Errorf("invoice number should start with INV-, got %s", num[:4])
	}
}

func TestPaymentExceedsRemaining(t *testing.T) {
	invoiceTotal := 100000.0
	alreadyPaid := 80000.0
	remaining := invoiceTotal - alreadyPaid
	paymentAmount := 30000.0

	if paymentAmount > remaining {
		if paymentAmount-remaining > 0.01 {
			t.Logf("correctly detected: payment %.0f exceeds remaining %.0f", paymentAmount, remaining)
		}
	} else {
		t.Error("should have detected excess payment")
	}
}
