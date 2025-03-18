package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"

	 _ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/DSSD-Madison/gmu/db"
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

	levelStr, exist := os.LookupEnv("LOG_LEVEL")
	var level slog.Level
	switch levelStr {
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
		Mode:  mode,
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

	host := os.Getenv("PROD_HOST")
	user := os.Getenv("PROD_USER")
	dbname := os.Getenv("PROD_DB")
	password := os.Getenv("PROD_PASSWORD")

	if host == "" || user == "" || dbname == "" || password == "" {
		log.Fatal("Database environment variables are not set properly")
	}

	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		user, password, host, dbname,
	)

	// Connect to PostgreSQL using pgxpool
	dbpool, err := pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	// Create a *sql.DB instance using the pgx driver
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Unable to initialize sql.DB: %v", err)
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)

	// Static file handlers
	e.Static("/images", "static/images")
	e.Static("/css", "static/css")
	e.Static("/svg", "static/svg")

	// Renderer
	e.Renderer = models.NewTemplate()

	// Routes
	routes.InitRoutes(e, queries)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
