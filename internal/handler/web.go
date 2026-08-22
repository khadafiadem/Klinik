package handler

import (
	"database/sql"
	"net/http"
	"strings"

	"klinik-app/internal/audit"
	"klinik-app/internal/auth"
	"klinik-app/internal/clinic"
	"klinik-app/internal/doctors"
	"klinik-app/internal/finance"
	"klinik-app/internal/medical_records"
	"klinik-app/internal/medicines"
	"klinik-app/internal/middleware"
	"klinik-app/internal/patients"
	"klinik-app/internal/prescriptions"
	"klinik-app/internal/queues"
	"klinik-app/internal/registrations"
	"klinik-app/internal/reports"
	"klinik-app/internal/staff"
	"klinik-app/internal/users"
)

type WebHandler struct {
	authService *auth.Service
	clinicSvc   *clinic.Service
	doctorSvc   *doctors.Service
	staffSvc    *staff.Service
	patientSvc  *patients.Service
	regSvc      *registrations.Service
	queueSvc    *queues.Service
	mrSvc       *medical_records.Service
	rxSvc       *prescriptions.Service
	medSvc      *medicines.Service
	finSvc      *finance.Service
	rptSvc      *reports.Service
	auditSvc    *audit.Service
	userSvc     *users.Service
	rl          *middleware.RateLimiter
}

func NewWebHandler(db *sql.DB, authService *auth.Service, rl *middleware.RateLimiter) *WebHandler {
	return &WebHandler{
		authService: authService,
		clinicSvc:   clinic.NewService(db),
		doctorSvc:   doctors.NewService(db),
		staffSvc:    staff.NewService(db),
		patientSvc:  patients.NewService(db),
		regSvc:      registrations.NewService(db),
		queueSvc:    queues.NewService(db),
		mrSvc:       medical_records.NewService(db),
		rxSvc:       prescriptions.NewService(db),
		medSvc:      medicines.NewService(db),
		finSvc:      finance.NewService(db),
		rptSvc:      reports.NewService(db),
		auditSvc:    audit.NewService(db),
		userSvc:     users.NewService(db),
		rl:          rl,
	}
}

func (h *WebHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	RenderTemplate(w, r, "auth/login", TemplateData{})
}

func (h *WebHandler) LoginPost(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	ip := r.RemoteAddr

	if username == "" || password == "" {
		RenderTemplate(w, r, "auth/login", TemplateData{Error: "Username dan password harus diisi"})
		return
	}

	if !h.rl.Allow(username) {
		RenderTemplate(w, r, "auth/login", TemplateData{Error: "Terlalu banyak percobaan login. Coba lagi dalam 5 menit."})
		return
	}

	resp, err := h.authService.Login(auth.LoginRequest{Username: username, Password: password})
	if err != nil {
		RenderTemplate(w, r, "auth/login", TemplateData{Error: err.Error()})
		return
	}

	h.auditSvc.Log(&resp.User.ID, "LOGIN", "users", &resp.User.ID, "Login berhasil", ip)

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    resp.Token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *WebHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *WebHandler) Dashboard(w http.ResponseWriter, r *http.Request, user *auth.User) {
	patientCount, _ := h.patientSvc.Count()
	doctorCount, _ := h.doctorSvc.Count()
	staffCount, _ := h.staffSvc.Count()
	regCount, _ := h.regSvc.Count()
	mrCount, _ := h.mrSvc.Count()

	data := map[string]interface{}{
		"total_patients":      patientCount,
		"total_doctors":       doctorCount,
		"total_staff":         staffCount,
		"today_registrations": regCount,
		"today_mr":            mrCount,
	}

	RenderTemplate(w, r, "dashboard/index", TemplateData{User: user, Data: data})
}
