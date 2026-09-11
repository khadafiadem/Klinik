package handler

import (
	"net/http"
	"strconv"
	"strings"

	"klinik-app/internal/auth"
	"klinik-app/internal/clinic"
	"klinik-app/internal/patients"
)

func (h *WebHandler) PatientsList(w http.ResponseWriter, r *http.Request, user *auth.User) {
	search := r.URL.Query().Get("search")
	patientsList, _, _ := h.patientSvc.GetAll(1, 100, search)
	RenderTemplate(w, r, "patients/list", TemplateData{User: user, Data: patientsList})
}

func (h *WebHandler) PatientForm(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/patients/")
	path = strings.TrimSuffix(path, "/edit")
	path = strings.TrimSuffix(path, "/new")

	var p *patients.Patient
	if path != "" && path != "new" {
		id, err := strconv.Atoi(path)
		if err == nil {
			p, _ = h.patientSvc.GetByID(id)
		}
	}
	if p == nil {
		p = &patients.Patient{}
	}
	providers := h.insuranceProvidersFor(p.InsuranceName)

	RenderTemplate(w, r, "patients/form", TemplateData{User: user, Data: p, Data2: providers})
}

func (h *WebHandler) PatientSave(w http.ResponseWriter, r *http.Request, user *auth.User) {
	idStr := r.FormValue("id")
	isEdit := idStr != ""

	p := &patients.Patient{
		FullName:                 r.FormValue("full_name"),
		NIK:                      r.FormValue("nik"),
		Gender:                   r.FormValue("gender"),
		DateOfBirth:              r.FormValue("date_of_birth"),
		BloodType:                r.FormValue("blood_type"),
		Phone:                    r.FormValue("phone"),
		Email:                    r.FormValue("email"),
		Address:                  r.FormValue("address"),
		City:                     r.FormValue("city"),
		Province:                 r.FormValue("province"),
		PostalCode:               r.FormValue("postal_code"),
		EmergencyContactName:     r.FormValue("emergency_contact_name"),
		EmergencyContactPhone:    r.FormValue("emergency_contact_phone"),
		EmergencyContactRelation: r.FormValue("emergency_contact_relation"),
		InsuranceName:            r.FormValue("insurance_name"),
		InsuranceNumber:          r.FormValue("insurance_number"),
		Allergies:                r.FormValue("allergies"),
		Notes:                    r.FormValue("notes"),
		IsActive:                 true,
	}

	var err error
	if isEdit {
		id, _ := strconv.Atoi(idStr)
		err = h.patientSvc.Update(id, p)
	} else {
		err = h.patientSvc.Create(p)
	}

	if err != nil {
		providers := h.insuranceProvidersFor(p.InsuranceName)
		RenderTemplate(w, r, "patients/form", TemplateData{User: user, Data: p, Data2: providers, Error: err.Error()})
		return
	}

	http.Redirect(w, r, "/patients", http.StatusSeeOther)
}

func (h *WebHandler) PatientDelete(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/patients/")
	path = strings.TrimSuffix(path, "/delete")

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Redirect(w, r, "/patients", http.StatusSeeOther)
		return
	}

	h.patientSvc.Delete(id)
	http.Redirect(w, r, "/patients", http.StatusSeeOther)
}

// insuranceProvidersFor mengembalikan daftar asuransi untuk dropdown form pasien,
// memastikan nilai yang sudah tersimpan tetap muncul meski tidak ada di daftar master.
func (h *WebHandler) insuranceProvidersFor(current string) []clinic.InsuranceProvider {
	providers, _ := h.clinicSvc.ListInsuranceProviders()
	if current == "" {
		return providers
	}
	for _, p := range providers {
		if strings.EqualFold(p.Name, current) {
			return providers
		}
	}
	return append(providers, clinic.InsuranceProvider{Name: current, SortOrder: 99})
}
