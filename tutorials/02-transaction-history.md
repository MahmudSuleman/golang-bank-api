# Tutorial 02 — Transaction history endpoint

> **Where you are:** Every deposit, withdrawal and transfer now writes rows to the `transactions`
> table, and you fixed the `accounts_id` typo in `GetByAccountId`. But no endpoint reads those rows
> yet, so the ledger can't be seen from outside the database.
>
> **Goal of this tutorial:** Add
>
> ```
> GET /accounts/{id}/transactions
> ```
>
> It returns the account's ledger, newest first. **Only the owner of the account** can see it.
> Anyone else gets a 404.
>
> Along the way you will hit a real Go problem, the **import cycle**, and solve it the idiomatic way.
> You will also write your **first unit test** in this project, without a database.

Estimated time: 60 minutes.

---

## Go concepts you'll practise

| Concept | Where it shows up |
|---|---|
| Package layout and import cycles | Why `transaction` can't import `account` |
| Interfaces defined by the *consumer* | `transaction.AccountOwnershipChecker` |
| Compile-time interface checks | `var _ Interface = (*Type)(nil)` |
| Sentinel errors per package | `transaction.ErrAccountNotFound` |
| Constructor wiring in `main.go` | Passing `accountService` into `transaction.NewService` |
| Table-driven tests + fakes | `service_test.go` |

---

## Step 0 — Design: where does this code live?

The data comes from the `transaction` package, so the handler and service should live there too:

```
internals/transaction/
├── model.go                 (exists)
├── repository.go            (exists)
├── postgres_repository.go   (exists — GetByAccountId is already written)
├── service.go               NEW
├── service_test.go          NEW
└── handler.go               NEW
```

There's one complication. Before returning the history, the service must check **"does this
account belong to the logged-in customer?"**. The `account` package already knows how to answer
that, so the obvious move is to import it:

```go
package transaction

import "bank-api/internals/account" // ❌
```

### 🧪 Try it and see the error

Create `internals/transaction/service.go` containing only:

```go
package transaction

import "bank-api/internals/account"

var _ = account.Account{}
```

Run `go build ./...`:

```
package bank-api
	imports bank-api/internals/account
	imports bank-api/internals/transaction
	imports bank-api/internals/account: import cycle not allowed
```

`account` already imports `transaction`, because `FinancialService` creates ledger rows. If
`transaction` imported `account` as well, the two packages would depend on each other. **Go
forbids import cycles completely.** Other languages allow them, but in Go it's a compile error.

### The idiomatic fix: let the consumer define the interface

You've already used this pattern. Look at `account/service.go`:

```go
type CustomerChecker interface {
	Exists(ctx context.Context, id int64) (bool, error)
}
```

The `account` package doesn't import a concrete customer service. It says *"give me anything with
an `Exists` method"*, and `main.go` passes in the customer repository.

You'll do the same here. `transaction` declares the small interface it needs, and
`account.Service` will satisfy it **without either package importing the other**. Go interfaces
are satisfied *implicitly* (there is no `implements` keyword), which is why this pattern works so
naturally in Go.

```
            main.go  (knows everybody, wires them together)
           /       \
     account  ───►  transaction
  (implements         (declares AccountOwnershipChecker,
   IsOwnedBy)          never imports account)
```

Empty out that experimental `service.go` before moving on. Step 2 fills it with the real code.

---

## Step 1 — Add an ownership check to `account.Service`

Open `internals/account/service.go` and add this method:

```go
// IsOwnedBy reports whether the account exists and belongs to the customer.
// A missing account returns false with no error.
func (s *Service) IsOwnedBy(ctx context.Context, accountId int64, customerId int64) (bool, error) {
	account, err := s.repository.GetById(ctx, accountId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return account.CustomerID == customerId, nil
}
```

Add `"github.com/jackc/pgx/v5"` to the imports. `PostgresRepository.GetById` returns the raw
`pgx.ErrNoRows` when nothing matches, so you have to check for it here.

> 💡 **Why `bool` and not `ErrForbidden`?** If `IsOwnedBy` returned `account.ErrForbidden`, the
> `transaction` package would need to import `account` to check it with `errors.Is`, and the
> import cycle would come back. A plain `bool` only uses built-in types, so it can cross the
> package boundary without creating a dependency.

---

## Step 2 — Write the transaction service

Now put real content in `internals/transaction/service.go`:

```go
package transaction

import (
	"context"
	"errors"
)

var ErrAccountNotFound = errors.New("account not found")

// AccountOwnershipChecker is implemented by account.Service.
// It is declared here so this package never has to import account.
type AccountOwnershipChecker interface {
	IsOwnedBy(ctx context.Context, accountId int64, customerId int64) (bool, error)
}

type Service struct {
	repository     Repository
	accountChecker AccountOwnershipChecker
}

func NewService(repository Repository, accountChecker AccountOwnershipChecker) *Service {
	return &Service{
		repository:     repository,
		accountChecker: accountChecker,
	}
}

func (s *Service) GetByAccountId(ctx context.Context, accountId int64, customerId int64) ([]Transaction, error) {
	owned, err := s.accountChecker.IsOwnedBy(ctx, accountId, customerId)
	if err != nil {
		return nil, err
	}

	if !owned {
		return nil, ErrAccountNotFound
	}

	return s.repository.GetByAccountId(ctx, accountId)
}
```

