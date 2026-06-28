package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TxRunner runs a callback inside a transaction with a transactional Querier.
type TxRunner interface {
	Run(ctx context.Context, fn func(q Querier) error) error
}

// PoolTxRunner implements TxRunner using a connection pool and queries.
type PoolTxRunner struct {
	pool    *pgxpool.Pool
	queries *Queries
}

// NewPoolTxRunner returns a TxRunner that uses the given pool and queries.
func NewPoolTxRunner(pool *pgxpool.Pool, queries *Queries) *PoolTxRunner {
	return &PoolTxRunner{pool: pool, queries: queries}
}

// Run executes fn in a transaction and commits when fn returns nil.
func (r *PoolTxRunner) Run(ctx context.Context, fn func(q Querier) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qTx := r.queries.WithTx(tx)
	if err := fn(qTx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
