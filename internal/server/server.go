package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"klinik-app/internal/auth"
	"klinik-app/internal/config"
	"klinik-app/internal/handler"
	"klinik-app/internal/logger"
	"klinik-app/internal/middleware"
)

type Server struct {
	cfg  *config.Config
	db   *sql.DB
	srv  *http.Server
	auth *auth.Service
	wh   *handler.WebHandler
	rl   *middleware.RateLimiter
}

type HealthResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Database  string `json:"database"`
	Timestamp string `json:"timestamp"`
}

func New(cfg *config.Config, db *sql.DB) *Server {
	s := &Server{cfg: cfg, db: db}

	s.auth = auth.NewService(db, cfg.JWTSecret, 24)
	s.rl = middleware.NewRateLimiter(5, 5*time.Minute)
	s.wh = handler.NewWebHandler(db, s.auth, s.rl)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", s.healthCheck)
	s.registerAPIRoutes(mux)
	s.registerWebRoutes(mux)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	s.srv = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.AppPort),
		Handler:      s.logRequests(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

func (s *Server) registerAPIRoutes(mux *http.ServeMux) {
	apiHandler := auth.NewHandler(s.auth)
	mux.HandleFunc("/api/auth/login", apiHandler.Login)
	mux.HandleFunc("/api/auth/logout", apiHandler.Logout)
}

func (s *Server) registerWebRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/login", s.webLogin)
	mux.HandleFunc("/logout", s.wh.Logout)

	protected := func(roles []string, next func(http.ResponseWriter, *http.Request, *auth.User)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token, err := r.Cookie("token")
			if err != nil || token.Value == "" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			claims, err := s.auth.ValidateToken(token.Value)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			user, err := s.auth.GetUserByID(claims.UserID)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			// RBAC: roles kosong berarti semua user terautentikasi diizinkan.
			if len(roles) > 0 && !userHasRole(user, roles...) {
				handler.RenderForbidden(w, r, user)
				return
			}
			next(w, r, user)
		}
	}

	var (
		all         = []string(nil)
		admin       = []string{"ADMIN"}
		doctor      = []string{"ADMIN", "DOCTOR"}
		doctorNurse = []string{"ADMIN", "DOCTOR", "NURSE"}
		nurse       = []string{"ADMIN", "NURSE"}
		pharmacist  = []string{"ADMIN", "PHARMACIST"}
		cashier     = []string{"ADMIN", "CASHIER"}
		management  = []string{"ADMIN", "OWNER"}
		patientsAcc = []string{"ADMIN", "DOCTOR", "NURSE", "CASHIER"}
		regAcc      = []string{"ADMIN", "DOCTOR", "NURSE"}
		medicalAcc  = []string{"ADMIN", "DOCTOR", "NURSE"}
		rxView      = []string{"ADMIN", "DOCTOR", "PHARMACIST"}
		medView     = []string{"ADMIN", "PHARMACIST", "DOCTOR"}
	)

	mux.HandleFunc("/dashboard", protected(all, s.wh.Dashboard))
	mux.HandleFunc("/profile", protected(all, s.wh.Profile))
	mux.HandleFunc("/profile/password", protected(all, s.wh.ProfilePasswordPost))
	mux.HandleFunc("/clinic-settings", protected(admin, s.wh.ClinicSettings))
	mux.HandleFunc("/settings/bpjs/doctors", protected(admin, s.wh.BPJSDoctorMapPost))
	mux.HandleFunc("/settings/bpjs", protected(admin, s.wh.BPJSSettings))

	mux.HandleFunc("/doctors", protected(doctorNurse, s.wh.DoctorsList))
	mux.HandleFunc("/doctors/new", protected(doctor, s.wh.DoctorForm))
	mux.HandleFunc("/doctors/save", protected(doctor, s.wh.DoctorSave))
	mux.HandleFunc("/doctors/", protected(doctor, s.wh.DoctorForm))
	mux.HandleFunc("/doctors/save/", protected(doctor, s.wh.DoctorSave))

	mux.HandleFunc("/staff", protected(admin, s.wh.StaffList))
	mux.HandleFunc("/staff/new", protected(admin, s.wh.StaffForm))
	mux.HandleFunc("/staff/save", protected(admin, s.wh.StaffSave))
	mux.HandleFunc("/staff/", protected(admin, s.wh.StaffForm))
	mux.HandleFunc("/staff/save/", protected(admin, s.wh.StaffSave))

	mux.HandleFunc("/patients", protected(patientsAcc, s.wh.PatientsList))
	mux.HandleFunc("/patients/new", protected(regAcc, s.wh.PatientForm))
	mux.HandleFunc("/patients/save", protected(regAcc, s.wh.PatientSave))
	mux.HandleFunc("/patients/", protected(patientsAcc, s.wh.PatientForm))
	mux.HandleFunc("/patients/save/", protected(regAcc, s.wh.PatientSave))

	mux.HandleFunc("/registrations", protected(regAcc, s.wh.RegistrationsList))
	mux.HandleFunc("/registrations/new", protected(nurse, s.wh.RegistrationForm))
	mux.HandleFunc("/registrations/save", protected(nurse, s.wh.RegistrationSave))
	mux.HandleFunc("/registrations/", protected(regAcc, s.wh.RegistrationAction))

	mux.HandleFunc("/queues", protected(regAcc, s.wh.QueuesPage))
	mux.HandleFunc("/queues/add", protected(nurse, s.wh.QueueAdd))
	mux.HandleFunc("/queues/", protected(regAcc, s.wh.QueueAction))

	mux.HandleFunc("/medical-records", protected(medicalAcc, s.wh.MedicalRecordsList))
	mux.HandleFunc("/medical-records/new", protected(doctor, s.wh.MedicalRecordForm))
	mux.HandleFunc("/medical-records/save", protected(doctor, s.wh.MedicalRecordSave))
	mux.HandleFunc("/medical-records/prescription/create", protected(doctor, s.wh.MRCreatePrescription))
	mux.HandleFunc("/medical-records/diagnosis/add", protected(doctor, s.wh.MRAddDiagnosis))
	mux.HandleFunc("/medical-records/diagnosis/", protected(doctor, s.wh.MRRemoveDiagnosis))
	mux.HandleFunc("/medical-records/treatment/add", protected(doctor, s.wh.MRAddTreatment))
	mux.HandleFunc("/medical-records/treatment/", protected(doctor, s.wh.MRRemoveTreatment))
	mux.HandleFunc("/medical-records/", protected(medicalAcc, s.wh.MedicalRecordView))

	mux.HandleFunc("/prescriptions", protected(rxView, s.wh.PrescriptionsList))
	mux.HandleFunc("/prescriptions/item/add", protected(doctor, s.wh.PrescriptionAddItem))
	mux.HandleFunc("/prescriptions/item/", protected(doctor, s.wh.PrescriptionRemoveItem))
	mux.HandleFunc("/prescriptions/", protected(rxView, s.wh.PrescriptionView))

	mux.HandleFunc("/medicines", protected(medView, s.wh.MedicinesList))
	mux.HandleFunc("/medicines/new", protected(pharmacist, s.wh.MedicineForm))
	mux.HandleFunc("/medicines/save", protected(pharmacist, s.wh.MedicineSave))
	mux.HandleFunc("/medicines/stock/add", protected(pharmacist, s.wh.MedicineStockAdd))
	mux.HandleFunc("/medicines/stock/reduce", protected(pharmacist, s.wh.MedicineStockReduce))
	mux.HandleFunc("/medicines/", protected(medView, s.wh.MedicinePage))

	mux.HandleFunc("/invoices", protected(cashier, s.wh.InvoicesList))
	mux.HandleFunc("/invoices/new", protected(cashier, s.wh.InvoiceForm))
	mux.HandleFunc("/invoices/save", protected(cashier, s.wh.InvoiceSave))
	mux.HandleFunc("/invoices/item/add", protected(cashier, s.wh.InvoiceAddItem))
	mux.HandleFunc("/invoices/item/", protected(cashier, s.wh.InvoiceRemoveItem))
	mux.HandleFunc("/invoices/", protected(cashier, s.wh.InvoiceView))

	mux.HandleFunc("/payments", protected(cashier, s.wh.PaymentsList))
	mux.HandleFunc("/payments/new", protected(cashier, s.wh.PaymentForm))
	mux.HandleFunc("/payments/save", protected(cashier, s.wh.PaymentSave))

	mux.HandleFunc("/reports", protected(management, s.wh.ReportsDashboard))
	mux.HandleFunc("/reports/patients", protected(management, s.wh.ReportsPatient))
	mux.HandleFunc("/reports/registrations", protected(management, s.wh.ReportsRegistration))
	mux.HandleFunc("/reports/doctors", protected(management, s.wh.ReportsDoctorActivity))
	mux.HandleFunc("/reports/revenue", protected(management, s.wh.ReportsRevenue))
	mux.HandleFunc("/reports/medicines", protected(management, s.wh.ReportsMedicineStock))
	mux.HandleFunc("/reports/export/patients", protected(management, s.wh.ExportPatientsCSV))
	mux.HandleFunc("/reports/export/registrations", protected(management, s.wh.ExportRegistrationsCSV))
	mux.HandleFunc("/reports/export/revenue", protected(management, s.wh.ExportRevenueCSV))

	mux.HandleFunc("/audit-logs", protected(admin, s.wh.AuditLogsList))

	mux.HandleFunc("/users", protected(admin, s.wh.UsersList))
	mux.HandleFunc("/users/new", protected(admin, s.wh.UserForm))
	mux.HandleFunc("/users/save", protected(admin, s.wh.UserSave))
	mux.HandleFunc("/users/delete", protected(admin, s.wh.UserDelete))
	mux.HandleFunc("/users/", protected(admin, s.wh.UserForm))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `<!DOCTYPE html><html lang="id"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>404 - Halaman Tidak Ditemukan - KlinikApp</title>
<link rel="preconnect" href="https://fonts.googleapis.com"><link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
<link rel="stylesheet" href="/static/css/style.css">
</head><body style="display:flex;align-items:center;justify-content:center;min-height:100vh;background:var(--bg-primary);font-family:'Inter',sans-serif;">
<div style="text-align:center;padding:40px;">
<div style="font-size:6rem;font-weight:800;color:var(--primary);line-height:1;">404</div>
<h1 style="margin:16px 0 8px;font-size:1.5rem;color:var(--text-primary);">Halaman Tidak Ditemukan</h1>
<p style="color:var(--text-muted);margin-bottom:24px;">Halaman yang Anda cari tidak tersedia atau sudah dipindahkan.</p>
<a href="/dashboard" class="btn btn-primary">Kembali ke Dashboard</a>
</div></body></html>`)
	})
}

