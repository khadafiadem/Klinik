package handler

import (
	"net/http"
	"strconv"

	"klinik-app/internal/auth"
)

func (h *WebHandler) ClinicSettings(w http.ResponseWriter, r *http.Request, user *auth.User) {
	if r.Method == http.MethodPost {
		h.clinicSettingsPost(w, r, user)
		return
	}

	settings, err := h.clinicSvc.Get()
	if err != nil {
		RenderTemplate(w, r, "clinic/settings", TemplateData{User: user, Error: err.Error()})
		return
	}

	RenderTemplate(w, r, "clinic/settings", TemplateData{User: user, Data: settings})
}

func (h *WebHandler) clinicSettingsPost(w http.ResponseWriter, r *http.Request, user *auth.User) {
	settings, _ := h.clinicSvc.Get()
	if settings == nil {
		RenderTemplate(w, r, "clinic/settings", TemplateData{User: user, Error: "Gagal memuat pengaturan"})
		return
	}

	settings.ClinicName = r.FormValue("clinic_name")
	settings.ClinicAddress = r.FormValue("clinic_address")
	settings.ClinicPhone = r.FormValue("clinic_phone")
	settings.ClinicEmail = r.FormValue("clinic_email")
	settings.OpeningTime = r.FormValue("opening_time")
	settings.ClosingTime = r.FormValue("closing_time")
	settings.Currency = r.FormValue("currency")

	if v, err := strconv.Atoi(r.FormValue("max_patients_per_day")); err == nil {
		settings.MaxPatientsPerDay = v
	}
	if v, err := strconv.ParseFloat(r.FormValue("registration_fee"), 64); err == nil {
		settings.RegistrationFee = v
	}
	if v, err := strconv.ParseFloat(r.FormValue("consultation_fee"), 64); err == nil {
		settings.ConsultationFee = v
	}
	if v, err := strconv.ParseFloat(r.FormValue("tax_percentage"), 64); err == nil {
		settings.TaxPercentage = v
	}

	if err := h.clinicSvc.Update(settings); err != nil {
		RenderTemplate(w, r, "clinic/settings", TemplateData{User: user, Error: err.Error()})
		return
	}

	RenderTemplate(w, r, "clinic/settings", TemplateData{User: user, Data: settings, Success: "Pengaturan berhasil disimpan"})
}
