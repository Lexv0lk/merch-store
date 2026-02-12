-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_transactions_from_user_id ON transactions(from_user_id);
CREATE INDEX idx_transactions_to_user_id ON transactions(to_user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_transactions_from_user_id;
DROP INDEX IF EXISTS idx_transactions_to_user_id;
-- +goose StatementEnd
