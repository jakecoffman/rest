package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Error struct {
	Error string `json:"error"`
}

type UserController struct {
	userService UserService
}

func NewUserController(r *mux.Router, s UserService) *UserController {
	cont := UserController{s}
	r.Handle("/users", cont)
	r.Handle("/users/{id}", cont)
	return &cont
}

func (u UserController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := http.StatusMethodNotAllowed
	var data interface{}

	defer func(c int) {
		log.Println(r.URL, "-", r.Method, "-", code, r.RemoteAddr)
	}(code)

	if r.URL.Path == "/users" {
		switch r.Method {
		case "GET":
			code, data = u.List(w, r)
		case "POST":
			code, data = u.Add(w, r)
		default:
			return
		}
	} else {
		switch r.Method {
		case "GET":
			code, data = u.Get(w, r)
		case "PUT":
			code, data = u.Update(w, r)
		case "DELETE":
			code, data = u.Delete(w, r)
		default:
			return
		}
	}

	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println("Failed to write data:", err)
		code = http.StatusInternalServerError
	}
}

func (u UserController) List(w http.ResponseWriter, r *http.Request) (int, interface{}) {
	return http.StatusOK, u.userService.List()
}

func (u UserController) Add(w http.ResponseWriter, r *http.Request) (int, interface{}) {
	if r.Body == nil {
		return http.StatusBadRequest, Error{"no payload"}
	}
	decoder := json.NewDecoder(r.Body)
	var user User
	err := decoder.Decode(&user)
	if err != nil {
		return http.StatusBadRequest, Error{"can't parse json payload"}
	}
	if user.Name == "" {
		return 422, Error{"please provide 'name'"}
	}

	u.userService.Add(&user)
	return http.StatusCreated, user
}

func (u UserController) Get(w http.ResponseWriter, r *http.Request) (int, interface{}) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return http.StatusBadRequest, Error{"id must be int64"}
	}

	user, err := u.userService.Get(id)
	if err != nil {
		return http.StatusNotFound, Error{err.Error()}
	}
	return http.StatusOK, user
}

func (u UserController) Update(w http.ResponseWriter, r *http.Request) (int, interface{}) {
	if r.Body == nil {
		return http.StatusBadRequest, Error{"no payload"}
	}
	vars := mux.Vars(r)
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return http.StatusBadRequest, Error{"can't parse json payload"}
	}
	if user.Name == "" {
		return 422, Error{"please provide 'name'"}
	}
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return http.StatusBadRequest, Error{"id must be int64"}
	}
	user.Id = id

	err = u.userService.Update(&user)

	if err != nil {
		return http.StatusNotFound, Error{err.Error()}
	}

	return http.StatusOK, map[string]string{"id": vars["id"]}
}

func (u UserController) Delete(w http.ResponseWriter, r *http.Request) (int, interface{}) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return http.StatusBadRequest, Error{"id should be int64"}
	}
	err = u.userService.Delete(id)
	if err != nil {
		return http.StatusNotFound, Error{err.Error()}
	}

	return http.StatusOK, map[string]string{"id": vars["id"]}
}
