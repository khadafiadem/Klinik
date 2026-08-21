package audit

import "time"

type AuditLog struct {
	ID          int       `json:"id"`
	UserID      *int      `json:"user_id"`
	UserName    string    `json:"user_name,omitempty"`
	Action      string    `json:"action"`
	Entity      string    `json:"entity"`
	EntityID    *int      `json:"entity_id"`
	Description string    `json:"description"`
	IPAddress   string    `json:"ip_address"`
	CreatedAt   time.Time `json:"created_at"`
}
