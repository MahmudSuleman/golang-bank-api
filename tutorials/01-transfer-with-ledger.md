# Tutorial 01 — Move Transfer into `FinancialService` and record it in the ledger

> **Where you are:** In the commit `refactor transaction logic` you created `FinancialService`,
> which wraps **Deposit** and **Withdraw** in a database transaction (`database.WithTx`) and writes
> a row to the `transactions` table (the *ledger*) for each one.
>
> **What is missing:** **Transfer** still uses the old approach. It lives in `account.Service`,
> the repository opens and commits its own `tx`, and **no ledger rows are written**. If someone
> moves money from account 1 to account 2, nothing in `transactions` records it.
>
> **Goal of this tutorial:** Move Transfer to the same pattern as Deposit and Withdraw:
>
> 1. The repository takes a `pgx.Tx` instead of opening its own transaction.
> 2. `FinancialService.Transfer` owns the transaction and writes **two** ledger rows:
>    a `TRANSFER_DEBIT` on the source and a `TRANSFER_CREDIT` on the destination.
> 3. The handler calls `financialService.Transfer` and returns the updated source account.

Estimated time: 45–60 minutes.

---

## Go concepts you'll practise

| Concept | Where it shows up |
|---|---|
| Interfaces as contracts | Changing the `Repository` interface makes the compiler point at every place to fix |
| Multiple return values | `Transfer` returns `(Account, Account, error)` |
| Closures that capture variables | The `func(tx pgx.Tx) error` passed to `WithTx` |
| `=` vs `:=` (shadowing) | The most common bug when you assign to outer variables inside a closure |
| Small helper functions | `lockAccount` and `ensureActive` remove duplicated code |
| Sentinel errors + `errors.Is` | Mapping domain errors to HTTP status codes |

---

## Step 0 — Understand the current flow

Read these files before you change anything:

- `internals/account/financial_service.go`: look at how `Deposit` uses `database.WithTx`.
- `internals/account/postgres_repository.go`: look at `Transfer` and see how it calls
  `r.db.Begin(ctx)`, `defer tx.Rollback(ctx)` and `tx.Commit(ctx)` itself.
- `internals/database/tx.go`: `WithTx` already does Begin, Rollback and Commit for you.

Today the call chain is:

```
Handler.Transfer ──► Service.Transfer (validation) ──► PostgresRepository.Transfer (opens its own tx)
```

After this tutorial it will be:

```
Handler.Transfer ──► FinancialService.Transfer ──► database.WithTx
                                                     ├─► accountRepository.Transfer(tx)   (lock + move money)
                                                     ├─► transactionRepository.Create(tx) (DEBIT row)
                                                     └─► transactionRepository.Create(tx) (CREDIT row)
```

The money movement and both ledger rows commit together or not at all. That guarantee is
called **atomicity**.

> 💡 **Why not keep the repo's own transaction?** `transactionRepository.Create` must run on the
> **same** `tx` as the balance updates. If the repository opens a private transaction, the service
> has no way to add more statements to it. The layer that knows the whole unit of work (the
> service) should own the transaction.

---

## Step 1 — Change the repository contract

Open `internals/account/repository.go` and change the `Transfer` line:

```go
// before
Transfer(ctx context.Context, fromAccountId int64, toAccountId int64, amount int64) error

// after
Transfer(ctx context.Context, tx pgx.Tx, fromAccountId int64, toAccountId int64, amount int64) (Account, Account, error)
```

It now returns the **source** and **destination** accounts after the update, because the service
needs both new balances for the ledger's `balance_after` column.

Now run:

```bash
go build ./...
```

The build **fails** with something like:

```
*PostgresRepository does not implement Repository (wrong type for method Transfer)
```

The failure is useful. In Go, a type satisfies an interface implicitly, so when you change the
interface the compiler lists every implementation that no longer matches. You can use the errors
as a to-do list.

---

## Step 2 — Add two small helpers in the repository

