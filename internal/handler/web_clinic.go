package handler

import (
	"net/http"
	"strconv"
	"strings"

	"klinik-app/internal/auth"
	"klinik-app/internal/logger"
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

	providers, _ := h.clinicSvc.ListInsuranceProviders()

	success, errorMsg := "", ""
	switch r.URL.Query().Get("msg") {
	case "providers-added":
		success = "Asuransi berhasil ditambahkan"
	case "provider-deleted":
		success = "Asuransi berhasil dihapus"
	case "provider-fail":
		errorMsg = "Gagal memproses asuransi"
	case "provider-bpjs":
		errorMsg = "Asuransi BPJS tidak dapat dihapus"
	}

	RenderTemplate(w, r, "clinic/settings", TemplateData{User: user, Data: settings, Data2: providers, Success: success, Error: errorMsg})
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
		providers, _ := h.clinicSvc.ListInsuranceProviders()
		RenderTemplate(w, r, "clinic/settings", TemplateData{User: user, Data: settings, Data2: providers, Error: err.Error()})
		return
	}

	providers, _ := h.clinicSvc.ListInsuranceProviders()
	RenderTemplate(w, r, "clinic/settings", TemplateData{User: user, Data: settings, Data2: providers, Success: "Pengaturan berhasil disimpan"})
}

func (h *WebHandler) InsuranceProviderAdd(w http.ResponseWriter, r *http.Request, user *auth.User) {
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		backToSettings(w, r, "provider-fail")
		return
	}
	if err := h.clinicSvc.AddInsuranceProvider(name); err != nil {
		logger.Error.Printf("Tambah asuransi %q gagal: %v", name, err)
		backToSettings(w, r, "provider-fail")
		return
	}
	backToSettings(w, r, "providers-added")
}

func (h *WebHandler) InsuranceProviderDelete(w http.ResponseWriter, r *http.Request, user *auth.User) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil || id <= 0 {
		backToSettings(w, r, "provider-fail")
		return
	}
	providers, _ := h.clinicSvc.ListInsuranceProviders()
	for _, p := range providers {
		if p.ID == id && strings.EqualFold(p.Name, "BPJS") {
			backToSettings(w, r, "provider-bpjs")
			return
		}
	}
	if err := h.clinicSvc.DeleteInsuranceProvider(id); err != nil {
		logger.Error.Printf("Hapus asuransi id %d gagal: %v", id, err)
		backToSettings(w, r, "provider-fail")
		return
	}
	backToSettings(w, r, "provider-deleted")
}

func backToSettings(w http.ResponseWriter, r *http.Request, msg string) {
	http.Redirect(w, r, "/clinic-settings?msg="+msg, http.StatusSeeOther)
}
