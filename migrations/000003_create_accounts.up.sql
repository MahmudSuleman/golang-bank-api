CREATE TABLE accounts
(
    id             BIGSERIAL PRIMARY KEY,

    customer_id    BIGINT      NOT NULL,

    account_number VARCHAR(20) NOT NULL UNIQUE,

    account_type   VARCHAR(20) NOT NULL,

    currency       VARCHAR(3)  NOT NULL DEFAULT 'GHS',

    balance        BIGINT      NOT NULL DEFAULT 0,

    status         VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',

    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_accounts_customer
        FOREIGN KEY (customer_id)
            REFERENCES customers (id)
            ON DELETE RESTRICT
);
