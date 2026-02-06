package tx

import (
	"context"
	"database/sql"
)

type Tx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Manager interface {
	WithTx(ctx context.Context, fn func(tx Tx) error) error
}
