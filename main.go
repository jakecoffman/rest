package main

import (
	"database/sql"
	"encoding/json"
	"flag"
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
	userService UserService
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

	r := router(&context{NewUserService(db)})
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

type Error struct {
	Error string `json:"error"`
}

// TODO: Only call on errors that are unrecoverable as the server goes down
func check(err error) {
	if err != nil {
		log.Println(err)
		debug.PrintStack()
		log.Fatal()
	}
}
