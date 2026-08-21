package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"klinik-app/internal/auth"
	"klinik-app/internal/medicines"
)

func (h *WebHandler) MedicinesList(w http.ResponseWriter, r *http.Request, user *auth.User) {
	search := r.URL.Query().Get("search")
	list, _, _ := h.medSvc.GetAll(1, 100, search)
	RenderTemplate(w, r, "medicines/list", TemplateData{
		User:   user,
		Data:   list,
		Search: search,
	})
}

func (h *WebHandler) MedicineForm(w http.ResponseWriter, r *http.Request, user *auth.User) {
	categories, _ := h.medSvc.GetAllCategories()
	units, _ := h.medSvc.GetAllUnits()
	RenderTemplate(w, r, "medicines/form", TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Categories": categories,
			"Units":      units,
		},
	})
}

func (h *WebHandler) MedicineSave(w http.ResponseWriter, r *http.Request, user *auth.User) {
	catID, _ := strconv.Atoi(r.FormValue("category_id"))
	unitID, _ := strconv.Atoi(r.FormValue("unit_id"))
	purchasePrice, _ := strconv.ParseFloat(r.FormValue("purchase_price"), 64)
	sellingPrice, _ := strconv.ParseFloat(r.FormValue("selling_price"), 64)
	stock, _ := strconv.Atoi(r.FormValue("stock"))
	minStock, _ := strconv.Atoi(r.FormValue("minimum_stock"))

	var catPtr, unitPtr *int
	if catID > 0 {
		catPtr = &catID
	}
	if unitID > 0 {
		unitPtr = &unitID
	}

	med := &medicines.Medicine{
		MedicineCode:  r.FormValue("medicine_code"),
		Name:          r.FormValue("name"),
		GenericName:   r.FormValue("generic_name"),
		CategoryID:    catPtr,
		UnitID:        unitPtr,
		Form:          r.FormValue("form"),
		PurchasePrice: purchasePrice,
		SellingPrice:  sellingPrice,
		Stock:         stock,
		MinimumStock:  minStock,
		IsActive:      r.FormValue("is_active") == "on" || r.FormValue("is_active") == "true",
	}

	var err error
	editID, _ := strconv.Atoi(r.FormValue("id"))
	if editID > 0 {
		med.ID = editID
		err = h.medSvc.Update(med)
	} else {
		err = h.medSvc.Create(med)
	}

	if err != nil {
		categories, _ := h.medSvc.GetAllCategories()
		units, _ := h.medSvc.GetAllUnits()
		RenderTemplate(w, r, "medicines/form", TemplateData{
			User:  user,
			Error: err.Error(),
			Data: map[string]interface{}{
				"Categories": categories,
				"Units":      units,
				"Form":       med,
			},
		})
		return
	}

	http.Redirect(w, r, "/medicines", http.StatusSeeOther)
}

func (h *WebHandler) MedicinePage(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/medicines/")
	parts := strings.SplitN(path, "/", 2)

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Redirect(w, r, "/medicines", http.StatusSeeOther)
		return
	}

	// /medicines/{id}/stock
	if len(parts) > 1 && parts[1] == "stock" {
		h.medicineStockPage(w, r, user, id)
		return
	}

	// /medicines/{id}/edit
	if len(parts) > 1 && parts[1] == "edit" {
		h.medicineEditPage(w, r, user, id)
		return
	}

	// Default: /medicines/{id} -> stock page
	h.medicineStockPage(w, r, user, id)
}

func (h *WebHandler) medicineStockPage(w http.ResponseWriter, r *http.Request, user *auth.User, id int) {
	med, err := h.medSvc.GetByID(id)
	if err != nil {
		http.Redirect(w, r, "/medicines", http.StatusSeeOther)
		return
	}

	transactions, _ := h.medSvc.GetStockTransactions(id)

	RenderTemplate(w, r, "medicines/stock", TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Medicine":     med,
			"Transactions": transactions,
		},
	})
}

func (h *WebHandler) medicineEditPage(w http.ResponseWriter, r *http.Request, user *auth.User, id int) {
	med, _ := h.medSvc.GetByID(id)
	categories, _ := h.medSvc.GetAllCategories()
	units, _ := h.medSvc.GetAllUnits()

	RenderTemplate(w, r, "medicines/form", TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Categories": categories,
			"Units":      units,
			"Form":       med,
			"Edit":       true,
		},
	})
}

func (h *WebHandler) MedicineStockAdd(w http.ResponseWriter, r *http.Request, user *auth.User) {
	medID, _ := strconv.Atoi(r.FormValue("medicine_id"))
	qty, _ := strconv.Atoi(r.FormValue("quantity"))
	batch := r.FormValue("batch_number")
	notes := r.FormValue("notes")

	if err := h.medSvc.AddStock(medID, qty, batch, notes); err != nil {
		http.Redirect(w, r, fmt.Sprintf("/medicines/%d/stock", medID), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/medicines/%d/stock", medID), http.StatusSeeOther)
}

func (h *WebHandler) MedicineStockReduce(w http.ResponseWriter, r *http.Request, user *auth.User) {
	medID, _ := strconv.Atoi(r.FormValue("medicine_id"))
	qty, _ := strconv.Atoi(r.FormValue("quantity"))
	batch := r.FormValue("batch_number")
	notes := r.FormValue("notes")

	if err := h.medSvc.ReduceStock(medID, qty, batch, notes); err != nil {
		http.Redirect(w, r, fmt.Sprintf("/medicines/%d/stock", medID), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/medicines/%d/stock", medID), http.StatusSeeOther)
}
