package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	mockdb "github.com/liorlavon/simplebank/db/mock"
	db "github.com/liorlavon/simplebank/db/sqlc"
	"github.com/liorlavon/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestCreateOwner(t *testing.T) {

	o := createRandomOwner()

	// create new mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create new mockDB store
	mStore := mockdb.NewMockStore(ctrl)

	testCases := []struct {
		name       string
		body       func() db.CreateOwnerParams
		buildStub  func()
		validation func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: func() db.CreateOwnerParams {
				return db.CreateOwnerParams{
					Firstname: o.Firstname,
					Lastname:  o.Lastname,
					Email:     o.Email,
				}
			},
			buildStub: func() {

				cop := db.CreateOwnerParams{
					Firstname: o.Firstname,
					Lastname:  o.Lastname,
					Email:     o.Email,
				}

				mStore.EXPECT().
					CreateOwner(gomock.Any(), gomock.Eq(cop)).
					Times(1).
					Return(o, nil)
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check status code
				require.Equal(t, http.StatusOK, recorder.Code)

				checkBodyResponse(t, recorder.Body, o)
			},
		},
		{
			name: "Validation",
			body: func() db.CreateOwnerParams {
				return db.CreateOwnerParams{
					//					Firstname: o.Firstname,
					Lastname: o.Lastname,
					Email:    o.Email,
				}
			},
			buildStub: func() {
				mStore.EXPECT().
					CreateOwner(gomock.Any(), gomock.Any()).
					Times(0)
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check status code
				require.Equal(t, http.StatusBadRequest, recorder.Code)

				//checkBodyResponse(t, recorder.Body, o)
			},
		},
		{
			name: "BadGateway",
			body: func() db.CreateOwnerParams {
				return db.CreateOwnerParams{
					Firstname: o.Firstname,
					Lastname:  o.Lastname,
					Email:     o.Email,
				}
			},
			buildStub: func() {
				mStore.EXPECT().
					CreateOwner(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Owner{}, sql.ErrConnDone)
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
			server := NewServer(mStore)
			recorder := httptest.NewRecorder()

			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(tc.body())
			require.NoError(t, err)

			url := "/api/v1/owners"
			request, err := http.NewRequest(http.MethodPost, url, &buf)
			require.NoError(t, err)

			// send the request to the router
			server.router.ServeHTTP(recorder, request)

			tc.validation(t, recorder)
		})
	}
}

