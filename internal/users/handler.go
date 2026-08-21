package users

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Method tidak diizinkan",
		})
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")

	users, total, err := h.service.GetAll(page, limit, search)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Gagal mengambil data user",
		})
		return
	}

	sendJSON(w, http.StatusOK, PaginatedResponse{
		Success: true,
		Message: "Berhasil mengambil data user",
		Data:    users,
		Total:   total,
		Page:    page,
		Limit:   limit,
	})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Method tidak diizinkan",
		})
		return
	}

	id, err := extractID(r)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "ID tidak valid",
		})
		return
	}

	user, err := h.service.GetByID(id)
	if err != nil {
		sendJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Berhasil mengambil data user",
		Data:    user,
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Method tidak diizinkan",
		})
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Format request tidak valid",
		})
		return
	}

	user, err := h.service.Create(req)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "sudah digunakan") {
			status = http.StatusConflict
		}
		sendJSON(w, status, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	sendJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "User berhasil dibuat",
		Data:    user,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Method tidak diizinkan",
		})
		return
	}

	id, err := extractID(r)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "ID tidak valid",
		})
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Format request tidak valid",
		})
		return
	}

	if err := h.service.Update(id, req); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "User berhasil diupdate",
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Method tidak diizinkan",
		})
		return
	}

	id, err := extractID(r)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "ID tidak valid",
		})
		return
	}

	if err := h.service.Delete(id); err != nil {
		sendJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "User berhasil dihapus",
	})
}

func extractID(r *http.Request) (int, error) {
	path := strings.TrimPrefix(r.URL.Path, "/api/users/")
	path = strings.TrimSuffix(path, "/")
	return strconv.Atoi(path)
}

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
