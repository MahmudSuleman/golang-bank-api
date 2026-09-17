package account

import "context"

type Repository interface {
	Create(ctx context.Context, account Account) (Account, error)
	GetById(ctx context.Context, id int64) (Account, error)
	GetByCustomerID(ctx context.Context, customerId int64) ([]Account, error)
	Deposit(ctx context.Context, id int64, amount int64) (Account, error)
}
