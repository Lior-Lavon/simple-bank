package gapi

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/liorlavon/simplebank/db/mock"
	db "github.com/liorlavon/simplebank/db/sqlc"
	pb "github.com/liorlavon/simplebank/pb"
	"github.com/liorlavon/simplebank/util"
	mockwk "github.com/liorlavon/simplebank/worker/mock"
	"github.com/stretchr/testify/require"
)

func randomUser() (db.User, string) {
	password := "secret"
	hp, _ := util.HashPassword(password)

	return db.User{
		Username:          "llavon",
		HashedPassword:    hp,
		Firstname:         "lior",
		Lastname:          "lavon",
		Email:             "lior.lavon@gmail.com",
		PasswordChangedAt: time.Time{},
		CreatedAt:         time.Time{},
	}, password
}

func TestCreateUser(t *testing.T) {

	user, rowPassword := randomUser()

	testCases := []struct {
		name          string
		req           *pb.CreateUserRequest
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, res *pb.CreateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateUserRequest{
				Username:  user.Username,
				Firstname: user.Firstname,
				Lastname:  user.Lastname,
				Email:     user.Email,
				Password:  rowPassword,
			},
			buildStub: func(store *mockdb.MockStore) {

				hp, _ := util.HashPassword(rowPassword)
				arg := db.CreateUserTxParams{
					CreateUserParams: db.CreateUserParams{
						Username:       user.Username,
						HashedPassword: hp,
						Firstname:      user.Firstname,
						Lastname:       user.Lastname,
						Email:          user.Email,
					},
				}

				store.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(arg, rowPassword)).
					Times(1).
					Return(db.CreateUserTxResult{User: user}, nil)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				createdUser := res.GetUser()
				require.Equal(t, user.Username, createdUser.Username)
				require.Equal(t, user.Firstname, createdUser.Firstname)
				require.Equal(t, user.Lastname, createdUser.Lastname)
				require.Equal(t, user.Email, createdUser.Email)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {

			// create new mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// create new mockDB store
			mStore := mockdb.NewMockStore(ctrl)

			// build stub
			tc.buildStub(mStore)

			taskDistributor := mockwk.NewMockTaskDistributor(ctrl)

			// create http server
			server := newTestServer(t, mStore, taskDistributor)

			res, err := server.CreateUser(context.Background(), tc.req)

			tc.checkResponse(t, res, err)
		})
	}
}

type eqCreateUserTxParamMatcher struct {
	arg      db.CreateUserTxParams
	password string // raw password
}

func (expected eqCreateUserTxParamMatcher) Matches(x interface{}) bool {

	// as x is an interface , convert x to db.CreateUserParams type
	actualArg, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}

	// check if the hashed password (x) match with the expected password or no (arg)
	err := util.CheckPassword(expected.password, actualArg.HashedPassword)
	if err != nil {
		return false
	}

	expected.arg.HashedPassword = actualArg.HashedPassword
	if !reflect.DeepEqual(expected.arg.CreateUserParams, actualArg.CreateUserParams) {
		return false
	}

	return true
}

func (e eqCreateUserTxParamMatcher) String() string {
	return fmt.Sprintf("match argX %v and password %v", e.arg, e.password)
}

func EqCreateUserTxParams(arg db.CreateUserTxParams, password string) gomock.Matcher {
	return eqCreateUserTxParamMatcher{arg, password}
}
