package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

var thingColumns = []string{"o_id", "o_name"}

var data = []struct {
	url          string
	method       string
	reqBody      io.Reader
	expectations func()
	expectedCode int
	expectedBody string
}{
	{
		"/things", "GET", nil, func() {
			sqlmock.ExpectQuery("select id, name from things").
				WillReturnRows(sqlmock.NewRows(thingColumns).
				FromCSVString("1,Bob\n2,Ted"))
		}, 200, `[{"id":1,"name":"Bob"},{"id":2,"name":"Ted"}]`,
	},
	{
		"/things", "POST", strings.NewReader(`{"name":"test"}`), func() {
			sqlmock.ExpectPrepare()
			sqlmock.ExpectExec("insert into things \\(name\\) values \\(\\?\\)").
				WithArgs("test").
				WillReturnResult(sqlmock.NewResult(7, 1))
		}, 201, `{"id":7,"name":"test"}`,
	},
	{
		"/things/3", "GET", nil, func() {
			sqlmock.ExpectPrepare()
			// TODO: I can't .WithArgs("3") here for some reason with the QueryRow call
			sqlmock.ExpectQuery("select id, name from things where id=\\?").
				WillReturnRows(sqlmock.NewRows(thingColumns).AddRow(3, "Jim"))
		}, 200, `{"id":3,"name":"Jim"}`,
	}, {
		"/things/2", "PUT", strings.NewReader(`{"name":"Bob"}`), func() {
			sqlmock.ExpectPrepare()
			sqlmock.ExpectExec("update things set name=\\? where id=\\?").
				WithArgs("Bob", "2").
				WillReturnResult(sqlmock.NewResult(0, 1))
		}, 200, `{"id":"2"}`,
	}, {
		"/things/3", "DELETE", nil, func() {
			sqlmock.ExpectPrepare()
			sqlmock.ExpectExec("delete from things where id=\\?").
				WithArgs("3").
				WillReturnResult(sqlmock.NewResult(0, 1))
		}, 200, `{"id":"3"}`,
	},
}

func TestAllTheThings(t *testing.T) {
	for _, d := range data {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(d.method, d.url, d.reqBody)
		db, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		d.expectations()
		router(&context{db}).ServeHTTP(w, r)
		if w.Code != d.expectedCode {
			t.Fatalf("expected %v got %v", d.expectedCode, w.Code)
		}
		if strings.TrimSpace(w.Body.String()) != d.expectedBody {
			t.Fatalf("expected %v got %v", d.expectedBody, w.Body.String())
		}
	}
}
