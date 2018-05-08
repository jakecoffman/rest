package rest

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GinRest interface {
	List(*gin.Context)
	Get(*gin.Context)
	Add(*gin.Context)
	Update(*gin.Context)
	Delete(*gin.Context)
}

func StdRoutes(group *gin.RouterGroup, rest GinRest) {
	if rest.List != nil {
		group.Handle("GET", "", rest.List)
	}
	if rest.Get != nil {
		group.Handle("GET", "/:id", rest.Get)
	}
	if rest.Add != nil {
		group.Handle("POST", "", rest.Add)
	}
	if rest.Update != nil {
		group.Handle("PUT", "/:id", rest.Update)
	}
	if rest.Delete != nil {
		group.Handle("DELETE", "/:id", rest.Delete)
	}
}

type Error struct {
	Error string `json:"error"`
}

var ErrNotFound = errors.New("not found")

type Resource interface {
	Get() Resource
}

type Crud interface {
	List(params url.Values) (interface{}, error)
	Create(resource interface{}) (interface{}, error)
	Get(id int) (interface{}, error)
	Update(id int, resource interface{}) (interface{}, error)
	Delete(id int) error
}

type controller struct {
	GinRest
	resource Resource
	restful  Crud
}

func NewController(restful Crud, resource Resource) *controller {
	return &controller{
		restful:  restful,
		resource: resource,
	}
}

func (c controller) List(ctx *gin.Context) {
	if resources, err := c.restful.List(ctx.Request.URL.Query()); err != nil {
		ctx.JSON(http.StatusInternalServerError, Error{err.Error()})
	} else {
		ctx.JSON(http.StatusOK, resources)
	}
}

func (c controller) Add(ctx *gin.Context) {
	resource := c.resource.Get()
	if err := ctx.ShouldBindJSON(resource); err != nil {
		log.Println(err)
		ctx.JSON(http.StatusBadRequest, Error{"Can't parse JSON payload"})
		return
	}

	if u, err := c.restful.Create(resource); err != nil {
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

	user, err := c.restful.Get(id)
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
	if u, err := c.restful.Update(id, user); err != nil {
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
	if err = c.restful.Delete(id); err != nil {
		if err == ErrNotFound {
			ctx.JSON(http.StatusNotFound, Error{err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, Error{err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