The current `Transfer` has the same `SELECT ... FOR UPDATE` + `Scan` block twice. Pull it into a
helper. At the bottom of `internals/account/postgres_repository.go`, next to
`accountIdsInLockOrder`, add:

```go
// lockAccount loads an account and takes a row lock on it until the tx ends.
func lockAccount(ctx context.Context, tx pgx.Tx, id int64) (Account, error) {
	var account Account

	err := tx.QueryRow(ctx, `
		SELECT id, customer_id, account_number, account_type, currency, balance, status
		FROM accounts
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(
		&account.ID,
		&account.CustomerID,
		&account.AccountNumber,
		&account.AccountType,
		&account.Currency,
		&account.Balance,
		&account.Status,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Account{}, ErrAccountNotFound
		}
		return Account{}, err
	}

	return account, nil
}

// ensureActive turns an account status into the matching domain error.
func ensureActive(account Account) error {
	switch account.Status {
	case AccountStatusActive:
		return nil
	case AccountStatusBlocked:
		return ErrAccountBlocked
	case AccountStatusClosed:
		return ErrAccountClosed
	default:
		return ErrAccountOperationFailed
	}
}
```

Notes:

- `FOR UPDATE` locks the row. Any other transaction that tries to lock or update the same row
  **waits** until this transaction commits or rolls back. Without the lock, two transfers running
  at the same time could both read a balance of 100 and both withdraw 100.
- `ensureActive` fixes a small bug in the old code, which returned `ErrAccountBlocked` for closed
  accounts too. A Go `switch` doesn't fall through by default, so you don't need any `break`
  statements.

---

## Step 3 — Rewrite `PostgresRepository.Transfer`

Replace the whole `Transfer` method with:

```go
func (r *PostgresRepository) Transfer(
	ctx context.Context,
	tx pgx.Tx,
	fromAccountId int64,
	toAccountId int64,
	amount int64,
) (Account, Account, error) {
	// Always lock rows in the same order (lowest id first) to avoid deadlocks.
	firstId, secondId := accountIdsInLockOrder(fromAccountId, toAccountId)

	firstAccount, err := lockAccount(ctx, tx, firstId)
	if err != nil {
		return Account{}, Account{}, err
	}

	secondAccount, err := lockAccount(ctx, tx, secondId)
	if err != nil {
		return Account{}, Account{}, err
	}

	// identify source and destination accounts
	source, destination := firstAccount, secondAccount
	if source.ID != fromAccountId {
		source, destination = destination, source
	}

	if err := ensureActive(source); err != nil {
		return Account{}, Account{}, err
	}

	if err := ensureActive(destination); err != nil {
		return Account{}, Account{}, err
	}

	if source.Currency != destination.Currency {
		return Account{}, Account{}, ErrDifferentCurrency
	}

	if source.Balance < amount {
		return Account{}, Account{}, ErrInsufficientBalance
	}

	// debit source account
	err = tx.QueryRow(ctx, `
		UPDATE accounts
		SET balance = balance - $1
		WHERE id = $2
		RETURNING balance
	`, amount, source.ID).Scan(&source.Balance)
	if err != nil {
		return Account{}, Account{}, err
	}

	// credit destination account
	err = tx.QueryRow(ctx, `
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2
		RETURNING balance
	`, amount, destination.ID).Scan(&destination.Balance)
	if err != nil {
		return Account{}, Account{}, err
	}

	return source, destination, nil
}
```

What changed and why:

| Old | New | Why |
|---|---|---|
| `tx, err := r.db.Begin(ctx)` | `tx` is a parameter | The service owns the transaction now |
| `defer tx.Rollback(ctx)` / `tx.Commit(ctx)` | removed | `database.WithTx` does both |
| `tx.Exec(...)` | `tx.QueryRow(... RETURNING balance).Scan(...)` | You need the new balance for the ledger |
| `if fromAccountId == firstAccount.ID { ... } else { ... }` | `source, destination = destination, source` | Go swaps two values in one line with a tuple assignment |

> 💡 **Deadlocks:** Suppose transfer A goes 1 → 2 and transfer B goes 2 → 1 at the same moment.
> If each one locked its *source* first, A would hold row 1 and wait for row 2 while B held row 2
> and waited for row 1, and neither could finish. `accountIdsInLockOrder` makes every transfer lock
> the lower id first, so they queue up instead of deadlocking. You already wrote this helper.

Run `go build ./...` again. The repository now compiles, and the errors move to
`service.go`, because `Service.Transfer` still calls the old signature. Fix that next.

---

## Step 4 — Delete `Service.Transfer`

Open `internals/account/service.go` and **delete** the whole `Transfer` method. Its validation
moves into `FinancialService`, the same way the Deposit and Withdraw validation moved in your
last commit.

`go build ./...` now complains in `handler.go` (`h.service.Transfer undefined`). That's expected,
and Step 6 fixes it.

---

## Step 5 — Add `FinancialService.Transfer`

Open `internals/account/financial_service.go` and add this method below `Withdraw`:

```go
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
```

### ⚠️ The `=` vs `:=` trap

Look at this line inside the closure:

```go
source, destination, err = s.accountRepository.Transfer(...)
```

It uses `=`, not `:=`. That matters because `source` was declared **outside** the closure
(`var source Account`). The closure *captures* that variable, so assigning to it inside the
closure changes the same variable the function returns at the end.

If you wrote `:=` instead:

```go
source, destination, err := s.accountRepository.Transfer(...) // BUG
```

Go would create a **new** `source` that exists only inside the closure and *shadows* the outer
one. The code compiles, the transfer succeeds, and the API returns an account where every field
is zero. Shadowing bugs like this are easy to miss.

That's why the closure declares `destination` and `err` up front with `var` and then uses plain
`=` for all three. Your `Deposit` method uses the same pattern (`var err error` and then
`account, err = ...`).

> 🧪 **Try it:** after Step 8 passes, change `=` to `:=` and run `go vet ./...`. vet will probably
> say nothing. Then call the endpoint and look at the response. Change it back afterwards.

---

## Step 6 — Update the handler

In `internals/account/handler.go`, inside `Transfer`:

1. Call the financial service and keep the returned account:

```go
// before
err = h.service.Transfer(r.Context(), accountID, request)

