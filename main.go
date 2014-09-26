package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type context struct {
	db *sql.DB
}

type appHandler struct {
	*context
	handler func(*context, http.ResponseWriter, *http.Request) (int, interface{})
}

func (t appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code, data := t.handler(t.context, w, r)
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	err := encoder.Encode(data)
	if err != nil {
		log.Println("Failed to write data:", err)
	}
	log.Println(r.URL, "-", r.Method, "-", code, r.RemoteAddr)
}

func main() {
	wd, err := os.Getwd()
	check(err)
	log.Println("Working directory", wd)

	os.Remove("./sqlite.db")
	db, err := sql.Open("sqlite3", "./sqlite.db")
	check(err)
	defer db.Close()

	_, err = db.Exec(`create table things (id integer not null primary key, name text);`)
	check(err)

	_, err = db.Exec("insert into things(name) values ('bob')")
	check(err)

	appCtx := &context{db}

	r := mux.NewRouter()
	r.Handle("/things", appHandler{appCtx, List})
	r.Handle("/things", appHandler{appCtx, Add})
	r.Handle("/things/{id}", appHandler{appCtx, Get})
	r.Handle("/things/{id}", appHandler{appCtx, Update})
	r.Handle("/things/{id}", appHandler{appCtx, Delete})
	http.ListenAndServe("0.0.0.0:8070", r)
}

type Thing struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func List(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	rows, err := c.db.Query("select id, name from things")
	check(err)
	defer rows.Close()

	things := []*Thing{}
	for rows.Next() {
		thing := &Thing{}
		rows.Scan(&thing.Id, &thing.Name)
		things = append(things, thing)
	}

	return http.StatusOK, things
}

func Add(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	stmt, err := c.db.Prepare("insert into things (name) values (?)")
	return http.StatusCreated, map[string]interface{}{"data": "1"}
}

func Get(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	vars := mux.Vars(r)
	// Get a specific item from the database
	return http.StatusOK, map[string]interface{}{"data": vars["id"]}
}

func Update(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	vars := mux.Vars(r)
	// Update an item in the database
	return http.StatusOK, map[string]interface{}{"data": vars["id"]}
}

func Delete(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	vars := mux.Vars(r)
	return http.StatusOK, map[string]string{"data": vars["id"]}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
