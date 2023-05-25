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

func createRandomUser(t *testing.T) User {

	hp, err := util.HashPassword("secret")
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       util.RandomUser(),
		HashedPassword: hp,
		Firstname:      util.RandomUser(),
		Lastname:       util.RandomUser(),
		Email:          util.RandEmail(),
	}

	user, err := testQueriers.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.Firstname, user.Firstname)
	require.Equal(t, arg.Lastname, user.Lastname)
	require.Equal(t, arg.Email, user.Email)

	// PasswordChangedAt is created with default 0 values , so IsZero should be true
	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	// create user
	u := createRandomUser(t)

	// get user
	user, err := testQueriers.GetUser(context.Background(), u.Username)

	// compare result
	assert.NoError(t, err)
	assert.NotEmpty(t, user)

	// check values
	assert.Equal(t, u.Username, user.Username)
	assert.Equal(t, u.HashedPassword, user.HashedPassword)
	assert.Equal(t, u.Firstname, user.Firstname)
	assert.Equal(t, u.Lastname, user.Lastname)
	assert.Equal(t, u.Email, user.Email)
	assert.WithinDuration(t, u.PasswordChangedAt, user.PasswordChangedAt, time.Second)
	assert.WithinDuration(t, u.CreatedAt, user.CreatedAt, time.Second)
}

func TestUpdateUserFirstName(t *testing.T) {
	oldUser := createRandomUser(t)

	newFirstName := util.RandomUser()

	updatedUser, err := testQueriers.UpdateUser(context.Background(), UpdateUserParams{
		Username:  oldUser.Username,
		Firstname: sql.NullString{String: newFirstName, Valid: true},
	})

	require.NoError(t, err)

	// check value change
	require.NotEqual(t, updatedUser.Firstname, oldUser.Firstname)
	require.Equal(t, newFirstName, updatedUser.Firstname)

	// hceck no change
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.Lastname, updatedUser.Lastname)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.WithinDuration(t, oldUser.PasswordChangedAt, updatedUser.PasswordChangedAt, time.Second)
	require.WithinDuration(t, oldUser.CreatedAt, updatedUser.CreatedAt, time.Second)
}

func TestUpdateUserEmail(t *testing.T) {
	oldUser := createRandomUser(t)

	newEmail := util.RandEmail()

	updatedUser, err := testQueriers.UpdateUser(context.Background(), UpdateUserParams{
		Username: oldUser.Username,
		Email:    sql.NullString{String: newEmail, Valid: true},
	})

	require.NoError(t, err)

	// check value change
	require.NotEqual(t, updatedUser.Email, oldUser.Email)
	require.Equal(t, newEmail, updatedUser.Email)

	// hceck no change
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.Firstname, updatedUser.Firstname)
	require.Equal(t, oldUser.Lastname, updatedUser.Lastname)
	require.WithinDuration(t, oldUser.PasswordChangedAt, updatedUser.PasswordChangedAt, time.Second)
	require.WithinDuration(t, oldUser.CreatedAt, updatedUser.CreatedAt, time.Second)
}

func TestUpdateUserPassword(t *testing.T) {
	oldUser := createRandomUser(t)

	newHashedPassword, err := util.HashPassword("new_secret")
	require.NoError(t, err)

	updatedUser, err := testQueriers.UpdateUser(context.Background(), UpdateUserParams{
		Username:       oldUser.Username,
		HashedPassword: sql.NullString{String: newHashedPassword, Valid: true},
	})

	require.NoError(t, err)

	// check value change
	require.NotEqual(t, updatedUser.HashedPassword, oldUser.HashedPassword)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)

	// hceck no change
	require.Equal(t, oldUser.Firstname, updatedUser.Firstname)
	require.Equal(t, oldUser.Lastname, updatedUser.Lastname)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.WithinDuration(t, oldUser.PasswordChangedAt, updatedUser.PasswordChangedAt, time.Second)
	require.WithinDuration(t, oldUser.CreatedAt, updatedUser.CreatedAt, time.Second)
}

func TestUpdateUserAllFields(t *testing.T) {
	oldUser := createRandomUser(t)

	newFirstName := util.RandomUser()
	newEmail := util.RandEmail()
	newHashedPassword, err := util.HashPassword("new_secret")
	require.NoError(t, err)

	updatedUser, err := testQueriers.UpdateUser(context.Background(), UpdateUserParams{
		Username:       oldUser.Username,
		Firstname:      sql.NullString{String: newFirstName, Valid: true},
		Email:          sql.NullString{String: newEmail, Valid: true},
		HashedPassword: sql.NullString{String: newHashedPassword, Valid: true},
	})

	require.NoError(t, err)

	// check value change
	require.NotEqual(t, updatedUser.Firstname, oldUser.Firstname)
	require.Equal(t, newFirstName, updatedUser.Firstname)

	// check value change
	require.NotEqual(t, updatedUser.Email, oldUser.Email)
	require.Equal(t, newEmail, updatedUser.Email)

	// check value change
	require.NotEqual(t, updatedUser.HashedPassword, oldUser.HashedPassword)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)

	// hceck no change
	require.Equal(t, oldUser.Lastname, updatedUser.Lastname)
	require.WithinDuration(t, oldUser.PasswordChangedAt, updatedUser.PasswordChangedAt, time.Second)
	require.WithinDuration(t, oldUser.CreatedAt, updatedUser.CreatedAt, time.Second)
}

func TestDeleteUser(t *testing.T) {
	u := createRandomUser(t)

	err := testQueriers.DeleteUser(context.Background(), u.Username)
	assert.NoError(t, err)

	// check that the user is deleted
	user, err := testQueriers.GetUser(context.Background(), u.Username)
	assert.Error(t, err)
	assert.EqualError(t, err, sql.ErrNoRows.Error()) // check the specific error
	assert.Empty(t, user)
}

func TestListUsers(t *testing.T) {

	var createdUsers []User
	// create random users
	for i := 0; i < 10; i++ {
		u := createRandomUser(t)
		createdUsers = append(createdUsers, u)
	}

	arg := ListUsersParams{
		Limit:  5,
		Offset: 5,
	}

	users, err := testQueriers.ListUsers(context.Background(), arg)
	assert.NoError(t, err)
	assert.Len(t, users, 5)

	for _, o := range users {
		assert.NotEmpty(t, o)
	}

}
