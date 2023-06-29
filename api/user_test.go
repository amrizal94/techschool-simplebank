package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	mockdb "github.com/amrizal94/simplebank/db/mock"
	db "github.com/amrizal94/simplebank/db/sqlc"
	"github.com/amrizal94/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type eqCreatedUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreatedUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword
	// In case, some value is nil
	if e.arg == (db.CreateUserParams{}) || arg == (db.CreateUserParams{}) {
		return reflect.DeepEqual(e.arg, arg)
	}

	// Check if types assignable and convert them to common type
	x1Val := reflect.ValueOf(e.arg)
	x2Val := reflect.ValueOf(arg)

	if x1Val.Type().AssignableTo(x2Val.Type()) {
		x1ValConverted := x1Val.Convert(x2Val.Type())
		return reflect.DeepEqual(x1ValConverted.Interface(), x2Val.Interface())
	}

	return false

}

func (e eqCreatedUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreatedUserParamsMatcher{arg, password}
}
func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser()
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	testCases := []struct {
		desc          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			desc: "OK",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username:       user.Username,
					FullName:       user.FullName,
					Email:          user.Email,
					HashedPassword: hashedPassword,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recoder.Code)
				requireBodyMatchUser(t, recoder.Body, user)
			},
		},
		{
			desc: "DuplicateUsername",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recoder.Code)
			},
		},
		{
			desc: "DuplicateEmail",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recoder.Code)
			},
		},
		{
			desc: "InternalError",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recoder.Code)
			},
		},
		{
			desc: "InvalidUsername",
			body: gin.H{
				"username":  "amr#1",
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recoder.Code)
			},
		},
		{
			desc: "InvalidEmail",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     "asdmexample.com",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recoder.Code)
			},
		},
		{
			desc: "PasswordTooShort",
			body: gin.H{
				"username":  user.Username,
				"password":  "amr",
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recoder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recoder.Code)
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctlr := gomock.NewController(t)
			defer ctlr.Finish()

			store := mockdb.NewMockStore(ctlr)
			tC.buildStubs(store)

			server := newTestServer(t, store)
			recoder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tC.body)
			require.NoError(t, err)

			url := "/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recoder, request)
			tC.checkResponse(recoder)
		})
	}
}

func randomUser() (user db.User, password string) {
	password = util.RandomString(6)

	user = db.User{
		Username: util.RandomString(5),
		Email:    util.RandomEmail(),
		FullName: util.RandomOwner(),
	}
	return
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	var data map[string]json.RawMessage
	err := json.Unmarshal(body.Bytes(), &data)
	require.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(data["user"], &gotUser)
	require.NoError(t, err)
	require.Equal(t, gotUser, user)
}
