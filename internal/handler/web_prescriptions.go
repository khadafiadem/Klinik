package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"klinik-app/internal/auth"
	"klinik-app/internal/logger"
	"klinik-app/internal/prescriptions"
)

// HasAnyRole memeriksa apakah user memiliki salah satu role yang diizinkan.
func HasAnyRole(u *auth.User, roles ...string) bool {
	for _, ur := range u.Roles {
		for _, r := range roles {
			if strings.EqualFold(ur.Name, r) {
				return true
			}
		}
	}
	return false
}

func (h *WebHandler) PrescriptionsList(w http.ResponseWriter, r *http.Request, user *auth.User) {
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")
	list, _, err := h.rxSvc.GetAll(1, 100, search, status)
	if err != nil {
		logger.Error.Printf("Gagal memuat resep: %v", err)
	}
	RenderTemplate(w, r, "prescriptions/list", TemplateData{
		User:   user,
		Data:   list,
		Search: search,
		Data2:  status,
	})
}

func (h *WebHandler) PrescriptionView(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/prescriptions/")
	parts := strings.SplitN(path, "/", 2)

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Redirect(w, r, "/prescriptions", http.StatusSeeOther)
		return
	}

	if len(parts) > 1 && parts[1] != "" {
		if parts[1] == "print" {
			h.PrescriptionPrint(w, r, id, user)
			return
		}
		h.prescriptionAction(w, r, id, parts[1], user)
		return
	}

	rx, err := h.rxSvc.GetByID(id)
	if err != nil {
		http.Redirect(w, r, "/prescriptions", http.StatusSeeOther)
		return
	}

	medicinesList, _, err := h.medSvc.GetAll(1, 100, "")
	if err != nil {
		logger.Error.Printf("Gagal memuat daftar obat: %v", err)
	}

	RenderTemplate(w, r, "prescriptions/view", TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Prescription": rx,
			"Medicines":    medicinesList,
		},
	})
}

func (h *WebHandler) PrescriptionPrint(w http.ResponseWriter, r *http.Request, id int, user *auth.User) {
	if !HasAnyRole(user, "ADMIN", "PHARMACIST", "DOCTOR") {
		RenderForbidden(w, r, user)
		return
	}

	rx, err := h.rxSvc.GetByID(id)
	if err != nil {
		http.Redirect(w, r, "/prescriptions", http.StatusSeeOther)
		return
	}

	settings, _ := h.clinicSvc.Get()
	RenderPrint(w, "prescriptions/print", map[string]interface{}{
		"Prescription": rx,
		"Clinic":       settings,
		"PrintedBy":    user.FullName,
	})
}

func (h *WebHandler) prescriptionAction(w http.ResponseWriter, r *http.Request, id int, action string, user *auth.User) {
	// Perubahan status resep adalah kewenangan apotek (atau admin).
	if !HasAnyRole(user, "ADMIN", "PHARMACIST") {
		RenderForbidden(w, r, user)
		return
	}

	var err error
	switch action {
	case "process":
		err = h.rxSvc.Process(id)
	case "complete":
		err = h.rxSvc.Complete(id, user.ID)
	case "cancel":
		err = h.rxSvc.Cancel(id)
	default:
		http.Redirect(w, r, fmt.Sprintf("/prescriptions/%d", id), http.StatusSeeOther)
		return
	}

	if err != nil {
		logger.Error.Printf("Aksi resep %d (%s) oleh %s gagal: %v", id, action, user.Username, err)
		rx, rxErr := h.rxSvc.GetByID(id)
		if rxErr != nil {
			http.Redirect(w, r, "/prescriptions", http.StatusSeeOther)
			return
		}
		medicinesList, _, _ := h.medSvc.GetAll(1, 100, "")
		RenderTemplate(w, r, "prescriptions/view", TemplateData{
			User:  user,
			Error: err.Error(),
			Data: map[string]interface{}{
				"Prescription": rx,
				"Medicines":    medicinesList,
			},
		})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/prescriptions/%d", id), http.StatusSeeOther)
}

func (h *WebHandler) PrescriptionAddItem(w http.ResponseWriter, r *http.Request, user *auth.User) {
	rxID, _ := strconv.Atoi(r.FormValue("prescription_id"))
	qty, _ := strconv.Atoi(r.FormValue("quantity"))
	var medID *int
	if v := r.FormValue("medicine_id"); v != "" {
		id, _ := strconv.Atoi(v)
		medID = &id
	}

	item := &prescriptions.PrescriptionItem{
		PrescriptionID: rxID,
		MedicineID:     medID,
		Quantity:       qty,
		Dosage:         r.FormValue("dosage"),
		Frequency:      r.FormValue("frequency"),
		Duration:       r.FormValue("duration"),
		Instructions:   r.FormValue("instructions"),
	}

	if err := h.rxSvc.AddItem(item); err != nil {
		logger.Error.Printf("Tambah item resep %d gagal: %v", rxID, err)
	}

	http.Redirect(w, r, fmt.Sprintf("/prescriptions/%d", rxID), http.StatusSeeOther)
}

func (h *WebHandler) PrescriptionRemoveItem(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/prescriptions/item/")
	parts := strings.SplitN(path, "/", 2)
	rxID, _ := strconv.Atoi(parts[0])
	itemID, _ := strconv.Atoi(parts[1])

	if err := h.rxSvc.RemoveItem(itemID); err != nil {
		logger.Error.Printf("Hapus item resep %d gagal: %v", itemID, err)
	}
	http.Redirect(w, r, fmt.Sprintf("/prescriptions/%d", rxID), http.StatusSeeOther)
}
