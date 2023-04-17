package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/liorlavon/simplebank/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) Account {
	// create Owner
	u := createRandomUser(t)

	arg := CreateAccountParams{
		Owner:    u.Username,
		Balance:  100,
		Currency: util.RandomCurrency(),
	}

	account, err := testQueriers.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	// check all values compare to input
	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)

	// check that ID is generated by PG
	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}

func deleteRandomAccount(acc Account) {
	testQueriers.DeleteAccount(context.Background(), acc.ID)
	testQueriers.DeleteUser(context.Background(), acc.Owner)

}

func TestCreateAccount(t *testing.T) {
	ac := createRandomAccount(t)
	// clear tests objects
	deleteRandomAccount(ac)
}

func TestGetAccount(t *testing.T) {
	acc1 := createRandomAccount(t)

	acc2, err := testQueriers.GetAccount(context.Background(), acc1.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, acc2)
	assert.NotZero(t, acc2.ID)

	assert.Equal(t, acc1.ID, acc2.ID)
	assert.Equal(t, acc1.Owner, acc2.Owner)
	assert.Equal(t, acc1.Balance, acc2.Balance)
	assert.Equal(t, acc1.Currency, acc2.Currency)
	assert.WithinDuration(t, acc1.CreatedAt, acc2.CreatedAt, time.Second)

	// clear tests objects
	deleteRandomAccount(acc2)
}

func TestListAccounts(t *testing.T) {
	var createdAccounts []Account
	for i := 0; i < 10; i++ {
		a := createRandomAccount(t)
		createdAccounts = append(createdAccounts, a)
	}

	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	accounts, err := testQueriers.ListAccounts(context.Background(), arg)
	assert.NoError(t, err)
	assert.Len(t, accounts, 5)

	for _, account := range accounts {
		assert.NotEmpty(t, account)
	}

	// cleare account
	for _, account := range createdAccounts {
		deleteRandomAccount(account)
	}

}

func TestUpdateAccount(t *testing.T) {
	// create random account
	acc1 := createRandomAccount(t)

	// update account
	arg := UpdateAccountParams{
		ID:      acc1.ID,
		Balance: util.RandomMoney(),
	}

	acc2, err := testQueriers.UpdateAccount(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, acc2)

	assert.Equal(t, acc1.ID, acc2.ID)
	assert.Equal(t, acc1.Owner, acc2.Owner)
	assert.Equal(t, arg.Balance, acc2.Balance)
	assert.Equal(t, acc1.Currency, acc2.Currency)
	assert.WithinDuration(t, acc1.CreatedAt, acc2.CreatedAt, time.Second)

	// clear tests objects
	deleteRandomAccount(acc2)
}

func TestDeleteAccount(t *testing.T) {
	a := createRandomAccount(t)

	err := testQueriers.DeleteAccount(context.Background(), a.ID)
	assert.NoError(t, err)
	testQueriers.DeleteUser(context.Background(), a.Owner)

	// try to get account
	acc, err := testQueriers.GetAccount(context.Background(), a.ID)
	assert.Error(t, err)
	assert.EqualError(t, err, sql.ErrNoRows.Error())
	assert.Empty(t, acc)

}
