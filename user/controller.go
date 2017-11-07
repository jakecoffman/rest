package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Error struct {
	Error string `json:"error"`
}

type controller struct {
	userService Service
}

func NewController(u Service) *controller {
	return &controller{
		userService: u,
	}
}

func (u controller) List(c *gin.Context) {
	if users, err := u.userService.List(); err != nil {
		c.JSON(http.StatusInternalServerError, Error{err.Error()})
	} else {
		c.JSON(http.StatusOK, users)
	}
}

func (u controller) Add(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, Error{"Can't parse JSON payload"})
		return
	}
	if user.Name == "" {
		c.JSON(http.StatusUnprocessableEntity, Error{"please provide 'name'"})
		return
	}

	if u, err := u.userService.Add(user); err != nil {
		c.JSON(http.StatusInternalServerError, Error{err.Error()})
	} else {
		c.JSON(http.StatusCreated, u)
	}
}

func (u controller) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Error{"id must be int"})
		return
	}

	user, err := u.userService.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, Error{err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (u controller) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Error{"id must be int"})
		return
	}
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, Error{"can't parse json payload"})
		return
	}
	if user.Name == "" {
		c.JSON(http.StatusUnprocessableEntity, Error{"please provide 'name'"})
		return
	}
	user.Id = id

	if u, err := u.userService.Update(user); err != nil {
		c.JSON(http.StatusNotFound, Error{err.Error()})
	} else {
		c.JSON(http.StatusOK, u)
	}
}

func (u controller) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Error{"id must be int"})
		return
	}
	if err = u.userService.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, Error{err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
