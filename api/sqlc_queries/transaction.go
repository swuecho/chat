package sqlc_queries

import (
	"context"
	"database/sql"
	"fmt"
)

type transactionBeginner interface {
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

// InTransaction runs fn with queries bound to one database transaction.
func (q *Queries) InTransaction(ctx context.Context, fn func(*Queries) error) error {
	beginner, ok := q.db.(transactionBeginner)
	if !ok {
		return fmt.Errorf("query connection does not support transactions")
	}
	tx, err := beginner.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(q.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit()
}
