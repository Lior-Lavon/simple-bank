package db

import (
	"context"
	"testing"
	"time"

	"github.com/liorlavon/simplebank/util"
	"github.com/stretchr/testify/assert"
)

func createRandomEntry(t *testing.T) Entry {

	o := createRandomOwner(t)

	argAcc := CreateAccountParams{
		OwnerID:  o.ID,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
	a, err := testQueriers.CreateAccount(context.Background(), argAcc)
	assert.NoError(t, err)

	argEntry := CreateEntryParams{
		AccountID: a.ID,
		Amount:    util.RandomMoney(),
	}

	entry, err := testQueriers.CreateEntry(context.Background(), argEntry)
	assert.NoError(t, err)
	assert.NotEmpty(t, entry)

	assert.NotZero(t, entry.ID)
	assert.Equal(t, a.ID, entry.AccountID)
	assert.Equal(t, argEntry.Amount, entry.Amount)
	assert.NotZero(t, entry.CreatedAt)

	return entry
}

func deleteRandomEntry(t *testing.T, entry Entry) {
	testQueriers.DeleteEntry(context.Background(), entry.ID)

	// get account
	ac, err := testQueriers.GetAccount(context.Background(), entry.AccountID)
	assert.NoError(t, err)

	testQueriers.DeleteAccount(context.Background(), ac.ID)
	testQueriers.DeleteOwner(context.Background(), ac.OwnerID)
}

func TestCreateEntry(t *testing.T) {
	entry := createRandomEntry(t)
	deleteRandomEntry(t, entry)
}

func TestGetEntry(t *testing.T) {

	// create test entry
	entry1 := createRandomEntry(t)

	// get the event
	entry2, err := testQueriers.GetEntry(context.Background(), entry1.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, entry2)

	assert.NotZero(t, entry2.ID)
	assert.Equal(t, entry1.AccountID, entry2.AccountID)
	assert.Equal(t, entry1.Amount, entry2.Amount)
	assert.WithinDuration(t, entry1.CreatedAt, entry2.CreatedAt, time.Second)

	deleteRandomEntry(t, entry2)
}

func TestListEntries(t *testing.T) {
	en1 := createRandomEntry(t)

	var testEntries []Entry

	for i := 0; i < 9; i++ {
		arg := CreateEntryParams{
			AccountID: en1.AccountID,
			Amount:    util.RandomMoney(),
		}

		en, err := testQueriers.CreateEntry(context.Background(), arg)
		assert.NoError(t, err)
		testEntries = append(testEntries, en)
	}

	//	test
	arg := ListEntriesParams{
		AccountID: en1.AccountID,
		Limit:     5,
		Offset:    5,
	}

	entries, err := testQueriers.ListEntries(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, entries)

	assert.Len(t, entries, 5)
	for _, en := range entries {
		assert.NotEmpty(t, en)
	}

	for _, en := range testEntries {
		testQueriers.DeleteEntry(context.Background(), en.ID)
	}
	deleteRandomEntry(t, en1)
}

func TestUpdateEntry(t *testing.T) {

	// create test entity
	ent1 := createRandomEntry(t)

	arg := UpdateEntryParams{
		ID:     ent1.ID,
		Amount: util.RandomMoney(),
	}

	ent2, err := testQueriers.UpdateEntry(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, ent2)

	assert.Equal(t, ent1.ID, ent2.ID)
	assert.Equal(t, ent1.AccountID, ent2.AccountID)
	assert.Equal(t, arg.Amount, ent2.Amount)
	assert.WithinDuration(t, ent1.CreatedAt, ent2.CreatedAt, time.Second)

	// clear resources
	deleteRandomEntry(t, ent1)
}

func TestDeleteEntry(t *testing.T) {

	ent := createRandomEntry(t)
	err := testQueriers.DeleteEntry(context.Background(), ent.ID)
	assert.NoError(t, err)

	// clear resources
	deleteRandomEntry(t, ent)
}
