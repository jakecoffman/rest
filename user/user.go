package user

import (
	"errors"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/jakecoffman/golang-rest-bootstrap/lib"
	"net/url"
)

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (u User) Get() lib.Resource {
	return User{}
}

func (u User) Valid() error {
	if u.Name == "" {
		return errors.New("please provide `name`")
	}
	return nil
}

type Repository struct {
	db *sqlx.DB
}

func NewResource(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (s Repository) List(params url.Values) (interface{}, error) {
	users := []User{}
	limit := params.Get("limit")
	if limit == "" {
		limit = "-1"
	}
	err := s.db.Select(&users, "select id, name from users limit ?", limit)
	if err != nil {
		log.Println(err)
	}
	return users, err
}

func (s Repository) Add(resource interface{}) (interface{}, error) {
	result, err := s.db.NamedExec("insert into users (name) values (:name)", resource)
	if err != nil {
		return resource, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return resource, err
	}
	return s.Get(int(id))
}

func (s Repository) Get(id int) (interface{}, error) {
	var user User
	err := s.db.Get(&user, "select id, name from users where id=?", id)
	if err != nil {
		log.Println(err)
	}
	return user, err
}

func (s Repository) Update(id int, resource interface{}) (interface{}, error) {
	user := resource.(User)
	user.Id = id
	result, err := s.db.NamedExec("update users set name=? where id=?", user)
	if err != nil {
		return resource, err
	}
	i, err := result.RowsAffected()
	if err != nil {
		return resource, err
	}
	if i == 0 {
		return resource, lib.ErrNotFound
	}
	return resource, nil
}

func (s Repository) Delete(id int) error {
	result, err := s.db.Exec("delete from users where id=?", id)
	if err != nil {
		return err
	}
	i, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if i == 0 {
		return lib.ErrNotFound
	}
	return nil
}
