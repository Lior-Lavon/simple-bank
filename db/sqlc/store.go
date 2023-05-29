package db

import (
	"context"
	"database/sql"
	"fmt"
)

// store interface provides all functions to ensure DB queries and transaction
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error)
}

// REAL store provides all functions to ensure DB queries and transaction
type SQL_Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQL_Store{
		db:      db,
		Queries: New(db),
	}
}

// Create and run a new database transaction
// execTx execute a function within a database transaction
func (store *SQL_Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return nil
	}

	q := New(tx) // get a new Queries object
	err = fn(q)
	if err != nil {
		// role back the tx
		rbErr := tx.Rollback()
		if rbErr != nil {
			fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
