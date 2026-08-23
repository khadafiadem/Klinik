package handler

import (
	"net/http"

	"klinik-app/internal/auth"
)

func (h *WebHandler) ReportsDashboard(w http.ResponseWriter, r *http.Request, user *auth.User) {
	patientSummary, _ := h.rptSvc.GetPatientSummary()
	regSummary, _ := h.rptSvc.GetRegistrationSummary()
	revenueSummary, _ := h.rptSvc.GetRevenueSummary()
	medicineSummary, _ := h.rptSvc.GetMedicineStockSummary()

	data := map[string]interface{}{
		"Patients":  patientSummary,
		"Regs":      regSummary,
		"Revenue":   revenueSummary,
		"Medicines": medicineSummary,
	}

	RenderTemplate(w, r, "reports/index", TemplateData{
		User: user,
		Data: data,
	})
}

func (h *WebHandler) ReportsPatient(w http.ResponseWriter, r *http.Request, user *auth.User) {
	from, to := parseReportRange(r)
	patientSummary, _ := h.rptSvc.GetPatientSummary()
	rows, err := h.rptSvc.GetPatientRows(from, to)
	if err != nil {
		rows = nil
	}
	RenderTemplate(w, r, "reports/patients", TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Summary": patientSummary,
			"Rows":    rows,
			"From":    from,
			"To":      to,
		},
	})
}

func (h *WebHandler) ReportsRegistration(w http.ResponseWriter, r *http.Request, user *auth.User) {
	from, to := parseReportRange(r)
	regSummary, _ := h.rptSvc.GetRegistrationSummary()
	rows, _ := h.rptSvc.GetVisitRows(from, to)
	RenderTemplate(w, r, "reports/registrations", TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Summary": regSummary,
			"Rows":    rows,
			"From":    from,
			"To":      to,
		},
	})
}

func (h *WebHandler) ReportsDoctorActivity(w http.ResponseWriter, r *http.Request, user *auth.User) {
	activity, _ := h.rptSvc.GetDoctorActivity()
	RenderTemplate(w, r, "reports/doctor_activity", TemplateData{
		User: user,
		Data: activity,
	})
}

func (h *WebHandler) ReportsRevenue(w http.ResponseWriter, r *http.Request, user *auth.User) {
	from, to := parseReportRange(r)
	revenueSummary, _ := h.rptSvc.GetRevenueSummary()
	paymentSummary, _ := h.rptSvc.GetPaymentSummary()
	detail, _ := h.rptSvc.GetRevenueDetail(from, to)
	data := map[string]interface{}{
		"Revenue":  revenueSummary,
		"Payments": paymentSummary,
		"Detail":   detail,
		"From":     from,
		"To":       to,
	}
	RenderTemplate(w, r, "reports/revenue", TemplateData{
		User: user,
		Data: data,
	})
}

func (h *WebHandler) ReportsMedicineStock(w http.ResponseWriter, r *http.Request, user *auth.User) {
	medSummary, _ := h.rptSvc.GetMedicineStockSummary()
	lowStock, _ := h.rptSvc.GetLowStockMedicines()
	data := map[string]interface{}{
		"Summary":  medSummary,
		"LowStock": lowStock,
	}
	RenderTemplate(w, r, "reports/medicine_stock", TemplateData{
		User: user,
		Data: data,
	})
}
