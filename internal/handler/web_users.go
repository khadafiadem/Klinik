package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"klinik-app/internal/auth"
	"klinik-app/internal/users"
)

func (h *WebHandler) UsersList(w http.ResponseWriter, r *http.Request, user *auth.User) {
	search := r.URL.Query().Get("search")
	userList, _, _ := h.userSvc.GetAll(1, 100, search)
	RenderTemplate(w, r, "users/list", TemplateData{User: user, Data: userList})
}

func (h *WebHandler) UserForm(w http.ResponseWriter, r *http.Request, user *auth.User) {
	roles, err := h.userSvc.GetAllRoles()
	if err != nil {
		roles = nil
	}

	var target *users.User
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	path = strings.TrimSuffix(path, "/edit")
	if path != "" && path != "new" {
		if id, err := strconv.Atoi(path); err == nil {
			target, _ = h.userSvc.GetByID(id)
		}
	}

	data := map[string]interface{}{
		"Target": target,
		"Roles":  roles,
		"Form":   map[string]string{},
	}
	RenderTemplate(w, r, "users/form", TemplateData{User: user, Data: data})
}

func (h *WebHandler) UserSave(w http.ResponseWriter, r *http.Request, user *auth.User) {
	idStr := strings.TrimSpace(r.FormValue("id"))
	isEdit := idStr != ""
	password := r.FormValue("password")
	roleID, _ := strconv.Atoi(r.FormValue("role_id"))

	var reload func(err error)
	reload = func(err error) {
		roles, _ := h.userSvc.GetAllRoles()
		var target *users.User
		if isEdit {
			id, _ := strconv.Atoi(idStr)
			target, _ = h.userSvc.GetByID(id)
		}
		data := map[string]interface{}{
			"Target": target,
			"Roles":  roles,
			"Form": map[string]string{
				"username": r.FormValue("username"),
				"email":    r.FormValue("email"),
				"fullname": r.FormValue("full_name"),
			},
		}
		RenderTemplate(w, r, "users/form", TemplateData{User: user, Data: data, Error: err.Error()})
	}

	if isEdit {
		id, _ := strconv.Atoi(idStr)

		activeVal := r.FormValue("is_active")
		if activeVal != "" {
			isActive := activeVal == "true"
			if id == user.ID && !isActive {
				reload(fmt.Errorf("tidak dapat menonaktifkan akun sendiri"))
				return
			}
			if err := h.userSvc.Update(id, users.UpdateUserRequest{IsActive: &isActive}); err != nil {
				reload(err)
				return
			}
		}

		req := users.UpdateUserRequest{
			Email:    strings.TrimSpace(r.FormValue("email")),
			FullName: strings.TrimSpace(r.FormValue("full_name")),
		}
		if req.Email != "" || req.FullName != "" {
			if err := h.userSvc.Update(id, req); err != nil {
				reload(err)
				return
			}
		}

		if password != "" {
			if err := h.userSvc.SetPassword(id, password); err != nil {
				reload(err)
				return
			}
		}

		if roleID > 0 {
			if id == user.ID {
				current := userHasRoleID(user, roleID)
				if !current {
					reload(fmt.Errorf("tidak dapat mengubah role akun sendiri"))
					return
				}
			} else if err := h.userSvc.AssignPrimaryRole(id, roleID); err != nil {
				reload(err)
				return
			}
		}

		h.auditSvc.Log(&user.ID, "UPDATE_USER", "users", &id, "Update data pengguna", r.RemoteAddr)
		http.Redirect(w, r, "/users", http.StatusSeeOther)
		return
	}

	req := users.CreateUserRequest{
		Username: strings.TrimSpace(r.FormValue("username")),
		Email:    strings.TrimSpace(r.FormValue("email")),
		Password: password,
		FullName: strings.TrimSpace(r.FormValue("full_name")),
		RoleID:   roleID,
	}
	created, err := h.userSvc.Create(req)
	if err != nil {
		reload(err)
		return
	}

	h.auditSvc.Log(&user.ID, "CREATE_USER", "users", &created.ID, "Tambah pengguna baru: "+created.Username, r.RemoteAddr)
	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func (h *WebHandler) UserDelete(w http.ResponseWriter, r *http.Request, user *auth.User) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil || id == user.ID {
		http.Redirect(w, r, "/users", http.StatusSeeOther)
		return
	}

	h.userSvc.Delete(id)
	h.auditSvc.Log(&user.ID, "DELETE_USER", "users", &id, "Hapus pengguna", r.RemoteAddr)
	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func userHasRoleID(u *auth.User, roleID int) bool {
	for _, role := range u.Roles {
		if role.ID == roleID {
			return true
		}
	}
	return false
}