func TestGetOwner(t *testing.T) {

	o := createRandomOwner()

	// create new mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create new mockDB store
	mStore := mockdb.NewMockStore(ctrl)

	testCases := []struct {
		name       string
		ownerID    int64
		buildStub  func()
		validation func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "OK",
			ownerID: o.ID,
			buildStub: func() {
				mStore.EXPECT().
					GetOwner(gomock.Any(), gomock.Eq(o.ID)).
					Times(1).
					Return(o, nil)
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check status code
				require.Equal(t, http.StatusOK, recorder.Code)

				checkBodyResponse(t, recorder.Body, o)
			},
		},
		{
			name:    "validation",
			ownerID: 0,
			buildStub: func() {
				mStore.EXPECT().
					GetOwner(gomock.Any(), gomock.Eq(o.ID)).
					Times(0)
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check status code
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:    "NotFound",
			ownerID: o.ID,
			buildStub: func() {
				mStore.EXPECT().
					GetOwner(gomock.Any(), gomock.Eq(o.ID)).
					Times(1).
					Return(db.Owner{}, sql.ErrNoRows)
			},
			validation: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check status code
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:    "BadGateway",
			ownerID: o.ID,
			buildStub: func() {
				mStore.EXPECT().
					GetOwner(gomock.Any(), gomock.Eq(o.ID)).
					Times(1).
					Return(db.Owner{}, sql.ErrConnDone)
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
			server := NewServer(mStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/owners/%d", tc.ownerID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// send the request to the router
			server.router.ServeHTTP(recorder, request)

			tc.validation(t, recorder)
		})
	}
}

func TestListOwners(t *testing.T) {
	// create random list owners
	var owners []db.Owner
	for i := 0; i < 5; i++ {
		owners = append(owners, randomOwner())
	}

	// define a list of test cases
	testCases := []struct {
		name           string // uniqe test name
		listownerParam func() db.ListOwnersParams
		buildStub      func(store *mockdb.MockStore, arg db.ListOwnersParams)
		checkResponse  func(t *testing.T, recorder *httptest.ResponseRecorder, limit int32)
	}{
		{
			name: "OK",
			listownerParam: func() db.ListOwnersParams {
				return db.ListOwnersParams{
					Limit:  5,
					Offset: 1,
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.ListOwnersParams) {
				arg.Offset--
				// build stub
				store.EXPECT().
					ListOwners(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(owners, nil)
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
			listownerParam: func() db.ListOwnersParams {
				return db.ListOwnersParams{
					Limit:  5,
					Offset: 0,
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.ListOwnersParams) {
				arg.Offset--
				// build stub
				store.EXPECT().
					ListOwners(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, limit int32) {
				// check statusCode response
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "listOwnerError",
			listownerParam: func() db.ListOwnersParams {
				return db.ListOwnersParams{
					Limit:  5,
					Offset: 1,
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.ListOwnersParams) {
				arg.Offset--
				// build stub
				store.EXPECT().
					ListOwners(gomock.Any(), gomock.Eq(arg)).
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
			arg := tc.listownerParam()

			tc.buildStub(mStore, arg)

			// start http server and send http request
			server := NewServer(mStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/owners?page_id=%d&page_size=%d", arg.Offset, arg.Limit)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// send the request to the server router, and response is record in the recorder
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder, arg.Limit)
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	// create random account
	account := randomAccount()
	// create updated account
	updatedAccount := db.Account(account)
	updatedAccount.Balance += 50

	// define a list of test cases
	testCases := []struct {
		name          string // uniqe test name
		url           func(id int32) string
		accountParam  func() db.UpdateAccountParams
		buildStub     func(store *mockdb.MockStore, arg db.UpdateAccountParams) // the getAccount stub for each test will be build differently
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)   // define a function that will check the output of the API
	}{
		{
			name: "OK",
			url: func(id int32) string {
				return fmt.Sprintf("/api/v1/accounts/%d", id)
			},
			accountParam: func() db.UpdateAccountParams {
				return db.UpdateAccountParams{
					ID:      account.ID,
					Balance: updatedAccount.Balance,
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.UpdateAccountParams) {
				gomock.InOrder(
					store.EXPECT().
						UpdateAccount(gomock.Any(), gomock.Eq(arg)).
						Times(1).
						Return(updatedAccount, nil),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusOK, recorder.Code)

				// check the response Body account response
				requierBodyMatchAccount(t, recorder.Body, updatedAccount)
			},
		},
		{
			name: "Validation",
			url: func(id int32) string {
				id = 0
				return fmt.Sprintf("/api/v1/accounts/%d", id)
			},
			accountParam: func() db.UpdateAccountParams {
				return db.UpdateAccountParams{
					ID:      account.ID,
					Balance: updatedAccount.Balance,
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.UpdateAccountParams) {
				gomock.InOrder(
					store.EXPECT().
						UpdateAccount(gomock.Any(), gomock.Eq(arg)).
						Times(0),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BindError",
			url: func(id int32) string {
				return fmt.Sprintf("/api/v1/accounts/%d", id)
			},
			accountParam: func() db.UpdateAccountParams {
				return db.UpdateAccountParams{
					ID: account.ID,
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.UpdateAccountParams) {
				gomock.InOrder(
					store.EXPECT().
						UpdateAccount(gomock.Any(), gomock.Eq(arg)).
						Times(0),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "IdNoMatch",
			url: func(id int32) string {
				id = 500
				return fmt.Sprintf("/api/v1/accounts/%d", id)
			},
			accountParam: func() db.UpdateAccountParams {
				return db.UpdateAccountParams{
					ID:      account.ID,
					Balance: updatedAccount.Balance,
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.UpdateAccountParams) {
				gomock.InOrder(
					store.EXPECT().
						UpdateAccount(gomock.Any(), gomock.Eq(arg)).
						Times(0),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusBadRequest, recorder.Code)

				// check the response Body account response
				requierBodyMatchIDNotMatch(t, recorder.Body, 500)
			},
		},
		{
			name: "UpdateAccountError",
			url: func(id int32) string {
				return fmt.Sprintf("/api/v1/accounts/%d", id)
			},
			accountParam: func() db.UpdateAccountParams {
				return db.UpdateAccountParams{
					ID:      account.ID,
					Balance: updatedAccount.Balance,
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.UpdateAccountParams) {
				gomock.InOrder(
					store.EXPECT().
						UpdateAccount(gomock.Any(), gomock.Eq(arg)).
						Times(1).
						Return(db.Account{}, sql.ErrConnDone),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
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

			uap := tc.accountParam()
			tc.buildStub(mStore, uap)

			// start http server and send http request
			server := NewServer(mStore)
			recorder := httptest.NewRecorder()

			// get the createAccountParag from the table
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(uap)
			require.NoError(t, err)

			url := tc.url(int32(uap.ID))
			request, err := http.NewRequest(http.MethodPut, url, &buf)
			require.NoError(t, err)

			// send the request to the server router, and response is record in the recorder
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteAccount(t *testing.T) {
	// create random account
	account := randomAccount()

	// define a list of test cases
	testCases := []struct {
		name          string                                                  // uniqe test name
		accountID     int64                                                   //  accountID that we want to get
		buildStub     func(store *mockdb.MockStore)                           // the getAccount stub for each test will be build differently
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder) // define a function that will check the output of the API
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStub: func(store *mockdb.MockStore) {
				// build stub
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(nil)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusOK, recorder.Code)

				// check the response Body account response
				requierBodyMatchResponse(t, recorder.Body, account.ID)
			},
		},
		{
			name:      "BadRequest",
			accountID: 0,
			buildStub: func(store *mockdb.MockStore) {
				// build stub
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: account.ID,
			buildStub: func(store *mockdb.MockStore) {
				// build stub
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
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
			server := NewServer(mStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			// send the request to the server router, and response is record in the recorder
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

// check the body of the response to equal to db.Owner
func checkBodyResponse(t *testing.T, body *bytes.Buffer, o db.Owner) {

	// read all data from response body
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var owner db.Owner
	err = json.Unmarshal(data, &owner)
	require.NoError(t, err)

	require.Equal(t, o, owner)
}

func createRandomOwner() db.Owner {
	return db.Owner{
		ID:        util.RandomInt(1, 1000),
		Firstname: util.RandomOwner(),
		Lastname:  util.RandomOwner(),
		Email:     util.RandEmail(),
		CreatedAt: util.GetTime(),
	}
}
