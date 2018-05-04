package lib

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"net/url"
	"errors"
)

type Rest interface {
	List(*gin.Context)
	Get(*gin.Context)
	Add(*gin.Context)
	Update(*gin.Context)
	Delete(*gin.Context)
}

func StdRoutes(group *gin.RouterGroup, rest Rest) {
	group.Handle("GET", "", rest.List)
	group.Handle("GET", "/:id", rest.Get)
	group.Handle("POST", "", rest.Add)
	group.Handle("PUT", "/:id", rest.Update)
	group.Handle("DELETE", "/:id", rest.Delete)
}

type Error struct {
	Error string `json:"error"`
}

var ErrNotFound = errors.New("not found")

type Resource interface {
	Get() Resource
	Valid() error
}

type Repository interface {
	List(params url.Values) (interface{}, error)
	Add(resource interface{}) (interface{}, error)
	Get(id int) (interface{}, error)
	Update(id int, resource interface{}) (interface{}, error)
	Delete(id int) error
}

type controller struct {
	resource Resource
	repository Repository
}

func NewController(repository Repository, resource Resource) *controller {
	return &controller{
		repository: repository,
		resource: resource,
	}
}

func (c controller) List(ctx *gin.Context) {
	if resources, err := c.repository.List(ctx.Request.URL.Query()); err != nil {
		ctx.JSON(http.StatusInternalServerError, Error{err.Error()})
	} else {
		ctx.JSON(http.StatusOK, resources)
	}
}

func (c controller) Add(ctx *gin.Context) {
	resource := c.resource.Get()
	if err := ctx.ShouldBindJSON(resource); err != nil {
		ctx.JSON(http.StatusBadRequest, Error{"Can't parse JSON payload"})
		return
	}
	if err := resource.Valid(); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, Error{err.Error()})
		return
	}

	if u, err := c.repository.Add(resource); err != nil {
		ctx.JSON(http.StatusInternalServerError, Error{err.Error()})
	} else {
		ctx.JSON(http.StatusCreated, u)
	}
}

func (c controller) Get(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Error{"id must be int"})
		return
	}

	user, err := c.repository.Get(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, Error{err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (c controller) Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Error{"id must be int"})
		return
	}
	user := c.resource.Get()
	if err := ctx.ShouldBindJSON(user); err != nil {
		ctx.JSON(http.StatusBadRequest, Error{"can't parse json payload"})
		return
	}
	if err := user.Valid(); err != nil {
		ctx.JSON(http.StatusUnprocessableEntity, Error{err.Error()})
		return
	}
	if u, err := c.repository.Update(id, user); err != nil {
		if err == ErrNotFound {
			ctx.JSON(http.StatusNotFound, Error{err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, Error{err.Error()})
		}
		return
	} else {
		ctx.JSON(http.StatusOK, u)
	}
}

func (c controller) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, Error{"id must be int"})
		return
	}
	if err = c.repository.Delete(id); err != nil {
		if err == ErrNotFound {
			ctx.JSON(http.StatusNotFound, Error{err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, Error{err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
