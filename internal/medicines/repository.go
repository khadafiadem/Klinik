package medicines

import (
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(page, limit int, search string) ([]Medicine, int, error) {
	offset := (page - 1) * limit
	args := []interface{}{}
	where := ""

	if search != "" {
		where = `WHERE (m.medicine_code ILIKE $1 OR m.name ILIKE $1 OR m.generic_name ILIKE $1)`
		args = append(args, "%"+search+"%")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM medicines m %s`, where)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	idx := len(args)
	query := fmt.Sprintf(`SELECT m.id, m.medicine_code, m.name, COALESCE(m.generic_name,''),
		COALESCE(mc.name,''), COALESCE(mu.name,''), COALESCE(m.form,''),
		m.purchase_price, m.selling_price, m.stock, m.minimum_stock, m.is_active,
		m.created_at, m.updated_at
		FROM medicines m
		LEFT JOIN medicine_categories mc ON m.category_id = mc.id
		LEFT JOIN medicine_units mu ON m.unit_id = mu.id
		%s ORDER BY m.name ASC LIMIT $%d OFFSET $%d`, where, idx-1, idx)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []Medicine
	for rows.Next() {
		var med Medicine
		if err := rows.Scan(&med.ID, &med.MedicineCode, &med.Name, &med.GenericName,
			&med.CategoryName, &med.UnitName, &med.Form,
			&med.PurchasePrice, &med.SellingPrice, &med.Stock, &med.MinimumStock, &med.IsActive,
			&med.CreatedAt, &med.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, med)
	}
	return list, total, nil
}

func (r *Repository) GetByID(id int) (*Medicine, error) {
	m := &Medicine{}
	query := `SELECT m.id, m.medicine_code, m.name, COALESCE(m.generic_name,''),
		m.category_id, COALESCE(mc.name,''), m.unit_id, COALESCE(mu.name,''), COALESCE(m.form,''),
		m.purchase_price, m.selling_price, m.stock, m.minimum_stock, m.is_active,
		m.created_at, m.updated_at
		FROM medicines m
		LEFT JOIN medicine_categories mc ON m.category_id = mc.id
		LEFT JOIN medicine_units mu ON m.unit_id = mu.id
		WHERE m.id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&m.ID, &m.MedicineCode, &m.Name, &m.GenericName,
		&m.CategoryID, &m.CategoryName, &m.UnitID, &m.UnitName, &m.Form,
		&m.PurchasePrice, &m.SellingPrice, &m.Stock, &m.MinimumStock, &m.IsActive,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("obat tidak ditemukan")
		}
		return nil, err
	}
	return m, nil
}

func (r *Repository) Create(m *Medicine) error {
	query := `INSERT INTO medicines (medicine_code, name, generic_name, category_id, unit_id, form,
		purchase_price, selling_price, stock, minimum_stock, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, m.MedicineCode, m.Name, m.GenericName, m.CategoryID, m.UnitID,
		m.Form, m.PurchasePrice, m.SellingPrice, m.Stock, m.MinimumStock, m.IsActive).
		Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

func (r *Repository) Update(m *Medicine) error {
	_, err := r.db.Exec(`UPDATE medicines SET name=$1, generic_name=$2, category_id=$3, unit_id=$4,
		form=$5, purchase_price=$6, selling_price=$7, minimum_stock=$8, is_active=$9
		WHERE id=$10`,
		m.Name, m.GenericName, m.CategoryID, m.UnitID, m.Form,
		m.PurchasePrice, m.SellingPrice, m.MinimumStock, m.IsActive, m.ID)
	return err
}

func (r *Repository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM medicines").Scan(&count)
	return count, err
}

func (r *Repository) CountLowStock() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM medicines WHERE stock <= minimum_stock AND is_active=true").Scan(&count)
	return count, err
}

func (r *Repository) GetLowStockMedicines() ([]Medicine, error) {
	query := `SELECT m.id, m.medicine_code, m.name, m.stock, m.minimum_stock
		FROM medicines m WHERE m.stock <= m.minimum_stock AND m.is_active=true ORDER BY m.stock ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Medicine
	for rows.Next() {
		var m Medicine
		if err := rows.Scan(&m.ID, &m.MedicineCode, &m.Name, &m.Stock, &m.MinimumStock); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

// Stock transactions
func (r *Repository) AddStockTransaction(tx *StockTransaction) error {
	query := `INSERT INTO medicine_stock_transactions (medicine_id, transaction_type, quantity,
		batch_number, expiration_date, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`
	return r.db.QueryRow(query, tx.MedicineID, tx.TransactionType, tx.Quantity,
		tx.BatchNumber, tx.ExpirationDate, tx.Notes, tx.CreatedBy).Scan(&tx.ID, &tx.CreatedAt)
}

func (r *Repository) UpdateMedicineStock(medicineID int, delta int) error {
	_, err := r.db.Exec("UPDATE medicines SET stock = stock + $1 WHERE id = $2 AND stock + $1 >= 0", delta, medicineID)
	return err
}

func (r *Repository) GetStockTransactions(medicineID int) ([]StockTransaction, error) {
	query := `SELECT id, medicine_id, transaction_type, quantity,
		COALESCE(batch_number,''), expiration_date, COALESCE(notes,''), created_at
		FROM medicine_stock_transactions WHERE medicine_id = $1 ORDER BY created_at DESC LIMIT 50`
	rows, err := r.db.Query(query, medicineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []StockTransaction
	for rows.Next() {
		var tx StockTransaction
		if err := rows.Scan(&tx.ID, &tx.MedicineID, &tx.TransactionType, &tx.Quantity,
			&tx.BatchNumber, &tx.ExpirationDate, &tx.Notes, &tx.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, tx)
	}
	return list, nil
}

// Categories
func (r *Repository) GetAllCategories() ([]Category, error) {
	rows, err := r.db.Query("SELECT id, name FROM medicine_categories ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *Repository) CreateCategory(name string) error {
	_, err := r.db.Exec("INSERT INTO medicine_categories (name) VALUES ($1) ON CONFLICT DO NOTHING", name)
	return err
}

// Units
func (r *Repository) GetAllUnits() ([]Unit, error) {
	rows, err := r.db.Query("SELECT id, name FROM medicine_units ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Unit
	for rows.Next() {
		var u Unit
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, nil
}

func (r *Repository) CreateUnit(name string) error {
	_, err := r.db.Exec("INSERT INTO medicine_units (name) VALUES ($1) ON CONFLICT DO NOTHING", name)
	return err
}
