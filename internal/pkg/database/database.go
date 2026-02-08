package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type QueryTxBeginner interface {
	Querier
	TxBeginner
}

type QueryExecuter interface {
	Querier
	Executor
}

// Querier defines an interface for executing SQL queries.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// Executor defines an interface for executing SQL commands.
type Executor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// TxBeginner defines an interface for beginning database transactions.
type TxBeginner interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

// PostgresSettings contains configuration parameters for PostgreSQL database connection.
type PostgresSettings struct {
	User       string
	Password   string
	Host       string
	Port       string
	DBName     string
	SSlEnabled bool
}

// GetUrl constructs and returns a PostgreSQL connection URL string.
// The URL format is: postgres://user:password@host:port/dbname
// If SSlEnabled is false, "?sslmode=disable" is appended to the URL.
// This method does not return any errors; it assumes all fields are properly set.
func (p *PostgresSettings) GetUrl() string {
	result := "postgres://" + p.User + ":" + p.Password + "@" + p.Host + ":" + p.Port + "/" + p.DBName
	if p.SSlEnabled == false {
		result += "?sslmode=disable"
	}

	return result
}
