package user

import (
	"database/sql"

	"github.com/jakecoffman/rest"
	"github.com/jmoiron/sqlx"
)

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (u User) Get() rest.Resource {
	return &User{}
}

type Repository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *rest.SqliteRepository {
	return &rest.SqliteRepository{
		Repository: Repository{db: db},
	}
}

func (r Repository) Lister(limit string) (interface{}, error) {
	users := []User{}
	err := r.db.Select(&users, "select id, name from users limit ?", limit)
	return users, err
}

func (r Repository) Getter(id int) (interface{}, error) {
	var user User
	err := r.db.Get(&user, "select id, name from users where id=?", id)
	return user, err
}

func (r Repository) Adder(resource interface{}) (sql.Result, error) {
	return r.db.NamedExec("insert into users (name) values (:name)", resource)
}

func (r Repository) Updater(id int, resource interface{}) (sql.Result, error) {
	user := resource.(*User)
	user.Id = id
	return r.db.Exec("update users set name=? where id=?", user.Name, id)
}

func (r Repository) Deleter(id int) (sql.Result, error) {
	return r.db.Exec("delete from users where id=?", id)
}
