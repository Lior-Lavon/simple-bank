package gapi

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/liorlavon/simplebank/db/mock"
	db "github.com/liorlavon/simplebank/db/sqlc"
	pb "github.com/liorlavon/simplebank/pb"
	"github.com/liorlavon/simplebank/token"
	"github.com/liorlavon/simplebank/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateUserAPI(t *testing.T) {

	user, _ := randomUser()

	newFirstName := util.RandomUser()
	newEmail := util.RandEmail()
	newPassword := "ab"
	invalidEmail := "invalid-email"

	testCases := []struct {
		name          string
		req           *pb.UpdateUserRequest
		buildStub     func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.UpdateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.UpdateUserRequest{
				Username:  user.Username,
				Firstname: &newFirstName,
				Email:     &newEmail,
			},
			buildStub: func(store *mockdb.MockStore) {

				arg := db.UpdateUserParams{
					Username: user.Username,
					Firstname: sql.NullString{
						String: newFirstName,
						Valid:  true,
					},
					Email: sql.NullString{
						String: newEmail,
						Valid:  true,
					},
				}

				updatedUser := db.User{
					Username:          arg.Username,
					HashedPassword:    user.HashedPassword,
					Firstname:         newFirstName,
					Lastname:          user.Lastname,
					IsEmailVerified:   user.IsEmailVerified,
					Email:             newEmail,
					PasswordChangedAt: user.PasswordChangedAt,
					CreatedAt:         user.PasswordChangedAt,
				}

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(updatedUser, nil)

			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				// add a access token to the context metadata
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				updatedUser := res.GetUser()
				require.Equal(t, user.Username, updatedUser.Username)
				require.Equal(t, newFirstName, updatedUser.Firstname)
				require.Equal(t, newEmail, updatedUser.Email)
			},
		},
		{
			name: "UserNotFound",
			req: &pb.UpdateUserRequest{
				Username:  user.Username,
				Firstname: &newFirstName,
				Email:     &newEmail,
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)

			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				// add a access token to the context metadata
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				// check the status code of the call
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "ExpiredToken",
			req: &pb.UpdateUserRequest{
				Username:  user.Username,
				Firstname: &newFirstName,
				Email:     &newEmail,
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)

			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				// add a access token to the context metadata
				return newContextWithBearerToken(t, tokenMaker, user.Username, -time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				// check the status code of the call
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "InvalidPasswordError",
			req: &pb.UpdateUserRequest{
				Username:  user.Username,
				Password:  &newPassword,
				Firstname: &newFirstName,
				Email:     &newEmail,
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)

			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				// add a access token to the context metadata
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				// check the status code of the call
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "NoAuthorization",
			req: &pb.UpdateUserRequest{
				Username:  user.Username,
				Firstname: &newFirstName,
				Email:     &newEmail,
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)

			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				// add a access token to the context metadata
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				// check the status code of the call
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "PermissionDeniedErrro",
			req: &pb.UpdateUserRequest{
				Username:  "moshe",
				Firstname: &newFirstName,
				Email:     &newEmail,
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)

			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				// add a access token to the context metadata
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				// check the status code of the call
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			name: "InvalidEmail",
			req: &pb.UpdateUserRequest{
				Username:  user.Username,
				Firstname: &newFirstName,
				Email:     &invalidEmail,
			},
			buildStub: func(store *mockdb.MockStore) {

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)

			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				// add a access token to the context metadata
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				// check the status code of the call
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {

			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			mStore := mockdb.NewMockStore(storeCtrl)

			// build stub
			tc.buildStub(mStore)

			// create http server
			server := newTestServer(t, mStore, nil)

			// get a context with auth token
			ctx := tc.buildContext(t, server.tokenMaker)

			res, err := server.UpdateUser(ctx, tc.req)

			tc.checkResponse(t, res, err)
		})
	}
}
