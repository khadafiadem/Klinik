package reports

import "database/sql"

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type PatientSummary struct {
	TotalPatients  int
	MalePatients   int
	FemalePatients int
	NewThisMonth   int
}

type RegistrationSummary struct {
	TotalToday      int
	TotalWeek       int
	TotalMonth      int
	CompletedToday  int
	CancelledToday  int
}

type DoctorActivity struct {
	DoctorName    string
	Specialization string
	VisitCount    int
	TodayCount    int
}

type RevenueSummary struct {
	TodayRevenue   float64
	WeekRevenue    float64
	MonthRevenue   float64
	TotalRevenue   float64
}

type MedicineStockSummary struct {
	TotalMedicines  int
	LowStockCount   int
	ExpiredCount    int
}

type PaymentSummary struct {
	CashTotal     float64
	TransferTotal float64
	QRISTotal     float64
	BPJSTotal     float64
	OtherTotal    float64
	CompletedCount int
	PendingCount   int
}

func (s *Service) GetPatientSummary() (*PatientSummary, error) {
	ps := &PatientSummary{}
	_ = s.db.QueryRow("SELECT COUNT(*) FROM patients").Scan(&ps.TotalPatients)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM patients WHERE gender='LAKI_LAKI'").Scan(&ps.MalePatients)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM patients WHERE gender='PEREMPUAN'").Scan(&ps.FemalePatients)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM patients WHERE created_at >= date_trunc('month', CURRENT_DATE)`).Scan(&ps.NewThisMonth)
	return ps, nil
}

func (s *Service) GetRegistrationSummary() (*RegistrationSummary, error) {
	rs := &RegistrationSummary{}
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM registrations WHERE registration_date = CURRENT_DATE`).Scan(&rs.TotalToday)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM registrations WHERE registration_date >= CURRENT_DATE - INTERVAL '7 days'`).Scan(&rs.TotalWeek)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM registrations WHERE registration_date >= date_trunc('month', CURRENT_DATE)`).Scan(&rs.TotalMonth)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM registrations WHERE registration_date = CURRENT_DATE AND status='COMPLETED'`).Scan(&rs.CompletedToday)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM registrations WHERE registration_date = CURRENT_DATE AND status='CANCELLED'`).Scan(&rs.CancelledToday)
	return rs, nil
}

func (s *Service) GetDoctorActivity() ([]DoctorActivity, error) {
	rows, err := s.db.Query(`SELECT d.full_name, COALESCE(d.specialization,''),
		COUNT(mr.id) AS total,
		COUNT(CASE WHEN mr.examination_date = CURRENT_DATE THEN 1 END) AS today
		FROM doctors d
		LEFT JOIN medical_records mr ON d.id = mr.doctor_id
		GROUP BY d.full_name, d.specialization
		ORDER BY total DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []DoctorActivity
	for rows.Next() {
		var da DoctorActivity
		if err := rows.Scan(&da.DoctorName, &da.Specialization, &da.VisitCount, &da.TodayCount); err != nil {
			return nil, err
		}
		list = append(list, da)
	}
	return list, nil
}

func (s *Service) GetRevenueSummary() (*RevenueSummary, error) {
	rs := &RevenueSummary{}
	_ = s.db.QueryRow(`SELECT COALESCE(SUM(amount),0) FROM payments WHERE payment_date = CURRENT_DATE AND status='COMPLETED'`).Scan(&rs.TodayRevenue)
	_ = s.db.QueryRow(`SELECT COALESCE(SUM(amount),0) FROM payments WHERE payment_date >= CURRENT_DATE - INTERVAL '7 days' AND status='COMPLETED'`).Scan(&rs.WeekRevenue)
	_ = s.db.QueryRow(`SELECT COALESCE(SUM(amount),0) FROM payments WHERE payment_date >= date_trunc('month', CURRENT_DATE) AND status='COMPLETED'`).Scan(&rs.MonthRevenue)
	_ = s.db.QueryRow(`SELECT COALESCE(SUM(amount),0) FROM payments WHERE status='COMPLETED'`).Scan(&rs.TotalRevenue)
	return rs, nil
}

func (s *Service) GetMedicineStockSummary() (*MedicineStockSummary, error) {
	ms := &MedicineStockSummary{}
	_ = s.db.QueryRow("SELECT COUNT(*) FROM medicines").Scan(&ms.TotalMedicines)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM medicines WHERE stock <= minimum_stock").Scan(&ms.LowStockCount)
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM medicines WHERE expiry_date IS NOT NULL AND expiry_date < CURRENT_DATE`).Scan(&ms.ExpiredCount)
	return ms, nil
}

func (s *Service) GetPaymentSummary() (*PaymentSummary, error) {
	ps := &PaymentSummary{}
	_ = s.db.QueryRow(`SELECT COALESCE(SUM(CASE WHEN payment_method='CASH' THEN amount ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN payment_method='BANK_TRANSFER' THEN amount ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN payment_method='QRIS' THEN amount ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN payment_method='BPJS' THEN amount ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN payment_method NOT IN ('CASH','BANK_TRANSFER','QRIS','BPJS') THEN amount ELSE 0 END),0),
		COUNT(CASE WHEN status='COMPLETED' THEN 1 END),
		COUNT(CASE WHEN status='PENDING' THEN 1 END)
		FROM payments`).Scan(
		&ps.CashTotal, &ps.TransferTotal, &ps.QRISTotal,
		&ps.BPJSTotal, &ps.OtherTotal, &ps.CompletedCount, &ps.PendingCount,
	)
	return ps, nil
}

type LowStockMedicine struct {
	ID            int
	Name          string
	MedicineCode  string
	Stock         int
	MinimumStock  int
}

func (s *Service) GetLowStockMedicines() ([]LowStockMedicine, error) {
	rows, err := s.db.Query(`SELECT id, name, medicine_code, stock, minimum_stock
		FROM medicines WHERE stock <= minimum_stock ORDER BY stock ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []LowStockMedicine
	for rows.Next() {
		var m LowStockMedicine
		if err := rows.Scan(&m.ID, &m.Name, &m.MedicineCode, &m.Stock, &m.MinimumStock); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}
