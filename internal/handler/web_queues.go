package handler

import (
	"net/http"
	"strconv"
	"strings"

	"klinik-app/internal/auth"
	"klinik-app/internal/bpjs"
	"klinik-app/internal/queues"
)

// syncBPJS menjalankan sinkronisasi BPJS secara asinkron (best-effort).
// Kegagalan tidak memengaruhi operasional klinik dan tercatat di bpjs_log.
func (h *WebHandler) syncBPJS(fn func(s *bpjs.Service)) {
	go func() {
		defer func() { _ = recover() }()
		fn(h.bpjsSvc)
	}()
}

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
	var bpjsStatus int
	var cancelReason string
	switch action {
	case "call":
		status, bpjsStatus = "DIPANGGIL", 1
	case "start":
		status, bpjsStatus = "SEDANG_DIPERIKSA", 2
	case "complete":
		status, bpjsStatus = "SELESAI", 3
	case "cancel":
		status = "DIBATALKAN"
		cancelReason = strings.TrimSpace(r.FormValue("alasan"))
		if cancelReason == "" {
			cancelReason = "Dibatalkan oleh petugas"
		}
	default:
		http.Redirect(w, r, "/queues", http.StatusSeeOther)
		return
	}

	h.queueSvc.UpdateStatus(id, status)

	switch {
	case bpjsStatus > 0:
		s := bpjsStatus
		h.syncBPJS(func(svc *bpjs.Service) { svc.OnQueueStatusChanged(id, s) })
	case action == "cancel":
		h.syncBPJS(func(svc *bpjs.Service) { svc.OnQueueCancelled(id, cancelReason) })
	}

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
	h.syncBPJS(func(svc *bpjs.Service) { svc.OnQueueCreated(q.ID) })
	http.Redirect(w, r, "/queues", http.StatusSeeOther)
}