// after
account, err := h.financialService.Transfer(r.Context(), accountID, request)
```

> Why `:=` here when `err` already exists? `:=` works as long as **at least one** variable on the
> left is new. `account` is new, so Go declares `account` and reuses the existing `err`. Both are
> in the same scope here, so nothing is shadowed.

2. Add a case for the currency error, which the old handler never handled. Put it in the
   `switch` next to `ErrSameAccount`:

```go
case errors.Is(err, ErrDifferentCurrency):
	response.ErrorJSON(w, http.StatusBadRequest, "DIFFERENT_CURRENCY", "Source and destination accounts must have the same currency")
```

3. Return the account instead of a plain message:

```go
// before
response.JSON(w, http.StatusOK, map[string]string{"message": "Transfer successful"})

// after
response.JSON(w, http.StatusOK, account)
```

4. Update the swagger comment above the function to match:

```go
// @Success 200 {object} Account
```

Run:

```bash
go build ./...
go vet ./...
```

Both should pass now.

---

## Step 7 — Remove `Transfer` from the old service's dependencies (check)

`Service` no longer needs anything related to transfers. Search to confirm nothing still calls the
old method:

```bash
grep -rn "service.Transfer\|\.Transfer(ctx, from" internals/
```

The only matches should be `financialService.Transfer` in the handler and `accountRepository.Transfer`
inside `FinancialService`.

Regenerate the swagger docs so `/swagger` shows the new response type:

```bash
swag init
```

---

## Step 8 — Test it manually

Start the app (`air` or `go run .`). Log in to get a token, and make sure you have **two active
accounts** with some balance in the first one (deposit into it first).

Add this to `api.http`:

```http
###
# transfer between accounts
POST {{baseUrl}}/accounts/1/transfer
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "destination_account_id": 2,
  "amount": 10000
}
```

### Expected

- **200** with the source account JSON, and its `balance` reduced by 10000.

Then check the ledger in psql or the IDE database console:

```sql
SELECT id, account_id, type, amount, balance_after, reference
FROM transactions
ORDER BY id DESC
LIMIT 2;
```

You should see two rows whose references share the same UUID: one ending in `-DR` on account 1
with `TRANSFER_DEBIT`, and one ending in `-CR` on account 2 with `TRANSFER_CREDIT`.

### Error cases to try

| Request | Expected |
|---|---|
| `destination_account_id` = same as path id | 400 `SAME_ACCOUNT` |
| `amount: 0` | 400 `INVALID_AMOUNT` |
| `amount` larger than balance | 400 `INSUFFICIENT_BALANCE` |
| `destination_account_id: 99999` | 404 `ACCOUNT_NOT_FOUND` |
| Set account 2's status to `CLOSED` in the DB, then transfer | 403 `ACCOUNT_CLOSED` (the old code said `ACCOUNT_BLOCKED`) |

---

## Step 9 — Prove atomicity

This experiment shows why the transaction matters.

1. Note the balances of accounts 1 and 2.
2. In `FinancialService.Transfer`, temporarily change the credit row's reference to
   `reference + "-DR"`, so it collides with the debit row's reference.
3. Run the transfer again.

The second `INSERT` breaks the `UNIQUE` constraint on `reference`, the closure returns an error,
and `WithTx` **rolls back everything**, including the two `UPDATE accounts` statements. You get a
500, and both balances in the DB are unchanged.

Undo the change when you're done.

> This is the main point of the refactor. In the old code the balance updates committed on their
> own, so a later ledger failure would leave the money moved with no record of it.

---

## Commit

```bash
git add internals/ docs/ api.http
git commit -m "move transfer into financial service and record ledger entries"
```

---

## Things I noticed while reading the code (fix them yourself as practice)

You asked me not to fix anything, so here are the problems as exercises. Each one is small.

1. **`financial_service.go` → `Withdraw`** records `Type: transaction.TypeDeposit`. It should be
   `transaction.TypeWithdrawal`. Your ledger currently labels withdrawals as deposits.
2. **`handler.go` → `Withdraw`**: the `if err != nil { switch ... }` block has no `return`, so after
   writing the error it also writes `response.JSON(w, 200, account)`. Add `return` after the
   switch, as `Deposit` does.
3. **`handler.go` → `Deposit`**: the invalid-id check is missing a `return`.
4. **`transaction/postgres_repository.go` → `GetByAccountId`**: the query says `WHERE accounts_id`.
   The column is `account_id`. Nothing calls this function yet, which is why you haven't seen the
   error. Tutorial 02 will use it.
5. **`handler.go` → `Create`** calls `w.WriteHeader(http.StatusCreated)` and then
   `response.JSON(...)`, which calls `WriteHeader` again. Go logs
   `http: superfluous response.WriteHeader call`. Remove the two manual header lines, because
   `response.JSON` already sets them. `GetByCustomerId` and `GetById` have the same redundant
   `Content-Type` line.
6. **Security, important:** Deposit, Withdraw and Transfer never check that the account in the URL
   belongs to the logged-in customer (`auth.CustomerIdKey`). Any logged-in user can currently
   transfer money **out of anyone's account**. That's the topic for a later tutorial.

---

## What's next

- **Tutorial 02 — Transaction history endpoint:** `GET /accounts/{id}/transactions` using
  `transactionRepository.GetByAccountId` (after fixing the typo above), with an ownership check.
- **Tutorial 03 — Ownership checks on money movement:** only the account owner can deposit,
  withdraw or transfer out of an account.
