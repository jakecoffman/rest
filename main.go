package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"
	"github.com/jakecoffman/golang-rest-bootstrap/users"
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	flag.Parse()
}

func router(db *sql.DB) *mux.Router {
	r := mux.NewRouter()
	users.NewUserController(r, users.NewUserService(db))
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

	r := router(db)
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

func check(err error) {
	if err != nil {
		log.Println(err)
		debug.PrintStack()
		log.Fatal()
	}
}
