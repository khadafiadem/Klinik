package handler

import (
	"net/http"
	"strconv"
	"strings"

	"klinik-app/internal/auth"
	"klinik-app/internal/immunizations"
	"klinik-app/internal/logger"
)

func (h *WebHandler) ImmunizationsList(w http.ResponseWriter, r *http.Request, user *auth.User) {
	search := r.URL.Query().Get("search")
	records, _, _ := h.immSvc.ListImmunizations(1, 100, search)
	schedules, _ := h.immSvc.ListSchedules()
	patientsList, _, _ := h.patientSvc.GetAll(1, 1000, "")

	RenderTemplate(w, r, "immunizations/list", TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Records":           records,
			"Schedules":         schedules,
			"Patients":          patientsList,
			"Search":            search,
			"SelectedPatientID": 0,
			"Err":               r.URL.Query().Get("err"),
		},
	})
}

func (h *WebHandler) ImmunizationSave(w http.ResponseWriter, r *http.Request, user *auth.User) {
	patientID, _ := strconv.Atoi(r.FormValue("patient_id"))
	var immID *int
	if v := r.FormValue("immunization_id"); v != "" {
		id, _ := strconv.Atoi(v)
		immID = &id
	}
	vaccineName := strings.TrimSpace(r.FormValue("vaccine_name"))
	if vaccineName == "" && immID != nil {
		schedules, _ := h.immSvc.ListSchedules()
		for _, s := range schedules {
			if s.ID == *immID {
				vaccineName = s.Name
				break
			}
		}
	}
	createdBy := user.ID

	im := &immunizations.Immunization{
		PatientID:       patientID,
		ImmunizationID:  immID,
		VaccinationDate: strings.TrimSpace(r.FormValue("vaccination_date")),
		VaccineName:     vaccineName,
		BatchNumber:     strings.TrimSpace(r.FormValue("batch_number")),
		ProviderName:    strings.TrimSpace(r.FormValue("provider_name")),
		Notes:           strings.TrimSpace(r.FormValue("notes")),
		CreatedBy:       &createdBy,
	}

	if err := h.immSvc.Create(im); err != nil {
		logger.Error.Printf("Simpan imunisasi gagal: %v", err)
		http.Redirect(w, r, "/immunizations?err=fail", http.StatusSeeOther)
		return
	}

	h.auditSvc.Log(&user.ID, "CREATE", "patient_immunizations", &im.ID,
		"Imunisasi "+im.VaccineName+" dicatat untuk pasien", r.RemoteAddr)

	http.Redirect(w, r, "/immunizations", http.StatusSeeOther)
}

func (h *WebHandler) ImmunizationDelete(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/immunizations/")
	path = strings.TrimSuffix(path, "/delete")

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Redirect(w, r, "/immunizations", http.StatusSeeOther)
		return
	}

	if err := h.immSvc.Delete(id); err != nil {
		logger.Error.Printf("Hapus imunisasi %d gagal: %v", id, err)
	}
	http.Redirect(w, r, "/immunizations", http.StatusSeeOther)
}
