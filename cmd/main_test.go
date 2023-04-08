package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/gorilla/mux"
)

const (
	testUsername = "admin"
	testPassword = "admin"
)

func TestSumNumbersInJSON(t *testing.T) {
	// create mux with /auth and /sum handlers
	r := mux.NewRouter()
	r.HandleFunc("/auth", authHandler).Methods("POST")
	r.HandleFunc("/sum", sumHandler).Methods("POST")

	// run server using httptest
	server := httptest.NewServer(r)
	defer server.Close()

	// create httpexpect instance
	e := httpexpect.Default(t, server.URL)

	// invalid request method
	e.GET("/sum").
		Expect().
		Status(http.StatusMethodNotAllowed)

	// no authorization
	e.POST("/sum").
		Expect().
		Status(http.StatusUnauthorized)

	// invalid authorization
	e.POST("/sum").
		WithHeader("Authorization", "Bearer blabla").
		Expect().
		Status(http.StatusUnauthorized)

	// get auth token
	token := getAuthToken(e)

	testSumError(e, "", token, http.StatusBadRequest)
	testSumError(e, "blabla", token, http.StatusBadRequest)

	// tests from README

	// [1,2,3,4]
	testSumSuccess(e, "[1,2,3,4]", token, "4a44dc15364204a80fe80e9039455cc1608281820fe2b24f1e5233ade6af1dd5")

	// {"a":6,"b":4}
	testSumSuccess(e, `{"a":6,"b":4}`, token, "4a44dc15364204a80fe80e9039455cc1608281820fe2b24f1e5233ade6af1dd5")

	// [[[2]]]
	testSumSuccess(e, "[[[2]]]", token, "d4735e3a265e16eee03f59718b9b5d03019c07d8b6c51f90da3a666eec13ab35")

	// {"a":{"b":4},"c":-2}
	testSumSuccess(e, `{"a":{"b":4},"c":-2}`, token, "d4735e3a265e16eee03f59718b9b5d03019c07d8b6c51f90da3a666eec13ab35")

	// {"a":[-1,1,"dark"]}
	testSumSuccess(e, `{"a":[-1,1,"dark"]}`, token, "5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9")

	// [-1,{"a":1, "b":"light"}]
	testSumSuccess(e, `[-1,{"a":1, "b":"light"}]`, token, "5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9")

	// []
	testSumSuccess(e, `[]`, token, "5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9")

	// {}
	testSumSuccess(e, `{}`, token, "5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9")

	// more tests

	// ["math.MaxInt64-1 math.MaxInt64+1 1"] => 1
	testSumSuccess(e, `["-9223372036854775808 9223372036854775808 1"]`, token, "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b")

	// ["1 2","3.14 0x4"] => 10
	testSumSuccess(e, `["1 2", "3.14 0x4"]`, token, "4a44dc15364204a80fe80e9039455cc1608281820fe2b24f1e5233ade6af1dd5")

	// [1E-10, -2E-10] => 0
	testSumSuccess(e, "[1E-10, -2E-10]", token, "5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9")

	// [2.99792458e8,5.99792458e8] => 7
	testSumSuccess(e, "[2.99792458e8,5.99792458e8]", token, "924742de0792204f5b8b73160987444bdb7422abe15ee43f10dcd7b3e919eb41")

	// {"a 1":-1,"b 2":-2} => 0
	testSumSuccess(e, `{"a 1":-1,"b 2":-2}`, token, "5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9")

	// {"a":"b","c":"d} => 0
	testSumSuccess(e, `{"a":"b","c":"d"}`, token, "5feceb66ffc86f38d952786c6d696c79c2dbc239dd4e91b46729d73a27fb57e9")

}

func TestAuth(t *testing.T) {
	// create http.Handler
	handler := http.HandlerFunc(authHandler)

	// run server using httptest
	server := httptest.NewServer(handler)
	defer server.Close()

	// create httpexpect instance
	e := httpexpect.Default(t, server.URL)

	// success
	e.POST("/auth").
		WithJSON(map[string]any{
			"username": testUsername,
			"password": testPassword,
		}).
		Expect().
		Status(http.StatusOK).Text().NotEmpty()

	// invalid request method
	e.GET("/auth").
		WithJSON(map[string]any{
			"username": testUsername,
			"password": testPassword,
		}).
		Expect().
		Status(http.StatusMethodNotAllowed)

	// bad request
	e.POST("/auth").
		WithBytes([]byte("blabla")).
		Expect().
		Status(http.StatusBadRequest)

	// empty password
	e.POST("/auth").
		WithJSON(map[string]any{
			"username": testUsername,
		}).
		Expect().
		Status(http.StatusBadRequest).Text().HasPrefix(errEmptyUserNameOrPassword.Error())

	// empty username
	e.POST("/auth").
		WithJSON(map[string]any{
			"password": testPassword,
		}).
		Expect().
		Status(http.StatusBadRequest).Text().HasPrefix(errEmptyUserNameOrPassword.Error())
}

func testSumSuccess(e *httpexpect.Expect, payload string, token, result string) {
	e.POST("/sum").
		WithHeader("Authorization", "Bearer "+token).
		WithBytes([]byte(payload)).
		Expect().
		Status(http.StatusOK).Text().IsEqual(result)
}

func testSumError(e *httpexpect.Expect, payload, token string, status int) {
	e.POST("/sum").
		WithHeader("Authorization", "Bearer "+token).
		WithBytes([]byte(payload)).
		Expect().
		Status(status)
}

func getAuthToken(e *httpexpect.Expect) string {
	return e.POST("/auth").
		WithJSON(map[string]any{
			"username": testUsername,
			"password": testPassword,
		}).
		Expect().
		Status(http.StatusOK).Text().NotEmpty().Raw()
}
