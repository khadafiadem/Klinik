package handler

import (
	"net/http"
	"strconv"
	"strings"

	"klinik-app/internal/auth"
	"klinik-app/internal/queues"
)

func (h *WebHandler) QueuesPage(w http.ResponseWriter, r *http.Request, user *auth.User) {
	date := r.URL.Query().Get("date")
	list, _ := h.queueSvc.GetAllByDate(date)
	waiting, inProgress, completed, _ := h.queueSvc.GetTodayStats()

	RenderTemplate(w, r, "queues/index", TemplateData{
		User:  user,
		Data: map[string]interface{}{
			"Queues":     list,
			"Date":       date,
			"Waiting":    waiting,
			"InProgress": inProgress,
			"Completed":  completed,
		},
	})
}

func (h *WebHandler) QueueAction(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/queues/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.Redirect(w, r, "/queues", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Redirect(w, r, "/queues", http.StatusSeeOther)
		return
	}

	action := parts[1]
	var status string
	switch action {
	case "call":
		status = "DIPANGGIL"
	case "start":
		status = "SEDANG_DIPERIKSA"
	case "complete":
		status = "SELESAI"
	case "cancel":
		status = "DIBATALKAN"
	default:
		http.Redirect(w, r, "/queues", http.StatusSeeOther)
		return
	}

	h.queueSvc.UpdateStatus(id, status)
	http.Redirect(w, r, "/queues", http.StatusSeeOther)
}

func (h *WebHandler) QueueAdd(w http.ResponseWriter, r *http.Request, user *auth.User) {
	registrationID, _ := strconv.Atoi(r.FormValue("registration_id"))
	patientID, _ := strconv.Atoi(r.FormValue("patient_id"))
	doctorID, _ := strconv.Atoi(r.FormValue("doctor_id"))

	q := &queues.Queue{
		RegistrationID: registrationID,
		PatientID:      patientID,
		DoctorID:       doctorID,
	}

	if err := h.queueSvc.Create(q); err != nil {
		http.Redirect(w, r, "/queues", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/queues", http.StatusSeeOther)
}
