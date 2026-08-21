package audit

import "database/sql"

type Service struct {
	repo *Repository
}

func NewService(db *sql.DB) *Service {
	return &Service{repo: NewRepository(db)}
}

func (s *Service) Log(userID *int, action, entity string, entityID *int, description, ipAddress string) {
	_ = s.repo.Log(userID, action, entity, entityID, description, ipAddress)
}

func (s *Service) GetAll(page, limit int, search string) ([]AuditLog, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 30
	}
	return s.repo.GetAll(page, limit, search)
}
