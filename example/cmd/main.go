package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/rest"
	"github.com/jakecoffman/rest/example/user"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	db, err := sqlx.Connect("sqlite3", "./gorunner.db")
	check(err)
	defer db.Close()

	router := gin.Default()

	// Bootstrap database
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users
(
    id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    name TEXT NOT NULL CHECK(name <> "")
);
CREATE UNIQUE INDEX IF NOT EXISTS users_id_uindex ON users (id);`)
	check(err)
	_, err = db.Exec("insert or ignore into users values (1, 'admin')")
	check(err)
	_, err = db.Exec("insert or ignore into users values (2, 'bob')")
	check(err)

	userService := user.NewUserRepository(db)
	userController := rest.NewController(userService, user.User{})

	router.Handle("GET", "/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/users")
	})
	rest.StdRoutes(router.Group("/users"), userController)

	port := "0.0.0.0:8099"
	router.Run(port)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
