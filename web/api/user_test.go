package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	mockdb "github.com/liorlavon/simplebank/db/mock"
	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/util"
	"github.com/stretchr/testify/require"
)

// ************************************************************************************
// Create Custom Matcher , based on Interface that implements Matches & String functions

type eqCreateUserParamMatcher struct {
	arg      db.CreateUserParams
	password string // raw password
}

func (e eqCreateUserParamMatcher) Matches(x interface{}) bool {

	// as x is an interface , convert x to db.CreateUserParams type
	argX, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	// check if the hashed password (x) match with the expected password or no (arg)
	err := util.CheckPassword(e.password, argX.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = argX.HashedPassword
	return reflect.DeepEqual(e.arg, argX)
}

func (e eqCreateUserParamMatcher) String() string {
	return fmt.Sprintf("match argX %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamMatcher{arg, password}
}

//************************************************************************************

func TestUserLogin(t *testing.T) {

	user, rowPassword := randomUser()

	testCases := []struct {
		name       string
		body       gin.H
		buildStub  func(store *mockdb.MockStore)
		validation func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"password": rowPassword,
			},
			buildStub: func(store *mockdb.MockStore) {

				gomock.InOrder(
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(1).
						Return(user, nil),
					store.EXPECT().
						CreateSession(gomock.Any(), gomock.Any()).
						Times(1),
				)
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "UserNotFound",
			body: gin.H{
				"username": user.Username,
				"password": rowPassword,
			},
			buildStub: func(store *mockdb.MockStore) {

				gomock.InOrder(
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(1).
						Return(db.User{}, sql.ErrNoRows),
				)
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check status code
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username": user.Username,
				"password": rowPassword,
			},
			buildStub: func(store *mockdb.MockStore) {

				gomock.InOrder(
					store.EXPECT().
						GetUser(gomock.Any(), gomock.Eq(user.Username)).
						Times(1).
						Return(db.User{}, sql.ErrConnDone),
				)
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check status code
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			// create http server
			server := newTestServer(t, mStore)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users/login"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			// send the request to the router
			server.router.ServeHTTP(recorder, request)

			tc.validation(t, recorder)
		})
	}
}

func TestCreateUser(t *testing.T) {

	user, rowPassword := randomUser()

	testCases := []struct {
		name       string
		body       gin.H
		buildStub  func(store *mockdb.MockStore)
		validation func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username":  user.Username,
				"password":  rowPassword,
				"firstname": user.Firstname,
				"lastname":  user.Lastname,
				"email":     user.Email,
			},
			buildStub: func(store *mockdb.MockStore) {

				arg := db.CreateUserParams{
					Username: user.Username,
					//					HashedPassword: user.HashedPassword,
					Firstname: user.Firstname,
					Lastname:  user.Lastname,
					Email:     user.Email,
				}

				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, rowPassword)).
					Times(1).
					Return(user, nil)
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check status code
				require.Equal(t, http.StatusOK, recorder.Code)

				arg := newUserResponse(user)

				checkBodyResponse(t, recorder.Body, arg)
			},
		},
		{
			name: "BadGateway",
			body: gin.H{
				"username":  user.Username,
				"password":  rowPassword,
				"firstname": user.Firstname,
				"lastname":  user.Lastname,
				"email":     user.Email,
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check status code
				require.Equal(t, http.StatusBadGateway, recorder.Code)
			},
		},
		{
			name: "DuplicateUsername",
			body: gin.H{
				"username":  user.Username,
				"password":  rowPassword,
				"firstname": user.Firstname,
				"lastname":  user.Lastname,
				"email":     user.Email,
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, &pq.Error{Code: "23505"})
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check status code
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InvalidEmail",
			// name: "ShortPassword",
			body: gin.H{
				"username":  user.Username,
				"password":  rowPassword,
				"firstname": user.Firstname,
				"lastname":  user.Lastname,
				"email":     "invalid-email",
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check status code
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ShortPassword",
			body: gin.H{
				"username":  user.Username,
				"password":  "123",
				"firstname": user.Firstname,
				"lastname":  user.Lastname,
				"email":     user.Email,
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check status code
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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

			// create http server
			server := newTestServer(t, mStore)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/api/v1/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			// send the request to the router
			server.router.ServeHTTP(recorder, request)

			tc.validation(t, recorder)
		})
	}
}

