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
	patientSummary, _ := h.rptSvc.GetPatientSummary()
	RenderTemplate(w, r, "reports/patients", TemplateData{
		User: user,
		Data: patientSummary,
	})
}

func (h *WebHandler) ReportsRegistration(w http.ResponseWriter, r *http.Request, user *auth.User) {
	regSummary, _ := h.rptSvc.GetRegistrationSummary()
	RenderTemplate(w, r, "reports/registrations", TemplateData{
		User: user,
		Data: regSummary,
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
	revenueSummary, _ := h.rptSvc.GetRevenueSummary()
	paymentSummary, _ := h.rptSvc.GetPaymentSummary()
	data := map[string]interface{}{
		"Revenue":  revenueSummary,
		"Payments": paymentSummary,
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
