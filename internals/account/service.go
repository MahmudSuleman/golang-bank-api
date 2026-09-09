package account

import (
	"bank-api/internals/customer"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrInvalidAccount = errors.New("invalid account")
var ErrAccountNotFound = errors.New("account not found")

type Service struct {
	repository         Repository
	customerRepository customer.Repository
}

func NewService(repository Repository, customerRepository customer.Repository) *Service {
	return &Service{
		repository:         repository,
		customerRepository: customerRepository,
	}
}

func (s *Service) Create(ctx context.Context, request CreateAccountRequest) (Account, error) {
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
		CustomerID:    request.CustomerID,
		AccountNumber: generateAccountNumber(),
		AccountType:   request.AccountType,
		Currency:      request.Currency,
		Balance:       0,
		Status:        AccountStatusActive,
	}

	_, err := s.customerRepository.GetById(ctx, request.CustomerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Account{}, customer.ErrCustomerNotFound
		}
		return Account{}, err
	}
	return s.repository.Create(ctx, account)
}

func (s *Service) GetById(ctx context.Context, customerId int64) (Account, error) {
	if customerId <= 0 {
		return Account{}, ErrInvalidAccount
	}

	return s.repository.GetById(ctx, customerId)
}
func (s *Service) GetByCustomerId(ctx context.Context, customerId int64) ([]Account, error) {
	if customerId <= 0 {
		return nil, ErrInvalidAccount
	}

	return s.repository.GetByCustomerID(ctx, customerId)
}
func generateAccountNumber() string {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	return fmt.Sprintf("%010d", random.Int63n(10000000000))
}
