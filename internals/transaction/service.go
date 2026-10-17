package transaction

import (
	"context"
	"errors"
)

var ErrAccountNotFound = errors.New("account not found")

type AccountOwnershipCheck interface {
	IsOwnedBy(ctx context.Context, accountId int64, customerId int64) (bool, error)
}

type Service struct {
	repository   Repository
	accountCheck AccountOwnershipCheck
}

func NewService(repository Repository, accountCheck AccountOwnershipCheck) *Service {
	return &Service{repository: repository, accountCheck: accountCheck}
}

func (s *Service) GetByAccountId(ctx context.Context, accountId int64, customerId int64) ([]Transaction, error) {
	owned, err := s.accountCheck.IsOwnedBy(ctx, accountId, customerId)
	if err != nil {
		return nil, err
	}
	if !owned {
		return nil, ErrAccountNotFound
	}
	return s.repository.GetByAccountId(ctx, accountId)
}
