package gapi

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/liorlavon/simplebank/db/mock"
	db "github.com/liorlavon/simplebank/db/sqlc"
	pb "github.com/liorlavon/simplebank/pb"
	"github.com/liorlavon/simplebank/util"
	"github.com/liorlavon/simplebank/worker"
	mockwk "github.com/liorlavon/simplebank/worker/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		buildStub     func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
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
			buildStub: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {

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
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(arg, rowPassword, user)).
					Times(1).
					Return(db.CreateUserTxResult{User: user}, nil)

				taskPayload := &worker.PayloadSendVerifyEmail{
					Username: user.Username,
				}
				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), taskPayload, gomock.Any()).
					Times(1).
					Return(nil)
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
		{
			name: "InternalError",
			req: &pb.CreateUserRequest{
				Username:  user.Username,
				Firstname: user.Firstname,
				Lastname:  user.Lastname,
				Email:     user.Email,
				Password:  rowPassword,
			},
			buildStub: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {

				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrConnDone)

				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {

			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			mStore := mockdb.NewMockStore(storeCtrl)

			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)

			// build stub
			tc.buildStub(mStore, taskDistributor)

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
	user     db.User
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

	// call the AfterCreate function here
	err = actualArg.AfterCreate(expected.user)
	return err == nil
}

func (e eqCreateUserTxParamMatcher) String() string {
	return fmt.Sprintf("match argX %v and password %v", e.arg, e.password)
}

func EqCreateUserTxParams(arg db.CreateUserTxParams, password string, user db.User) gomock.Matcher {
	return eqCreateUserTxParamMatcher{arg, password, user}
}
