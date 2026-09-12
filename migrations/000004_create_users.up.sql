CREATE TABLE users
(
    id            BIGSERIAL PRIMARY KEY,

    customer_id   BIGINT       NOT NULL UNIQUE,

    email         VARCHAR(255) NOT NULL UNIQUE,

    password_hash TEXT         NOT NULL,

    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_users_customer
        FOREIGN KEY (customer_id)
            REFERENCES customers (id)
            ON DELETE RESTRICT
);