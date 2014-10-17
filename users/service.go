package users

import (
	"database/sql"
	"errors"
	"log"
	"runtime/debug"
)

type User struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type UserService interface {
	List() []*User
	Add(*User)
	Get(int64) (*User, error)
	Update(*User) error
	Delete(int64) error
}

type userService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) UserService {
	return &userService{db}
}

func (s userService) List() []*User {
	rows, err := s.db.Query("select id, name from users")
	check(err)
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		User := &User{}
		rows.Scan(&User.Id, &User.Name)
		users = append(users, User)
	}
	return users
}

func (s userService) Add(user *User) {
	stmt, err := s.db.Prepare("insert into users (name) values (?)") // TODO: limit
	check(err)
	result, err := stmt.Exec(user.Name)
	check(err)
	id, err := result.LastInsertId()
	check(err)
	user.Id = id
}

func (s userService) Get(id int64) (*User, error) {
	stmt, err := s.db.Prepare("select id, name from users where id=?")
	check(err)
	var user User
	rows, err := stmt.Query(id)
	check(err)
	defer rows.Close()
	if !rows.Next() {
		return nil, errors.New("User not found")
	}
	check(rows.Scan(&user.Id, &user.Name))
	return &user, nil
}

func (s userService) Update(user *User) error {
	stmt, err := s.db.Prepare("update users set name=? where id=?")
	check(err)
	result, err := stmt.Exec(user.Name, user.Id)
	check(err)
	i, err := result.RowsAffected()
	check(err)
	if i == 0 {
		return errors.New("User not found")
	}
	return nil
}

func (s userService) Delete(id int64) error {
	stmt, err := s.db.Prepare("delete from users where id=?")
	check(err)
	result, err := stmt.Exec(id)
	check(err)
	i, err := result.RowsAffected()
	check(err)
	if i == 0 {
		return errors.New("User not found")
	}
	return nil
}

// TODO: Only call on errors that are unrecoverable as the server goes down
// or handle panics and panic here
func check(err error) {
	if err != nil {
		log.Println(err)
		debug.PrintStack()
		log.Fatal()
	}
}
