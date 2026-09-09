package customer

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidCustomer  = errors.New("invalid customer")
	ErrCustomerNotFound = errors.New("customer not found")
)

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

func (s *Service) Create(ctx context.Context, request CreateCustomerRequest) (Customer, error) {

	request.FirstName = strings.TrimSpace(request.FirstName)
	request.LastName = strings.TrimSpace(request.LastName)
	request.Email = strings.TrimSpace(request.Email)

	if request.FirstName == "" || request.LastName == "" || request.Email == "" {
		return Customer{}, ErrInvalidCustomer
	}

	customer := Customer{
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Email:     request.Email,
	}
	return s.repository.Create(ctx, customer)
}
