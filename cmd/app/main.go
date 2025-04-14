package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	"github.com/DSSD-Madison/gmu/pkg/config"
	"github.com/DSSD-Madison/gmu/pkg/db"
	db_util "github.com/DSSD-Madison/gmu/pkg/db/util"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/routes"
)

func main() {
	// --- Configuration ---
	appConfig, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading app config: %v", err)
	}
	dbConfig, err := db_util.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading database config: %v", err)
		os.Exit(1)
	}
	kendraConfig, err := awskendra.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading kendra config: %v", err)
		os.Exit(1)
	}

	// --- Logger Initialization ---
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
		Mode:      appConfig.Mode,
		Level:     level,
		AddSource: appConfig.Mode == "dev",
	}
	appLogger := logger.New(&loggerOpts)
	appLogger.Info("Logger initialized", "mode", appConfig.Mode, "level", level.String())

	// --- Database Initialization ---
	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		dbConfig.DBUser, dbConfig.DBPassword, dbConfig.DBHost, dbConfig.DBName,
	)

	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		appLogger.Error("Unable to initialize sql.DB", "error", err)
		os.Exit(1)
	}
	defer func(sqlDB *sql.DB) {
		if err := sqlDB.Close(); err != nil {
			appLogger.Error("Failed to close sql.DB", "error", err)
		}
	}(sqlDB)

	if err := sqlDB.PingContext(context.Background()); err != nil {
		appLogger.Error("Unable to ping database", "error", err)
		os.Exit(1)
	}
	appLogger.Info("Database connection established")

	dbQuerier := db.New(sqlDB)

	// --- AWS Kendra Initialization ---
	kendraClient, err := awskendra.NewKendraClient(*kendraConfig)
	if err != nil {
		appLogger.Error("Could not initialize kendra client", "err", err)
		os.Exit(1)
	}
	appLogger.Info("Kendra client initialized")

	queryQueue := awskendra.NewKendraQueryQueue(kendraClient, 2, 5)
	appLogger.Info("Kendra query queue initialized")

	// --- Echo Setup ---
	e := echo.New()

	// --- Middleware ---
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogRemoteIP: true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				appLogger.InfoContext(c.Request().Context(), "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("ip", v.RemoteIP),
				)
			} else {
				appLogger.ErrorContext(c.Request().Context(), "REQUEST ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
					slog.String("ip", v.RemoteIP),
				)
			}
			return nil
		},
	}))

	// --- Handler Initialization ---
	appLogger.Info("Initializing Handlers...")

	routesHandler := routes.NewHandler(dbQuerier, queryQueue, kendraClient, appLogger)

	appLogger.Info("Handlers initialized")

	// --- Routes Initialization ---
	routes.InitRoutes(e, routesHandler)

	e.Static("/images", "web/assets/images")
	e.Static("/css", "web/assets/css")
	e.Static("/svg", "web/assets/svg")
	e.Static("/js", "web/assets/js")

	// --- Start Server ---
	address := ":8080"
	appLogger.Info("Starting Server", "address", address)
	if err := e.Start(address); err != nil && err != http.ErrServerClosed {
		appLogger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
