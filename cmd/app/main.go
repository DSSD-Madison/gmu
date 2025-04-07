package main

import (
	"log"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	"github.com/DSSD-Madison/gmu/pkg/config"
	db_util "github.com/DSSD-Madison/gmu/pkg/db/util"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/routes"
)

func main() {
	var logHandler *slog.Logger

	appConfig, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Error loading .env file")
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

	loggerOpts := logger.HandlerOptions{
		Mode:  appConfig.Mode,
		Level: level,
	}

	logHandler = slog.New(logger.NewHandler(&loggerOpts))

	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogRemoteIP: true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logHandler.LogAttrs(c.Request().Context(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("ip", v.RemoteIP),
				)
			} else {
				logHandler.LogAttrs(c.Request().Context(), slog.LevelError, "REQUEST ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
					slog.String("ip", v.RemoteIP),
				)
			}
			return nil
		},
	}))

	dbh, err := db_util.NewDBHandler()
	if err != nil {
		logHandler.Error("Failed to initialize DB Handler", "err", err)
		os.Exit(1)
	}

	defer func(dbh *db_util.DBHandler) {
		err := dbh.Close()
		if err != nil {
			logHandler.Error("Failed to close DB connections", "err", err)
		}
	}(dbh)

	kendraConfig, err := awskendra.LoadConfig()
	if err != nil {
		logHandler.Error("Could not load AWS Kendra config", "err", err)
		os.Exit(1)
	}

	kendraClient, err := awskendra.NewKendraClient(*kendraConfig)
	if err != nil {
		logHandler.Error("Could not initialize kendra client", "err", err)
		os.Exit(1)
	}

	routesHandler := routes.NewHandler(dbh.Querier, kendraClient, logHandler)

	// Static file handlers
	e.Static("/images", "web/assets/images")
	e.Static("/css", "web/assets/css")
	e.Static("/svg", "web/assets/svg")

	// Routes
	routes.InitRoutes(e, routesHandler)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
