package account

import (
	"bank-api/internals/database"
	"bank-api/internals/transaction"

	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FinancialService struct {
	db                    *pgxpool.Pool
	accountRepository     Repository
	transactionRepository transaction.Repository
}

func NewFinancialService(
	db *pgxpool.Pool,
	accountRepository Repository,
	transactionRepository transaction.Repository,
) *FinancialService {
	return &FinancialService{
		db:                    db,
		accountRepository:     accountRepository,
		transactionRepository: transactionRepository,
	}
}

func (s *FinancialService) Deposit(ctx context.Context, accountId int64, request MoneyRequest) (Account, error) {

	if accountId <= 0 {
		return Account{}, ErrInvalidAccount
	}

	if request.Amount <= 0 {
		return Account{}, ErrInvalidAmount
	}

	var account Account
	err := database.WithTx(
		ctx,
		s.db,
		func(tx pgx.Tx) error {
			var err error

			account, err = s.accountRepository.Deposit(ctx, tx, accountId, request.Amount)
			if err != nil {
				return err
			}

			_, err = s.transactionRepository.Create(
				ctx,
				tx,
				transaction.Transaction{
					AccountId:    accountId,
					Type:         transaction.TypeDeposit,
					Amount:       request.Amount,
					BalanceAfter: account.Balance,
					Reference:    uuid.NewString(),
				},
			)
			return err
		})

	if err != nil {
		return Account{}, err
	}

	return account, nil

}

func (s *FinancialService) Withdraw(
	ctx context.Context,
	accountId int64,
	request MoneyRequest,
) (Account, error) {
	if accountId <= 0 {
		return Account{}, ErrInvalidAccount
	}

	if request.Amount <= 0 {
		return Account{}, ErrInvalidAmount
	}

	var account Account

	err := database.WithTx(
		ctx,
		s.db,
		func(tx pgx.Tx) error {
			var err error
			account, err = s.accountRepository.Withdraw(
				ctx,
				tx,
				accountId,
				request.Amount,
			)

			if err != nil {
				return err
			}

			_, err = s.transactionRepository.Create(
				ctx,
				tx,
				transaction.Transaction{
					AccountId:    accountId,
					Type:         transaction.TypeWithdrawal,
					Amount:       request.Amount,
					BalanceAfter: account.Balance,
					Reference:    uuid.NewString(),
				},
			)
			return err

		})

	if err != nil {
		return Account{}, err
	}

	return account, nil

}

func (s *FinancialService) Transfer(
	ctx context.Context,
	fromAccountId int64,
	request TransferRequest,
) (Account, error) {
	if fromAccountId <= 0 || request.DestinationAccountId <= 0 {
		return Account{}, ErrInvalidAccount
	}

	if fromAccountId == request.DestinationAccountId {
		return Account{}, ErrSameAccount
	}

	if request.Amount <= 0 {
		return Account{}, ErrInvalidAmount
	}

	// Both ledger rows share one transfer id so they can be matched later.
	// `reference` is UNIQUE in the table, so each row gets its own suffix.
	reference := uuid.NewString()

	var source Account

	err := database.WithTx(
		ctx,
		s.db,
		func(tx pgx.Tx) error {
			var destination Account
			var err error

			source, destination, err = s.accountRepository.Transfer(
				ctx,
				tx,
				fromAccountId,
				request.DestinationAccountId,
				request.Amount,
			)
			if err != nil {
				return err
			}

			_, err = s.transactionRepository.Create(
				ctx,
				tx,
				transaction.Transaction{
					AccountId:    source.ID,
					Type:         transaction.TypeTransferDebit,
					Amount:       request.Amount,
					BalanceAfter: source.Balance,
					Reference:    reference + "-DR",
				},
			)
			if err != nil {
				return err
			}

			_, err = s.transactionRepository.Create(
				ctx,
				tx,
				transaction.Transaction{
					AccountId:    destination.ID,
					Type:         transaction.TypeTransferCredit,
					Amount:       request.Amount,
					BalanceAfter: destination.Balance,
					Reference:    reference + "-CR",
				},
			)
			return err
		})

	if err != nil {
		return Account{}, err
	}

	return source, nil
}
