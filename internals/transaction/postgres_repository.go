package transaction

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func (r *PostgresRepository) Create(ctx context.Context, transaction Transaction) (Transaction, error) {
	err := r.db.QueryRow(ctx,
		`
		INSERT INTO transactions(account_id, type, amount, balance_after, reference)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, account_id, type, amount, balance_after, reference, created_at
`, transaction.AccountId, transaction.Type, transaction.Amount, transaction.BalanceAfter, transaction.Reference).
		Scan(&transaction.Id, &transaction.AccountId, &transaction.Type, &transaction.Amount, &transaction.BalanceAfter, &transaction.Reference, &transaction.CreatedAt)
	if err != nil {
		return Transaction{}, err
	}

	return transaction, nil
}

func (r *PostgresRepository) GetByAccountId(ctx context.Context, accountId int64) ([]Transaction, error) {
	rows, err := r.db.Query(ctx,
		`
			SELECT id, account_id, type, amount, balance_after, reference, created_at
			FROM transactions
			WHERE accounts_id = $1
			ORDER BY created_at DESC
			`, accountId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	transactions := make([]Transaction, 0)

	for rows.Next() {
		var trasaction Transaction
		err := rows.Scan(
			&trasaction.Id,
			&trasaction.AccountId,
			&trasaction.Type,
			&trasaction.Amount,
			&trasaction.BalanceAfter,
			&trasaction.Reference,
			&trasaction.CreatedAt,
		)

		if err != nil{
			return nil, err
		}

		transactions = append(transactions, trasaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}
