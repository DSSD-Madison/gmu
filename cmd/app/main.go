package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	"github.com/DSSD-Madison/gmu/pkg/config"
	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
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

	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "form:_csrf",
		CookieName:     "csrf",
		CookiePath:     "/",
		CookieDomain:   "",
		ContextKey:     "csrf",
		CookieSameSite: http.SameSiteStrictMode,
		CookieSecure:   appConfig.Mode == "prod", // Only set secure cookies in prod
		Skipper: func(c echo.Context) bool {
			switch c.Path() {
			case "/", "/search", "/search/suggestions", "/login", "/logout":
				return true
			default:
				return false
			}
		},
	}))

	dbConfig, err := db_util.LoadConfig()
	if err != nil {
		logHandler.Error("Unable to load db config", "err", err)
		os.Exit(1)
	}

	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		dbConfig.DBUser, dbConfig.DBPassword, dbConfig.DBHost, dbConfig.DBName,
	)

	// Connect to PostgreSQL using pgxpool
	dbpool, err := pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		logHandler.Error("Unable to connect to database", "err", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	// Create a *sql.DB instance using the pgx driver
	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		logHandler.Error("Unable to initialize sql.DB", "err", err)
		os.Exit(1)
	}
	defer func(sqlDB *sql.DB) {
		err := sqlDB.Close()
		if err != nil {
			logHandler.Error("Failed to close sql.DB", "err", err)
		}
	}(sqlDB)

	dbQuerier := db.New(sqlDB)

	awsConfig, err := awskendra.LoadConfig()
	if err != nil {
		logHandler.Error("Could not load AWS Kendra config", "err", err)
		os.Exit(1)
	}

	kendraClient, err := awskendra.NewKendraClient(*awsConfig)
	bedrockClient, err := awskendra.NewBedrockClient(*awsConfig)
	if err != nil {
		logHandler.Error("Could not initialize kendra client", "err", err)
		os.Exit(1)
	}

	routesHandler := routes.NewHandler(dbQuerier, kendraClient, bedrockClient, logHandler)

	// Static file handlers
	e.Static("/images", "web/assets/images")
	e.Static("/css", "web/assets/css")
	e.Static("/svg", "web/assets/svg")
	e.Static("/js", "web/assets/js")

	// Routes
	routes.InitRoutes(e, routesHandler)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
