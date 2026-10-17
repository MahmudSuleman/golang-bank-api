package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

type fakeRepository struct {
	transactions []Transaction
	called       bool
}

func (f *fakeRepository) Create(ctx context.Context, tx pgx.Tx, transaction Transaction) (Transaction, error) {
	return transaction, nil
}

func (f *fakeRepository) GetByAccountId(ctx context.Context, accountId int64) ([]Transaction, error) {
	f.called = true
	return f.transactions, nil
}

type fakeChecker struct {
	owned bool
	err   error
}

func (f fakeChecker) IsOwnedBy(ctx context.Context, accountId int64, customerId int64) (bool, error) {
	return f.owned, f.err
}

func TestService_GetByAccountId(t *testing.T) {
	dbErr := errors.New("db down")

	tests := []struct {
		name           string
		checker        fakeChecker
		wantErr        error
		wantRepoCalled bool
	}{
		{name: "owner gets history", checker: fakeChecker{owned: true}, wantErr: nil, wantRepoCalled: true},
		{name: "non owner gets not found", checker: fakeChecker{owned: false}, wantErr: ErrAccountNotFound, wantRepoCalled: false},
		{name: "checker error is returned", checker: fakeChecker{err: dbErr}, wantErr: dbErr, wantRepoCalled: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{transactions: []Transaction{{Id: 1, AccountId: 7}}}
			service := NewService(repo, tt.checker)

			got, err := service.GetByAccountId(context.Background(), 7, 42)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}

			if repo.called != tt.wantRepoCalled {
				t.Fatalf("repository called = %v, want %v", repo.called, tt.wantRepoCalled)
			}

			if tt.wantErr == nil && len(got) != 1 {
				t.Fatalf("got %d transactions, want 1", len(got))
			}
		})
	}
}
