package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"klinik-app/internal/appointments"
	"klinik-app/internal/auth"
)

func (h *WebHandler) AppointmentsList(w http.ResponseWriter, r *http.Request, user *auth.User) {
	search := r.URL.Query().Get("search")
	date := r.URL.Query().Get("date")

	list, _, err := h.apptSvc.GetAll(1, 100, search, date)
	if err != nil {
		RenderTemplate(w, r, "appointments/list", TemplateData{User: user, Error: err.Error()})
		return
	}

	RenderTemplate(w, r, "appointments/list", TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Rows":      list,
			"Date":      date,
			"Search":    search,
			"ErrStatus": r.URL.Query().Get("err") == "status",
		},
	})
}

func (h *WebHandler) AppointmentForm(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/appointments/")
	path = strings.TrimSuffix(path, "/edit")
	path = strings.TrimSuffix(path, "/new")

	var a *appointments.Appointment
	if path != "" && path != "new" {
		id, err := strconv.Atoi(path)
		if err == nil {
			a, _ = h.apptSvc.GetByID(id)
		}
	}
	if a == nil {
		a = &appointments.Appointment{}
	}

	doctorsList, _, _ := h.doctorSvc.GetAll(1, 200, "")
	patientsList, _, _ := h.patientSvc.GetAll(1, 1000, "")

	RenderTemplate(w, r, "appointments/form", TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Appt":     a,
			"Doctors":  doctorsList,
			"Patients": patientsList,
		},
	})
}

func (h *WebHandler) AppointmentSave(w http.ResponseWriter, r *http.Request, user *auth.User) {
	idStr := r.FormValue("id")
	isEdit := idStr != ""

	patientID, _ := strconv.Atoi(r.FormValue("patient_id"))
	doctorID, _ := strconv.Atoi(r.FormValue("doctor_id"))
	createdBy := user.ID

	a := &appointments.Appointment{
		PatientID:       patientID,
		DoctorID:        doctorID,
		AppointmentDate: strings.TrimSpace(r.FormValue("appointment_date")),
		StartTime:       strings.TrimSpace(r.FormValue("start_time")),
		Purpose:         strings.TrimSpace(r.FormValue("purpose")),
		Notes:           strings.TrimSpace(r.FormValue("notes")),
		Status:          strings.TrimSpace(r.FormValue("status")),
		CreatedBy:       &createdBy,
	}
	if a.Status == "" {
		a.Status = "TERJADWAL"
	}

	rerender := func(errMsg string) {
		doctorsList, _, _ := h.doctorSvc.GetAll(1, 200, "")
		patientsList, _, _ := h.patientSvc.GetAll(1, 1000, "")
		RenderTemplate(w, r, "appointments/form", TemplateData{
			User:  user,
			Error: errMsg,
			Data: map[string]interface{}{
				"Appt":     a,
				"Doctors":  doctorsList,
				"Patients": patientsList,
			},
		})
	}

	var err error
	if isEdit {
		id, _ := strconv.Atoi(idStr)
		a.ID = id
		err = h.apptSvc.Update(id, a)
	} else {
		err = h.apptSvc.Create(a)
	}
	if err != nil {
		rerender(err.Error())
		return
	}

	http.Redirect(w, r, "/appointments", http.StatusSeeOther)
}

func (h *WebHandler) AppointmentAction(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/appointments/")

	if strings.HasSuffix(path, "/status") {
		idStr := strings.TrimSuffix(path, "/status")
		idStr = strings.TrimSuffix(idStr, "/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Redirect(w, r, "/appointments", http.StatusSeeOther)
			return
		}
		status := strings.TrimSpace(r.FormValue("status"))
		if err := h.apptSvc.UpdateStatus(id, status); err != nil {
			http.Redirect(w, r, "/appointments?err=status", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/appointments", http.StatusSeeOther)
		return
	}

	// /appointments/{id}  -> form edit
	id, err := strconv.Atoi(strings.TrimSuffix(path, "/"))
	if err != nil {
		http.Redirect(w, r, "/appointments", http.StatusSeeOther)
		return
	}
	h.AppointmentForm(w, r, user)
	_ = id
}

// AppointmentDashboard mengembalikan jumlah janji temu mendatang.
func (h *WebHandler) appointmentTodayCount() int {
	today := time.Now().Format("2006-01-02")
	n, _ := h.apptSvc.CountUpcoming(today)
	return n
}
