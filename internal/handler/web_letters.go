package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"klinik-app/internal/auth"
	"klinik-app/internal/letters"
	"klinik-app/internal/logger"
)

func (h *WebHandler) LettersList(w http.ResponseWriter, r *http.Request, user *auth.User) {
	templates, _ := h.letterSvc.ListActiveTemplates()
	recent, _ := h.letterSvc.ListRecent(20)
	patientsList, _, _ := h.patientSvc.GetAll(1, 1000, "")
	doctorsList, _, _ := h.doctorSvc.GetAll(1, 200, "")

	RenderTemplate(w, r, "letters/list", TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Templates": templates,
			"Recent":    recent,
			"Patients":  patientsList,
			"Doctors":   doctorsList,
			"Err":       r.URL.Query().Get("err"),
		},
	})
}

func (h *WebHandler) LetterGenerate(w http.ResponseWriter, r *http.Request, user *auth.User) {
	code := strings.TrimSpace(r.FormValue("code"))
	patientID, _ := strconv.Atoi(r.FormValue("patient_id"))
	var doctorID *int
	if v := r.FormValue("doctor_id"); v != "" {
		id, _ := strconv.Atoi(v)
		if id > 0 {
			doctorID = &id
		}
	}
	notes := strings.TrimSpace(r.FormValue("notes"))

	tmpl, err := h.letterSvc.GetTemplateByCode(code)
	if err != nil {
		logger.Error.Printf("Generate surat: template %q tidak ditemukan: %v", code, err)
		http.Redirect(w, r, "/surat?err=fail", http.StatusSeeOther)
		return
	}

	iss := &letters.Issuance{
		TemplateID: tmpl.ID,
		PatientID:  patientID,
		DoctorID:   doctorID,
		Notes:      notes,
		IssuedBy:   &user.ID,
	}
	if err := h.letterSvc.CreateIssuance(iss); err != nil {
		logger.Error.Printf("Generate surat gagal: %v", err)
		http.Redirect(w, r, "/surat?err=fail", http.StatusSeeOther)
		return
	}

	h.auditSvc.Log(&user.ID, "CREATE", "letter_issuances", &iss.ID,
		"Surat "+tmpl.Name+" diterbitkan untuk pasien", r.RemoteAddr)

	http.Redirect(w, r, "/surat/print/"+strconv.Itoa(iss.ID), http.StatusSeeOther)
}

func (h *WebHandler) LetterPrint(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/surat/print/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Redirect(w, r, "/surat", http.StatusSeeOther)
		return
	}

	iss, err := h.letterSvc.GetIssuance(id)
	if err != nil {
		http.Redirect(w, r, "/surat", http.StatusSeeOther)
		return
	}

	p, _ := h.patientSvc.GetByID(iss.PatientID)
	s, _ := h.clinicSvc.Get()
	clinicName := "Klinik"
	var clinicAddr, clinicPhone, clinicEmail string
	if s != nil {
		clinicName = s.ClinicName
		clinicAddr = s.ClinicAddress
		clinicPhone = s.ClinicPhone
		clinicEmail = s.ClinicEmail
	}

	ttl := p.DateOfBirth + ""
	if p.DateOfBirth != "" {
		if t, err := time.Parse("2006-01-02", p.DateOfBirth); err == nil {
			ttl = t.Format("02 January 2006")
		}
	}
	jk := "Laki-laki"
	if p.Gender == "PEREMPUAN" {
		jk = "Perempuan"
	}

	data := map[string]string{
		"{{nama}}":    p.FullName,
		"{{nik}}":     p.NIK,
		"{{ttl}}":     ttl,
		"{{jk}}":      jk,
		"{{alamat}}":  p.Address,
		"{{no_rm}}":   p.MedicalRecordNumber,
		"{{tanggal}}": time.Now().Format("02 January 2006"),
		"{{dokter}}":  iss.DoctorName,
		"{{klinik}}":  clinicName,
	}

	body := iss.BodyTemplate
	body = strings.ReplaceAll(body, `\n`, "\n")
	for k, v := range data {
		body = strings.ReplaceAll(body, k, v)
	}

	render := map[string]interface{}{
		"ClinicName":    clinicName,
		"ClinicAddress": clinicAddr,
		"ClinicPhone":   clinicPhone,
		"ClinicEmail":   clinicEmail,
		"LetterNumber":  iss.LetterNumber,
		"Subject":       iss.Subject,
		"Date":          time.Now().Format("02 January 2006"),
		"Body":          body,
		"DoctorName":    iss.DoctorName,
		"PatientName":   iss.PatientName,
	}
	RenderPrint(w, "letters/print", render)
}
