package clinic

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

func (r *Repository) Get() (*ClinicSettings, error) {
	s := &ClinicSettings{}
	query := `SELECT id, clinic_name, COALESCE(clinic_address,''), COALESCE(clinic_phone,''), 
		COALESCE(clinic_email,''), COALESCE(clinic_logo,''),
		TO_CHAR(opening_time,'HH24:MI'), TO_CHAR(closing_time,'HH24:MI'),
		max_patients_per_day, registration_fee, consultation_fee, tax_percentage, currency,
		created_at, updated_at
		FROM clinic_settings ORDER BY id LIMIT 1`
	err := r.db.QueryRow(query).Scan(
		&s.ID, &s.ClinicName, &s.ClinicAddress, &s.ClinicPhone,
		&s.ClinicEmail, &s.ClinicLogo, &s.OpeningTime, &s.ClosingTime,
		&s.MaxPatientsPerDay, &s.RegistrationFee, &s.ConsultationFee,
		&s.TaxPercentage, &s.Currency, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil pengaturan klinik: %w", err)
	}
	return s, nil
}

func (r *Repository) Update(s *ClinicSettings) error {
	query := `UPDATE clinic_settings SET
		clinic_name=$1, clinic_address=$2, clinic_phone=$3, clinic_email=$4,
		opening_time=$5, closing_time=$6, max_patients_per_day=$7,
		registration_fee=$8, consultation_fee=$9, tax_percentage=$10, currency=$11
		WHERE id=$12`
	_, err := r.db.Exec(query,
		s.ClinicName, s.ClinicAddress, s.ClinicPhone, s.ClinicEmail,
		s.OpeningTime, s.ClosingTime, s.MaxPatientsPerDay,
		s.RegistrationFee, s.ConsultationFee, s.TaxPercentage, s.Currency, s.ID,
	)
	if err != nil {
		return fmt.Errorf("gagal update pengaturan klinik: %w", err)
	}
	return nil
}

func (r *Repository) ListInsuranceProviders() ([]InsuranceProvider, error) {
	rows, err := r.db.Query(`SELECT id, name, sort_order, is_active
		FROM insurance_providers
		WHERE is_active = TRUE
		ORDER BY sort_order ASC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar asuransi: %w", err)
	}
	defer rows.Close()

	var providers []InsuranceProvider
	for rows.Next() {
		var p InsuranceProvider
		if err := rows.Scan(&p.ID, &p.Name, &p.SortOrder, &p.IsActive); err != nil {
			return nil, fmt.Errorf("gagal membaca daftar asuransi: %w", err)
		}
		providers = append(providers, p)
	}
	return providers, nil
}

func (r *Repository) AddInsuranceProvider(name string) error {
	count := 0
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM insurance_providers WHERE name = $1`, name).Scan(&count); err != nil {
		return fmt.Errorf("gagal memeriksa nama asuransi: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("asuransi %q sudah ada", name)
	}

	nextSort := 0
	_ = r.db.QueryRow(`SELECT COALESCE(MAX(sort_order), 0) + 1 FROM insurance_providers`).Scan(&nextSort)

	_, err := r.db.Exec(`INSERT INTO insurance_providers (name, sort_order) VALUES ($1, $2)`, name, nextSort)
	if err != nil {
		return fmt.Errorf("gagal menambah asuransi: %w", err)
	}
	return nil
}

func (r *Repository) DeleteInsuranceProvider(id int) error {
	result, err := r.db.Exec(`DELETE FROM insurance_providers WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus asuransi: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("asuransi tidak ditemukan")
	}
	return nil
}
