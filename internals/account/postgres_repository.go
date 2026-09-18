package account

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func (r PostgresRepository) Withdraw(ctx context.Context, id int64, amount int64) (Account, error) {
	var account Account

	err := r.db.QueryRow(ctx, `
		UPDATE accounts 
		SET balance = balance - $1
		WHERE id = $2
		  AND status = 'ACTIVE'
		  AND balance >= $1
		RETURNING 
		id, customer_id, account_number, account_type, currency, status, balance
`, amount, id).Scan(
		&account.ID, &account.CustomerID, &account.AccountNumber, &account.AccountType, &account.Currency, &account.Status, &account.Balance)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Account{}, ErrWithdrawalFailed
		}
		return account, err
	}
	return account, nil
}

func (r PostgresRepository) Deposit(ctx context.Context, id int64, amount int64) (Account, error) {
	var account Account

	err := r.db.QueryRow(ctx, `
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2
		AND status = 'ACTIVE'
		RETURNING id, customer_id, account_number, account_type, currency, balance, status
	`, amount, id).Scan(
		&account.ID,
		&account.CustomerID,
		&account.AccountNumber,
		&account.AccountType,
		&account.Currency,
		&account.Balance,
		&account.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Account{}, ErrAccountOperationFailed
		}
		return account, err
	}
	return account, nil
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}
func (r PostgresRepository) Create(ctx context.Context, account Account) (Account, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO accounts (
			customer_id,
			account_number,
		  account_type,
		  currency,
			balance,
		  status
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
		`,
		account.CustomerID,
		account.AccountNumber,
		account.AccountType,
		account.Currency,
		account.Balance,
		account.Status).Scan(&account.ID)
	if err != nil {
		return Account{}, err
	}
	return account, nil
}

func (r PostgresRepository) GetById(ctx context.Context, id int64) (Account, error) {
	var account Account

	err := r.db.QueryRow(ctx, `
		SELECT id, customer_id, account_number, account_type, currency, balance, status
		FROM accounts
		WHERE id = $1`,
		id).Scan(
		&account.ID,
		&account.CustomerID,
		&account.AccountNumber,
		&account.AccountType,
		&account.Currency,
		&account.Balance,
		&account.Status,
	)

	if err != nil {
		return Account{}, err
	}
	return account, nil
}

func (r PostgresRepository) GetByCustomerID(ctx context.Context, customerId int64) ([]Account, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, customer_id, account_number, account_type, currency, balance, status
		FROM accounts
		WHERE customer_id = $1
		ORDER BY id
		`, customerId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var accounts []Account
	for rows.Next() {
		var account Account
		err := rows.Scan(
			&account.ID,
			&account.CustomerID,
			&account.AccountNumber,
			&account.AccountType,
			&account.Currency,
			&account.Balance,
			&account.Status,
		)

		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)

	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}
