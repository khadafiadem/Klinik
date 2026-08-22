package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"klinik-app/internal/auth"
	"klinik-app/internal/config"
	"klinik-app/internal/database"
	"klinik-app/internal/logger"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "up":
		runUp()
	case "status":
		runStatus()
	case "seed":
		runSeed()
	default:
		fmt.Printf("Perintah tidak dikenal: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Sistem Manajemen Klinik - Database Migration")
	fmt.Println()
	fmt.Println("Penggunaan:")
	fmt.Println("  go run ./cmd/migrate up       Menjalankan semua migration")
	fmt.Println("  go run ./cmd/migrate status   Melihat status migration")
	fmt.Println("  go run ./cmd/migrate seed     Membuat admin user default")
	fmt.Println()
	fmt.Println("Pastikan file .env sudah dikonfigurasi dengan benar.")
}

func runUp() {
	logger.Init("info")
	logger.Info.Println("Menjalankan database migration...")

	cfg, err := config.Load()
	if err != nil {
		logger.Error.Printf("Gagal memuat konfigurasi: %v", err)
		os.Exit(1)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error.Printf("Gagal terhubung ke database: %v", err)
		os.Exit(1)
	}
	defer database.Close(db)

	migrator := database.NewMigrator(db, "migrations")
	if err := migrator.Up(); err != nil {
		logger.Error.Printf("Migration gagal: %v", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Migration selesai!")
}

func runStatus() {
	logger.Init("info")

	cfg, err := config.Load()
	if err != nil {
		logger.Error.Printf("Gagal memuat konfigurasi: %v", err)
		os.Exit(1)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error.Printf("Gagal terhubung ke database: %v", err)
		os.Exit(1)
	}
	defer database.Close(db)

	migrator := database.NewMigrator(db, "migrations")
	statuses, err := migrator.Status()
	if err != nil {
		logger.Error.Printf("Gagal mendapatkan status: %v", err)
		os.Exit(1)
	}

	fmt.Println("Status Database Migration")
	fmt.Println("=========================")

	if len(statuses) == 0 {
		fmt.Println("Tidak ada file migration yang ditemukan.")
		return
	}

	for _, s := range statuses {
		status := "BELUM DIJALANKAN"
		if s.Applied {
			status = "SUDAH DIJALANKAN"
		}
		fmt.Printf("[%s] %s - %s\n", status, s.Version, s.Filename)
	}

	appliedCount := 0
	for _, s := range statuses {
		if s.Applied {
			appliedCount++
		}
	}
	fmt.Printf("\nTotal: %d migration, %d sudah dijalankan, %d belum dijalankan\n",
		len(statuses), appliedCount, len(statuses)-appliedCount)
}

func runSeed() {
	logger.Init("info")
	logger.Info.Println("Membuat admin user default...")

	cfg, err := config.Load()
	if err != nil {
		logger.Error.Printf("Gagal memuat konfigurasi: %v", err)
		os.Exit(1)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error.Printf("Gagal terhubung ke database: %v", err)
		os.Exit(1)
	}
	defer database.Close(db)

	adminUser := "admin"
	adminEmail := "admin@klinik.com"
	adminPassword := "admin123"
	adminFullName := "Administrator"

	exists, err := checkUserExists(db, adminUser)
	if err != nil {
		logger.Error.Printf("Gagal mengecek user: %v", err)
		os.Exit(1)
	}

	if exists {
		fmt.Println("Admin user sudah ada, melewati pembuatan.")
	} else {
		passwordHash, err := auth.HashPassword(adminPassword)
		if err != nil {
			logger.Error.Printf("Gagal hash password: %v", err)
			os.Exit(1)
		}

		var userID int
		query := `INSERT INTO users (username, email, password_hash, full_name) 
			VALUES ($1, $2, $3, $4) RETURNING id`
		err = db.QueryRow(query, adminUser, adminEmail, passwordHash, adminFullName).Scan(&userID)
		if err != nil {
			logger.Error.Printf("Gagal membuat admin user: %v", err)
			os.Exit(1)
		}

		roleQuery := `INSERT INTO user_roles (user_id, role_id) 
			SELECT $1, id FROM roles WHERE name = 'ADMIN' ON CONFLICT DO NOTHING`
		_, err = db.Exec(roleQuery, userID)
		if err != nil {
			logger.Error.Printf("Gagal assign role ADMIN: %v", err)
			os.Exit(1)
		}

		fmt.Println()
		fmt.Println("Admin user berhasil dibuat!")
		fmt.Printf("  Username: %s\n", adminUser)
		fmt.Printf("  Email:    %s\n", adminEmail)
		fmt.Printf("  Password: %s\n", adminPassword)
		fmt.Println()
		fmt.Println("PERINGATAN: Segera setelah login, ubah password default!")
	}

	seedDummyData(db)
}

func seedDummyData(db *sql.DB) {
	fmt.Println()
	fmt.Println("Memasukkan data dummy...")

	doctors := []struct {
		Code, Name, Spec, Phone, Email string
		Fee                            float64
	}{
		{"D001", "dr. Ahmad Suharto, Sp.PD", "Penyakit Dalam", "081234567890", "ahmad@klinik.com", 250000},
		{"D002", "dr. Siti Nurhaliza, Sp.A", "Anak", "081234567891", "siti@klinik.com", 300000},
		{"D003", "dr. Budi Santoso, Sp.BED", "Bedah Umum", "081234567892", "budi@klinik.com", 350000},
		{"D004", "dr. Dewi Lestari, Sp.KK", "Kulit & Kelamin", "081234567893", "dewi@klinik.com", 275000},
		{"D005", "dr. Eko Prasetyo, Sp.THT", "THT-KL", "081234567894", "eko@klinik.com", 250000},
	}

	for _, d := range doctors {
		_, err := db.Exec(`INSERT INTO doctors (doctor_code, full_name, specialization, phone, email, consultation_fee) 
			VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT (doctor_code) DO NOTHING`,
			d.Code, d.Name, d.Spec, d.Phone, d.Email, d.Fee)
		if err != nil {
			fmt.Printf("  [SKIP] Dokter %s: %v\n", d.Name, err)
		}
	}
	fmt.Printf("  %d dokter ditambahkan\n", len(doctors))

	staffs := []struct {
		Code, Name, Position, Phone, Email string
	}{
		{"S001", "Rina Wulandari", "Resepsionis", "082111111111", "rina@klinik.com"},
		{"S002", "Maya Putri", "Perawat", "082111111112", "maya@klinik.com"},
		{"S003", "Andi Kurniawan", "Admin", "082111111113", "andi@klinik.com"},
		{"S004", "Lestari Budiman", "Perawat", "082111111114", "lestari@klinik.com"},
		{"S005", "Farhan Maulana", "Apoteker", "082111111115", "farhan@klinik.com"},
	}

	for _, s := range staffs {
		_, err := db.Exec(`INSERT INTO staff (staff_code, full_name, position, phone, email) 
			VALUES ($1,$2,$3,$4,$5) ON CONFLICT (staff_code) DO NOTHING`,
			s.Code, s.Name, s.Position, s.Phone, s.Email)
		if err != nil {
			fmt.Printf("  [SKIP] Staff %s: %v\n", s.Name, err)
		}
	}
	fmt.Printf("  %d staff ditambahkan\n", len(staffs))

	patients := []struct {
		MRN, Name, NIK, Gender, DOB, Blood, Phone, Email, Address, City, Allergies string
	}{
		{"MR-000001", "Andi Saputra", "3201234567890001", "LAKI_LAKI", "1990-05-15", "O", "083111111111", "andi.s@email.com", "Jl. Merdeka No. 10, Kel. Sukamaju", "Jakarta Selatan", "Tidak ada"},
		{"MR-000002", "Sari Dewi", "3201234567890002", "PEREMPUAN", "1985-08-20", "A", "083111111112", "sari.d@email.com", "Jl. Gatot Subroto No. 25", "Bandung", "Penisilin"},
		{"MR-000003", "Rizki Pratama", "3201234567890003", "LAKI_LAKI", "2000-01-10", "B", "083111111113", "rizki.p@email.com", "Jl. Asia Afrika No. 5", "Jakarta Pusat", "Debu, Udara dingin"},
		{"MR-000004", "Putri Ayu", "3201234567890004", "PEREMPUAN", "1995-12-25", "AB", "083111111114", "", "Jl. Pahlawan No. 8, RT 02/05", "Surabaya", "Makanan laut"},
		{"MR-000005", "Hendra Wijaya", "3201234567890005", "LAKI_LAKI", "1978-03-08", "O", "083111111115", "hendra.w@email.com", "Jl. Sudirman No. 100, Lantai 3", "Yogyakarta", "Tidak ada"},
	}

	for _, p := range patients {
		_, err := db.Exec(`INSERT INTO patients (medical_record_number, full_name, nik, gender, date_of_birth, blood_type, phone, email, address, city, allergies) 
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT (medical_record_number) DO NOTHING`,
			p.MRN, p.Name, p.NIK, p.Gender, p.DOB, p.Blood, p.Phone, p.Email, p.Address, p.City, p.Allergies)
		if err != nil {
			fmt.Printf("  [SKIP] Pasien %s: %v\n", p.Name, err)
		}
	}
	fmt.Printf("  %d pasien ditambahkan\n", len(patients))

	seedRegistrations(db)
	seedQueues(db)
	seedMedicalRecords(db)
	seedDiagnoses(db)
	seedTreatments(db)
	seedPrescriptions(db)
	seedMedicines(db)
	seedInvoices(db)
	seedTodayTransactions(db)
	seedUsers(db)

	fmt.Println()
	fmt.Println("Data dummy berhasil dimasukkan!")
}

func seedRegistrations(db *sql.DB) {
	type reg struct {
		RegNum, PatientMRN, DoctorCode, RegDate, RegType, Complaint, Status string
	}
	regs := []reg{
		{"REG-20260821-001", "MR-000001", "D001", "2026-08-21", "UMUM", "Demam 3 hari, batuk pilek", "TERDAFTAR"},
		{"REG-20260821-002", "MR-000002", "D002", "2026-08-21", "BPJS", "Anak demam tinggi 39C", "SEDANG_DIPERIKSA"},
		{"REG-20260821-003", "MR-000003", "D005", "2026-08-21", "UMUM", "Sakit tenggorokan, pilek", "TERDAFTAR"},
		{"REG-20260821-004", "MR-000004", "D004", "2026-08-21", "ASURANSI", "Ruam kulit pada tangan", "TERDAFTAR"},
		{"REG-20260821-005", "MR-000005", "D003", "2026-08-21", "UMUM", "Luka sobek pada kaki", "SELESAI"},
	}
	for i := range regs {
		var pid, did int
		_ = db.QueryRow("SELECT id FROM patients WHERE medical_record_number=$1", regs[i].PatientMRN).Scan(&pid)
		_ = db.QueryRow("SELECT id FROM doctors WHERE doctor_code=$1", regs[i].DoctorCode).Scan(&did)
		if pid == 0 || did == 0 {
			fmt.Printf("  [SKIP] Registrasi %s: data tidak ditemukan\n", regs[i].RegNum)
			continue
		}
		_, err := db.Exec(
			`INSERT INTO registrations (registration_number, patient_id, doctor_id, registration_date, registration_type, complaint, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (registration_number) DO NOTHING`,
			regs[i].RegNum, pid, did, regs[i].RegDate, regs[i].RegType, regs[i].Complaint, regs[i].Status)
		if err != nil {
			fmt.Printf("  [SKIP] Registrasi %s: %v\n", regs[i].RegNum, err)
		}
	}
	fmt.Printf("  %d registrasi ditambahkan\n", len(regs))
}

func seedQueues(db *sql.DB) {
	type q struct {
		QueueNum, RegNum, PatientMRN, DoctorCode, Date, Status string
	}
	queues := []q{
		{"A-001", "REG-20260821-001", "MR-000001", "D001", "2026-08-21", "MENUNGGU"},
		{"A-002", "REG-20260821-002", "MR-000002", "D002", "2026-08-21", "SEDANG_DIPERIKSA"},
		{"A-003", "REG-20260821-003", "MR-000003", "D005", "2026-08-21", "MENUNGGU"},
		{"A-004", "REG-20260821-004", "MR-000004", "D004", "2026-08-21", "MENUNGGU"},
		{"A-005", "REG-20260821-005", "MR-000005", "D003", "2026-08-21", "SELESAI"},
	}
	for i := range queues {
		var regID, pid, did int
		_ = db.QueryRow("SELECT id FROM registrations WHERE registration_number=$1", queues[i].RegNum).Scan(&regID)
		_ = db.QueryRow("SELECT id FROM patients WHERE medical_record_number=$1", queues[i].PatientMRN).Scan(&pid)
		_ = db.QueryRow("SELECT id FROM doctors WHERE doctor_code=$1", queues[i].DoctorCode).Scan(&did)
		if regID == 0 || pid == 0 || did == 0 {
			fmt.Printf("  [SKIP] Antrian %s: data tidak ditemukan\n", queues[i].QueueNum)
			continue
		}
		_, err := db.Exec(
			`INSERT INTO queues (queue_number, registration_id, patient_id, doctor_id, queue_date, status)
			VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT DO NOTHING`,
			queues[i].QueueNum, regID, pid, did, queues[i].Date, queues[i].Status)
		if err != nil {
			fmt.Printf("  [SKIP] Antrian %s: %v\n", queues[i].QueueNum, err)
		}
	}
	fmt.Printf("  %d antrian ditambahkan\n", len(queues))
}

func seedMedicalRecords(db *sql.DB) {
	type mr struct {
		MRNum, PatientMRN, DoctorCode, ExamDate, Complaint, Anamnesis, PhysExam, Status string
	}
	records := []mr{
		{"RM-20260821-001", "MR-000001", "D001", "2026-08-21", "Demam 3 hari, batuk pilek", "Pasien mengeluh demam sejak 3 hari lalu, disertai batuk dan pilek", "Suhu 38.5C, TD 120/80, RR 20", "DRAFT"},
		{"RM-20260821-002", "MR-000002", "D002", "2026-08-21", "Anak demam tinggi", "Ibu membawa anak usia 5 tahun dengan demam tinggi 39C sejak semalam", "Suhu 39C, TD 90/60, RR 24", "FINAL"},
		{"RM-20260821-003", "MR-000005", "D003", "2026-08-21", "Luka sobek kaki kanan", "Terjatuh dari sepeda, luka sobek pada kaki kanan sepanjang 5cm", "Luka sobek 5cm kaki kanan, tidak ada patah tulang", "FINAL"},
	}
	for i := range records {
		var pid, did int
		_ = db.QueryRow("SELECT id FROM patients WHERE medical_record_number=$1", records[i].PatientMRN).Scan(&pid)
		_ = db.QueryRow("SELECT id FROM doctors WHERE doctor_code=$1", records[i].DoctorCode).Scan(&did)
		if pid == 0 || did == 0 {
			continue
		}
		vs := `{"temperature":"38.5","heart_rate":"80","blood_pressure":"120/80","respiratory_rate":"20","weight":"70","height":"170"}`
		num := fmt.Sprintf("RM-20260821-%03d", i+1)
		_, err := db.Exec(`INSERT INTO medical_records
			(medical_record_number, patient_id, doctor_id, examination_date, chief_complaint, anamnesis, physical_examination, vital_signs, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9) ON CONFLICT (medical_record_number) DO NOTHING`,
			num, pid, did, records[i].ExamDate, records[i].Complaint, records[i].Anamnesis, records[i].PhysExam, vs, records[i].Status)
		if err != nil {
			fmt.Printf("  [SKIP] RM %s: %v\n", num, err)
		}
	}
	fmt.Printf("  %d rekam medis ditambahkan\n", len(records))
}

func seedDiagnoses(db *sql.DB) {
	diagnoses := []struct{ Code, Name string }{
		{"A09", "Infeksi saluran pencernaan"},
		{"J06.9", "Infeksi saluran pernapasan atas"},
		{"J18.9", "Pneumonia"},
		{"L30.9", "Dermatitis"},
		{"S81.0", "Luka terbuka kaki"},
		{"E11.9", "Diabetes Mellitus tipe 2"},
		{"I10", "Hipertensi esensial"},
		{"M54.5", "Nyeri punggung bawah"},
		{"N39.0", "Infeksi saluran kemih"},
		{"K29.7", "Gastritis"},
	}
	for _, d := range diagnoses {
		_, err := db.Exec(`INSERT INTO diagnoses (diagnosis_code, name) VALUES ($1,$2) ON CONFLICT (diagnosis_code) DO NOTHING`,
			d.Code, d.Name)
		if err != nil {
			fmt.Printf("  [SKIP] Diagnosa %s: %v\n", d.Code, err)
		}
	}
	fmt.Printf("  %d diagnosa ditambahkan\n", len(diagnoses))
}

func seedTreatments(db *sql.DB) {
	treatments := []struct{ Code, Name string; Cost float64 }{
		{"T001", "Pemeriksaan Umum", 250000},
		{"T002", "Pemeriksaan Anak", 300000},
		{"T003", "Pembersihan Luka", 150000},
		{"T004", "Jahit Luka", 500000},
		{"T005", "Suntik antibiotik", 100000},
		{"T006", "Nebulisasi", 200000},
		{"T007", "EKG", 350000},
		{"T008", "Tindakan bedah minor", 1000000},
	}
	for _, t := range treatments {
		_, err := db.Exec(`INSERT INTO treatments (treatment_code, name, default_cost) VALUES ($1,$2,$3) ON CONFLICT (treatment_code) DO NOTHING`,
			t.Code, t.Name, t.Cost)
		if err != nil {
			fmt.Printf("  [SKIP] Tindakan %s: %v\n", t.Code, err)
		}
	}
	fmt.Printf("  %d tindakan ditambahkan\n", len(treatments))
}

func seedPrescriptions(db *sql.DB) {
	// Get medical record ID for patient MR-000002
	var mrID, pid, did int
	_ = db.QueryRow(`SELECT mr.id, mr.patient_id, mr.doctor_id FROM medical_records mr
		JOIN patients p ON mr.patient_id = p.id WHERE p.medical_record_number='MR-000002'`).Scan(&mrID, &pid, &did)
	if mrID == 0 || pid == 0 || did == 0 {
		fmt.Println("  [SKIP] Resep: data tidak ditemukan")
		return
	}
	num := "RX-20260821-001"
	_, err := db.Exec(`INSERT INTO prescriptions (prescription_number, medical_record_id, patient_id, doctor_id, prescription_date, status)
		VALUES ($1,$2,$3,$4,'2026-08-21','COMPLETED') ON CONFLICT (prescription_number) DO NOTHING`,
		num, mrID, pid, did)
	if err != nil {
		fmt.Printf("  [SKIP] Resep: %v\n", err)
		return
	}

	var rxID int
	_ = db.QueryRow("SELECT id FROM prescriptions WHERE prescription_number=$1", num).Scan(&rxID)
	items := []struct {
		Name, Code, Dosage, Freq, Dur, Instr string
		Qty                                   int
	}{
		{"Parasetamol 500mg", "PAR-001", "1 tablet", "3x sehari", "5 hari", "Diminum setelah makan", 15},
		{"Amoxicillin 500mg", "AMX-001", "1 kapsul", "3x sehari", "7 hari", "Diminum sebelum makan", 21},
		{"OBH Combi", "OBH-001", "15ml", "3x sehari", "5 hari", "Diminum setelah makan", 1},
	}
	for _, item := range items {
		_, _ = db.Exec(`INSERT INTO prescription_items (prescription_id, medicine_name, medicine_code, quantity, dosage, frequency, duration, instructions)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			rxID, item.Name, item.Code, item.Qty, item.Dosage, item.Freq, item.Dur, item.Instr)
	}
	fmt.Printf("  1 resep dengan %d item obat ditambahkan\n", len(items))
}

func seedMedicines(db *sql.DB) {
	categories := []string{"Analgesik", "Antibiotik", "Antipiretik", "Antihistamin", "Vitamin", "Obat Lambung", "Obat Batuk", "Salep"}
	for _, c := range categories {
		_, _ = db.Exec("INSERT INTO medicine_categories (name) VALUES ($1) ON CONFLICT DO NOTHING", c)
	}
	units := []string{"Tablet", "Kapsul", "Botol", "Tube", "Ampul", "Strip"}
	for _, u := range units {
		_, _ = db.Exec("INSERT INTO medicine_units (name) VALUES ($1) ON CONFLICT DO NOTHING", u)
	}

	type medData struct {
		Code, Name, Generic, Form, CatName, UnitName string
		BuyPrice, SellPrice                          float64
		Stock, MinStock                              int
	}
	meds := []medData{
		{"OBT-001", "Parasetamol 500mg", "Paracetamol", "Tablet", "Antipiretik", "Tablet", 200, 500, 500, 50},
		{"OBT-002", "Amoxicillin 500mg", "Amoxicillin", "Kapsul", "Antibiotik", "Kapsul", 1500, 3000, 200, 30},
		{"OBT-003", "OBH Combi Plus", "Dextromethorphan", "Sirup", "Obat Batuk", "Botol", 12000, 22000, 50, 10},
		{"OBT-004", "Cetirizine 10mg", "Cetirizine", "Tablet", "Antihistamin", "Tablet", 800, 2000, 150, 20},
		{"OBT-005", "Vitamin C 1000mg", "Ascorbic Acid", "Tablet", "Vitamin", "Tablet", 500, 1500, 300, 50},
		{"OBT-006", "Omeprazole 20mg", "Omeprazole", "Kapsul", "Obat Lambung", "Kapsul", 2000, 5000, 100, 15},
		{"OBT-007", "Salbutamol Nebul", "Salbutamol", "Injeksi", "Analgesik", "Ampul", 8000, 15000, 20, 5},
		{"OBT-008", "Miconazole Cream", "Miconazole", "Salep", "Salep", "Tube", 5000, 12000, 30, 10},
		{"OBT-009", "Ibuprofen 400mg", "Ibuprofen", "Tablet", "Analgesik", "Tablet", 300, 800, 250, 30},
		{"OBT-010", "Parasetamol Drop", "Paracetamol", "Tetes", "Antipiretik", "Botol", 8000, 15000, 25, 5},
		{"OBT-011", "ORS Sachet", "Oralit", "Lainnya", "Vitamin", "Strip", 500, 1500, 100, 20},
		{"OBT-012", "Methergin", "Methylergometrine", "Tablet", "Analgesik", "Tablet", 3000, 8000, 10, 5},
	}
	for _, m := range meds {
		_, err := db.Exec(`INSERT INTO medicines (medicine_code, name, generic_name, form, category_id, unit_id,
			purchase_price, selling_price, stock, minimum_stock, is_active)
			SELECT $1,$2,$3,$4,mc.id,mu.id,$5,$6,$7,$8,true
			FROM medicine_categories mc, medicine_units mu
			WHERE mc.name=$9 AND mu.name=$10
			ON CONFLICT (medicine_code) DO NOTHING`,
			m.Code, m.Name, m.Generic, m.Form, m.BuyPrice, m.SellPrice, m.Stock, m.MinStock, m.CatName, m.UnitName)
		if err != nil {
			fmt.Printf("  [SKIP] Obat %s: %v\n", m.Code, err)
		}
	}
	fmt.Printf("  %d obat ditambahkan\n", len(meds))
}

func seedInvoices(db *sql.DB) {
	type invData struct {
		InvNum, PatientMRN, Date, Status string
		Discount                         float64
		items                            []struct {
			Desc string
			Qty  int
			Price float64
			Type  string
		}
	}
	invoices := []invData{
		{"INV-20260821-001", "MR-000001", "2026-08-21", "SUDAH_BAYAR", 0, []struct {
			Desc string; Qty int; Price float64; Type string
		}{
			{"Biaya Konsultasi Umum", 1, 250000, "KONSULTASI"},
			{"Parasetamol 500mg (15 tablet)", 1, 7500, "OBAT"},
		}},
		{"INV-20260821-002", "MR-000002", "2026-08-21", "SEBAGIAN", 0, []struct {
			Desc string; Qty int; Price float64; Type string
		}{
			{"Biaya Konsultasi Anak", 1, 300000, "KONSULTASI"},
			{"Amoxicillin 500mg (21 kapsul)", 1, 63000, "OBAT"},
			{"Parasetamol Drop", 1, 15000, "OBAT"},
		}},
		{"INV-20260821-003", "MR-000005", "2026-08-21", "BELUM_BAYAR", 0, []struct {
			Desc string; Qty int; Price float64; Type string
		}{
			{"Biaya Konsultasi Bedah", 1, 350000, "KONSULTASI"},
			{"Pembersihan Luka", 1, 150000, "TINDAKAN"},
			{"Jahit Luka", 1, 500000, "TINDAKAN"},
		}},
	}
	for _, inv := range invoices {
		var pid int
		_ = db.QueryRow("SELECT id FROM patients WHERE medical_record_number=$1", inv.PatientMRN).Scan(&pid)
		if pid == 0 {
			continue
		}

		var subtotal float64
		for _, item := range inv.items {
			subtotal += float64(item.Qty) * item.Price
		}
		total := subtotal - inv.Discount

		var invID int
		err := db.QueryRow(`INSERT INTO invoices (invoice_number, patient_id, invoice_date, subtotal, discount, total, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
			inv.InvNum, pid, inv.Date, subtotal, inv.Discount, total, inv.Status).Scan(&invID)
		if err != nil {
			fmt.Printf("  [SKIP] Invoice %s: %v\n", inv.InvNum, err)
			continue
		}

		for _, item := range inv.items {
			itemTotal := float64(item.Qty) * item.Price
			_, _ = db.Exec(`INSERT INTO invoice_items (invoice_id, description, quantity, unit_price, total_price, item_type)
				VALUES ($1,$2,$3,$4,$5,$6)`,
				invID, item.Desc, item.Qty, item.Price, itemTotal, item.Type)
		}

		if inv.Status == "SUDAH_BAYAR" {
			_, _ = db.Exec(`INSERT INTO payments (payment_number, invoice_id, patient_id, payment_date, amount, payment_method, status)
				VALUES ($1,$2,$3,$4,$5,'CASH','COMPLETED')`,
				fmt.Sprintf("PAY-20260821-%03d", invID), invID, pid, inv.Date, total)
		} else if inv.Status == "SEBAGIAN" {
			_, _ = db.Exec(`INSERT INTO payments (payment_number, invoice_id, patient_id, payment_date, amount, payment_method, status)
				VALUES ($1,$2,$3,$4,$5,'CASH','COMPLETED')`,
				fmt.Sprintf("PAY-20260821-%03d", invID), invID, pid, inv.Date, 300000.0)
		}
	}
	fmt.Printf("  %d tagihan ditambahkan\n", len(invoices))
}

func seedTodayTransactions(db *sql.DB) {
	today := time.Now().Format("2006-01-02")

	// Registrasi hari ini
	type reg struct {
		RegNum, PatientMRN, DoctorCode, RegType, Complaint, Status string
	}
	regs := []reg{
		{"REG-" + today + "-101", "MR-000001", "D001", "UMUM", "Kontrol demam", "TERDAFTAR"},
		{"REG-" + today + "-102", "MR-000002", "D002", "BPJS", "Vaksinasi anak", "SELESAI"},
		{"REG-" + today + "-103", "MR-000003", "D005", "UMUM", "Telinga berdenging", "SEDANG_DIPERIKSA"},
		{"REG-" + today + "-104", "MR-000004", "D004", "ASURANSI", "Gatal pada kulit", "TERDAFTAR"},
		{"REG-" + today + "-105", "MR-000005", "D003", "KONTROL", "Kontrol luka pasca jahit", "TERDAFTAR"},
	}
	for _, r := range regs {
		var pid, did int
		_ = db.QueryRow("SELECT id FROM patients WHERE medical_record_number=$1", r.PatientMRN).Scan(&pid)
		_ = db.QueryRow("SELECT id FROM doctors WHERE doctor_code=$1", r.DoctorCode).Scan(&did)
		if pid == 0 || did == 0 {
			continue
		}
		var exists int
		_ = db.QueryRow("SELECT COUNT(*) FROM registrations WHERE registration_number=$1", r.RegNum).Scan(&exists)
		if exists > 0 {
			continue
		}
		_, _ = db.Exec(`INSERT INTO registrations (registration_number, patient_id, doctor_id, registration_date, registration_type, complaint, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`, r.RegNum, pid, did, today, r.RegType, r.Complaint, r.Status)
	}

	// Antrian hari ini
	type q struct {
		QueueNum, RegNum, PatientMRN, DoctorCode, Status string
	}
	queues := []q{
		{"B-001", "REG-" + today + "-101", "MR-000001", "D001", "MENUNGGU"},
		{"B-002", "REG-" + today + "-102", "MR-000002", "D002", "SELESAI"},
		{"B-003", "REG-" + today + "-103", "MR-000003", "D005", "SEDANG_DIPERIKSA"},
		{"B-004", "REG-" + today + "-104", "MR-000004", "D004", "MENUNGGU"},
		{"B-005", "REG-" + today + "-105", "MR-000005", "D003", "MENUNGGU"},
	}
	for _, qq := range queues {
		var regID, pid, did int
		_ = db.QueryRow("SELECT id FROM registrations WHERE registration_number=$1", qq.RegNum).Scan(&regID)
		_ = db.QueryRow("SELECT id FROM patients WHERE medical_record_number=$1", qq.PatientMRN).Scan(&pid)
		_ = db.QueryRow("SELECT id FROM doctors WHERE doctor_code=$1", qq.DoctorCode).Scan(&did)
		if regID == 0 || pid == 0 || did == 0 {
			continue
		}
		var exists int
		_ = db.QueryRow("SELECT COUNT(*) FROM queues WHERE queue_number=$1 AND queue_date=$2", qq.QueueNum, today).Scan(&exists)
		if exists > 0 {
			continue
		}
		_, _ = db.Exec(`INSERT INTO queues (queue_number, registration_id, patient_id, doctor_id, queue_date, status)
			VALUES ($1,$2,$3,$4,$5,$6)`, qq.QueueNum, regID, pid, did, today, qq.Status)
	}

	// Rekam medis untuk pasien MR-000003 dan MR-000004
	mrs := []struct {
		MRNum, PatientMRN, DoctorCode, Complaint, Anamnesis, PhysExam, Status string
	}{
		{"RM-" + today + "-101", "MR-000003", "D005", "Telinga berdenging sejak 2 hari", "Pasien mengeluh telinga kanan berdenging setelah berenang", "Tympani sedang, kanal telinga bersih", "FINAL"},
		{"RM-" + today + "-102", "MR-000004", "D004", "Gatal dan ruam pada lengan", "Ruam timbul setelah makan seafood, gatal menjalar", "Eritema papular di forearm bilateral", "DRAFT"},
	}
	for _, mr := range mrs {
		var pid, did int
		_ = db.QueryRow("SELECT id FROM patients WHERE medical_record_number=$1", mr.PatientMRN).Scan(&pid)
		_ = db.QueryRow("SELECT id FROM doctors WHERE doctor_code=$1", mr.DoctorCode).Scan(&did)
		if pid == 0 || did == 0 {
			continue
		}
		var exists int
		_ = db.QueryRow("SELECT COUNT(*) FROM medical_records WHERE medical_record_number=$1", mr.MRNum).Scan(&exists)
		if exists > 0 {
			continue
		}
		_, _ = db.Exec(`INSERT INTO medical_records (medical_record_number, patient_id, doctor_id, examination_date, chief_complaint, anamnesis, physical_examination, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, mr.MRNum, pid, did, today, mr.Complaint, mr.Anamnesis, mr.PhysExam, mr.Status)
	}

	// Resep hari ini
	rxs := []struct {
		RxNum, PatientMRN, DoctorCode, Status string
		Items                                 []struct {
			Name, Code, Dosage, Freq, Dur, Instr string
			Qty                                  int
		}
	}{
		{"RX-" + today + "-101", "MR-000003", "D005", "PENDING", []struct {
			Name, Code, Dosage, Freq, Dur, Instr string
			Qty                                  int
		}{{
			"Cetirizine 10mg", "OBT-004", "1 tablet", "1x malam", "7 hari", "Diminum sebelum tidur", 7,
		}},
		},
		{"RX-" + today + "-102", "MR-000004", "D004", "PROCESSING", []struct {
			Name, Code, Dosage, Freq, Dur, Instr string
			Qty                                  int
		}{{
			"Miconazole Cream", "OBT-008", "Apply tipis", "2x sehari", "14 hari", "Oleskan pada area ruam", 1,
		}, {
			"Vitamin C 1000mg", "OBT-005", "1 tablet", "1x sehari", "7 hari", "Diminum setelah makan", 7,
		}},
		},
		{"RX-" + today + "-103", "MR-000001", "D001", "COMPLETED", []struct {
			Name, Code, Dosage, Freq, Dur, Instr string
			Qty                                  int
		}{{
			"Parasetamol 500mg", "OBT-001", "1 tablet", "3x sehari", "3 hari", "Diminum bila demam", 9,
		}},
		},
		{"RX-" + today + "-104", "MR-000002", "D002", "PENDING", []struct {
			Name, Code, Dosage, Freq, Dur, Instr string
			Qty                                  int
		}{{
			"Parasetamol Drop", "OBT-010", "0.6ml", "4x sehari", "3 hari", "Diberikan sesuai takaran", 1,
		}, {
			"ORS Sachet", "OBT-011", "1 sachet", "2x sehari", "2 hari", "Larutkan dalam 200ml air", 4,
		}},
		},
	}
	for _, rx := range rxs {
		var pid, did int
		_ = db.QueryRow("SELECT id FROM patients WHERE medical_record_number=$1", rx.PatientMRN).Scan(&pid)
		_ = db.QueryRow("SELECT id FROM doctors WHERE doctor_code=$1", rx.DoctorCode).Scan(&did)
		if pid == 0 || did == 0 {
			continue
		}
		var mrID int
		_ = db.QueryRow(`SELECT mr.id FROM medical_records mr WHERE mr.patient_id=$1 ORDER BY mr.examination_date DESC LIMIT 1`, pid).Scan(&mrID)
		if mrID == 0 {
			continue
		}
		var exists int
		_ = db.QueryRow("SELECT COUNT(*) FROM prescriptions WHERE prescription_number=$1", rx.RxNum).Scan(&exists)
		if exists > 0 {
			continue
		}
		var rxID int
		err := db.QueryRow(`INSERT INTO prescriptions (prescription_number, medical_record_id, patient_id, doctor_id, prescription_date, status)
			VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`, rx.RxNum, mrID, pid, did, today, rx.Status).Scan(&rxID)
		if err != nil {
			fmt.Printf("  [SKIP] Resep %s: %v\n", rx.RxNum, err)
			continue
		}
		for _, item := range rx.Items {
			_, _ = db.Exec(`INSERT INTO prescription_items (prescription_id, medicine_name, medicine_code, quantity, dosage, frequency, duration, instructions)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, rxID, item.Name, item.Code, item.Qty, item.Dosage, item.Freq, item.Dur, item.Instr)
		}
	}

	// Invoice & pembayaran hari ini
	type invItem struct {
		Desc  string
		Qty   int
		Price float64
		Type  string
	}
	invs := []struct {
		InvNum, PatientMRN, Status, PayMethod string
		PayAmount                             float64
		items                                 []invItem
	}{
		{"INV-" + today + "-101", "MR-000001", "SUDAH_BAYAR", "CASH", -1, []invItem{
			{"Biaya Konsultasi Umum", 1, 250000, "KONSULTASI"},
			{"Parasetamol 500mg (9 tablet)", 1, 4500, "OBAT"},
		}},
		{"INV-" + today + "-102", "MR-000002", "SUDAH_BAYAR", "QRIS", -1, []invItem{
			{"Biaya Konsultasi Anak", 1, 300000, "KONSULTASI"},
			{"Parasetamol Drop", 1, 15000, "OBAT"},
			{"ORS Sachet", 4, 1500, "OBAT"},
		}},
		{"INV-" + today + "-103", "MR-000003", "BELUM_BAYAR", "", 0, []invItem{
			{"Biaya Konsultasi THT", 1, 250000, "KONSULTASI"},
			{"Cetirizine 10mg (7 tablet)", 1, 2000, "OBAT"},
		}},
		{"INV-" + today + "-104", "MR-000004", "SEBAGIAN", "BANK_TRANSFER", 100000, []invItem{
			{"Biaya Konsultasi Kulit", 1, 275000, "KONSULTASI"},
			{"Miconazole Cream", 1, 12000, "OBAT"},
			{"Vitamin C 1000mg (7 tablet)", 1, 10500, "OBAT"},
		}},
		{"INV-" + today + "-105", "MR-000005", "SUDAH_BAYAR", "DEBIT", -1, []invItem{
			{"Biaya Konsultasi Bedah (kontrol)", 1, 350000, "KONSULTASI"},
			{"Perawatan Luka", 1, 100000, "TINDAKAN"},
		}},
	}
	addedInv := 0
	for _, inv := range invs {
		var pid int
		_ = db.QueryRow("SELECT id FROM patients WHERE medical_record_number=$1", inv.PatientMRN).Scan(&pid)
		if pid == 0 {
			continue
		}
		var exists int
		_ = db.QueryRow("SELECT COUNT(*) FROM invoices WHERE invoice_number=$1", inv.InvNum).Scan(&exists)
		if exists > 0 {
			continue
		}

		var subtotal float64
		for _, item := range inv.items {
			subtotal += float64(item.Qty) * item.Price
		}
		total := subtotal

		var invID int
		err := db.QueryRow(`INSERT INTO invoices (invoice_number, patient_id, invoice_date, subtotal, discount, total, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`, inv.InvNum, pid, today, subtotal, 0.0, total, inv.Status).Scan(&invID)
		if err != nil {
			fmt.Printf("  [SKIP] Invoice %s: %v\n", inv.InvNum, err)
			continue
		}
		for _, item := range inv.items {
			itemTotal := float64(item.Qty) * item.Price
			_, _ = db.Exec(`INSERT INTO invoice_items (invoice_id, description, quantity, unit_price, total_price, item_type)
				VALUES ($1,$2,$3,$4,$5,$6)`, invID, item.Desc, item.Qty, item.Price, itemTotal, item.Type)
		}
		if inv.PayAmount != 0 {
			amount := inv.PayAmount
			if amount < 0 {
				amount = total
			}
			_, _ = db.Exec(`INSERT INTO payments (payment_number, invoice_id, patient_id, payment_date, amount, payment_method, status)
				VALUES ($1,$2,$3,$4,$5,$6,'COMPLETED')`,
				fmt.Sprintf("PAY-%s-%03d", strings.ReplaceAll(today, "-", ""), invID), invID, pid, today, amount, inv.PayMethod)
		}
		addedInv++
	}
	fmt.Printf("  Transaksi hari ini: %d registrasi, 5 antrian, 2 rekam medis, 4 resep, %d invoice ditambahkan\n", len(regs), addedInv)
}

func seedUsers(db *sql.DB) {
	fmt.Println()
	fmt.Println("Membuat akun login pengguna...")

	type acc struct {
		Username, FullName, Email, Role, Keterangan string
	}
	accounts := []acc{
		{"d001", "dr. Ahmad Suharto, Sp.PD", "ahmad@klinik.com", "DOCTOR", "Dokter Penyakit Dalam"},
		{"d002", "dr. Siti Nurhaliza, Sp.A", "siti@klinik.com", "DOCTOR", "Dokter Anak"},
		{"d003", "dr. Budi Santoso, Sp.BED", "budi@klinik.com", "DOCTOR", "Dokter Bedah"},
		{"d004", "dr. Dewi Lestari, Sp.KK", "dewi@klinik.com", "DOCTOR", "Dokter Kulit"},
		{"d005", "dr. Eko Prasetyo, Sp.THT", "eko@klinik.com", "DOCTOR", "Dokter THT"},
		{"rina", "Rina Wulandari", "rina@klinik.com", "NURSE", "Resepsionis"},
		{"maya", "Maya Putri", "maya@klinik.com", "NURSE", "Perawat"},
		{"lestari", "Lestari Budiman", "lestari@klinik.com", "NURSE", "Perawat"},
		{"andi", "Andi Kurniawan", "andi@klinik.com", "ADMIN", "Admin"},
		{"farhan", "Farhan Maulana", "farhan@klinik.com", "PHARMACIST", "Apoteker"},
		{"kasir01", "Kasir Klinik", "kasir@klinik.com", "CASHIER", "Kasir"},
	}

	created := 0
	for _, a := range accounts {
		exists, err := checkUserExists(db, a.Username)
		if err != nil {
			fmt.Printf("  [SKIP] %s: %v\n", a.Username, err)
			continue
		}
		if exists {
			continue
		}

		var roleID int
		if err := db.QueryRow("SELECT id FROM roles WHERE name=$1", a.Role).Scan(&roleID); err != nil {
			fmt.Printf("  [SKIP] %s: role %s tidak ditemukan\n", a.Username, a.Role)
			continue
		}

		password := a.Username + "123"
		hash, err := auth.HashPassword(password)
		if err != nil {
			fmt.Printf("  [SKIP] %s: %v\n", a.Username, err)
			continue
		}

		var userID int
		err = db.QueryRow(`INSERT INTO users (username, email, password_hash, full_name)
			VALUES ($1,$2,$3,$4) RETURNING id`, a.Username, a.Email, hash, a.FullName).Scan(&userID)
		if err != nil {
			fmt.Printf("  [SKIP] %s: %v\n", a.Username, err)
			continue
		}
		_, _ = db.Exec(`INSERT INTO user_roles (user_id, role_id) VALUES ($1,$2)`, userID, roleID)
		fmt.Printf("  %-8s (%s) password: %s\n", a.Username, a.Keterangan, password)
		created++
	}
	if created == 0 {
		fmt.Println("  Semua akun sudah ada.")
	} else {
		fmt.Printf("  %d akun dibuat.\n", created)
	}
}

func checkUserExists(db *sql.DB, username string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE username = $1`
	err := db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func init() {
	_ = strings.TrimSpace
}
