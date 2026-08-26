package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"klinik-app/internal/auth"
	"klinik-app/internal/bpjs"
	"klinik-app/internal/queues"
)

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
	paused, _ := h.queueSvc.IsPaused()

	RenderTemplate(w, r, "queues/index", TemplateData{
		User:  user,
		Data: map[string]interface{}{
			"Queues":     list,
			"Date":       date,
			"Waiting":    waiting,
			"InProgress": inProgress,
			"Completed":  completed,
			"Paused":     paused,
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

	if action == "call" {
		h.queueSvc.UpdateStatusCalledBy(id, status, user.ID)
	} else {
		h.queueSvc.UpdateStatus(id, status)
	}

	// Auto-call next waiting patient after starting examination
	var autoCalled *queues.Queue
	if action == "start" {
		autoCalled, _ = h.queueSvc.CallNextPatient()
	}

	switch {
	case bpjsStatus > 0:
		s := bpjsStatus
		h.syncBPJS(func(svc *bpjs.Service) { svc.OnQueueStatusChanged(id, s) })
	case action == "cancel":
		h.syncBPJS(func(svc *bpjs.Service) { svc.OnQueueCancelled(id, cancelReason) })
	}

	_ = autoCalled

	http.Redirect(w, r, "/queues", http.StatusSeeOther)
}

func (h *WebHandler) QueueAdd(w http.ResponseWriter, r *http.Request, user *auth.User) {
	registrationID, _ := strconv.Atoi(r.FormValue("registration_id"))
	patientID, _ := strconv.Atoi(r.FormValue("patient_id"))
	doctorID, _ := strconv.Atoi(r.FormValue("doctor_id"))

	q := &queues.Queue{
		RegistrationID: &registrationID,
		PatientID:      &patientID,
		DoctorID:       &doctorID,
	}

	if err := h.queueSvc.Create(q); err != nil {
		http.Redirect(w, r, "/queues", http.StatusSeeOther)
		return
	}
	h.syncBPJS(func(svc *bpjs.Service) { svc.OnQueueCreated(q.ID) })
	http.Redirect(w, r, "/queues", http.StatusSeeOther)
}

// KioskPage menampilkan layar sentuh untuk pengambilan nomor antrian.
func (h *WebHandler) KioskPage(w http.ResponseWriter, r *http.Request) {
	clinic, _ := h.clinicSvc.Get()
	RenderTemplate(w, r, "queues/kiosk", TemplateData{
		Data: map[string]interface{}{
			"Clinic": clinic,
		},
	})
}

// KioskTakeNumber API endpoint untuk mengambil nomor antrian dari kiosk.
func (h *WebHandler) KioskTakeNumber(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q, err := h.queueSvc.CreateKiosk()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Gagal mengambil nomor antrian: " + err.Error(),
		})
		return
	}

	clinic, _ := h.clinicSvc.Get()
	clinicName := "Klinik"
	if clinic != nil {
		clinicName = clinic.ClinicName
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"message":    "Nomor antrian berhasil diambil",
		"queue":      q,
		"clinicName": clinicName,
	})
}

// DisplayPage menampilkan layar monitoring antrian untuk TV.
func (h *WebHandler) DisplayPage(w http.ResponseWriter, r *http.Request) {
	clinic, _ := h.clinicSvc.Get()
	RenderTemplate(w, r, "queues/display", TemplateData{
		Data: map[string]interface{}{
			"Clinic": clinic,
		},
	})
}

// DisplayAPI endpoint JSON untuk auto-refresh TV display.
func (h *WebHandler) DisplayAPI(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		date = r.URL.Query().Get("d")
	}

	queuesList, err := h.queueSvc.GetMonitorData(date)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	waiting, inProgress, completed, _ := h.queueSvc.GetTodayStats()

	clinic, _ := h.clinicSvc.Get()
	clinicName := "Klinik"
	clinicAddress := ""
	if clinic != nil {
		clinicName = clinic.ClinicName
		clinicAddress = clinic.ClinicAddress
	}

	// Find currently called queue
	var currentCalled *queues.Queue
	for i := range queuesList {
		if queuesList[i].Status == "DIPANGGIL" {
			currentCalled = &queuesList[i]
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"queues":      queuesList,
		"waiting":     waiting,
		"inProgress":  inProgress,
		"completed":   completed,
		"clinicName":  clinicName,
		"clinicAddress": clinicAddress,
		"currentCalled": currentCalled,
		"paused":      false,
	})
}

func (h *WebHandler) QueueTogglePause(w http.ResponseWriter, r *http.Request, user *auth.User) {
	paused, _ := h.queueSvc.IsPaused()
	newState := !paused
	h.queueSvc.SetPaused(newState)

	http.Redirect(w, r, "/queues", http.StatusSeeOther)
}

func (h *WebHandler) QueueCallNext(w http.ResponseWriter, r *http.Request, user *auth.User) {
	http.Redirect(w, r, "/queues", http.StatusSeeOther)
}

// QueueCallNextAPI panggil pasien berikutnya (untuk AJAX manual call).
func (h *WebHandler) QueueCallNextAPI(w http.ResponseWriter, r *http.Request, user *auth.User) {
	w.Header().Set("Content-Type", "application/json")

	called, err := h.queueSvc.CallNextPatient()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if called == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Tidak ada pasien menunggu atau sistem dalam posisi jeda",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"queue":   called,
	})
}
