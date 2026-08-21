package audit

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

func (r *Repository) Log(userID *int, action, entity string, entityID *int, description, ipAddress string) error {
	_, err := r.db.Exec(`INSERT INTO audit_logs (user_id, action, entity, entity_id, description, ip_address)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		userID, action, entity, entityID, description, ipAddress)
	return err
}

func (r *Repository) GetAll(page, limit int, search string) ([]AuditLog, int, error) {
	offset := (page - 1) * limit
	args := []interface{}{}
	where := ""

	if search != "" {
		where = `WHERE (al.action ILIKE $1 OR al.entity ILIKE $1 OR al.description ILIKE $1 OR COALESCE(u.full_name,'') ILIKE $1)`
		args = append(args, "%"+search+"%")
	}

	countQuery := `SELECT COUNT(*) FROM audit_logs al LEFT JOIN users u ON al.user_id = u.id ` + where
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	idx := len(args)
	query := fmt.Sprintf(`SELECT al.id, al.user_id, COALESCE(u.full_name,'System'), al.action, al.entity,
		al.entity_id, COALESCE(al.description,''), COALESCE(al.ip_address,''), al.created_at
		FROM audit_logs al LEFT JOIN users u ON al.user_id = u.id
		%s ORDER BY al.id DESC LIMIT $%d OFFSET $%d`, where, idx-1, idx)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []AuditLog
	for rows.Next() {
		var al AuditLog
		if err := rows.Scan(&al.ID, &al.UserID, &al.UserName, &al.Action, &al.Entity,
			&al.EntityID, &al.Description, &al.IPAddress, &al.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, al)
	}
	return list, total, nil
}
