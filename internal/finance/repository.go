package finance

import (
	"database/sql"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// --- Invoices ---

func (r *Repository) GetAllInvoices(page, limit int, search string) ([]Invoice, int, error) {
	offset := (page - 1) * limit
	args := []interface{}{}
	where := ""

	if search != "" {
		where = `WHERE (inv.invoice_number ILIKE $1 OR p.full_name ILIKE $1 OR p.medical_record_number ILIKE $1)`
		args = append(args, "%"+search+"%")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM invoices inv
		JOIN patients p ON inv.patient_id = p.id %s`, where)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	idx := len(args)
	query := fmt.Sprintf(`SELECT inv.id, inv.invoice_number, inv.patient_id, p.full_name, p.medical_record_number,
		inv.invoice_date, inv.subtotal, inv.discount, inv.total,
		COALESCE(inv.notes,''), inv.status, inv.created_by, inv.created_at, inv.updated_at
		FROM invoices inv
		JOIN patients p ON inv.patient_id = p.id
		%s ORDER BY inv.id DESC LIMIT $%d OFFSET $%d`, where, idx-1, idx)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Invoice
	for rows.Next() {
		var inv Invoice
		if err := rows.Scan(&inv.ID, &inv.InvoiceNumber, &inv.PatientID, &inv.PatientName, &inv.PatientMRN,
			&inv.InvoiceDate, &inv.Subtotal, &inv.Discount, &inv.Total,
			&inv.Notes, &inv.Status, &inv.CreatedBy, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, inv)
	}
	return list, total, nil
}

func (r *Repository) GetInvoiceByID(id int) (*Invoice, error) {
	inv := &Invoice{}
	query := `SELECT inv.id, inv.invoice_number, inv.patient_id, p.full_name, p.medical_record_number,
		inv.invoice_date, inv.subtotal, inv.discount, inv.total,
		COALESCE(inv.notes,''), inv.status, inv.created_by, inv.created_at, inv.updated_at
		FROM invoices inv
		JOIN patients p ON inv.patient_id = p.id
		WHERE inv.id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&inv.ID, &inv.InvoiceNumber, &inv.PatientID, &inv.PatientName, &inv.PatientMRN,
		&inv.InvoiceDate, &inv.Subtotal, &inv.Discount, &inv.Total,
		&inv.Notes, &inv.Status, &inv.CreatedBy, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tagihan tidak ditemukan")
		}
		return nil, err
	}
	inv.Items, _ = r.GetInvoiceItems(id)
	inv.Payments, _ = r.GetPaymentsByInvoiceID(id)
	return inv, nil
}

