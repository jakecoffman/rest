package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSomething(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/things", nil)

	db, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	columns := []string{"o_id", "o_name"}
	sqlmock.ExpectQuery("").WillReturnRows(sqlmock.NewRows(columns).FromCSVString("1,Bob"))

	c := &context{db}
	code, data := List(c, w, r)
	if code != 200 {
		t.Errorf("expected 200 got %v", code)
	}
	things := data.([]*Thing)
	if len(things) != 1 && things[0].Id != 1 && things[0].Name != "Bob" {
		t.Error(things)
	}
}
