package handler

import (
	"net/http"
	"strconv"
	"strings"

	"klinik-app/internal/auth"
	"klinik-app/internal/staff"
)

func (h *WebHandler) StaffList(w http.ResponseWriter, r *http.Request, user *auth.User) {
	search := r.URL.Query().Get("search")
	staffList, _, _ := h.staffSvc.GetAll(1, 100, search)
	RenderTemplate(w, r, "staff/list", TemplateData{User: user, Data: staffList})
}

func (h *WebHandler) StaffForm(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/staff/")
	path = strings.TrimSuffix(path, "/edit")
	path = strings.TrimSuffix(path, "/new")

	var s *staff.Staff
	if path != "" && path != "new" {
		id, err := strconv.Atoi(path)
		if err == nil {
			s, _ = h.staffSvc.GetByID(id)
		}
	}

	RenderTemplate(w, r, "staff/form", TemplateData{User: user, Data: s})
}

func (h *WebHandler) StaffSave(w http.ResponseWriter, r *http.Request, user *auth.User) {
	idStr := r.FormValue("id")
	isEdit := idStr != ""

	s := &staff.Staff{
		StaffCode: strings.TrimSpace(r.FormValue("staff_code")),
		FullName:  r.FormValue("full_name"),
		Position:  r.FormValue("position"),
		Phone:     r.FormValue("phone"),
		Email:     r.FormValue("email"),
	}
	s.IsActive = r.FormValue("is_active") != "false"

	var err error
	if isEdit {
		id, _ := strconv.Atoi(idStr)
		err = h.staffSvc.Update(id, s)
	} else {
		err = h.staffSvc.Create(s)
	}

	if err != nil {
		RenderTemplate(w, r, "staff/form", TemplateData{User: user, Data: s, Error: err.Error()})
		return
	}

	http.Redirect(w, r, "/staff", http.StatusSeeOther)
}

func (h *WebHandler) StaffDelete(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/staff/")
	path = strings.TrimSuffix(path, "/delete")

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Redirect(w, r, "/staff", http.StatusSeeOther)
		return
	}

	h.staffSvc.Delete(id)
	http.Redirect(w, r, "/staff", http.StatusSeeOther)
}
