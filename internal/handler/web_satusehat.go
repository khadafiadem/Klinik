package handler

import (
	"net/http"
	"strings"

	"klinik-app/internal/auth"
	"klinik-app/internal/logger"
)

func (h *WebHandler) SatusehatIndex(w http.ResponseWriter, r *http.Request, user *auth.User) {
	cfg, err := h.satusehatSvc.Get()
	if err != nil {
		RenderTemplate(w, r, "satusehat/index", TemplateData{User: user, Error: err.Error()})
		return
	}

	success, errorMsg := "", ""
	switch r.URL.Query().Get("msg") {
	case "saved":
		success = "Pengaturan SATUSEHAT berhasil disimpan"
	case "fail":
		errorMsg = "Gagal menyimpan pengaturan SATUSEHAT"
	}

	masked := maskSecret(cfg.ClientSecret)
	view := &struct {
		Environment  string
		BaseURL      string
		OrgID        string
		OrgName      string
		LocationID   string
		ClientID     string
		ClientSecret string
		HasSecret    bool
		IsActive     bool
		Configured   bool
		Status       string
		StatusClass  string
	}{
		Environment:  cfg.Environment,
		BaseURL:      cfg.BaseURL,
		OrgID:        cfg.OrgID,
		OrgName:      cfg.OrgName,
		LocationID:   cfg.LocationID,
		ClientID:     cfg.ClientID,
		ClientSecret: masked,
		HasSecret:    masked != "",
		IsActive:     cfg.IsActive,
	}
	view.Configured = strings.TrimSpace(cfg.OrgID) != "" && strings.TrimSpace(cfg.ClientID) != ""
	view.Status = "Belum Dikonfigurasi"
	view.StatusClass = "badge-secondary"
	if view.Configured {
		view.Status = "Terkonfigurasi"
		view.StatusClass = "badge-info"
	}
	if cfg.IsActive {
		view.Status = "Aktif"
		view.StatusClass = "badge-success"
	}

	RenderTemplate(w, r, "satusehat/index", TemplateData{
		User:    user,
		Data:    view,
		Success: success,
		Error:   errorMsg,
	})
}

func (h *WebHandler) SatusehatSave(w http.ResponseWriter, r *http.Request, user *auth.User) {
	cfg, err := h.satusehatSvc.Get()
	if err != nil || cfg == nil {
		http.Redirect(w, r, "/satusehat?msg=fail", http.StatusSeeOther)
		return
	}

	env := strings.TrimSpace(r.FormValue("environment"))
	if env != "SANDBOX" && env != "PROD" {
		env = "PROD"
	}
	cfg.Environment = env
	if v := strings.TrimSpace(r.FormValue("base_url")); v != "" {
		cfg.BaseURL = v
	}
	cfg.OrgID = strings.TrimSpace(r.FormValue("org_id"))
	cfg.OrgName = strings.TrimSpace(r.FormValue("org_name"))
	cfg.LocationID = strings.TrimSpace(r.FormValue("location_id"))
	cfg.ClientID = strings.TrimSpace(r.FormValue("client_id"))
	if v := strings.TrimSpace(r.FormValue("client_secret")); v != "" {
		cfg.ClientSecret = v
	}
	cfg.IsActive = r.FormValue("is_active") == "on"

	if err := h.satusehatSvc.Update(cfg); err != nil {
		logger.Error.Printf("Simpan pengaturan SATUSEHAT gagal: %v", err)
		http.Redirect(w, r, "/satusehat?msg=fail", http.StatusSeeOther)
		return
	}

	h.auditSvc.Log(&user.ID, "UPDATE", "satusehat_settings", &cfg.ID,
		"Pengaturan SATUSEHAT diperbarui", r.RemoteAddr)

	http.Redirect(w, r, "/satusehat?msg=saved", http.StatusSeeOther)
}

func maskSecret(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + strings.Repeat("*", len(s)-4)
}
