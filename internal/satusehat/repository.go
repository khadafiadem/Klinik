package satusehat

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

func (r *Repository) Get() (*Settings, error) {
	s := &Settings{}
	query := `SELECT id, environment, COALESCE(base_url,''), COALESCE(org_id,''),
		COALESCE(org_name,''), COALESCE(location_id,''), COALESCE(client_id,''),
		COALESCE(client_secret,''), is_active
		FROM satusehat_settings ORDER BY id LIMIT 1`
	err := r.db.QueryRow(query).Scan(
		&s.ID, &s.Environment, &s.BaseURL, &s.OrgID,
		&s.OrgName, &s.LocationID, &s.ClientID, &s.ClientSecret, &s.IsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil pengaturan SATUSEHAT: %w", err)
	}
	return s, nil
}

func (r *Repository) Update(s *Settings) error {
	query := `UPDATE satusehat_settings SET
		environment=$1, base_url=$2, org_id=$3, org_name=$4,
		location_id=$5, client_id=$6, client_secret=$7, is_active=$8
		WHERE id=$9`
	_, err := r.db.Exec(query,
		s.Environment, s.BaseURL, s.OrgID, s.OrgName,
		s.LocationID, s.ClientID, s.ClientSecret, s.IsActive, s.ID,
	)
	if err != nil {
		return fmt.Errorf("gagal update pengaturan SATUSEHAT: %w", err)
	}
	return nil
}
