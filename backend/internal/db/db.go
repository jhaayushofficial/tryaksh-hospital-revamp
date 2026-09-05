package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX is the interface that groups methods required to execute database operations.
// It is implemented by *pgxpool.Pool and pgx.Tx.
type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

func New(pool *pgxpool.Pool) *Queries {
	return &Queries{pool: pool}
}

type Queries struct {
	pool *pgxpool.Pool
}

// WithTx returns a new Queries instance that uses the provided transaction instead of the pool.
func (q *Queries) WithTx(tx pgx.Tx) *QueriesTx {
	return &QueriesTx{
		tx: tx,
	}
}

type QueriesTx struct {
	tx pgx.Tx
}

func (q *Queries) db() DBTX {
	return q.pool
}

func (q *QueriesTx) db() DBTX {
	return q.tx
}
