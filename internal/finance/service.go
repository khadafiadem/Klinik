package finance

import (
	"database/sql"
	"fmt"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(db *sql.DB) *Service {
	return &Service{repo: NewRepository(db)}
}

func (s *Service) GetAllInvoices(page, limit int, search string) ([]Invoice, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.GetAllInvoices(page, limit, search)
}

func (s *Service) GetInvoiceByID(id int) (*Invoice, error) {
	return s.repo.GetInvoiceByID(id)
}

func (s *Service) CreateInvoice(inv *Invoice) error {
	if inv.PatientID == 0 {
		return fmt.Errorf("pasien wajib dipilih")
	}

	num, err := s.repo.GenerateInvoiceNumber()
	if err != nil {
		return err
	}
	inv.InvoiceNumber = num

	if strings.TrimSpace(inv.Status) == "" {
		inv.Status = "BELUM_BAYAR"
	}

	return s.repo.CreateInvoice(inv)
}

func (s *Service) AddInvoiceItem(ii *InvoiceItem) error {
	if ii.Quantity < 1 {
		return fmt.Errorf("jumlah minimal 1")
	}
	if err := s.repo.AddInvoiceItem(ii); err != nil {
		return err
	}
	return s.repo.RecalculateInvoiceTotal(ii.InvoiceID)
}

func (s *Service) RemoveInvoiceItem(id, invoiceID int) error {
	if err := s.repo.RemoveInvoiceItem(id); err != nil {
		return err
	}
	return s.repo.RecalculateInvoiceTotal(invoiceID)
}

func (s *Service) ProcessPayment(pay *Payment) error {
	if pay.InvoiceID == 0 {
		return fmt.Errorf("tagihan wajib dipilih")
	}
	if pay.Amount <= 0 {
		return fmt.Errorf("jumlah pembayaran harus lebih dari 0")
	}

	inv, err := s.repo.GetInvoiceByID(pay.InvoiceID)
	if err != nil {
		return err
	}

	paid, _ := s.repo.GetTotalPaidForInvoice(pay.InvoiceID)
	remaining := inv.Total - paid

	if pay.Amount > remaining {
		return fmt.Errorf("jumlah pembayaran (%.0f) melebihi sisa tagihan (%.0f)", pay.Amount, remaining)
	}

	num, err := s.repo.GeneratePaymentNumber()
	if err != nil {
		return err
	}
	pay.PaymentNumber = num
	pay.PatientID = inv.PatientID
	if strings.TrimSpace(pay.Status) == "" {
		pay.Status = "COMPLETED"
	}

	if err := s.repo.CreatePayment(pay); err != nil {
		return err
	}

	newPaid, _ := s.repo.GetTotalPaidForInvoice(pay.InvoiceID)
	if newPaid >= inv.Total {
		s.repo.UpdateInvoiceStatus(pay.InvoiceID, "SUDAH_BAYAR")
	} else if newPaid > 0 {
		s.repo.UpdateInvoiceStatus(pay.InvoiceID, "SEBAGIAN")
	}

	return nil
}

func (s *Service) GetAllPayments(page, limit int, search string) ([]Payment, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.GetAllPayments(page, limit, search)
}

func (s *Service) InvoiceCount() (int, error) {
	return s.repo.InvoiceCount()
}

func (s *Service) TodayRevenue() (float64, error) {
	return s.repo.TodayRevenue()
}

func (s *Service) UnpaidCount() (int, error) {
	return s.repo.UnpaidCount()
}
