package reports

import (
	"testing"
)

func TestPatientSummaryStructure(t *testing.T) {
	ps := PatientSummary{
		TotalPatients:  100,
		MalePatients:   45,
		FemalePatients: 55,
		NewThisMonth:   12,
	}

	if ps.TotalPatients != ps.MalePatients+ps.FemalePatients {
		t.Error("total patients should equal male + female")
	}
}

func TestRevenueSummaryNonNegative(t *testing.T) {
	rs := RevenueSummary{
		TodayRevenue: 1500000,
		WeekRevenue:  8000000,
		MonthRevenue: 35000000,
		TotalRevenue: 500000000,
	}

	if rs.TodayRevenue < 0 {
		t.Error("today revenue should not be negative")
	}
	if rs.WeekRevenue < rs.TodayRevenue {
		t.Error("week revenue should be >= today revenue")
	}
	if rs.MonthRevenue < rs.WeekRevenue {
		t.Error("month revenue should be >= week revenue")
	}
	if rs.TotalRevenue < rs.MonthRevenue {
		t.Error("total revenue should be >= month revenue")
	}
}

func TestPaymentSummaryMethods(t *testing.T) {
	ps := PaymentSummary{
		CashTotal:      5000000,
		TransferTotal:  3000000,
		QRISTotal:      1000000,
		BPJSTotal:      2000000,
		OtherTotal:     500000,
		CompletedCount: 50,
		PendingCount:   5,
	}

	if ps.CompletedCount < 0 {
		t.Error("completed count should not be negative")
	}
	if ps.PendingCount < 0 {
		t.Error("pending count should not be negative")
	}
}
