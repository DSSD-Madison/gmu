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
	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	db_util "github.com/DSSD-Madison/gmu/pkg/db/util"
	"github.com/DSSD-Madison/gmu/pkg/handlers"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/services"
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
	awsConfig, err := awskendra.LoadConfig()
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
	kendraClient, err := awskendra.New(*awsConfig, appLogger)
	if err != nil {
		appLogger.Error("Could not initialize kendra client", "error", err)
		os.Exit(1)
	}
	appLogger.Info("Kendra client initialized")

	// TODO: Add DI and make an interface
	bedrockClient, err := awskendra.NewBedrockClient(*awsConfig)
	if err != nil {
		appLogger.Error("Could not initialize bedrock client", "error", err)
		os.Exit(1)
	}

	s3Client, err := awskendra.NewS3Client(*awsConfig)
	if err != nil {
		log.Fatalf("failed to create S3 client: %v", err)
		os.Exit(1)
	}
	
	appLogger.Info("Bedrock client initialized")

	appLogger.Info("Initializing services...")

	searchService := services.NewSearchService(appLogger, kendraClient, dbQuerier)
	suggestionService := services.NewSuggestionService(appLogger, kendraClient)
	// loginService := services.NewLoginService(appLogger, dbQuerier)
	authenticationService := services.NewLoginService(appLogger, dbQuerier)
	bedrockService := services.NewBedrockService(appLogger, *bedrockClient)
	userService := services.NewUserService(appLogger, dbQuerier)
	fileManagerService := services.NewFilemanagerService(appLogger, s3Client)

	appLogger.Info("Services initialized")

	// --- Handler Initialization ---
	appLogger.Info("Initializing Handlers...")

	homeHandler := handlers.NewHomeHandler(appLogger)
	searchHandler := handlers.NewSearchHandler(appLogger, searchService)
	authHandler := handlers.NewAuthenticationHandler(appLogger, userService, authenticationService)
	suggestionsHandler := handlers.NewSuggestionsHandler(appLogger, suggestionService)
	uploadHandler := handlers.NewUploadHandler(appLogger, dbQuerier, bedrockService, fileManagerService)
	userManagementHandler := handlers.NewUserManagementHandler(appLogger, dbQuerier)
	databaseHandler := handlers.NewDatabaseHandler(appLogger, dbQuerier)

	appLogger.Info("Handlers initialized")

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
			case "/", "/search", "/search/suggestions", "/login", "/logout", "/document-search":
				return true
			default:
				return false
			}
		},
	}))
	// --- Routes Initialization ---
	routes.RegisterAuthenticationRoutes(e, authHandler)
	routes.RegisterDatabaseRoutes(e, databaseHandler)
	routes.RegisterHomeRoutes(e, homeHandler)
	routes.RegisterSearchRoutes(e, searchHandler)
	routes.RegisterSuggestionsRoutes(e, suggestionsHandler)
	routes.RegisterUploadRoutes(e, uploadHandler)
	routes.RegisterUserManagementRoutes(e, userManagementHandler)
	appLogger.Info("Routes initialized")

	e.Static("/images", "web/assets/images")
	e.Static("/css", "web/assets/css")
	e.Static("/svg", "web/assets/svg")
	e.Static("/js", "web/assets/js")
	e.Static("/favicon", "web/assets/favicon")

	// --- Start Server ---
	address := ":8080"
	appLogger.Info("Starting Server", "address", address)
	if err := e.Start(address); err != nil && err != http.ErrServerClosed {
		appLogger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
