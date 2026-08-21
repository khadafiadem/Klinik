package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"klinik-app/internal/auth"
	"klinik-app/internal/finance"
)

func (h *WebHandler) InvoicesList(w http.ResponseWriter, r *http.Request, user *auth.User) {
	search := r.URL.Query().Get("search")
	list, _, _ := h.finSvc.GetAllInvoices(1, 100, search)
	RenderTemplate(w, r, "finance/invoices", TemplateData{
		User:   user,
		Data:   list,
		Search: search,
	})
}

func (h *WebHandler) InvoiceForm(w http.ResponseWriter, r *http.Request, user *auth.User) {
	patientsList, _, _ := h.patientSvc.GetAll(1, 1000, "")
	RenderTemplate(w, r, "finance/invoice_form", TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Patients": patientsList,
		},
	})
}

func (h *WebHandler) InvoiceSave(w http.ResponseWriter, r *http.Request, user *auth.User) {
	patientID, _ := strconv.Atoi(r.FormValue("patient_id"))
	var regID, mrID *int
	if v := r.FormValue("registration_id"); v != "" {
		id, _ := strconv.Atoi(v)
		regID = &id
	}
	if v := r.FormValue("medical_record_id"); v != "" {
		id, _ := strconv.Atoi(v)
		mrID = &id
	}
	discount, _ := strconv.ParseFloat(r.FormValue("discount"), 64)

	inv := &finance.Invoice{
		PatientID:       patientID,
		RegistrationID:  regID,
		MedicalRecordID: mrID,
		InvoiceDate:     r.FormValue("invoice_date"),
		Discount:        discount,
		Notes:           r.FormValue("notes"),
	}

	if err := h.finSvc.CreateInvoice(inv); err != nil {
		patientsList, _, _ := h.patientSvc.GetAll(1, 1000, "")
		RenderTemplate(w, r, "finance/invoice_form", TemplateData{
			User:  user,
			Error: err.Error(),
			Data: map[string]interface{}{
				"Patients": patientsList,
				"Form":     inv,
			},
		})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/invoices/%d", inv.ID), http.StatusSeeOther)
}

func (h *WebHandler) InvoiceView(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/invoices/")
	parts := strings.SplitN(path, "/", 2)

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Redirect(w, r, "/invoices", http.StatusSeeOther)
		return
	}

	inv, err := h.finSvc.GetInvoiceByID(id)
	if err != nil {
		http.Redirect(w, r, "/invoices", http.StatusSeeOther)
		return
	}

	RenderTemplate(w, r, "finance/invoice_view", TemplateData{
		User: user,
		Data: inv,
	})
}

func (h *WebHandler) InvoiceAddItem(w http.ResponseWriter, r *http.Request, user *auth.User) {
	invID, _ := strconv.Atoi(r.FormValue("invoice_id"))
	qty, _ := strconv.Atoi(r.FormValue("quantity"))
	price, _ := strconv.ParseFloat(r.FormValue("unit_price"), 64)

	item := &finance.InvoiceItem{
		InvoiceID:   invID,
		Description: r.FormValue("description"),
		Quantity:    qty,
		UnitPrice:   price,
		ItemType:    r.FormValue("item_type"),
	}

	if err := h.finSvc.AddInvoiceItem(item); err != nil {
		http.Redirect(w, r, fmt.Sprintf("/invoices/%d", invID), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/invoices/%d", invID), http.StatusSeeOther)
}

func (h *WebHandler) InvoiceRemoveItem(w http.ResponseWriter, r *http.Request, user *auth.User) {
	path := strings.TrimPrefix(r.URL.Path, "/invoices/item/")
	parts := strings.SplitN(path, "/", 2)
	invID, _ := strconv.Atoi(parts[0])
	itemID, _ := strconv.Atoi(parts[1])

	h.finSvc.RemoveInvoiceItem(itemID, invID)
	http.Redirect(w, r, fmt.Sprintf("/invoices/%d", invID), http.StatusSeeOther)
}

// --- Payments ---

func (h *WebHandler) PaymentsList(w http.ResponseWriter, r *http.Request, user *auth.User) {
	search := r.URL.Query().Get("search")
	list, _, _ := h.finSvc.GetAllPayments(1, 100, search)
	RenderTemplate(w, r, "finance/payments", TemplateData{
		User:   user,
		Data:   list,
		Search: search,
	})
}

func (h *WebHandler) PaymentForm(w http.ResponseWriter, r *http.Request, user *auth.User) {
	RenderTemplate(w, r, "finance/payment_form", TemplateData{
		User: user,
		Data: nil,
	})
}

func (h *WebHandler) PaymentSave(w http.ResponseWriter, r *http.Request, user *auth.User) {
	invID, _ := strconv.Atoi(r.FormValue("invoice_id"))
	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)

	pay := &finance.Payment{
		InvoiceID:       invID,
		PaymentDate:     r.FormValue("payment_date"),
		Amount:          amount,
		PaymentMethod:   r.FormValue("payment_method"),
		ReferenceNumber: r.FormValue("reference_number"),
		Notes:           r.FormValue("notes"),
	}

	if err := h.finSvc.ProcessPayment(pay); err != nil {
		RenderTemplate(w, r, "finance/payment_form", TemplateData{
			User:  user,
			Error: err.Error(),
			Data:  pay,
		})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/invoices/%d", invID), http.StatusSeeOther)
}
