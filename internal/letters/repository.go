package letters

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

func (r *Repository) ListActiveTemplates() ([]Template, error) {
	rows, err := r.db.Query(`SELECT id, name, code, COALESCE(subject,''), body_template, is_active, created_at, updated_at
		FROM letter_templates WHERE is_active = TRUE ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Template
	for rows.Next() {
		var t Template
		if err := rows.Scan(&t.ID, &t.Name, &t.Code, &t.Subject, &t.BodyTemplate, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *Repository) GetTemplateByCode(code string) (*Template, error) {
	t := &Template{}
	err := r.db.QueryRow(`SELECT id, name, code, COALESCE(subject,''), body_template, is_active, created_at, updated_at
		FROM letter_templates WHERE code = $1`, code).
		Scan(&t.ID, &t.Name, &t.Code, &t.Subject, &t.BodyTemplate, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template surat tidak ditemukan")
		}
		return nil, err
	}
	return t, nil
}

func (r *Repository) CreateIssuance(iss *Issuance) error {
	query := `INSERT INTO letter_issuances (template_id, letter_number, patient_id, doctor_id, notes, issued_by)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	return r.db.QueryRow(query, iss.TemplateID, iss.LetterNumber, iss.PatientID, iss.DoctorID, iss.Notes, iss.IssuedBy).
		Scan(&iss.ID, &iss.CreatedAt)
}

func (r *Repository) CountIssuances(templateID int) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM letter_issuances WHERE template_id = $1`, templateID).Scan(&count)
	return count, err
}

func (r *Repository) GetIssuance(id int) (*Issuance, error) {
	iss := &Issuance{}
	query := `SELECT li.id, li.template_id, li.letter_number, li.patient_id, p.full_name, p.medical_record_number,
		li.doctor_id, COALESCE(d.full_name,''), COALESCE(li.notes,''), li.issued_by, li.created_at,
		lt.name, lt.code, COALESCE(lt.subject,''), lt.body_template
		FROM letter_issuances li
		JOIN letter_templates lt ON li.template_id = lt.id
		JOIN patients p ON li.patient_id = p.id
		LEFT JOIN doctors d ON li.doctor_id = d.id
		WHERE li.id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&iss.ID, &iss.TemplateID, &iss.LetterNumber, &iss.PatientID, &iss.PatientName, &iss.PatientMRN,
		&iss.DoctorID, &iss.DoctorName, &iss.Notes, &iss.IssuedBy, &iss.CreatedAt,
		&iss.TemplateName, &iss.TemplateCode, &iss.Subject, &iss.BodyTemplate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("surat tidak ditemukan")
		}
		return nil, err
	}
	return iss, nil
}

func (r *Repository) ListRecent(limit int) ([]Issuance, error) {
	rows, err := r.db.Query(`SELECT li.id, li.template_id, li.letter_number, li.patient_id, p.full_name, p.medical_record_number,
		li.doctor_id, COALESCE(d.full_name,''), COALESCE(li.notes,''), li.issued_by, li.created_at,
		lt.name, lt.code, COALESCE(lt.subject,''), lt.body_template
		FROM letter_issuances li
		JOIN letter_templates lt ON li.template_id = lt.id
		JOIN patients p ON li.patient_id = p.id
		LEFT JOIN doctors d ON li.doctor_id = d.id
		ORDER BY li.id DESC LIMIT $1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Issuance
	for rows.Next() {
		var iss Issuance
		if err := rows.Scan(&iss.ID, &iss.TemplateID, &iss.LetterNumber, &iss.PatientID, &iss.PatientName, &iss.PatientMRN,
			&iss.DoctorID, &iss.DoctorName, &iss.Notes, &iss.IssuedBy, &iss.CreatedAt,
			&iss.TemplateName, &iss.TemplateCode, &iss.Subject, &iss.BodyTemplate); err != nil {
			return nil, err
		}
		list = append(list, iss)
	}
	return list, nil
}
