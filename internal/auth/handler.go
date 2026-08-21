package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"klinik-app/internal/logger"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Method tidak diizinkan",
		})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Format request tidak valid",
		})
		return
	}

	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		sendJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Username dan password harus diisi",
		})
		return
	}

	resp, err := h.service.Login(req)
	if err != nil {
		logger.Error.Printf("Login gagal: %v", err)
		sendJSON(w, http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	logger.Info.Printf("Login berhasil: %s", req.Username)
	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Login berhasil",
		Data:    resp,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Method tidak diizinkan",
		})
		return
	}

	// Di sisi server, JWT stateless tidak perlu invalidasi token
	// Client cukup menghapus token dari storage
	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Logout berhasil",
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Method tidak diizinkan",
		})
		return
	}

	claims, ok := r.Context().Value("claims").(*Claims)
	if !ok {
		sendJSON(w, http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "Tidak terautentikasi",
		})
		return
	}

	user, err := h.service.GetUserByID(claims.UserID)
	if err != nil {
		sendJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "User tidak ditemukan",
		})
		return
	}

	roles, err := h.service.GetUserRoles(claims.UserID)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Gagal mengambil data role",
		})
		return
	}

	permissions, err := h.service.GetUserPermissions(claims.UserID)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Gagal mengambil data permission",
		})
		return
	}

	data := map[string]interface{}{
		"user":        user,
		"roles":       roles,
		"permissions": permissions,
	}

	sendJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Berhasil mengambil data user",
		Data:    data,
	})
}

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
