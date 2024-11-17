package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/DSSD-Madison/gmu/models"
	"github.com/DSSD-Madison/gmu/routes"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	// Static file handlers
	e.Static("/images", "static/images")
	e.Static("/css", "static/css")

	// Renderer
	e.Renderer = models.NewTemplate()

	// Routes
	routes.InitRoutes(e)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
