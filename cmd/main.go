package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/golang-rest-bootstrap/user"
	_ "github.com/mattn/go-sqlite3"
	"github.com/jmoiron/sqlx"
)

func main() {
	db, err := sqlx.Connect("sqlite3", "./gorunner.db")
	check(err)
	defer db.Close()

	router := gin.Default()

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

	userService := user.NewService(db)
	userController := user.NewController(userService)

	router.Handle("GET", "/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/users")
	})
	router.Handle("GET", "/users", userController.List)
	router.Handle("GET", "/users/:id", userController.Get)
	router.Handle("POST", "/users", userController.Add)
	router.Handle("PUT", "/users/:id", userController.Update)
	router.Handle("DELETE", "/users/:id", userController.Delete)

	port := "0.0.0.0:8099"
	router.Run(port)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
