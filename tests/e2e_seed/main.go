package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type IDs struct {
	PatientID       int `json:"patient_id"`
	RegistrationID  int `json:"registration_id"`
	MedicalRecordID int `json:"medical_record_id"`
	PrescriptionID  int `json:"prescription_id"`
	InvoiceID       int `json:"invoice_id"`
	DoctorID        int `json:"doctor_id"`
	DiagnosisID     int `json:"diagnosis_id"`
	TreatmentID     int `json:"treatment_id"`
	MedicineID      int `json:"medicine_id"`
	KioskID         int `json:"kiosk_id"`
}

const seedDoctorName = "E2E Dokter Siklus"
const seedDiagnosisName = "E2E Diagnosa Siklus"
const seedTreatmentName = "E2E Tindakan Siklus"
const seedMedicineName = "E2E Obat Paracetamol"

func dbURL() string {
	url := os.Getenv("DATABASE_URL")
	if url != "" {
		return url
	}
	b, _ := os.ReadFile(".env")
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, "DATABASE_URL=") {
			return strings.TrimSpace(strings.TrimPrefix(l, "DATABASE_URL="))
		}
	}
	fmt.Println("DATABASE_URL tidak ditemukan")
	os.Exit(1)
	return ""
}

func getIDByName(db *sql.DB, table, nameCol, name string) (int, error) {
	var id int
	err := db.QueryRow(fmt.Sprintf("SELECT id FROM %s WHERE %s = $1", table, nameCol), name).Scan(&id)
	return id, err
}

func seed(db *sql.DB) {
	ids := IDs{}

	// Dokter (aktif, dengan tarif)
	if id, err := getIDByName(db, "doctors", "full_name", seedDoctorName); err == nil {
		ids.DoctorID = id
	} else {
		_, err := db.Exec(`INSERT INTO doctors (doctor_code, full_name, specialization, consultation_fee, is_active)
			VALUES (($1 || '-' || substr(md5(random()::text), 1, 6)), $2, 'Umum', 100000, TRUE)`, "E2EDK", seedDoctorName)
		if err != nil {
			panic(err)
		}
		if id, err := getIDByName(db, "doctors", "full_name", seedDoctorName); err == nil {
			ids.DoctorID = id
		}
	}

	// Diagnosa
	if id, err := getIDByName(db, "diagnoses", "name", seedDiagnosisName); err == nil {
		ids.DiagnosisID = id
	} else {
		code := fmt.Sprintf("E2E%x", time.Now().UnixNano()%900000+100000)
		if _, err := db.Exec(`INSERT INTO diagnoses (diagnosis_code, name, is_active) VALUES ($1, $2, TRUE)`, code, seedDiagnosisName); err != nil {
			panic(err)
		}
		if id, err := getIDByName(db, "diagnoses", "name", seedDiagnosisName); err == nil {
			ids.DiagnosisID = id
		}
	}

	// Tindakan
	if id, err := getIDByName(db, "treatments", "name", seedTreatmentName); err == nil {
		ids.TreatmentID = id
	} else {
		_, err := db.Exec(`INSERT INTO treatments (treatment_code, name, default_cost, is_active) VALUES ($1, $2, 50000, TRUE)`,
			"E2E-TND", seedTreatmentName)
		if err != nil {
			panic(err)
		}
		if id, err := getIDByName(db, "treatments", "name", seedTreatmentName); err == nil {
			ids.TreatmentID = id
		}
	}

	// Obat (stok cukup untuk ujicoba)
	if id, err := getIDByName(db, "medicines", "name", seedMedicineName); err == nil {
		ids.MedicineID = id
	} else {
		_, err := db.Exec(`INSERT INTO medicines (medicine_code, name, generic_name, form, purchase_price, selling_price, stock, minimum_stock, is_active)
			VALUES ($1, $2, 'Paracetamol 500mg', 'Tablet', 3000, 5000, 100, 10, TRUE)`,
			"E2E-OBT-1", seedMedicineName)
		if err != nil {
			panic(err)
		}
		if id, err := getIDByName(db, "medicines", "name", seedMedicineName); err == nil {
			ids.MedicineID = id
		}
	}

	out, _ := json.Marshal(ids)
	fmt.Println(string(out))
}

