package handler

import (
	"net/http"
	"strconv"
	"strings"

	"klinik-app/internal/auth"
	"klinik-app/internal/bpjs"
)

func (h *WebHandler) BPJSSettings(w http.ResponseWriter, r *http.Request, user *auth.User) {
	if r.Method == http.MethodPost {
		h.bpjsSettingsPost(w, r, user)
		return
	}
	h.renderBPJSSettings(w, r, user, "", "")
}

func (h *WebHandler) renderBPJSSettings(w http.ResponseWriter, r *http.Request, user *auth.User, success, errMsg string) {
	cfg, err := h.bpjsSvc.LoadConfig()
	if err != nil {
		RenderTemplate(w, r, "clinic/bpjs", TemplateData{User: user, Error: err.Error()})
		return
	}

	logs, _ := h.bpjsSvc.RecentLogs(25)
	doctorMaps, _ := h.bpjsSvc.ListDoctorMaps()

	data := map[string]interface{}{
		"Config":     cfg,
		"Logs":       logs,
		"DoctorMaps": doctorMaps,
	}

	td := TemplateData{User: user, Data: data}
	if success != "" {
		td.Success = success
	}
	if errMsg != "" {
		td.Error = errMsg
	}
	RenderTemplate(w, r, "clinic/bpjs", td)
}

func (h *WebHandler) bpjsSettingsPost(w http.ResponseWriter, r *http.Request, user *auth.User) {
	cfg, err := h.bpjsSvc.LoadConfig()
	if err != nil {
		h.renderBPJSSettings(w, r, user, "", err.Error())
		return
	}

	cfg.Enabled = r.FormValue("enabled") == "on"
	cfg.Mode = strings.ToUpper(strings.TrimSpace(r.FormValue("mode")))
	if cfg.Mode != bpjs.ModeProduction {
		cfg.Mode = bpjs.ModeSandbox
	}
	cfg.BaseURL = strings.TrimSpace(r.FormValue("base_url"))
	cfg.ConsID = strings.TrimSpace(r.FormValue("cons_id"))
	cfg.SecretKey = strings.TrimSpace(r.FormValue("secret_key"))
	cfg.UserKey = strings.TrimSpace(r.FormValue("user_key"))
	cfg.KodePPK = strings.TrimSpace(r.FormValue("kode_ppk"))
	cfg.NamaPPK = strings.TrimSpace(r.FormValue("nama_ppk"))
	cfg.KodePoli = strings.TrimSpace(r.FormValue("kode_poli"))
	cfg.NamaPoli = strings.TrimSpace(r.FormValue("nama_poli"))
	cfg.JamPraktek = strings.TrimSpace(r.FormValue("jam_praktek"))

	if err := h.bpjsSvc.SaveConfig(cfg); err != nil {
		h.renderBPJSSettings(w, r, user, "", err.Error())
		return
	}

	h.auditSvc.Log(&user.ID, "UPDATE", "bpjs_config", nil, "Perubahan konfigurasi BPJS", r.RemoteAddr)
	h.renderBPJSSettings(w, r, user, "Konfigurasi BPJS berhasil disimpan", "")
}

// BPJSDoctorMapPost menyimpan/memperbarui mapping kode dokter BPJS.
func (h *WebHandler) BPJSDoctorMapPost(w http.ResponseWriter, r *http.Request, user *auth.User) {
	doctorID, _ := strconv.Atoi(r.FormValue("doctor_id"))
	code := strings.TrimSpace(r.FormValue("bpjs_code"))

	if err := h.bpjsSvc.UpsertDoctorMap(doctorID, code); err != nil {
		h.renderBPJSSettings(w, r, user, "", err.Error())
		return
	}

	http.Redirect(w, r, "/settings/bpjs", http.StatusSeeOther)
}
