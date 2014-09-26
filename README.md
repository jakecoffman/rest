golang-rest-bootstrap
=====================

[![Build Status](https://secure.travis-ci.org/jakecoffman/golang-rest-bootstrap.png?branch=master)](http://travis-ci.org/jakecoffman/golang-rest-bootstrap)


Bootstrap a RESTful web server in Golang

```
$ go get github.com/jakecoffman/golang-rest-bootstrap
$ cd $GOPATH/src/github.com/jakecoffman/golang-rest-bootstrap
$ go run main.go
```

This will serve static files from the static directory under the 
path /static and also serve /static/index.html from /.

This is a perfect start for an angular project.