> 🔒 **Why 404 and not 403 for someone else's account?** A 403 says *"this account exists, but
> it's not yours"*. An attacker could loop through ids `1, 2, 3, …` and learn which accounts
> exist. If "doesn't exist" and "not yours" give the same 404, nothing leaks. This is why the
> service returns the same `ErrAccountNotFound` in both cases.

> Notice `transaction.ErrAccountNotFound` is a *different variable* from `account.ErrAccountNotFound`
> even though they have the same text. `errors.Is` compares the **error value**, not the message,
> so each package owns its own sentinel errors.

---

## Step 3 — Prove `account.Service` satisfies the interface at compile time

Because interfaces are implicit, a typo such as `IsOwnedby` (lowercase b) would compile
fine in `account`. The mistake would only show up later in `main.go`, where the compiler error is
harder to read. You can make the compiler check it right where the method lives.

At the top of `internals/account/service.go`, below the error variables, add:

```go
var _ transaction.AccountOwnershipChecker = (*Service)(nil)
```

and import `"bank-api/internals/transaction"`. That's allowed, because `account` → `transaction`
is the direction that already exists.

How to read that line:

- `(*Service)(nil)` is a nil pointer of type `*Service`. It costs nothing at runtime.
- Assigning it to a variable of the interface type makes the compiler confirm that `*Service`
  has every method of `AccountOwnershipChecker`.
- `_` throws the variable away. Only the check matters.

🧪 Rename `IsOwnedBy` to `IsOwnedBy2` and run `go build ./...` to see the error, then change it
back.

---

## Step 4 — Write the handler

Create `internals/transaction/handler.go`:

```go
package transaction

import (
	"bank-api/internals/auth"
	"bank-api/internals/response"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetByAccountId godoc
// @Summary Get account transactions
// @Description Returns the ledger entries for an account owned by the authenticated customer, newest first.
// @Tags Transactions
// @Produce json
// @Param id path int64 true "Account ID"
// @Success 200 {array} Transaction
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /accounts/{id}/transactions [get]
func (h *Handler) GetByAccountId(w http.ResponseWriter, r *http.Request) {
	customerId, ok := r.Context().Value(auth.CustomerIdKey).(int64)
	if !ok {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	accountId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || accountId <= 0 {
		response.ErrorJSON(w, http.StatusBadRequest, "INVALID_ACCOUNT_ID", "Invalid account id")
		return
	}

	transactions, err := h.service.GetByAccountId(r.Context(), accountId, customerId)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			response.ErrorJSON(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Account not found")
			return
		}

		response.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, transactions)
}
```

Notes:

- `r.Context().Value(auth.CustomerIdKey).(int64)` is a **type assertion**. `Value` returns
  `any`, and `.(int64)` says *"I expect an int64"*. The two-value form `v, ok := ...` doesn't
  panic if you're wrong; `ok` is just `false`. The one-value form `v := x.(int64)` **panics** on
  a mismatch, so always use the two-value form for data you don't control.
- If an account has no transactions, the response is `[]`, not `null`. That's because the
  repository starts with `make([]Transaction, 0)`. A `nil` slice encodes to `null` in JSON and an
  empty slice encodes to `[]`. Clients prefer `[]`. Compare with `account.GetByCustomerID`, which
  uses `var accounts []Account`: open `GET /customers/{id}/accounts` for a customer with no
  accounts and you'll see `null`.

---

## Step 5 — Fix the ordering tie

Both ledger rows from a transfer are inserted in the same DB transaction, and Postgres's `NOW()`
returns the **same timestamp for the whole transaction**. So `ORDER BY created_at DESC` can't
decide which of the two comes first, and the order may differ between calls.

In `internals/transaction/postgres_repository.go`, inside `GetByAccountId`, add a tiebreaker:

```sql
ORDER BY created_at DESC, id DESC
```

`id` comes from a sequence, so it always increases, which makes the order stable.

While you're in that file, rename the loop variable `trasaction` → `transaction`. Typos in names
spread when other code gets copied from them.

---

## Step 6 — Wire it together

### `main.go`

`transactionRepository` is already created. After `accountService` and `transactionRepository`
exist, add:

```go
transactionService := transaction.NewService(transactionRepository, accountService)
transactionHandler := transaction.NewHandler(transactionService)
```

Here `accountService` (a `*account.Service`) is passed where an
`AccountOwnershipChecker` is expected. `main.go` is the only place that knows about both packages,
and it's the right place to connect them.

Then pass the new handler to the router:

```go
router := server.NewRouter(customerHandler, accountHandler, transactionHandler, userHandler, jwtManager)
```

