package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/LythianOlyx/Zyra/internal/data/nplusone"
)

// Tx wraps sqlx.Tx with Zyra context tracking (N+1 query detection, tenant awareness).
type Tx struct {
	*sqlx.Tx
	ctx context.Context
}

// ExecContext executes a query within the transaction, logging for N+1 detection in dev mode.
func (tx *Tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if tracker, ok := nplusone.FromContext(ctx); ok {
		tracker.TrackQuery(query)
	}
	return tx.Tx.ExecContext(ctx, query, args...)
}

// GetContext fetches a single row into dest within the transaction.
func (tx *Tx) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	if tracker, ok := nplusone.FromContext(ctx); ok {
		tracker.TrackQuery(query)
	}
	return tx.Tx.GetContext(ctx, dest, query, args...)
}

// SelectContext fetches multiple rows into dest within the transaction.
func (tx *Tx) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	if tracker, ok := nplusone.FromContext(ctx); ok {
		tracker.TrackQuery(query)
	}
	return tx.Tx.SelectContext(ctx, dest, query, args...)
}

// Context returns the transaction's context.
func (tx *Tx) Context() context.Context {
	return tx.ctx
}

// Transaction executes fn inside a database transaction.
// Automatically commits on success, and rolls back on error or panic.
func (db *DB) Transaction(ctx context.Context, fn func(tx *Tx) error) (err error) {
	sqlxTx, err := db.SqlxDB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("zyra/db: failed to begin transaction: %w", err)
	}

	tx := &Tx{
		Tx:  sqlxTx,
		ctx: ctx,
	}

	defer func() {
		if p := recover(); p != nil {
			_ = sqlxTx.Rollback()
			panic(p) // re-throw panic after rollback
		} else if err != nil {
			_ = sqlxTx.Rollback()
		} else {
			if commitErr := sqlxTx.Commit(); commitErr != nil {
				err = fmt.Errorf("zyra/db: failed to commit transaction: %w", commitErr)
			}
		}
	}()

	err = fn(tx)
	return err
}
