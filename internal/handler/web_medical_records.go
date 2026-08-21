package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"klinik-app/internal/auth"
	"klinik-app/internal/logger"
	"klinik-app/internal/medical_records"
	"klinik-app/internal/prescriptions"
)

func (h *WebHandler) MedicalRecordsList(w http.ResponseWriter, r *http.Request, user *auth.User) {
	search := r.URL.Query().Get("search")
	list, _, _ := h.mrSvc.GetAll(1, 100, search)
	RenderTemplate(w, r, "medical_records/list", TemplateData{
		User:   user,
		Data:   list,
		Search: search,
	})
}

func (h *WebHandler) MedicalRecordForm(w http.ResponseWriter, r *http.Request, user *auth.User) {
	doctorsList, _, _ := h.doctorSvc.GetAll(1, 100, "")
	patientsList, _, _ := h.patientSvc.GetAll(1, 1000, "")
	RenderTemplate(w, r, "medical_records/form", TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Doctors":  doctorsList,
			"Patients": patientsList,
		},
	})
}

func (h *WebHandler) MedicalRecordSave(w http.ResponseWriter, r *http.Request, user *auth.User) {
	patientID, _ := strconv.Atoi(r.FormValue("patient_id"))
	doctorID, _ := strconv.Atoi(r.FormValue("doctor_id"))
	var regID *int
	if v := r.FormValue("registration_id"); v != "" {
		id, _ := strconv.Atoi(v)
		regID = &id
	}

	vitalSigns := map[string]string{
		"temperature": r.FormValue("temperature"),
		"heart_rate":  r.FormValue("heart_rate"),
		"blood_pressure": r.FormValue("blood_pressure"),
		"respiratory_rate": r.FormValue("respiratory_rate"),
		"weight": r.FormValue("weight"),
		"height": r.FormValue("height"),
	}
	vsJSON, _ := json.Marshal(vitalSigns)

	mr := struct {
		PatientID        int
		DoctorID         int
		RegistrationID   *int
		ExaminationDate  string
		ChiefComplaint   string
		VitalSigns       string
		Anamnesis        string
		PhysicalExam     string
		Notes            string
	}{
		PatientID:       patientID,
		DoctorID:        doctorID,
		RegistrationID:  regID,
		ExaminationDate: r.FormValue("examination_date"),
		ChiefComplaint:  r.FormValue("chief_complaint"),
		VitalSigns:      string(vsJSON),
		Anamnesis:       r.FormValue("anamnesis"),
		PhysicalExam:    r.FormValue("physical_examination"),
		Notes:           r.FormValue("notes"),
	}

	// Use a temp struct matching model
	// Use the service's MRInput struct
	input := &medical_records.MRInput{
		PatientID:       patientID,
		DoctorID:        doctorID,
		RegistrationID:  regID,
		ExaminationDate: mr.ExaminationDate,
		ChiefComplaint:  mr.ChiefComplaint,
		VitalSigns:      mr.VitalSigns,
		Anamnesis:       mr.Anamnesis,
		PhysicalExam:    mr.PhysicalExam,
		Notes:           mr.Notes,
	}

	err := h.mrSvc.CreateFromInput(input)
	if err != nil {
		doctorsList, _, _ := h.doctorSvc.GetAll(1, 100, "")
		patientsList, _, _ := h.patientSvc.GetAll(1, 1000, "")
		RenderTemplate(w, r, "medical_records/form", TemplateData{
			User:  user,
			Error: err.Error(),
			Data: map[string]interface{}{
				"Doctors":  doctorsList,
				"Patients": patientsList,
				"Form":     input,
			},
		})
		return
	}

	http.Redirect(w, r, "/medical-records", http.StatusSeeOther)
}

