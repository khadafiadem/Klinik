package main

import (
	"fmt"
	"os"

	"klinik-app/internal/config"
	"klinik-app/internal/database"
	"klinik-app/internal/logger"
)

func main() {
	logger.Init("info")
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close(db)

	catMap := map[string]int{}
	rows, _ := db.Query("SELECT id, name FROM medicine_categories")
	if rows != nil {
		for rows.Next() {
			var id int
			var name string
			rows.Scan(&id, &name)
			catMap[name] = id
		}
		rows.Close()
	}

	unitMap := map[string]int{}
	rows2, _ := db.Query("SELECT id, name FROM medicine_units")
	if rows2 != nil {
		for rows2.Next() {
			var id int
			var name string
			rows2.Scan(&id, &name)
			unitMap[name] = id
		}
		rows2.Close()
	}

	type medData struct {
		Code, Name, Generic, Form, CatName, UnitName string
		BuyPrice, SellPrice                          float64
		Stock, MinStock                              int
	}
	meds := []medData{
		{"OBT-004", "Cetirizine 10mg", "Cetirizine", "Tablet", "Antihistamin", "Tablet", 800, 2000, 150, 20},
		{"OBT-005", "Vitamin C 1000mg", "Ascorbic Acid", "Tablet", "Vitamin", "Tablet", 500, 1500, 300, 50},
		{"OBT-006", "Omeprazole 20mg", "Omeprazole", "Kapsul", "Obat Lambung", "Kapsul", 2000, 5000, 100, 15},
		{"OBT-011", "ORS Sachet", "Oralit", "Lainnya", "Vitamin", "Strip", 500, 1500, 100, 20},
	}

	for _, m := range meds {
		catID := catMap[m.CatName]
		unitID := unitMap[m.UnitName]
		if catID == 0 || unitID == 0 {
			fmt.Printf("  [SKIP] %s: cat=%d unit=%d\n", m.Code, catID, unitID)
			continue
		}
		_, err := db.Exec(`INSERT INTO medicines (medicine_code, name, generic_name, form, category_id, unit_id,
			purchase_price, selling_price, stock, minimum_stock, is_active)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,true) ON CONFLICT (medicine_code) DO NOTHING`,
			m.Code, m.Name, m.Generic, m.Form, catID, unitID, m.BuyPrice, m.SellPrice, m.Stock, m.MinStock)
		if err != nil {
			fmt.Printf("  [SKIP] %s: %v\n", m.Code, err)
		} else {
			fmt.Printf("  [OK] %s\n", m.Code)
		}
	}
}
