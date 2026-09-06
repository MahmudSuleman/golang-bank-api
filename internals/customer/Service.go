package customer

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetAll(ctx context.Context) ([]Customer, error) {
	return s.repository.GetAll(ctx)
}

func (s *Service) GetById(ctx context.Context, id int64) (Customer, error) {
	return s.repository.GetById(ctx, id)
}

func (s *Service) Create(ctx context.Context, customer Customer) (Customer, error) {
	return s.repository.Create(ctx, customer)
}
