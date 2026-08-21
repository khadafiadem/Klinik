package handler

import (
	"net/http"

	"klinik-app/internal/auth"
)

func (h *WebHandler) AuditLogsList(w http.ResponseWriter, r *http.Request, user *auth.User) {
	search := r.URL.Query().Get("search")
	list, _, _ := h.auditSvc.GetAll(1, 100, search)
	RenderTemplate(w, r, "audit/logs", TemplateData{
		User:   user,
		Data:   list,
		Search: search,
	})
}
