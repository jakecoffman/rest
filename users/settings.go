package users

import (
	"database/sql"

	"github.com/gorilla/mux"
)

func Init(r *mux.Router, db *sql.DB) {
	// Bootstrap database
	_, err := db.Exec(`create table if not exists users (
		id integer not null primary key,
		name text,
		CHECK(name <> ''),
		UNIQUE(id, name)
	);`)
	check(err)
	_, err = db.Exec("insert or ignore into users values (1, 'admin')")
	check(err)

	NewUserController(r, NewUserService(db))
}
