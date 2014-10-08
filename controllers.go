package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func List(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	return http.StatusOK, c.userService.List()
}

func Add(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
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

	c.userService.Add(&user)
	return http.StatusCreated, user
}

func Get(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return http.StatusBadRequest, Error{"id must be int64"}
	}

	user, err := c.userService.Get(id)
	if err != nil {
		return http.StatusNotFound, Error{err.Error()}
	}
	return http.StatusOK, user
}

func Update(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
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

	err = c.userService.Update(&user)

	if err != nil {
		return http.StatusNotFound, Error{err.Error()}
	}

	return http.StatusOK, map[string]string{"id": vars["id"]}
}

func Delete(c *context, w http.ResponseWriter, r *http.Request) (int, interface{}) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return http.StatusBadRequest, Error{"id should be int64"}
	}
	err = c.userService.Delete(id)
	if err != nil {
		return http.StatusNotFound, Error{err.Error()}
	}

	return http.StatusOK, map[string]string{"id": vars["id"]}
}
