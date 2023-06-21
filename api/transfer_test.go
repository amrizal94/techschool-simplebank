package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/amrizal94/simplebank/db/mock"
	db "github.com/amrizal94/simplebank/db/sqlc"
	"github.com/amrizal94/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestTransferAPI(t *testing.T) {
	amount := int64(10)
	user1 := randomAccount()
	user2 := randomAccount()
	user3 := randomAccount()

	user1.Currency = util.USD
	user2.Currency = util.USD
	user3.Currency = util.CAD

	testCases := []struct {
		name          string
		amount        int64
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			amount: amount,
			body: gin.H{
				"from_account_id": user1.ID,
				"to_account_id":   user2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user1.ID)).Times(1).
					Return(user1, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user2.ID)).Times(1).
					Return(user2, nil)

				arg := db.TranferTxParams{
					FromAccountID: user1.ID,
					ToAccountID:   user2.ID,
					Amount:        amount,
				}
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(arg)).Times(1)
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recoder.Code)
			},
		},
		{
			name:   "InternalError",
			amount: amount,
			body: gin.H{
				"from_account_id": user1.ID,
				"to_account_id":   user2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user1.ID)).Times(1).
					Return(user1, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user2.ID)).Times(1).
					Return(user2, nil)

				arg := db.TranferTxParams{
					FromAccountID: user1.ID,
					ToAccountID:   user2.ID,
					Amount:        amount,
				}
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(arg)).Times(1).
					Return(db.TransferTxResult{}, sql.ErrConnDone)

			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recoder.Code)
			},
		},
		{
			name:   "FromAccountNotFound",
			amount: amount,
			body: gin.H{
				"from_account_id": user1.ID,
				"to_account_id":   user2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user1.ID)).Times(1).
					Return(db.Account{}, sql.ErrNoRows)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user2.ID)).Times(0)

				arg := db.TranferTxParams{
					FromAccountID: user1.ID,
					ToAccountID:   user2.ID,
					Amount:        amount,
				}

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(arg)).Times(0)

			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recoder.Code)
			},
		},
		{
			name:   "ToAccountNotFound",
			amount: amount,
			body: gin.H{
				"from_account_id": user1.ID,
				"to_account_id":   user2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user1.ID)).Times(1).
					Return(user1, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user2.ID)).Times(1).
					Return(db.Account{}, sql.ErrNoRows)

				arg := db.TranferTxParams{
					FromAccountID: user1.ID,
					ToAccountID:   user2.ID,
					Amount:        amount,
				}

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(arg)).Times(0)

			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recoder.Code)
			},
		},
		{
			name:   "InvalidCurrency",
			amount: amount,
			body: gin.H{
				"from_account_id": user1.ID,
				"to_account_id":   user2.ID,
				"amount":          amount,
				"currency":        "Invalid",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user1.ID)).Times(0)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user2.ID)).Times(0)

				arg := db.TranferTxParams{
					FromAccountID: user1.ID,
					ToAccountID:   user2.ID,
					Amount:        amount,
				}

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(arg)).Times(0)

			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recoder.Code)
			},
		},
		{
			name:   "NegativeAmount",
			amount: amount,
			body: gin.H{
				"from_account_id": user1.ID,
				"to_account_id":   user2.ID,
				"amount":          -amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user1.ID)).Times(0)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user2.ID)).Times(0)

				arg := db.TranferTxParams{
					FromAccountID: user1.ID,
					ToAccountID:   user2.ID,
					Amount:        amount,
				}

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(arg)).Times(0)

			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recoder.Code)
			},
		},
		{
			name:   "FromAccountCurrencyMismatch",
			amount: amount,
			body: gin.H{
				"from_account_id": user3.ID,
				"to_account_id":   user2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user3.ID)).Times(1).
					Return(user3, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user2.ID)).Times(0)
				arg := db.TranferTxParams{
					FromAccountID: user3.ID,
					ToAccountID:   user2.ID,
					Amount:        amount,
				}
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(arg)).Times(0)
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recoder.Code)
			},
		},
		{
			name:   "ToAccountCurrencyMismatch",
			amount: amount,
			body: gin.H{
				"from_account_id": user1.ID,
				"to_account_id":   user3.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user1.ID)).Times(1).
					Return(user1, nil)

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(user3.ID)).Times(1).
					Return(user3, nil)
				arg := db.TranferTxParams{
					FromAccountID: user1.ID,
					ToAccountID:   user3.ID,
					Amount:        amount,
				}
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(arg)).Times(0)
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recoder.Code)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			testCase.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			// Marshal body data into JSON
			data, err := json.Marshal(testCase.body)
			require.NoError(t, err)

			url := "/transfers"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			testCase.checkResponse(recorder)

		})

	}

}
