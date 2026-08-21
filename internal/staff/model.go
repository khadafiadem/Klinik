package staff

import "time"

type Staff struct {
	ID        int       `json:"id"`
	UserID    *int      `json:"user_id,omitempty"`
	StaffCode string    `json:"staff_code"`
	FullName  string    `json:"full_name"`
	Position  string    `json:"position"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
