package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/codegangsta/negroni"
	"github.com/goincremental/negroni-sessions"
	"github.com/gorilla/mux"
	"github.com/jakecoffman/golang-rest-bootstrap/users"
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	flag.Parse() // TODO
}

func main() {
	db, err := sql.Open("sqlite3", "./gorunner.db")
	check(err)
	defer db.Close()

	// handle all requests by serving a file of the same name
	fileHandler := http.FileServer(http.Dir("static/"))

	r := mux.NewRouter()
	users.Init(r, db)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static", fileHandler))
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/404.html")
	})

	n := negroni.Classic()
	store := sessions.NewCookieStore([]byte("secret"))
	n.Use(sessions.Sessions("my_session", store))

	n.UseHandler(r)

	// enables gin autoreloading
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	n.Run(":" + port)
}

func check(err error) {
	if err != nil {
		log.Println(err)
		debug.PrintStack()
		log.Fatal()
	}
}
