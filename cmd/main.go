package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/jakecoffman/golang-rest-bootstrap/users"
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	flag.Parse()
}

func main() {
	db, err := sql.Open("sqlite3", "./gorunner.db")
	check(err)
	defer db.Close()

	// handle all requests by serving a file of the same name
	fileHandler := http.FileServer(http.Dir("../static/"))

	r := mux.NewRouter()
	users.Init(r, db)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../static/index.html")
	})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static", fileHandler))
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../static/404.html")
	})

	port := os.Getenv("PORT")
	log.Println("Serving on", ":"+port)
	http.ListenAndServe(":"+port, context.ClearHandler(r))
}

func check(err error) {
	if err != nil {
		log.Println(err)
		debug.PrintStack()
		log.Fatal()
	}
}
