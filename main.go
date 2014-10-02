package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	flag.Parse()
}

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
	r.Handle("/users", appHandler{appCtx, List}).Methods("GET")
	r.Handle("/users", appHandler{appCtx, Add}).Methods("POST")
	r.Handle("/users/{id}", appHandler{appCtx, Get}).Methods("GET")
	r.Handle("/users/{id}", appHandler{appCtx, Update}).Methods("PUT")
	r.Handle("/users/{id}", appHandler{appCtx, Delete}).Methods("DELETE")
	return r
}

func main() {
	db, err := sql.Open("sqlite3", "./gorunner.db")
	check(err)
	defer db.Close()

	// Bootstrap database
	_, err = db.Exec(`create table if not exists users (
		id integer not null primary key,
		name text,
		CHECK(name <> ''),
		UNIQUE(id, name)
	);`)
	check(err)
	_, err = db.Exec("insert or ignore into users values (1, 'admin')")
	check(err)

	// handle all requests by serving a file of the same name
	fileHandler := http.FileServer(http.Dir("static/"))

	r := router(&context{db})
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static", fileHandler))
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/404.html")
	})

	log.Println("Serving on", "0.0.0.0:8070")
	http.ListenAndServe("0.0.0.0:8070", r)
}

type User struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type Error struct {
	Error string `json:"error"`
}

func List(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	rows, err := c.db.Query("select id, name from users")
	check(err)
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		User := &User{}
		rows.Scan(&User.Id, &User.Name)
		users = append(users, User)
	}

	return http.StatusOK, users
}

func Add(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	if r.Body == nil {
		return http.StatusBadRequest, Error{"no payload"}
	}
	decoder := json.NewDecoder(r.Body)
	var User User
	err := decoder.Decode(&User)
	if err != nil {
		return http.StatusBadRequest, Error{"can't parse json payload"}
	}
	if User.Name == "" {
		return 422, Error{"please provide 'name'"}
	}

	stmt, err := c.db.Prepare("insert into users (name) values (?)") // TODO: limit
	check(err)
	result, err := stmt.Exec(User.Name)
	check(err)
	id, err := result.LastInsertId()
	check(err)
	User.Id = id
	return http.StatusCreated, User
}

func Get(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	vars := mux.Vars(r)

	stmt, err := c.db.Prepare("select id, name from users where id=?")
	check(err)
	var user User
	rows, err := stmt.Query(vars["id"])
	check(err)
	defer rows.Close()
	if !rows.Next() {
		return http.StatusNotFound, Error{fmt.Sprintf("can't find user with id %v", vars["id"])}
	}
	check(rows.Scan(&user.Id, &user.Name))

	return http.StatusOK, user
}

func Update(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	if r.Body == nil {
		return http.StatusBadRequest, Error{"no payload"}
	}
	vars := mux.Vars(r)
	var User User
	err := json.NewDecoder(r.Body).Decode(&User)
	if err != nil {
		return http.StatusBadRequest, Error{"can't parse json payload"}
	}
	if User.Name == "" {
		return 422, Error{"please provide 'name'"}
	}

	stmt, err := c.db.Prepare("update users set name=? where id=?")
	check(err)
	result, err := stmt.Exec(User.Name, vars["id"])
	check(err)
	i, err := result.RowsAffected()
	check(err)
	if i == 0 {
		return http.StatusNotFound, Error{fmt.Sprintf("can't find user with id %v", vars["id"])}
	}

	return http.StatusOK, map[string]string{"id": vars["id"]}
}

func Delete(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	vars := mux.Vars(r)

	stmt, err := c.db.Prepare("delete from users where id=?")
	check(err)
	result, err := stmt.Exec(vars["id"])
	check(err)
	i, err := result.RowsAffected()
	check(err)
	if i == 0 {
		return http.StatusNotFound, Error{fmt.Sprintf("can't find user with id %v", vars["id"])}
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
