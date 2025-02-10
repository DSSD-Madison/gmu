package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/DSSD-Madison/gmu/internal"
	"github.com/DSSD-Madison/gmu/models"
	"github.com/DSSD-Madison/gmu/routes"
)

func main() {

	var logger *slog.Logger

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mode, exist := os.LookupEnv("MODE")
	if !exist {
		mode = "dev"
	}

	switch mode {
	case "dev":
		logger = slog.New(internal.NewPrettyHandler(nil))
	case "prod":
		logger = slog.New(internal.NewHandler(nil))
	}

	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogRemoteIP: true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("ip", v.RemoteIP),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
					slog.String("ip", v.RemoteIP),
				)
			}
			return nil
		},
	}))
	// Static file handlers
	e.Static("/images", "static/images")
	e.Static("/css", "static/css")
	e.Static("/svg", "static/svg")

	// Renderer
	e.Renderer = models.NewTemplate()

	// Routes
	routes.InitRoutes(e)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
