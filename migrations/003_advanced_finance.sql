-- +goose Up
-- +goose StatementBegin

CREATE TABLE budgets (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_name VARCHAR(100) NOT NULL,
    monthly_limit BIGINT NOT NULL CHECK (monthly_limit > 0),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, category_name)
);

CREATE INDEX idx_budgets_user_id ON budgets(user_id);

CREATE TABLE wallets (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);

ALTER TABLE users ADD COLUMN active_wallet_id UUID REFERENCES wallets(id) ON DELETE SET NULL;

ALTER TABLE transactions ADD COLUMN wallet_id UUID REFERENCES wallets(id) ON DELETE SET NULL;

CREATE INDEX idx_transactions_wallet ON transactions(wallet_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_transactions_wallet;
ALTER TABLE transactions DROP COLUMN IF EXISTS wallet_id;
ALTER TABLE users DROP COLUMN IF EXISTS active_wallet_id;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS budgets;

-- +goose StatementEnd
