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

func TestCreateAccount(t *testing.T) {
	// create random owner & account
	owner := randomOwner()
	account := db.Account{
		ID:       1,
		OwnerID:  owner.ID,
		Balance:  100,
		Currency: "USD",
	}

	// define a list of test cases
	testCases := []struct {
		name          string // uniqe test name
		ownerID       int64
		accountParam  func() db.CreateAccountParams
		buildStub     func(store *mockdb.MockStore, arg db.CreateAccountParams) // the getAccount stub for each test will be build differently
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)   // define a function that will check the output of the API
	}{
		{
			name:    "OK",
			ownerID: owner.ID,
			accountParam: func() db.CreateAccountParams {
				return db.CreateAccountParams{
					OwnerID:  owner.ID,
					Balance:  100,
					Currency: "USD",
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.CreateAccountParams) {
				gomock.InOrder(
					store.EXPECT().
						GetOwner(gomock.Any(), gomock.Eq(owner.ID)).
						Times(1).
						Return(owner, nil),
					store.EXPECT().
						CreateAccount(gomock.Any(), arg).
						Times(1).
						Return(account, nil),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusOK, recorder.Code)

				// check the response Body account response
				requierBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:    "ValidationIssue",
			ownerID: owner.ID,
			accountParam: func() db.CreateAccountParams {
				return db.CreateAccountParams{
					// OwnerID:  owner.ID,
					Balance:  100,
					Currency: "USD",
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.CreateAccountParams) {
				gomock.InOrder(
					store.EXPECT().
						GetOwner(gomock.Any(), gomock.Eq(owner.ID)).
						Times(0),
					store.EXPECT().
						CreateAccount(gomock.Any(), arg).
						Times(0),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:    "OwnerNotExist",
			ownerID: owner.ID,
			accountParam: func() db.CreateAccountParams {
				return db.CreateAccountParams{
					OwnerID:  owner.ID,
					Balance:  100,
					Currency: "USD",
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.CreateAccountParams) {
				gomock.InOrder(
					store.EXPECT().
						GetOwner(gomock.Any(), gomock.Eq(owner.ID)).
						Times(1).
						Return(db.Owner{}, sql.ErrNoRows),
					store.EXPECT().
						CreateAccount(gomock.Any(), arg).
						Times(0),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusNotFound, recorder.Code)

				// check the response Body
				requierBodyMatchOwnerDoesNotExist(t, recorder.Body, owner.ID)
			},
		},
		{
			name:    "GetOwnerBadGateway",
			ownerID: owner.ID,
			accountParam: func() db.CreateAccountParams {
				return db.CreateAccountParams{
					OwnerID:  owner.ID,
					Balance:  100,
					Currency: "USD",
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.CreateAccountParams) {
				gomock.InOrder(
					store.EXPECT().
						GetOwner(gomock.Any(), gomock.Eq(owner.ID)).
						Times(1).
						Return(db.Owner{}, sql.ErrConnDone),
					store.EXPECT().
						CreateAccount(gomock.Any(), arg).
						Times(0),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// check statusCode response
				require.Equal(t, http.StatusBadGateway, recorder.Code)
			},
		},
		{
			name:    "CreateAccountBadGateway",
			ownerID: owner.ID,
			accountParam: func() db.CreateAccountParams {
				return db.CreateAccountParams{
					OwnerID:  owner.ID,
					Balance:  100,
					Currency: "USD",
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.CreateAccountParams) {
				gomock.InOrder(
					store.EXPECT().
						GetOwner(gomock.Any(), gomock.Eq(owner.ID)).
						Times(1).
						Return(owner, nil),
					store.EXPECT().
						CreateAccount(gomock.Any(), arg).
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

			cap := tc.accountParam()
			tc.buildStub(mStore, cap)

			// start http server and send http request
			server := NewServer(mStore)
			recorder := httptest.NewRecorder()

			// get the createAccountParag from the table
			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(cap)
			require.NoError(t, err)

			url := "/api/v1/accounts"
			request, err := http.NewRequest(http.MethodPost, url, &buf)
			require.NoError(t, err)

			// send the request to the server router, and response is record in the recorder
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccount(t *testing.T) {
	// create random list accounts
	var accounts []db.Account
	for i := 0; i < 5; i++ {
		accounts = append(accounts, randomAccount())
	}

	// define a list of test cases
	testCases := []struct {
		name             string // uniqe test name
		listAccountParam func() db.ListAccountsParams
		buildStub        func(store *mockdb.MockStore, arg db.ListAccountsParams)             // the getAccount stub for each test will be build differently
		checkResponse    func(t *testing.T, recorder *httptest.ResponseRecorder, limit int32) // define a function that will check the output of the API
	}{
		{
			name: "OK",
			listAccountParam: func() db.ListAccountsParams {
				return db.ListAccountsParams{
					Limit:  5,
					Offset: 1,
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.ListAccountsParams) {
				arg.Offset--
				// build stub
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, limit int32) {
				// check statusCode response
				require.Equal(t, http.StatusOK, recorder.Code)

				// check the response Body account response
				requierBodyMatchExpected(t, recorder.Body)
			},
		},
		{
			name: "Validation",
			listAccountParam: func() db.ListAccountsParams {
				return db.ListAccountsParams{
					Limit:  5,
					Offset: 0,
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.ListAccountsParams) {
				arg.Offset--
				// build stub
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, limit int32) {
				// check statusCode response
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "listAccountError",
			listAccountParam: func() db.ListAccountsParams {
				return db.ListAccountsParams{
					Limit:  5,
					Offset: 1,
				}
			},
			buildStub: func(store *mockdb.MockStore, arg db.ListAccountsParams) {
				arg.Offset--
				// build stub
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
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
			arg := tc.listAccountParam()

			tc.buildStub(mStore, arg)

			// start http server and send http request
			server := NewServer(mStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/api/v1/accounts?page_id=%d&page_size=%d", arg.Offset, arg.Limit)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// send the request to the server router, and response is record in the recorder
			server.router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder, arg.Limit)
		})
	}

}

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

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		OwnerID:  util.RandomInt(1, 1000),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func randomOwner() db.Owner {
	return db.Owner{
		ID:        util.RandomInt(1, 1000),
		Firstname: util.RandomOwner(),
		Lastname:  util.RandomOwner(),
		Email:     util.RandEmail(),
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

func requierBodyMatchOwnerDoesNotExist(t *testing.T, body *bytes.Buffer, ownerID int64) {
	// read all data from the response body
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var response struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)

	// compare the input account and the returened account
	require.Equal(t, fmt.Sprintf("owner %d does not exist", ownerID), response.Error)
}

func requierBodyMatchListAccountOK(t *testing.T, body *bytes.Buffer, limit int32) {
	// read all data from the response body
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var listAccount []db.Account
	err = json.Unmarshal(data, &listAccount)
	require.NoError(t, err)

	// compare the input account and the returened account
	require.Equal(t, len(listAccount), limit)
}

func requierBodyMatchExpected(t *testing.T, body *bytes.Buffer) {
	// read all data from the response body
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var listAccount []db.Account
	err = json.Unmarshal(data, &listAccount)
	require.NoError(t, err)

	// compare the input account and the returened account
	require.Equal(t, len(listAccount), 5)
}

func requierBodyMatchIDNotMatch(t *testing.T, body *bytes.Buffer, accountId int64) {
	// read all data from the response body
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var response struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)

	// compare the input account and the returened account
	require.Equal(t, fmt.Sprintf("id %d does not match", accountId), response.Error)
}