func userHasRole(u *auth.User, allowed ...string) bool {
	for _, ur := range u.Roles {
		for _, a := range allowed {
			if strings.EqualFold(ur.Name, a) {
				return true
			}
		}
	}
	return false
}

func (s *Server) webLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		token, err := r.Cookie("token")
		if err == nil && token.Value != "" {
			if _, err := s.auth.ValidateToken(token.Value); err == nil {
				http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
				return
			}
		}
		s.wh.LoginPage(w, r)
		return
	}
	s.wh.LoginPost(w, r)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dbStatus := "terhubung"
	if s.db == nil {
		dbStatus = "tidak dikonfigurasi"
	} else if err := s.db.Ping(); err != nil {
		dbStatus = "terputus"
	}

	resp := HealthResponse{
		Status:    "ok",
		Message:   "Sistem manajemen klinik berjalan normal",
		Database:  dbStatus,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/static/") {
			logger.Info.Printf("%s %s", r.Method, r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Handler() http.Handler {
	return s.srv.Handler
}

func (s *Server) Start() error {
	logger.Info.Printf("Server starting on port %d", s.cfg.AppPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error.Printf("Server error: %v", err)
		}
	}()

	<-quit
	logger.Info.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	logger.Info.Println("Server stopped gracefully")
	return nil
}