func (r *Repository) CreateInvoice(inv *Invoice) error {
	query := `INSERT INTO invoices (invoice_number, patient_id, registration_id, medical_record_id,
		invoice_date, subtotal, discount, total, notes, status, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, inv.InvoiceNumber, inv.PatientID, inv.RegistrationID, inv.MedicalRecordID,
		inv.InvoiceDate, inv.Subtotal, inv.Discount, inv.Total, inv.Notes, inv.Status, inv.CreatedBy).
		Scan(&inv.ID, &inv.CreatedAt, &inv.UpdatedAt)
}

func (r *Repository) UpdateInvoiceStatus(id int, status string) error {
	_, err := r.db.Exec("UPDATE invoices SET status=$1 WHERE id=$2", status, id)
	return err
}

func (r *Repository) GetInvoiceItems(invoiceID int) ([]InvoiceItem, error) {
	rows, err := r.db.Query(`SELECT id, invoice_id, description, quantity, unit_price, total_price,
		COALESCE(item_type,''), reference_id FROM invoice_items WHERE invoice_id = $1`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []InvoiceItem
	for rows.Next() {
		var ii InvoiceItem
		if err := rows.Scan(&ii.ID, &ii.InvoiceID, &ii.Description, &ii.Quantity,
			&ii.UnitPrice, &ii.TotalPrice, &ii.ItemType, &ii.ReferenceID); err != nil {
			return nil, err
		}
		list = append(list, ii)
	}
	return list, nil
}

func (r *Repository) AddInvoiceItem(ii *InvoiceItem) error {
	ii.TotalPrice = float64(ii.Quantity) * ii.UnitPrice
	query := `INSERT INTO invoice_items (invoice_id, description, quantity, unit_price, total_price, item_type, reference_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`
	return r.db.QueryRow(query, ii.InvoiceID, ii.Description, ii.Quantity,
		ii.UnitPrice, ii.TotalPrice, ii.ItemType, ii.ReferenceID).Scan(&ii.ID)
}

func (r *Repository) RemoveInvoiceItem(id int) error {
	_, err := r.db.Exec("DELETE FROM invoice_items WHERE id=$1", id)
	return err
}

func (r *Repository) RecalculateInvoiceTotal(invoiceID int) error {
	_, err := r.db.Exec(`UPDATE invoices SET subtotal = COALESCE((SELECT SUM(total_price) FROM invoice_items WHERE invoice_id=$1),0),
		total = COALESCE((SELECT SUM(total_price) FROM invoice_items WHERE invoice_id=$1),0) - discount
		WHERE id=$1`, invoiceID)
	return err
}

func (r *Repository) GenerateInvoiceNumber() (string, error) {
	var next int
	err := r.db.QueryRow("SELECT nextval('inv_seq')").Scan(&next)
	if err != nil {
		return "", fmt.Errorf("gagal generate nomor invoice: %w", err)
	}
	var dateStr string
	if err := r.db.QueryRow("SELECT TO_CHAR(CURRENT_DATE, 'YYYYMMDD')").Scan(&dateStr); err != nil || dateStr == "" {
		dateStr = time.Now().Format("20060102")
	}
	return fmt.Sprintf("INV-%s-%03d", dateStr, next), nil
}

func (r *Repository) InvoiceCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM invoices WHERE invoice_date = CURRENT_DATE").Scan(&count)
	return count, err
}

func (r *Repository) TodayRevenue() (float64, error) {
	var total float64
	err := r.db.QueryRow(`SELECT COALESCE(SUM(amount),0) FROM payments
		WHERE payment_date = CURRENT_DATE AND status = 'COMPLETED'`).Scan(&total)
	return total, err
}

func (r *Repository) UnpaidCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM invoices WHERE status IN ('BELUM_BAYAR','SEBAGIAN')").Scan(&count)
	return count, err
}

// --- Payments ---

func (r *Repository) GetAllPayments(page, limit int, search string) ([]Payment, int, error) {
	offset := (page - 1) * limit
	args := []interface{}{}
	where := ""

	if search != "" {
		where = `WHERE (pay.payment_number ILIKE $1 OR p.full_name ILIKE $1 OR inv.invoice_number ILIKE $1)`
		args = append(args, "%"+search+"%")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM payments pay
		JOIN patients p ON pay.patient_id = p.id
		JOIN invoices inv ON pay.invoice_id = inv.id %s`, where)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	idx := len(args)
	query := fmt.Sprintf(`SELECT pay.id, pay.payment_number, pay.invoice_id, inv.invoice_number,
		pay.patient_id, p.full_name, pay.payment_date, pay.amount,
		pay.payment_method, COALESCE(pay.reference_number,''), COALESCE(pay.notes,''),
		pay.status, pay.created_at
		FROM payments pay
		JOIN patients p ON pay.patient_id = p.id
		JOIN invoices inv ON pay.invoice_id = inv.id
		%s ORDER BY pay.id DESC LIMIT $%d OFFSET $%d`, where, idx-1, idx)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Payment
	for rows.Next() {
		var pay Payment
		if err := rows.Scan(&pay.ID, &pay.PaymentNumber, &pay.InvoiceID, &pay.PaymentNumber,
			&pay.PatientID, &pay.PatientName, &pay.PaymentDate, &pay.Amount,
			&pay.PaymentMethod, &pay.ReferenceNumber, &pay.Notes, &pay.Status, &pay.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, pay)
	}
	return list, total, nil
}

func (r *Repository) GetPaymentsByInvoiceID(invoiceID int) ([]Payment, error) {
	rows, err := r.db.Query(`SELECT id, payment_number, invoice_id, patient_id,
		payment_date, amount, payment_method, COALESCE(reference_number,''), COALESCE(notes,''), status, created_at
		FROM payments WHERE invoice_id = $1 ORDER BY created_at DESC`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Payment
	for rows.Next() {
		var pay Payment
		if err := rows.Scan(&pay.ID, &pay.PaymentNumber, &pay.InvoiceID, &pay.PatientID,
			&pay.PaymentDate, &pay.Amount, &pay.PaymentMethod, &pay.ReferenceNumber,
			&pay.Notes, &pay.Status, &pay.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, pay)
	}
	return list, nil
}

func (r *Repository) CreatePayment(pay *Payment) error {
	query := `INSERT INTO payments (payment_number, invoice_id, patient_id, payment_date,
		amount, payment_method, reference_number, notes, status, received_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id, created_at`
	return r.db.QueryRow(query, pay.PaymentNumber, pay.InvoiceID, pay.PatientID,
		pay.PaymentDate, pay.Amount, pay.PaymentMethod, pay.ReferenceNumber,
		pay.Notes, pay.Status, pay.ReceivedBy).Scan(&pay.ID, &pay.CreatedAt)
}

func (r *Repository) GeneratePaymentNumber() (string, error) {
	var next int
	err := r.db.QueryRow("SELECT nextval('pay_seq')").Scan(&next)
	if err != nil {
		return "PAY-20260821-001", nil
	}
	var dateStr string
	_ = r.db.QueryRow("SELECT TO_CHAR(CURRENT_DATE, 'YYYYMMDD')").Scan(&dateStr)
	if dateStr == "" {
		dateStr = "20260821"
	}
	return fmt.Sprintf("PAY-%s-%03d", dateStr, next), nil
}

func (r *Repository) GetTotalPaidForInvoice(invoiceID int) (float64, error) {
	var total float64
	err := r.db.QueryRow(`SELECT COALESCE(SUM(amount),0) FROM payments WHERE invoice_id=$1 AND status='COMPLETED'`, invoiceID).Scan(&total)
	return total, err
}
