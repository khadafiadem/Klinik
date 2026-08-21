package medicines

import (
	"database/sql"
	"time"
)

type Medicine struct {
	ID             int       `json:"id"`
	MedicineCode   string    `json:"medicine_code"`
	Name           string    `json:"name"`
	GenericName    string    `json:"generic_name"`
	CategoryID     *int      `json:"category_id,omitempty"`
	CategoryName   string    `json:"category_name,omitempty"`
	UnitID         *int      `json:"unit_id,omitempty"`
	UnitName       string    `json:"unit_name,omitempty"`
	Form           string    `json:"form"`
	PurchasePrice  float64   `json:"purchase_price"`
	SellingPrice   float64   `json:"selling_price"`
	Stock          int       `json:"stock"`
	MinimumStock   int       `json:"minimum_stock"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Unit struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type StockTransaction struct {
	ID              int            `json:"id"`
	MedicineID      int            `json:"medicine_id"`
	MedicineName    string         `json:"medicine_name,omitempty"`
	TransactionType string         `json:"transaction_type"`
	Quantity        int            `json:"quantity"`
	BatchNumber     string         `json:"batch_number"`
	ExpirationDate  sql.NullString `json:"expiration_date"`
	Notes           string         `json:"notes"`
	CreatedBy       *int           `json:"created_by,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
}
