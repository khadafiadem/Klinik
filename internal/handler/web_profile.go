package handler

import (
	"fmt"
	"net/http"
	"strings"

	"klinik-app/internal/auth"
)

func (h *WebHandler) Profile(w http.ResponseWriter, r *http.Request, user *auth.User) {
	data := map[string]interface{}{
		"Profile": user,
	}
	td := TemplateData{User: user, Data: data}

	if r.URL.Query().Get("updated") == "1" {
		td.Success = "Password berhasil diubah"
	}

	RenderTemplate(w, r, "profile/profile", td)
}

func (h *WebHandler) ProfilePasswordPost(w http.ResponseWriter, r *http.Request, user *auth.User) {
	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	fail := func(msg string) {
		data := map[string]interface{}{
			"Profile": user,
			"Form": map[string]string{},
		}
		RenderTemplate(w, r, "profile/profile", TemplateData{User: user, Data: data, Error: msg})
	}

	if strings.TrimSpace(newPassword) == "" || strings.TrimSpace(confirmPassword) == "" {
		fail("Password baru dan konfirmasi password wajib diisi")
		return
	}

	if newPassword != confirmPassword {
		fail("Konfirmasi password tidak sesuai dengan password baru")
		return
	}

	if err := h.userSvc.ChangeOwnPassword(user.ID, currentPassword, newPassword); err != nil {
		h.auditSvc.Log(&user.ID, "CHANGE_PASSWORD_FAILED", "users", &user.ID, err.Error(), r.RemoteAddr)
		fail(err.Error())
		return
	}

	h.auditSvc.Log(&user.ID, "CHANGE_PASSWORD", "users", &user.ID, fmt.Sprintf("User %s mengubah password sendiri", user.Username), r.RemoteAddr)
	http.Redirect(w, r, "/profile?updated=1", http.StatusSeeOther)
}
