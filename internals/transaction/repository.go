package transaction

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository interface {
	Create(ctx context.Context, tx pgx.Tx, transaction Transaction) (Transaction, error)
	GetByAccountId(ctx context.Context, accountId int64) ([]Transaction, error)
}