func (h *WebHandler) MedicalRecordView(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/medical-records/")
	parts := strings.SplitN(path, "/", 2)

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Redirect(w, r, "/medical-records", http.StatusSeeOther)
		return
	}

	if len(parts) > 1 {
		action := parts[1]
		if action == "finalize" {
			h.mrSvc.UpdateStatus(id, "FINAL")
		}
		http.Redirect(w, r, fmt.Sprintf("/medical-records/%d", id), http.StatusSeeOther)
		return
	}

	mr, err := h.mrSvc.GetByID(id)
	if err != nil {
		http.Redirect(w, r, "/medical-records", http.StatusSeeOther)
		return
	}

	diagnosesList, _ := h.mrSvc.GetAllDiagnoses()
	treatmentsList, _ := h.mrSvc.GetAllTreatments()

	RenderTemplate(w, r, "medical_records/view", TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Record":     mr,
			"Diagnoses":  diagnosesList,
			"Treatments": treatmentsList,
		},
	})
}

// MRCreatePrescription membuat resep PENDING dari rekam medis, lalu mengarahkan
// dokter ke halaman resep untuk menambahkan obat.
func (h *WebHandler) MRCreatePrescription(w http.ResponseWriter, r *http.Request, user *auth.User) {
	mrID, _ := strconv.Atoi(r.FormValue("medical_record_id"))

	mr, err := h.mrSvc.GetByID(mrID)
	if err != nil {
		logger.Error.Printf("Buat resep: rekam medis %d tidak ditemukan: %v", mrID, err)
		http.Redirect(w, r, "/medical-records", http.StatusSeeOther)
		return
	}

	rx := &prescriptions.Prescription{
		MedicalRecordID:  mr.ID,
		PatientID:        mr.PatientID,
		DoctorID:         mr.DoctorID,
		PrescriptionDate: time.Now().Format("2006-01-02"),
	}

	if err := h.rxSvc.Create(rx); err != nil {
		logger.Error.Printf("Gagal membuat resep dari rekam medis %d: %v", mrID, err)
		http.Redirect(w, r, fmt.Sprintf("/medical-records/%d", mrID), http.StatusSeeOther)
		return
	}

	h.auditSvc.Log(&user.ID, "CREATE", "prescriptions", &rx.ID,
		fmt.Sprintf("Resep %s dibuat dari rekam medis %s", rx.PrescriptionNumber, mr.MedicalRecordNumber), r.RemoteAddr)

	http.Redirect(w, r, fmt.Sprintf("/prescriptions/%d", rx.ID), http.StatusSeeOther)
}

func (h *WebHandler) MRAddDiagnosis(w http.ResponseWriter, r *http.Request, user *auth.User) {
	mrID, _ := strconv.Atoi(r.FormValue("medical_record_id"))
	diagID, _ := strconv.Atoi(r.FormValue("diagnosis_id"))
	diagType := r.FormValue("diagnosis_type")
	if diagType == "" {
		diagType = "UTAMA"
	}

	h.mrSvc.AddDiagnosis(mrID, diagID, diagType, "")
	http.Redirect(w, r, fmt.Sprintf("/medical-records/%d", mrID), http.StatusSeeOther)
}

func (h *WebHandler) MRRemoveDiagnosis(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/medical-records/diagnosis/")
	parts := strings.SplitN(path, "/", 2)
	mrID, _ := strconv.Atoi(parts[0])
	diagEntryID, _ := strconv.Atoi(parts[1])

	h.mrSvc.RemoveDiagnosis(diagEntryID)
	http.Redirect(w, r, fmt.Sprintf("/medical-records/%d", mrID), http.StatusSeeOther)
}

func (h *WebHandler) MRAddTreatment(w http.ResponseWriter, r *http.Request, user *auth.User) {
	mrID, _ := strconv.Atoi(r.FormValue("medical_record_id"))
	treatID, _ := strconv.Atoi(r.FormValue("treatment_id"))
	cost, _ := strconv.ParseFloat(r.FormValue("cost"), 64)

	h.mrSvc.AddTreatment(mrID, treatID, cost, "")
	http.Redirect(w, r, fmt.Sprintf("/medical-records/%d", mrID), http.StatusSeeOther)
}

func (h *WebHandler) MRRemoveTreatment(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/medical-records/treatment/")
	parts := strings.SplitN(path, "/", 2)
	mrID, _ := strconv.Atoi(parts[0])
	treatEntryID, _ := strconv.Atoi(parts[1])

	h.mrSvc.RemoveTreatment(treatEntryID)
	http.Redirect(w, r, fmt.Sprintf("/medical-records/%d", mrID), http.StatusSeeOther)
}
