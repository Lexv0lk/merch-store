-- +goose Up
-- +goose StatementBegin
CREATE TABLE balances (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) UNIQUE,
    balance INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE goods (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    price INTEGER NOT NULL
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    from_user_id INTEGER REFERENCES users(id),
    to_user_id INTEGER REFERENCES users(id),
    amount INTEGER NOT NULL
);

CREATE TABLE purchases (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
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