### `internals/server/server.go`

Add the parameter (and import `"bank-api/internals/transaction"`):

```go
func NewRouter(customerHandler *customer.Handler,
	accountHandler *account.Handler,
	transactionHandler *transaction.Handler,
	userHandler *user.Handler,
	jwtManager *auth.JWTManager,
) http.Handler {
```

and register the route inside the existing `r.Route("/accounts", ...)` block:

```go
r.Get("/{id}/transactions", transactionHandler.GetByAccountId)
```

It's inside the `r.Group` that uses `jwtManager.Authenticate`, so the middleware has already put
the customer id into the context before your handler runs.

Build:

```bash
go build ./...
go vet ./...
```

---

## Step 7 — Your first unit test

The interface you added in Step 2 also makes the service **testable without Postgres**. You can
pass in a fake checker and a fake repository.

Create `internals/transaction/service_test.go`:

```go
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
```

Run it:

```bash
go test ./internals/transaction/ -v
```

Expected:

```
=== RUN   TestService_GetByAccountId
=== RUN   TestService_GetByAccountId/owner_gets_history
=== RUN   TestService_GetByAccountId/non_owner_gets_not_found
=== RUN   TestService_GetByAccountId/checker_error_is_returned
--- PASS: TestService_GetByAccountId (0.00s)
...
PASS
```

What's going on:

- **Table-driven tests** are the standard Go style. Each case is a row in a slice of anonymous
  structs, and `t.Run` gives each row its own named subtest. To add a case, add a row.
- **`package transaction`** (not `transaction_test`) means the test can see unexported names. Use
  it for unit tests like this one.
- **`fakeRepository` uses a pointer receiver** (`*fakeRepository`) so `GetByAccountId` can set
  `f.called = true` on the original value. With a value receiver it would set the field on a
  *copy*, and the test would never see it. Try changing it to `(f fakeRepository)` and see what
  the compiler and the test do.
- `errors.Is(nil, nil)` is `true`, so the success row works with `wantErr: nil`.
- The "non owner" row is the one that matters most: it proves **the repository is never even
  queried** for someone else's account.

🧪 Break the service on purpose: delete the `if !owned { ... }` block and rerun the tests. The
"non owner" case should fail. Then put it back.

---

## Step 8 — Test it manually

Regenerate swagger and start the app:

```bash
swag init
air        # or: go run .
```

Add to `api.http`:

```http
###
# account transaction history
GET {{baseUrl}}/accounts/1/transactions
Authorization: Bearer {{token}}
```

### Expected

```json
[
  {
    "id": 5,
    "account_id": 1,
    "type": "TRANSFER_DEBIT",
    "amount": 10000,
    "balance_after": 40000,
    "reference": "3f1c...-DR",
    "created_at": "2026-10-16T12:01:44.123Z"
  },
  {
    "id": 1,
    "account_id": 1,
    "type": "DEPOSIT",
    "amount": 50000,
    "balance_after": 50000,
    "reference": "9a7e...",
    "created_at": "2026-10-16T11:58:02.456Z"
  }
]
```

A quick sanity check: the newest row's `balance_after` should equal the account's current
`balance` from `GET /accounts/1`.

### Error cases to try

| Request | Expected |
|---|---|
| No `Authorization` header | 401 `UNAUTHORIZED` (from the middleware) |
| `/accounts/abc/transactions` | 400 `INVALID_ACCOUNT_ID` |
| `/accounts/99999/transactions` | 404 `ACCOUNT_NOT_FOUND` |
| An account that belongs to **another customer** | 404 `ACCOUNT_NOT_FOUND` (same as above, on purpose) |
| A fresh account with no activity | 200 with `[]` |

To test the "another customer" row, register and log in a second user for a different customer.
Then call the endpoint with their token on your account id.

---

## Commit

```bash
git add internals/ main.go docs/ api.http
git commit -m "add account transaction history endpoint"
```

---

## Practice exercises (optional)

1. **Pagination.** An account with years of history would return thousands of rows. Add
   `?limit=20&offset=0` query parameters. Read them with `r.URL.Query().Get("limit")`, default to
   20, cap at 100, and pass them through the service to the repository (`LIMIT $2 OFFSET $3`).
   Add a test row to check that a `limit` of 500 is capped.
2. **Reuse `IsOwnedBy`.** `account.Service.GetByIdForCustomer` does almost the same check, and its
   handler turns *every* error into 404, including database failures. Change it so only
   "not found / not yours" gives 404 and anything else gives 500.
3. **Consistent empty lists.** Make `account.GetByCustomerID` return `[]` instead of `null` for a
   customer with no accounts. The fix is one line.

---

## What's next

- **Tutorial 03 — Ownership checks on money movement:** right now **any logged-in customer can
  deposit to, withdraw from, or transfer out of any account**. You'll reuse the `IsOwnedBy` idea
  inside `FinancialService` so only the owner can move money out. Fixing this security hole is the
  most important remaining item before this API could go anywhere near production.
