package main

import (
	"log"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/DSSD-Madison/gmu/db"
	"github.com/DSSD-Madison/gmu/internal"
	"github.com/DSSD-Madison/gmu/models"
	"github.com/DSSD-Madison/gmu/routes"
)

func main() {
	var logger *slog.Logger

	appConfig, err := internal.LoadConfig()
	if err != nil {
		log.Fatal("Error loading config: %q", err)
		os.Exit(1)
	}

	var level slog.Level
	switch appConfig.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	loggerOpts := internal.HandlerOptions{
		Mode:  appConfig.Mode,
		Level: level,
	}

	logger = slog.New(internal.NewHandler(&loggerOpts))

	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogRemoteIP: true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(c.Request().Context(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("ip", v.RemoteIP),
				)
			} else {
				logger.LogAttrs(c.Request().Context(), slog.LevelError, "REQUEST ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
					slog.String("ip", v.RemoteIP),
				)
			}
			return nil
		},
	}))

	dbConfig, err := db.LoadConfig()
	if err != nil {
		log.Fatalf("Unable to load db config: %q", err)
	}

	databaseURL := db.DBUrl(dbConfig)
	dbpool, err := db.NewPool(databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	sqlDB, err := db.NewDB(dbpool, databaseURL)
	defer sqlDB.Close()

	db_querier := db.NewQuerier(sqlDB)

	// Static file handlers
	e.Static("/images", "static/images")
	e.Static("/css", "static/css")
	e.Static("/svg", "static/svg")


	kendraConfig, err := models.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load AWS/Kendra config: %q", err)
	}
	kendraClient, err := models.NewKendraClient(*kendraConfig)
	if err != nil {
		log.Fatalf("Could not initialize kendra client: %q", err)
	}

	routesHandler := routes.NewHandler(db_querier, kendraClient)

	// Routes
	routes.InitRoutes(e, routesHandler)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
