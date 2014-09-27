package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"

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
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println("Failed to write data:", err)
	}
	log.Println(r.URL, "-", r.Method, "-", code, r.RemoteAddr)
}

func router(appCtx *context) *mux.Router {
	r := mux.NewRouter()
	// There are actually 2 options here: 1 - have 2 functions that handle both cases and
	// internally switch on the method or 2 - have a function for each method
	r.Handle("/things", appHandler{appCtx, List}).Methods("GET")
	r.Handle("/things", appHandler{appCtx, Add}).Methods("POST")
	r.Handle("/things/{id}", appHandler{appCtx, Get}).Methods("GET")
	r.Handle("/things/{id}", appHandler{appCtx, Update}).Methods("PUT")
	r.Handle("/things/{id}", appHandler{appCtx, Delete}).Methods("DELETE")
	return r
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
	r := router(appCtx)
	http.ListenAndServe("0.0.0.0:8070", r)
}

type Thing struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type Error struct {
	Error string `json:"error"`
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
	if r.Body == nil {
		return http.StatusBadRequest, Error{"no payload"}
	}
	decoder := json.NewDecoder(r.Body)
	var thing Thing
	err := decoder.Decode(&thing)
	if err != nil {
		return http.StatusBadRequest, Error{"can't parse json payload"}
	}
	if thing.Name == "" {
		return 422, Error{"please provide 'name'"}
	}

	stmt, err := c.db.Prepare("insert into things (name) values (?)") // TODO: limit
	check(err)
	result, err := stmt.Exec(thing.Name)
	check(err)
	id, err := result.LastInsertId()
	check(err)
	thing.Id = id
	return http.StatusCreated, thing
}

func Get(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	vars := mux.Vars(r)

	stmt, err := c.db.Prepare("select id, name from things where id=?")
	check(err)
	var thing Thing
	err = stmt.QueryRow(vars["id"]).Scan(&thing.Id, &thing.Name)
	check(err)

	return http.StatusOK, thing
}

func Update(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	if r.Body == nil {
		return http.StatusBadRequest, Error{"no payload"}
	}
	vars := mux.Vars(r)
	var thing Thing
	err := json.NewDecoder(r.Body).Decode(&thing)
	if err != nil {
		return http.StatusBadRequest, Error{"can't parse json payload"}
	}
	if thing.Name == "" {
		return 422, Error{"please provide 'name'"}
	}

	stmt, err := c.db.Prepare("update things set name=? where id=?")
	check(err)
	result, err := stmt.Exec(thing.Name, vars["id"])
	check(err)
	i, err := result.RowsAffected()
	check(err)
	if i == 0 {
		return http.StatusNotFound, Error{fmt.Sprintf("can't find thing with id %v", vars["id"])}
	}

	return http.StatusOK, map[string]string{"id": vars["id"]}
}

func Delete(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	vars := mux.Vars(r)

	stmt, err := c.db.Prepare("delete from things where id=?")
	check(err)
	result, err := stmt.Exec(vars["id"])
	check(err)
	i, err := result.RowsAffected()
	check(err)
	if i == 0 {
		return http.StatusNotFound, Error{fmt.Sprintf("can't find thing with id %v", vars["id"])}
	}

	return http.StatusOK, map[string]string{"id": vars["id"]}
}

// TODO: Only call on errors that are unrecoverable as the server goes down
func check(err error) {
	if err != nil {
		log.Println(err)
		debug.PrintStack()
		log.Fatal()
	}
}
