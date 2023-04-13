package db

import (
	"context"
	"database/sql"
	"fmt"
)

// store provides all functions to ensure DB queries and transaction
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
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

// TransferTxParams contains all neceery input parameters to transfare money between two accounts
type TransferTxParams struct {
	FromAccountId int64 `json:"from_account_id"`
	ToAccountId   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult contains the result of the transfer transaction
type TransferTxResult struct {
	Transfare   Transfer `json:"trasfer"`
	FromAccount Account  `json:"from_account"` // after the blance updated
	ToAccount   Account  `json:"to_account"`   // after the blance updated
	FromEntry   Entry    `json:"from_entry"`   // the entry from the money is moving out
	ToEntry     Entry    `json:""to_entry`     // the entry from the money is moving in
}

// ckeare an empty struct for the key/value pairs
var txKey = struct{}{}

// Transaction performs a money transfare from one account to another account
// - create a transfer record
// - add two ccount entries
// - update account balance
// with in a single DB transaction
func (store *SQL_Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		// implement the body of the transaction
		var err error

		// read the name from the context assigned in the _test
		//txName := ctx.Value(txKey)

		//fmt.Println(txName, "CreateTransfer")
		result.Transfare, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountId,
			ToAccountID:   arg.ToAccountId,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		//fmt.Println(txName, "CreateEntry From")
		// Create Entry for Account 1
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountId,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		// fmt.Println(txName, "CreateEntry To")
		// Create Entry for Account 2
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountId,
			Amount:    +arg.Amount,
		})
		if err != nil {
			return err
		}

		if arg.FromAccountId < arg.ToAccountId {

			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountId, -arg.Amount, arg.ToAccountId, arg.Amount)
			if err != nil {
				return err
			}

		} else {

			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountId, arg.Amount, arg.FromAccountId, -arg.Amount)
			if err != nil {
				return err
			}

		}

		return nil
	})

	return result, err
}

func addMoney(ctx context.Context, q *Queries, accountId1 int64, amount1 int64, accountId2 int64, amount2 int64) (account1 Account, account2 Account, err error) {

	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		Amount: amount1,
		ID:     accountId1,
	})
	if err != nil {
		return // we can just return here as we defined the valiables in the return statement
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		Amount: amount2,
		ID:     accountId2,
	})
	return // we can just return here as we defined the valiables in the return statement
}
