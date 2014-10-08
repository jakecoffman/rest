package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

var userColumns = []string{"o_id", "o_name"}

var data = []struct {
	url          string
	method       string
	reqBody      io.Reader
	expectations func()
	expectedCode int
	expectedBody string
}{
	{
		"/users", "GET", nil, func() {
			sqlmock.ExpectQuery("select id, name from users").
				WillReturnRows(sqlmock.NewRows(userColumns).
				FromCSVString("1,Bob\n2,Ted"))
		}, 200, `[{"id":1,"name":"Bob"},{"id":2,"name":"Ted"}]`,
	},
	{
		"/users", "POST", strings.NewReader(`{"name":"test"}`), func() {
			sqlmock.ExpectPrepare()
			sqlmock.ExpectExec("insert into users \\(name\\) values \\(\\?\\)").
				WithArgs("test").
				WillReturnResult(sqlmock.NewResult(7, 1))
		}, 201, `{"id":7,"name":"test"}`,
	},
	{
		"/users/3", "GET", nil, func() {
			sqlmock.ExpectPrepare()
			// TODO: I can't .WithArgs("3") here for some reason with the QueryRow call
			sqlmock.ExpectQuery("select id, name from users where id=\\?").
				WillReturnRows(sqlmock.NewRows(userColumns).AddRow(3, "Jim"))
		}, 200, `{"id":3,"name":"Jim"}`,
	}, {
		"/users/2", "PUT", strings.NewReader(`{"name":"Bob"}`), func() {
			sqlmock.ExpectPrepare()
			sqlmock.ExpectExec("update users set name=\\? where id=\\?").
				WithArgs("Bob", 2).
				WillReturnResult(sqlmock.NewResult(0, 1))
		}, 200, `{"id":"2"}`,
	}, {
		"/users/3", "DELETE", nil, func() {
			sqlmock.ExpectPrepare()
			sqlmock.ExpectExec("delete from users where id=\\?").
				WithArgs(3).
				WillReturnResult(sqlmock.NewResult(0, 1))
		}, 200, `{"id":"3"}`,
	}, {
		"/users", "POST", nil, func() {}, 400, `{"error":"no payload"}`,
	}, {
		"/users", "POST", strings.NewReader(`{"name":`), func() {}, 400, `{"error":"can't parse json payload"}`,
	}, {
		"/users", "POST", strings.NewReader(`{"name":""}`), func() {}, 422, `{"error":"please provide 'name'"}`,
	}, {
		"/users/4", "PUT", nil, func() {}, 400, `{"error":"no payload"}`,
	}, {
		"/users/4", "PUT", strings.NewReader(`{"name":`), func() {}, 400, `{"error":"can't parse json payload"}`,
	}, {
		"/users/4", "PUT", strings.NewReader(`{"name":""}`), func() {}, 422, `{"error":"please provide 'name'"}`,
	}, {
		"/users/4", "PUT", strings.NewReader(`{"name":"asdf"}`), func() {
			sqlmock.ExpectPrepare()
			sqlmock.ExpectExec("update users set name=\\? where id=\\?").
				WithArgs("asdf", 4).
				WillReturnResult(sqlmock.NewResult(0, 0))
		}, 404, `{"error":"User not found"}`,
	}, {
		"/users/6", "DELETE", strings.NewReader(`{"name":"asdf"}`), func() {
			sqlmock.ExpectPrepare()
			sqlmock.ExpectExec("delete from users where id=\\?").
				WithArgs(6).
				WillReturnResult(sqlmock.NewResult(0, 0))
		}, 404, `{"error":"User not found"}`,
	}, {
		"/users/asdf", "GET", nil, func() {
			sqlmock.ExpectPrepare()
			sqlmock.ExpectQuery("select id, name from users where id=\\?").
				WithArgs("asdf").
				WillReturnRows(sqlmock.NewRows(userColumns).
				FromCSVString(""))
		}, 400, `{"error":"id must be int64"}`,
	}, {
		"/nonexistant", "GET", nil, func() {}, 404, `404 page not found`,
	},
}

func TestAllTheUsers(t *testing.T) {
	for _, d := range data {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(d.method, d.url, d.reqBody)
		db, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		d.expectations()
		router(&context{NewUserService(db)}).ServeHTTP(w, r)
		if w.Code != d.expectedCode {
			t.Errorf("expected %v got %v", d.expectedCode, w.Code)
		}
		if strings.TrimSpace(w.Body.String()) != d.expectedBody {
			t.Errorf("expected %v got %v", d.expectedBody, w.Body.String())
		}
	}
}
