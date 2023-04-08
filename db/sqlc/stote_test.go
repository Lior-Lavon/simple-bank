package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {

	store := NewStore(testDB)

	// create two random accounts
	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)
	//fmt.Println(">> before : ", acc1.Balance, acc2.Balance)

	// to test the concurency, run in several go routies
	n := 5              // 5 transactions
	amount := int64(10) //each will transfer an amount of 10

	// channel is designed to connect cocurent go routiens, and comunication
	// use two chanels to check for err or TransferTxResult result
	errChan := make(chan error)
	resultsChan := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		// start new bconcurent routien
		txName := fmt.Sprintf("tx %d", i+1)
		go func() {
			// add name to the context, and read it from the store.go
			ctx := context.WithValue(context.Background(), txKey, txName)
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountId: acc1.ID,
				ToAccountId:   acc2.ID,
				Amount:        amount,
			})

			// send result to the main process from inside the go routien
			errChan <- err
			resultsChan <- result
		}()
	}

	// check the errs and results

	for i := 0; i < n; i++ {
		// receive the err from the chanel
		err := <-errChan
		require.NoError(t, err)

		// receive the result from the channel
		result := <-resultsChan
		require.NotEmpty(t, result)

		// check the content of result
		transfer := result.Transfare
		require.NotEmpty(t, transfer)
		require.NotZero(t, transfer.ID)
		require.Equal(t, acc1.ID, transfer.FromAccountID)
		require.Equal(t, acc2.ID, transfer.ToAccountID)
		require.Equal(t, acc1.ID, transfer.FromAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.CreatedAt)

		// check that the record is in the database
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)
		// check the account entries of the result
		fromEntry := result.FromEntry
		//		fmt.Printf("fromEntry: %+v\n", fromEntry)
		require.NotEmpty(t, fromEntry)
		require.Equal(t, fromEntry.AccountID, acc1.ID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		// get the accountentry from the database to check that it got created
		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)

		require.Equal(t, acc2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		// get the accountentry from the database to check that it got created
		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check account1
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, acc1.ID, fromAccount.ID)
		require.Equal(t, acc1.Currency, fromAccount.Currency)
		require.NotZero(t, fromAccount.CreatedAt)

		// check account1
		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, acc2.ID, toAccount.ID)
		require.Equal(t, acc2.Currency, toAccount.Currency)
		require.NotZero(t, toAccount.CreatedAt)

		fmt.Println(">> tx : ", fromAccount.Balance, toAccount.Balance)

		diff1 := acc1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - acc2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff2 > 0)

		require.True(t, diff1%amount == 0)
	}

	// check the final updated amount of the two accounts
	updatedAccoun1, err := testQueriers.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)

	updatedAccoun2, err := testQueriers.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)

	fmt.Println(">> after : ", updatedAccoun1.Balance, updatedAccoun2.Balance)

	require.Equal(t, acc1.Balance-int64(n)*amount, updatedAccoun1.Balance)
	require.Equal(t, acc2.Balance+int64(n)*amount, updatedAccoun2.Balance)

}

func TestTransferTxDeadLock(t *testing.T) {

	store := NewStore(testDB)

	// create two random accounts
	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)
	fmt.Println(">> before : ", acc1.Balance, acc2.Balance)

	// to test the concurency, run in several go routies
	n := 10             // 5 transactions
	amount := int64(10) //each will transfer an amount of 10

	// channel is designed to connect cocurent go routiens, and comunication
	// use two chanels to check for err or TransferTxResult result
	errChan := make(chan error)

	for i := 0; i < n; i++ {

		fromAccountId := acc1.ID
		toAccountId := acc2.ID

		// start new bconcurent routien
		txName := fmt.Sprintf("tx %d", i+1)

		if i%2 == 1 {
			fromAccountId = acc2.ID
			toAccountId = acc1.ID
		}

		go func() {
			// add name to the context, and read it from the store.go
			ctx := context.WithValue(context.Background(), txKey, txName)
			_, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountId: fromAccountId,
				ToAccountId:   toAccountId,
				Amount:        amount,
			})

			// send result to the main process from inside the go routien
			errChan <- err
		}()
	}

	// check the errs and results

	for i := 0; i < n; i++ {
		// receive the err from the chanel
		err := <-errChan
		require.NoError(t, err)

	}

	// check the final updated amount of the two accounts
	updatedAccoun1, err := testQueriers.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)

	updatedAccoun2, err := testQueriers.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)

	fmt.Println(">> after : ", updatedAccoun1.Balance, updatedAccoun2.Balance)

	require.Equal(t, acc1.Balance, updatedAccoun1.Balance)
	require.Equal(t, acc2.Balance, updatedAccoun2.Balance)

}
