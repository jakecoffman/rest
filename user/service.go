package user

import (
	"errors"
	"log"

	"github.com/jmoiron/sqlx"
)

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Service interface {
	List() ([]User, error)
	Add(User) (User, error)
	Get(int) (User, error)
	Update(User) (User, error)
	Delete(int) error
}

type service struct {
	db *sqlx.DB
}

func NewService(db *sqlx.DB) Service {
	return &service{db}
}

func (s service) List() ([]User, error) {
	users := []User{}
	err := s.db.Select(&users, "select id, name from users")
	if err != nil {
		log.Println(err)
	}
	return users, err
}

func (s service) Add(user User) (User, error) {
	stmt, err := s.db.Prepare("insert into users (name) values (?)") // TODO: limit
	if err != nil {
		return user, err
	}
	result, err := stmt.Exec(user.Name)
	if err != nil {
		return user, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return user, err
	}
	user.Id = int(id)
	return user, err
}

func (s service) Get(id int) (User, error) {
	var u User
	err := s.db.Get(&u, "select id, name from users where id=?", id)
	if err != nil {
		log.Println(err)
	}
	return u, err
}

func (s service) Update(user User) (User, error) {
	stmt, err := s.db.Prepare("update users set name=? where id=?")
	if err != nil {
		return user, err
	}
	result, err := stmt.Exec(user.Name, user.Id)
	if err != nil {
		return user, err
	}
	i, err := result.RowsAffected()
	if err != nil {
		return user, err
	}
	if i == 0 {
		return user, errors.New("user not found")
	}
	return user, nil
}

func (s service) Delete(id int) error {
	stmt, err := s.db.Prepare("delete from users where id=?")
	if err != nil {
		return err
	}
	result, err := stmt.Exec(id)
	if err != nil {
		return err
	}
	i, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if i == 0 {
		return errors.New("user not found")
	}
	return nil
}
