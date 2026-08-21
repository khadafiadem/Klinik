package medicines

import (
	"database/sql"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(db *sql.DB) *Service {
	return &Service{repo: NewRepository(db)}
}

func (s *Service) GetAll(page, limit int, search string) ([]Medicine, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.GetAll(page, limit, search)
}

func (s *Service) GetByID(id int) (*Medicine, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(m *Medicine) error {
	if m.Name == "" {
		return fmt.Errorf("nama obat wajib diisi")
	}
	if m.MedicineCode == "" {
		return fmt.Errorf("kode obat wajib diisi")
	}
	return s.repo.Create(m)
}

func (s *Service) Update(m *Medicine) error {
	return s.repo.Update(m)
}

func (s *Service) Count() (int, error) {
	return s.repo.Count()
}

func (s *Service) CountLowStock() (int, error) {
	return s.repo.CountLowStock()
}

func (s *Service) GetLowStockMedicines() ([]Medicine, error) {
	return s.repo.GetLowStockMedicines()
}

func (s *Service) AddStock(medicineID, quantity int, batchNumber, notes string) error {
	if quantity <= 0 {
		return fmt.Errorf("jumlah harus lebih dari 0")
	}

	m, err := s.repo.GetByID(medicineID)
	if err != nil {
		return err
	}
	if !m.IsActive {
		return fmt.Errorf("obat tidak aktif")
	}

	tx := &StockTransaction{
		MedicineID:      medicineID,
		TransactionType: "MASUK",
		Quantity:        quantity,
		BatchNumber:     batchNumber,
		Notes:           notes,
	}
	if err := s.repo.AddStockTransaction(tx); err != nil {
		return err
	}
	return s.repo.UpdateMedicineStock(medicineID, quantity)
}

func (s *Service) ReduceStock(medicineID, quantity int, batchNumber, notes string) error {
	if quantity <= 0 {
		return fmt.Errorf("jumlah harus lebih dari 0")
	}

	m, err := s.repo.GetByID(medicineID)
	if err != nil {
		return err
	}
	if m.Stock < quantity {
		return fmt.Errorf("stok tidak mencukupi (stok: %d)", m.Stock)
	}

	tx := &StockTransaction{
		MedicineID:      medicineID,
		TransactionType: "KELUAR",
		Quantity:        quantity,
		BatchNumber:     batchNumber,
		Notes:           notes,
	}
	if err := s.repo.AddStockTransaction(tx); err != nil {
		return err
	}
	return s.repo.UpdateMedicineStock(medicineID, -quantity)
}

func (s *Service) GetStockTransactions(medicineID int) ([]StockTransaction, error) {
	return s.repo.GetStockTransactions(medicineID)
}

func (s *Service) GetAllCategories() ([]Category, error) {
	return s.repo.GetAllCategories()
}

func (s *Service) GetAllUnits() ([]Unit, error) {
	return s.repo.GetAllUnits()
}

func (s *Service) CreateCategory(name string) error {
	if name == "" {
		return fmt.Errorf("nama kategori wajib diisi")
	}
	return s.repo.CreateCategory(name)
}

func (s *Service) CreateUnit(name string) error {
	if name == "" {
		return fmt.Errorf("nama satuan wajib diisi")
	}
	return s.repo.CreateUnit(name)
}
