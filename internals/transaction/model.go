package transaction

import "time"

const (
	TypeDeposit        = "DEPOSIT"
	TypeWithdrawal     = "WITHDRAWAL"
	TypeTransferDebit  = "TRANSFER_DEBIT"
	TypeTransferCredit = "TRANSFER_CREDIT"
)

type Transaction struct {
	Id           int64     `json:"id"`
	AccountId    int64     `json:"account_id"`
	Type         string    `json:"type"`
	Amount       int64     `json:"amount"`
	BalanceAfter int64     `json:"balance_after"`
	Reference    string    `json:"reference"`
	CreatedAt    time.Time `json:"created_at"`
}
