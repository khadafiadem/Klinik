package handler

import (
	"net/http"
	"strconv"
	"strings"

	"klinik-app/internal/auth"
	"klinik-app/internal/registrations"
)

func (h *WebHandler) RegistrationsList(w http.ResponseWriter, r *http.Request, user *auth.User) {
	search := r.URL.Query().Get("search")
	list, _, _ := h.regSvc.GetAll(1, 100, search)
	RenderTemplate(w, r, "registrations/list", TemplateData{
		User:   user,
		Data:   list,
		Search: search,
	})
}

func (h *WebHandler) RegistrationForm(w http.ResponseWriter, r *http.Request, user *auth.User) {
	doctorsList, _, _ := h.doctorSvc.GetAll(1, 100, "")
	patientsList, _, _ := h.patientSvc.GetAll(1, 1000, "")
	kioskQueues, _ := h.queueSvc.GetKioskQueuesToday()
	RenderTemplate(w, r, "registrations/form", TemplateData{
		User:  user,
		Data: map[string]interface{}{
			"Doctors":    doctorsList,
			"Patients":   patientsList,
			"KioskQueues": kioskQueues,
		},
	})
}

func (h *WebHandler) RegistrationSave(w http.ResponseWriter, r *http.Request, user *auth.User) {
	patientID, _ := strconv.Atoi(r.FormValue("patient_id"))
	doctorID, _ := strconv.Atoi(r.FormValue("doctor_id"))

	reg := &registrations.Registration{
		PatientID:        patientID,
		DoctorID:         doctorID,
		RegistrationDate: r.FormValue("registration_date"),
		RegistrationType: r.FormValue("registration_type"),
		Complaint:        r.FormValue("complaint"),
		Notes:            r.FormValue("notes"),
	}

	if err := h.regSvc.Create(reg); err != nil {
		doctorsList, _, _ := h.doctorSvc.GetAll(1, 100, "")
		patientsList, _, _ := h.patientSvc.GetAll(1, 1000, "")
		kioskQueues, _ := h.queueSvc.GetKioskQueuesToday()
		RenderTemplate(w, r, "registrations/form", TemplateData{
			User:  user,
			Error: err.Error(),
			Data: map[string]interface{}{
				"Doctors":    doctorsList,
				"Patients":   patientsList,
				"KioskQueues": kioskQueues,
				"Form":       reg,
			},
		})
		return
	}

	queueIDStr := r.FormValue("queue_id")
	if queueIDStr != "" {
		queueID, _ := strconv.Atoi(queueIDStr)
		if queueID > 0 {
			_ = h.queueSvc.LinkToRegistration(queueID, reg.ID, patientID, doctorID)
		}
	}

	http.Redirect(w, r, "/registrations", http.StatusSeeOther)
}

func (h *WebHandler) RegistrationAction(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/registrations/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.Redirect(w, r, "/registrations", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Redirect(w, r, "/registrations", http.StatusSeeOther)
		return
	}

	action := parts[1]
	if action == "cancel" {
		h.regSvc.UpdateStatus(id, "DIBATALKAN")
	}
	http.Redirect(w, r, "/registrations", http.StatusSeeOther)
}
