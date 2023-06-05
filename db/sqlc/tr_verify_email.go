package db

import (
	"context"
	"database/sql"
)

// create a transaction to create a new user

type VerifyEmailTxParams struct {
	EmailId    int64
	SecretCode string
}

type VerifyEmailTxResult struct {
	User        User        // updated user
	VerifyEmail VerifyEmail // updated email record
}

func (store *SQL_Store) VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error) {
	var result VerifyEmailTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		// implement the body of the transaction
		var err error

		// get the verification email by (email_id && secret_code)
		// and update its is_used filed to true,
		verifyEmail, err := q.UpdateVerifyEmail(ctx, UpdateVerifyEmailParams{
			ID:         arg.EmailId,
			SecretCode: arg.SecretCode,
		})
		if err != nil {
			return err
		}

		// update the isEmailVerified filed in User table based on email to true
		user, err := q.UpdateUser(ctx, UpdateUserParams{
			Username:        verifyEmail.Username,
			IsEmailVerified: sql.NullBool{Bool: true, Valid: true},
		})
		if err != nil {
			return err
		}

		result.VerifyEmail = verifyEmail
		result.User = user

		// execute the callbacl function to send the producer message to Radis
		return err
	})

	return result, err
}
