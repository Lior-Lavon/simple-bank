package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordSuccess(t *testing.T) {
	password := RandomString(6)

	hashPassword, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashPassword)

	err = CheckPassword(password, hashPassword)
	require.NoError(t, err)
}

func TestPasswordFailed(t *testing.T) {
	password1 := RandomString(6)
	password2 := RandomString(6)

	hashPassword, err := HashPassword(password1)
	require.NoError(t, err)
	require.NotEmpty(t, hashPassword)

	err = CheckPassword(password2, hashPassword)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
}

func TestHasedTwicedProducedTowDifferentValues(t *testing.T) {
	password := RandomString(6)

	hash1, err := HashPassword(password)
	require.NoError(t, err)

	hash2, err := HashPassword(password)
	require.NoError(t, err)

	require.NotEqual(t, hash1, hash2)
}
