CREATE TABLE transactions
(
    id            BIGSERIAL PRIMARY KEY,

    account_id    BIGINT       NOT NULL,

    type          VARCHAR(20)  NOT NULL,

    amount        BIGINT       NOT NULL,

    balance_after BIGINT       NOT NULL,

    reference     VARCHAR(100) NOT NULL UNIQUE,

    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_transactions_account
        FOREIGN KEY (account_id)
            REFERENCES accounts (id)
            ON DELETE RESTRICT,

    CONSTRAINT chk_transactions_amount
        CHECK (amount > 0),

    CONSTRAINT chk_transactions_type
        CHECK (
            type IN (
                     'DEPOSIT',
                     'WITHDRAWAL',
                     'TRANSFER_DEBIT',
                     'TRANSFER_CREDIT'
                )
            )
);