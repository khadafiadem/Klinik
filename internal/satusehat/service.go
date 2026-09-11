package satusehat

import "database/sql"

type Service struct {
	repo *Repository
}

func NewService(db *sql.DB) *Service {
	return &Service{repo: NewRepository(db)}
}

func (s *Service) Get() (*Settings, error) {
	return s.repo.Get()
}

func (s *Service) Update(settings *Settings) error {
	return s.repo.Update(settings)
}
