package rest

import (
	"database/sql"
	"log"
	"net/url"
)

type Repository interface {
	Lister(limit string) (interface{}, error)
	Getter(id int) (interface{}, error)
	Adder(resource interface{}) (sql.Result, error)
	Updater(id int, resource interface{}) (sql.Result, error)
	Deleter(id int) (sql.Result, error)
}

type SqliteRepository struct {
	Repository
}

func (r SqliteRepository) List(params url.Values) (interface{}, error) {
	limit := params.Get("limit")
	if limit == "" {
		limit = "-1"
	}
	results, err := r.Lister(limit)
	if err != nil {
		log.Println(err)
	}
	return results, err
}

func (r SqliteRepository) Create(resource interface{}) (interface{}, error) {
	result, err := r.Adder(resource)
	if err != nil {
		return resource, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return resource, err
	}
	return r.Get(int(id))
}

func (r SqliteRepository) Get(id int) (interface{}, error) {
	result, err := r.Getter(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return result, ErrNotFound
		}
		log.Println(err)
	}
	return result, err
}

func (r SqliteRepository) Update(id int, resource interface{}) (interface{}, error) {
	result, err := r.Updater(id, resource)
	if err != nil {
		return resource, err
	}
	i, err := result.RowsAffected()
	if err != nil {
		return resource, err
	}
	if i == 0 {
		return resource, ErrNotFound
	}
	return resource, nil
}

func (r SqliteRepository) Delete(id int) error {
	result, err := r.Deleter(id)
	if err != nil {
		return err
	}
	i, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if i == 0 {
		return ErrNotFound
	}
	return nil
}
