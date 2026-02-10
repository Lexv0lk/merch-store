-- +goose Up
-- +goose StatementBegin
CREATE TABLE balances (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE,
    balance INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE goods (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    price INTEGER NOT NULL
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    from_user_id INTEGER REFERENCES balances(user_id),
    to_user_id INTEGER REFERENCES balances(user_id),
    amount INTEGER NOT NULL
);

CREATE TABLE purchases (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES balances(user_id),
    good_id INTEGER REFERENCES goods(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS purchases;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS goods;
DROP TABLE IF EXISTS balances;
-- +goose StatementEnd
