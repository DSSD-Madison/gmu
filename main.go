package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/", func(e echo.Context) error {
		return e.String(200, "Hello")
	})

	e.GET("/test", func(e echo.Context) error {
		return e.String(200, "New element")
	})
}