func TestGetUser(t *testing.T) {

	user, _ := randomUser()

	// create new mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create new mockDB store
	mStore := mockdb.NewMockStore(ctrl)

	testCases := []struct {
		name       string
		username   string
		buildStub  func()
		validation func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			username: user.Username,
			buildStub: func() {
				mStore.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check status code
				require.Equal(t, http.StatusOK, recorder.Code)

				arg := newUserResponse(user)

				checkBodyResponse(t, recorder.Body, arg)
			},
		},
		{
			name:     "NotFound",
			username: user.Username,
			buildStub: func() {
				mStore.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check status code
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:     "BadGateway",
			username: user.Username,
			buildStub: func() {
				mStore.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check status code
				require.Equal(t, http.StatusBadGateway, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// build stub
			tc.buildStub()

			// create http server
			server := newTestServer(t, mStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/users/%s", tc.username)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			addAuthenticationHeader(t, server.tokenMaker, request, user.Username)

			// send the request to the router
			server.router.ServeHTTP(recorder, request)

			tc.validation(t, recorder)
		})
	}
}

func TestListUsers(t *testing.T) {
	// create random list Users
	var users []db.User
	for i := 0; i < 5; i++ {
		u, _ := randomUser()
		users = append(users, u)
	}

	// define a list of test cases
	testCases := []struct {
		name          string // uniqe test name
		listuserParam func() db.ListUsersParams
		buildStub     func(store *mockdb.MockStore, arg db.ListUsersParams)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder, limit int32)
	}{
		{
			name: "OK",
			listuserParam: func() db.ListUsersParams {
				return db.ListUsersParams{
					Limit:  5,
					Offset: 1,
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.ListUsersParams) {
				arg.Offset--
				// build stub
				store.EXPECT().
					ListUsers(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(users, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, limit int32) {
				// check statusCode response
				require.Equal(t, http.StatusOK, recorder.Code)

				// check the response Body account response
				//requierBodyMatchExpected(t, recorder.Body)
			},
		},
		{
			name: "Validation",
			listuserParam: func() db.ListUsersParams {
				return db.ListUsersParams{
					Limit:  5,
					Offset: 0,
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.ListUsersParams) {
				arg.Offset--
				// build stub
				store.EXPECT().
					ListUsers(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, limit int32) {
				// check statusCode response
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "listUserError",
			listuserParam: func() db.ListUsersParams {
				return db.ListUsersParams{
					Limit:  5,
					Offset: 1,
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.ListUsersParams) {
				arg.Offset--
				// build stub
				store.EXPECT().
					ListUsers(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, limit int32) {
				// check statusCode response
				require.Equal(t, http.StatusBadGateway, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			// create mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// create a new mockDB store
			mStore := mockdb.NewMockStore(ctrl)
			arg := tc.listuserParam()

			tc.buildStub(mStore, arg)

			// start http server and send http request
			server := newTestServer(t, mStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/users?page_id=%d&page_size=%d", arg.Offset, arg.Limit)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// send the request to the server router, and response is record in the recorder
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder, arg.Limit)
		})
	}
}

// func TestUpdateUser(t *testing.T) {
// 	// create random user
// 	user, _ := randomUser()
// 	// create updated user
// 	//updatedUser := db.User(user)
// 	//	updatedUser.Firstname += "HHH"

// 	// define a list of test cases
// 	testCases := []struct {
// 		name          string // uniqe test name
// 		url           func(username string) string
// 		userParam     func() db.UpdateUserParams
// 		buildStub     func(store *mockdb.MockStore, up db.UpdateUserParams)   // the getAccount stub for each test will be build differently
// 		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder) // define a function that will check the output of the API
// 	}{
// 		{
// 			name: "OK",
// 			url: func(username string) string {
// 				return fmt.Sprintf("/api/v1/users/%s", username)
// 			},
// 			userParam: func() db.UpdateUserParams {
// 				return db.UpdateUserParams{
// 					Username:  user.Username,
// 					Firstname: sql.NullString{String: user.Firstname, Valid: true},
// 					Lastname:  sql.NullString{String: user.Lastname, Valid: true},
// 					Email:     sql.NullString{String: user.Email, Valid: true},
// 				}
// 			},
// 			buildStub: func(store *mockdb.MockStore, up db.UpdateUserParams) {
// 				gomock.InOrder(
// 					store.EXPECT().
// 						UpdateUser(gomock.Any(), gomock.Eq(up)).
// 						Times(1).
// 						Return(user, nil),
// 				)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				// check statusCode response
// 				require.Equal(t, http.StatusOK, recorder.Code)

// 				// check the response Body account response
// 				requierBodyMatchUser(t, recorder.Body, user)
// 			},
// 		},
// 		{
// 			name: "BindError",
// 			url: func(username string) string {
// 				return fmt.Sprintf("/api/v1/users/%s", username)
// 			},
// 			userParam: func() db.UpdateUserParams {
// 				return db.UpdateUserParams{
// 					Username: user.Username,
// 				}
// 			},
// 			buildStub: func(store *mockdb.MockStore, arg db.UpdateUserParams) {
// 				store.EXPECT().
// 					UpdateUser(gomock.Any(), gomock.Eq(arg)).
// 					Times(0)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				// check statusCode response
// 				require.Equal(t, http.StatusBadRequest, recorder.Code)
// 			},
// 		},
// 		{
// 			name: "UpdateUserError",
// 			url: func(username string) string {
// 				return fmt.Sprintf("/api/v1/users/%s", username)
// 			},
// 			userParam: func() db.UpdateUserParams {
// 				return db.UpdateUserParams{
// 					Username:  user.Username,
// 					Firstname: sql.NullString{String: user.Firstname, Valid: true},
// 					Lastname:  sql.NullString{String: user.Lastname, Valid: true},
// 					Email:     sql.NullString{String: user.Email, Valid: true},
// 				}
// 			},
// 			buildStub: func(store *mockdb.MockStore, arg db.UpdateUserParams) {
// 				gomock.InOrder(
// 					store.EXPECT().
// 						UpdateUser(gomock.Any(), gomock.Eq(arg)).
// 						Times(1).
// 						Return(db.User{}, sql.ErrConnDone),
// 				)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				// check statusCode response
// 				require.Equal(t, http.StatusBadGateway, recorder.Code)
// 			},
// 		},
// 	}

// 	for i := range testCases {
// 		tc := testCases[i]

// 		t.Run(tc.name, func(t *testing.T) {
// 			// create mock
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			// create a new mockDB store
// 			mStore := mockdb.NewMockStore(ctrl)

// 			uap := tc.userParam()
// 			tc.buildStub(mStore, uap)

// 			// start http server and send http request
// 			server := newTestServer(t, mStore)
// 			recorder := httptest.NewRecorder()

// 			// get the createAccountParag from the table
// 			var buf bytes.Buffer
// 			err := json.NewEncoder(&buf).Encode(uap)
// 			require.NoError(t, err)

// 			url := tc.url(uap.Username)
// 			request, err := http.NewRequest(http.MethodPut, url, &buf)
// 			require.NoError(t, err)

// 			addAuthenticationHeader(t, server.tokenMaker, request, user.Username)

// 			// send the request to the server router, and response is record in the recorder
// 			server.router.ServeHTTP(recorder, request)

// 			tc.checkResponse(t, recorder)
// 		})
// 	}
// }

func TestDeleteUser(t *testing.T) {
	// create random User
	user, _ := randomUser()

	// define a list of test cases
	testCases := []struct {
		name          string
		url           string
		buildStub     func(store *mockdb.MockStore)                           // the getAccount stub for each test will be build differently
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder) // define a function that will check the output of the API
	}{
		{
			name: "OK",
			url:  fmt.Sprintf("/api/v1/users/%s", user.Username),
			buildStub: func(store *mockdb.MockStore) {
				// build stub
				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusOK, recorder.Code)

				// check the response Body account response
				requierBodyMatchDeleteUserResponse(t, recorder.Body, user.Username)
			},
		},
		{
			name: "InvalidID",
			url:  fmt.Sprintf("/api/v1/users/%s", user.Username),
			buildStub: func(store *mockdb.MockStore) {
				// build stub
				store.EXPECT().
					DeleteUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {

			// create mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// create a new mockDB store
			mStore := mockdb.NewMockStore(ctrl)

			tc.buildStub(mStore)

			// start http server and send http request
			server := newTestServer(t, mStore)
			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodDelete, tc.url, nil)
			require.NoError(t, err)

			// send the request to the server router, and response is record in the recorder
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

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

// check the body of the response to equal to db.User
func checkBodyResponse(t *testing.T, body *bytes.Buffer, u userResponse) {

	// read all data from response body
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var user userResponse
	err = json.Unmarshal(data, &user)
	require.NoError(t, err)

	require.Equal(t, u, user)
}

func requierBodyMatchDeleteUserResponse(t *testing.T, body *bytes.Buffer, username string) {
	// read all data from the response body
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var response struct {
		Response string `json:"response"`
	}
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)

	expected := fmt.Sprintf("user %s deleted", username)
	// compare the input account and the returened account
	require.Equal(t, expected, response.Response)
}

func requierBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	// read all data from the response body
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var bodyUser db.User
	err = json.Unmarshal(data, &bodyUser)
	require.NoError(t, err)

	// compare the input account and the returened account
	require.Equal(t, user, bodyUser)

}
