package handler

import (
	"net/http"
	"strconv"
	"strings"

	"klinik-app/internal/auth"
	"klinik-app/internal/doctors"
)

func (h *WebHandler) DoctorsList(w http.ResponseWriter, r *http.Request, user *auth.User) {
	search := r.URL.Query().Get("search")
	doctorsList, _, err := h.doctorSvc.GetAll(1, 100, search)
	if err != nil {
		RenderTemplate(w, r, "doctors/list", TemplateData{User: user, Error: err.Error()})
		return
	}
	RenderTemplate(w, r, "doctors/list", TemplateData{User: user, Data: doctorsList})
}

func (h *WebHandler) DoctorForm(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/doctors/")
	path = strings.TrimSuffix(path, "/edit")
	path = strings.TrimSuffix(path, "/new")

	var d *doctors.Doctor
	if path != "" && path != "new" {
		id, err := strconv.Atoi(path)
		if err == nil {
			d, _ = h.doctorSvc.GetByID(id)
		}
	}

	RenderTemplate(w, r, "doctors/form", TemplateData{User: user, Data: d})
}

func (h *WebHandler) DoctorSave(w http.ResponseWriter, r *http.Request, user *auth.User) {
	idStr := r.FormValue("id")
	isEdit := idStr != ""

	d := &doctors.Doctor{
		DoctorCode:     strings.TrimSpace(r.FormValue("doctor_code")),
		FullName:       r.FormValue("full_name"),
		Specialization: r.FormValue("specialization"),
		LicenseNumber:  r.FormValue("license_number"),
		Phone:          r.FormValue("phone"),
		Email:          r.FormValue("email"),
	}

	if v, err := strconv.ParseFloat(r.FormValue("consultation_fee"), 64); err == nil {
		d.ConsultationFee = v
	}
	d.IsActive = r.FormValue("is_active") != "false"

	var err error
	if isEdit {
		id, _ := strconv.Atoi(idStr)
		err = h.doctorSvc.Update(id, d)
	} else {
		err = h.doctorSvc.Create(d)
	}

	if err != nil {
		RenderTemplate(w, r, "doctors/form", TemplateData{User: user, Data: d, Error: err.Error()})
		return
	}

	http.Redirect(w, r, "/doctors", http.StatusSeeOther)
}

func (h *WebHandler) DoctorDelete(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/doctors/")
	path = strings.TrimSuffix(path, "/delete")

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Redirect(w, r, "/doctors", http.StatusSeeOther)
		return
	}

	h.doctorSvc.Delete(id)
	http.Redirect(w, r, "/doctors", http.StatusSeeOther)
}
