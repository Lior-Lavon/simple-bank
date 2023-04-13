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

// single test with out case table
func TestGetAccountAPI(t *testing.T) {
	// create random account
	account := randomAccount()

	// create mock_store
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create a new mockDB store
	mStore := mockdb.NewMockStore(ctrl)

	// build stub
	mStore.EXPECT().
		GetAccount(gomock.Any(), gomock.Eq(account.ID)).
		Times(1). // expect the GetAccount to be called exactly once
		Return(account, nil)

	// start http server and send http request
	server := NewServer(mStore)
	recorder := httptest.NewRecorder()

	url := fmt.Sprintf("/api/v1/accounts/%d", account.ID)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	// send the request to the server router, and response is record in the recorder
	server.router.ServeHTTP(recorder, request)

	// check statusCode response
	require.Equal(t, http.StatusOK, recorder.Code)

	// check the response Body account response, and compare it with the given account
	requierBodyMatchAccount(t, recorder.Body, account)
}

// using Table DrivenTest

func TestGetAccountAPIUsingTableDrivenTest(t *testing.T) {
	// create random account
	account := randomAccount()

	// table driven test set to cover all possible senarios
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
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1). // expect the GetAccount to be called exactly once
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusOK, recorder.Code)
				// check the response Body account response, and compare it with the given account
				requierBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStub: func(store *mockdb.MockStore) {
				// build stub
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1). // expect the GetAccount to be called exactly once
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStub: func(store *mockdb.MockStore) {
				// build stub
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).                             // expect the GetAccount to be called exactly once
					Return(db.Account{}, sql.ErrConnDone) // one possible error that a DB can return
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			buildStub: func(store *mockdb.MockStore) {
				// build stub
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i] // get each test case

		t.Run(tc.name, func(t *testing.T) {
			// create mock_store
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// create a new mockDB store
			mStore := mockdb.NewMockStore(ctrl)
			tc.buildStub(mStore)

			// start http server and send http request
			server := NewServer(mStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
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

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		OwnerID:  util.RandomInt(1, 1000),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

// compare the response body with the account to compare
func requierBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	// read all data from the response body
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var getAccount db.Account
	err = json.Unmarshal(data, &getAccount)
	require.NoError(t, err)

	// compare the input account and the returened account
	require.Equal(t, account, getAccount)
}

// compare the response body with the account to compare
func requierBodyMatchResponse(t *testing.T, body *bytes.Buffer, accountID int64) {
	// read all data from the response body
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var response struct {
		Response string `json:"response"`
	}
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)

	// compare the input account and the returened account
	require.Equal(t, fmt.Sprintf("account %d deleted", accountID), response.Response)
}
