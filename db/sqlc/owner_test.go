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

func createRandomOwner(t *testing.T) Owner {
	arg := CreateOwnerParams{
		Firstname: util.RandomOwner(),
		Lastname:  util.RandomOwner(),
		Email:     util.RandEmail(),
	}

	owner, err := testQueriers.CreateOwner(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, owner)

	require.Equal(t, arg.Firstname, owner.Firstname)
	require.Equal(t, arg.Lastname, owner.Lastname)
	require.Equal(t, arg.Email, owner.Email)

	require.NotZero(t, owner.ID)
	require.NotZero(t, owner.CreatedAt)

	return owner
}

func TestCreateOwner(t *testing.T) {
	o := createRandomOwner(t)
	testQueriers.DeleteOwner(context.Background(), o.ID)
}

func TestGetOwner(t *testing.T) {
	// create owner
	o := createRandomOwner(t)

	// get owner
	owner, err := testQueriers.GetOwner(context.Background(), o.ID)

	// compare result
	assert.NoError(t, err)
	assert.NotEmpty(t, owner)

	// check values
	assert.Equal(t, o.ID, owner.ID)
	assert.Equal(t, o.Firstname, owner.Firstname)
	assert.Equal(t, o.Lastname, owner.Lastname)
	assert.Equal(t, o.Email, owner.Email)
	assert.WithinDuration(t, o.CreatedAt.Time, owner.CreatedAt.Time, time.Second)

	testQueriers.DeleteOwner(context.Background(), o.ID)
}

func TestUpdateOwner(t *testing.T) {
	o := createRandomOwner(t)

	arg := UpdateOwnerParams{
		ID:        o.ID,
		Firstname: util.RandomOwner(),
		Lastname:  util.RandomOwner(),
		Email:     util.RandEmail(),
	}

	owner, err := testQueriers.UpdateOwner(context.Background(), arg)
	assert.NoError(t, err)
	assert.NotEmpty(t, owner)

	// check values
	assert.Equal(t, arg.ID, owner.ID)
	assert.Equal(t, arg.Firstname, owner.Firstname)
	assert.Equal(t, arg.Lastname, owner.Lastname)
	assert.Equal(t, arg.Email, owner.Email)
	assert.WithinDuration(t, o.CreatedAt.Time, owner.CreatedAt.Time, time.Second)

	testQueriers.DeleteOwner(context.Background(), o.ID)
}

func TestDeleteOwner(t *testing.T) {
	o := createRandomOwner(t)

	err := testQueriers.DeleteOwner(context.Background(), o.ID)
	assert.NoError(t, err)

	// check that the owner is deleted
	owner, err := testQueriers.GetOwner(context.Background(), o.ID)
	assert.Error(t, err)
	assert.EqualError(t, err, sql.ErrNoRows.Error()) // check the specific error
	assert.Empty(t, owner)
}

func TestListOwners(t *testing.T) {

	var createdOwners []Owner
	// create random owners
	for i := 0; i < 10; i++ {
		o := createRandomOwner(t)
		createdOwners = append(createdOwners, o)
	}

	arg := ListOwnersParams{
		Limit:  5,
		Offset: 5,
	}

	owners, err := testQueriers.ListOwners(context.Background(), arg)
	assert.NoError(t, err)
	assert.Len(t, owners, 5)

	for _, o := range owners {
		assert.NotEmpty(t, o)
	}

	for _, ow := range createdOwners {
		testQueriers.DeleteOwner(context.Background(), ow.ID)
	}

}
