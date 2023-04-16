package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmailValidation(t *testing.T) {

	// define a list of test cases
	testCases := []struct {
		name     string
		email    string
		validate func(t *testing.T, email string)
	}{
		{
			name:  "Valid",
			email: "lior.lavon@gmail.com",
			validate: func(t *testing.T, email string) {
				// check statusCode response
				result := IsEmailValid(email)
				require.True(t, result)
			},
		},
		{
			name:  "NotValid",
			email: "lior.lavon.gmail.com",
			validate: func(t *testing.T, email string) {
				// check statusCode response
				result := IsEmailValid(email)
				require.False(t, result)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			email := tc.email
			tc.validate(t, email)
		})
	}
}

func TestCurrencyValidation(t *testing.T) {
	// define a list of test cases
	testCases := []struct {
		name     string
		currency string
		validate func(t *testing.T, email string)
	}{
		{
			name:     "Valid_USD",
			currency: "USD",
			validate: func(t *testing.T, email string) {
				// check statusCode response
				result := IsSuportedCurrency(email)
				require.True(t, result)
			},
		},
		{
			name:     "Valid_EUR",
			currency: "EUR",
			validate: func(t *testing.T, email string) {
				// check statusCode response
				result := IsSuportedCurrency(email)
				require.True(t, result)
			},
		},
		{
			name:     "Valid_CAD",
			currency: "CAD",
			validate: func(t *testing.T, email string) {
				// check statusCode response
				result := IsSuportedCurrency(email)
				require.True(t, result)
			},
		},
		{
			name:     "NotValid_ISL",
			currency: "ISL",
			validate: func(t *testing.T, email string) {
				// check statusCode response
				result := IsSuportedCurrency(email)
				require.False(t, result)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			currency := tc.currency
			tc.validate(t, currency)
		})
	}

}