func cleanup(db *sql.DB, ids IDs) {
	del := func(query string, args ...interface{}) {
		if _, err := db.Exec(query, args...); err != nil {
			fmt.Println("  !", err)
		}
	}

	// 1. Pembayaran + item + invoice
	sqlInvoiceID := ids.InvoiceID
	del(`DELETE FROM payments WHERE invoice_id = $1`, sqlInvoiceID)
	del(`DELETE FROM invoice_items WHERE invoice_id = $1`, sqlInvoiceID)
	del(`DELETE FROM invoices WHERE id = $1`, sqlInvoiceID)

	// 2. Resep + item obat
	sqlRxID := ids.PrescriptionID
	del(`DELETE FROM prescription_items WHERE prescription_id = $1`, sqlRxID)
	del(`DELETE FROM prescriptions WHERE id = $1`, sqlRxID)

	// 3. Rekam medis + turunannya
	sqlMR := ids.MedicalRecordID
	del(`DELETE FROM medical_record_diagnoses WHERE medical_record_id = $1`, sqlMR)
	del(`DELETE FROM medical_record_treatments WHERE medical_record_id = $1`, sqlMR)
	del(`DELETE FROM pain_assessments WHERE medical_record_id = $1`, sqlMR)
	del(`DELETE FROM medical_records WHERE id = $1`, sqlMR)

	// 4. Antrian (kiosk + pendaftaran) & pendaftaran & pasien
	if ids.KioskID != 0 {
		del(`DELETE FROM queues WHERE id = $1`, ids.KioskID)
	}
	if ids.RegistrationID != 0 {
		del(`DELETE FROM queues WHERE registration_id = $1`, ids.RegistrationID)
	}
	if ids.PatientID != 0 {
		del(`DELETE FROM queues WHERE patient_id = $1`, ids.PatientID)
	}
	del(`DELETE FROM registrations WHERE id = $1`, ids.RegistrationID)
	if ids.PatientID != 0 {
		del(`DELETE FROM registrations WHERE patient_id = $1`, ids.PatientID)
	}
	del(`DELETE FROM patients WHERE id = $1`, ids.PatientID)

	// 5. Master seed
	del(`DELETE FROM medicine_stock_transactions WHERE medicine_id = $1`, ids.MedicineID)
	del(`DELETE FROM medicines WHERE id = $1`, ids.MedicineID)
	del(`DELETE FROM treatments WHERE id = $1`, ids.TreatmentID)
	del(`DELETE FROM diagnoses WHERE id = $1`, ids.DiagnosisID)
	del(`DELETE FROM doctors WHERE id = $1`, ids.DoctorID)

	fmt.Println("cleanup-selesai")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: e2e_seed seed | e2e_seed cleanup <ids.json> | e2e_seed peek")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", dbURL())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	switch os.Args[1] {
	case "seed":
		seed(db)
	case "cleanup":
		if len(os.Args) < 3 {
			fmt.Println("usage: e2e_seed cleanup <ids.json>")
			os.Exit(1)
		}
		raw, err := os.ReadFile(os.Args[2])
		if err != nil {
			panic(err)
		}
		var ids IDs
		if err := json.Unmarshal(raw, &ids); err != nil {
			panic(err)
		}
		cleanup(db, ids)
	case "peek":
		rows, err := db.Query(`SELECT q.id, q.queue_number, COALESCE(p.full_name, 'NULL'), q.status,
			COALESCE(q.registration_id, 0) FROM queues q
			LEFT JOIN patients p ON p.id = q.patient_id
			WHERE q.queue_date = CURRENT_DATE ORDER BY q.id`)
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		for rows.Next() {
			var id int
			var num, name, status string
			var reg int
			if err := rows.Scan(&id, &num, &name, &status, &reg); err != nil {
				panic(err)
			}
			fmt.Printf("q=%d num=%s patient=%q status=%s reg=%d\n", id, num, name, status, reg)
		}
	default:
		fmt.Println("perintah tidak dikenal:", os.Args[1])
		os.Exit(1)
	}
}
