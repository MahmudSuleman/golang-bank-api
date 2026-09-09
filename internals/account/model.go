package account

type Account struct {
	ID            int64  `json:"id"`
	CustomerID    int64  `json:"customer_id"`
	AccountNumber string `json:"account_number"`
	AccountType   string `json:"account_type"`
	Currency      string `json:"currency"`
	Balance       int64  `json:"balance"`
	Status        string `json:"status"`
}

type CreateAccountRequest struct {
	CustomerID  int64  `json:"customer_id"`
	AccountType string `json:"account_type"`
	Currency    string `json:"currency"`
}

const (
	AccountTypeSavings = "SAVINGS"
	AccountTypeCurrent = "CURRENT"
)

const (
	AccountStatusActive  = "ACTIVE"
	AccountStatusBlocked = "BLOCKED"
	AccountStatusClosed  = "CLOSED"
)

const CurrencyGHS = "GHS"
