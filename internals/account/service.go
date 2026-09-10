package account

import (
	"bank-api/internals/customer"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var (
	ErrInvalidAccount   = errors.New("invalid account")
	ErrAccountNotFound  = errors.New("account not found")
	ErrCustomerNotFound = errors.New("customer not found")
)

type CustomerChecker interface {
	Exists(ctx context.Context, id int64) (bool, error)
}
type Service struct {
	repository         Repository
	customerRepository customer.Repository
	customerChecker    CustomerChecker
}

func NewService(repository Repository, customerChecker CustomerChecker) *Service {
	return &Service{
		repository:      repository,
		customerChecker: customerChecker,
	}
}

func (s *Service) Create(ctx context.Context,customerId int64, request CreateAccountRequest) (Account, error) {

	exists , err := s.customerChecker.Exists(ctx, customerId)
	if err != nil {
		return Account{}, err
	}
	if !exists {
		return Account{}, ErrCustomerNotFound
	}
	request.AccountType = strings.ToUpper(strings.TrimSpace(request.AccountType))

	request.Currency = strings.ToUpper(strings.TrimSpace(request.Currency))

	if request.CustomerID <= 0 {
		return Account{}, ErrInvalidAccount
	}

	if request.AccountType != AccountTypeCurrent && request.AccountType != AccountTypeSavings {
		return Account{}, ErrInvalidAccount
	}

	if request.Currency == "" {
		request.Currency = CurrencyGHS
	}

	if request.Currency != CurrencyGHS {
		return Account{}, ErrInvalidAccount
	}

	account := Account{
		CustomerID:    customerId,
		AccountNumber: generateAccountNumber(),
		AccountType:   request.AccountType,
		Currency:      request.Currency,
		Balance:       0,
		Status:        AccountStatusActive,
	}
	return s.repository.Create(ctx, account)
}

func (s *Service) GetById(ctx context.Context, customerId int64) (Account, error) {
	return s.repository.GetById(ctx, customerId)
}
func (s *Service) GetByCustomerId(ctx context.Context, customerId int64) ([]Account, error) {
	exists, err := s.customerChecker.Exists(ctx, customerId)

	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrCustomerNotFound
	}

	return s.repository.GetByCustomerID(ctx, customerId)
}
func generateAccountNumber() string {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	return fmt.Sprintf("%010d", random.Int63n(10000000000))
}
