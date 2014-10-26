package users

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type errorResponse struct {
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

	if r.URL.Path == "/users" {
		switch r.Method {
		case "GET":
			code, data = u.List(w, r)
		case "POST":
			code, data = u.Add(w, r)
		default:
			w.WriteHeader(code)
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
			w.WriteHeader(code)
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
		return http.StatusBadRequest, errorResponse{"no payload"}
	}
	decoder := json.NewDecoder(r.Body)
	var user User
	err := decoder.Decode(&user)
	if err != nil {
		return http.StatusBadRequest, errorResponse{"can't parse json payload"}
	}
	if user.Name == "" {
		return 422, errorResponse{"please provide 'name'"}
	}

	u.userService.Add(&user)
	return http.StatusCreated, user
}

func (u UserController) Get(w http.ResponseWriter, r *http.Request) (int, interface{}) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return http.StatusBadRequest, errorResponse{"id must be int64"}
	}

	user, err := u.userService.Get(id)
	if err != nil {
		return http.StatusNotFound, errorResponse{err.Error()}
	}
	return http.StatusOK, user
}

func (u UserController) Update(w http.ResponseWriter, r *http.Request) (int, interface{}) {
	if r.Body == nil {
		return http.StatusBadRequest, errorResponse{"no payload"}
	}
	vars := mux.Vars(r)
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return http.StatusBadRequest, errorResponse{"can't parse json payload"}
	}
	if user.Name == "" {
		return 422, errorResponse{"please provide 'name'"}
	}
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return http.StatusBadRequest, errorResponse{"id must be int64"}
	}
	user.Id = id

	err = u.userService.Update(&user)

	if err != nil {
		return http.StatusNotFound, errorResponse{err.Error()}
	}

	return http.StatusOK, map[string]string{"id": vars["id"]}
}

func (u UserController) Delete(w http.ResponseWriter, r *http.Request) (int, interface{}) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return http.StatusBadRequest, errorResponse{"id should be int64"}
	}
	err = u.userService.Delete(id)
	if err != nil {
		return http.StatusNotFound, errorResponse{err.Error()}
	}

	return http.StatusOK, map[string]string{"id": vars["id"]}
}
