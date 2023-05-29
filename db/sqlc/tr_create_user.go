package db

import "context"

// create a transaction to create a new user

type CreateUserTxParams struct {
	CreateUserParams                       // param to create a user
	AfterCreate      func(user User) error // Callback , this function will be executed after the user is inserted
	// we will use the error to decide if the transaction should be rooled back

}

type CreateUserTxResult struct {
	User User // the created user record
}

func (store *SQL_Store) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {
	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		// implement the body of the transaction
		var err error

		user, err := q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err // roleback
		}

		result.User = user

		// execute the callbacl function to send the producer message to Radis
		return arg.AfterCreate(result.User)
	})

	return result, err
}
