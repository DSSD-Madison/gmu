package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/DSSD-Madison/gmu/internal/application"
	"github.com/DSSD-Madison/gmu/internal/infra/aws/bedrock"
	"github.com/DSSD-Madison/gmu/internal/infra/aws/kendra"
	"github.com/DSSD-Madison/gmu/internal/infra/aws/s3"
	db "github.com/DSSD-Madison/gmu/internal/infra/database/sqlc/generated"
	"github.com/DSSD-Madison/gmu/internal/infra/http/handlers"
	"github.com/DSSD-Madison/gmu/internal/infra/http/routes"
	"github.com/DSSD-Madison/gmu/pkg/config"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/ratelimiter"
)

const (
	sessionCookieName = "gmu_session"

	ipMaxAttempts   = 10
	ipBlockDuration = 5 * time.Minute
	ipWindow        = 1 * time.Minute

	userMaxAttempts   = 5
	userBlockDuration = 15 * time.Minute
	userWindow        = 5 * time.Minute
)

func main() {
	// --- Configuration ---
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading app config: %v", err)
	}

	// --- Logger Initialization ---
	var level slog.Level
	switch cfg.LogLevel {
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
		Mode:      cfg.Mode,
		Level:     level,
		AddSource: cfg.Mode == "dev",
	}
	appLogger := logger.New(&loggerOpts)
	appLogger.Info("Logger initialized", "mode", cfg.Mode, "level", level.String())

	// --- Database Initialization ---
	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Name,
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

	dbClient := db.New(sqlDB)

	// --- AWS Kendra Initialization ---
	kendraClient, err := kendra.NewClient(*cfg, appLogger)
	if err != nil {
		appLogger.Error("Could not initialize kendra client", "error", err)
		os.Exit(1)
	}
	appLogger.Info("Kendra client initialized")

	// TODO: Add DI and make an interface
	bedrockClient, err := bedrock.NewBedrockClient(*cfg)
	if err != nil {
		appLogger.Error("Could not initialize bedrock client", "error", err)
		os.Exit(1)
	}

	s3Client, err := s3.NewS3Client(*cfg)
	if err != nil {
		log.Fatalf("failed to create S3 client: %v", err)
		os.Exit(1)
	}

	appLogger.Info("Bedrock client initialized")

	appLogger.Info("Initializing Session Store...")
	sessionSecretKey := os.Getenv("SESSION_SECRET_KEY")
	if sessionSecretKey == "" {
		if cfg.Mode == "prod" {
			appLogger.Error("SESSION_SECRET_KEY environment variable not set in prod. Exiting now.")
			os.Exit(1)
		}
		appLogger.Warn("SESSION_SECRET_KEY environment variable not set. Using insecure default (dev only).")
		sessionSecretKey = "insecure-default-key-for-dev-only-change-me"
	}

	cookieStore := sessions.NewCookieStore([]byte(sessionSecretKey))
	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   cfg.Mode == "prod",
		SameSite: http.SameSiteLaxMode,
	}
	appLogger.Info("Session Store initialized", "name", sessionCookieName, "secure", cookieStore.Options.Secure)

	// --- Service Initialization ---
	appLogger.Info("Initializing services...")

	// Rate Limiters
	ipRateLimiter := ratelimiter.NewInMemoryRateLimiter(appLogger, ipMaxAttempts, ipBlockDuration, ipWindow)
	userRateLimiter := ratelimiter.NewInMemoryRateLimiter(appLogger, userMaxAttempts, userBlockDuration, userWindow)
	appLogger.Debug("Rate Limiters initialized")

	sessionManager, err := application.NewGorillaSessionManager(cookieStore, sessionCookieName, appLogger, dbClient)
	if err != nil {
		appLogger.Error("Failed to create session manager", "error", err)
		os.Exit(1)
	}

	userService := application.NewUserService(appLogger, dbClient)
	authenticationService := application.NewLoginService(appLogger, ipRateLimiter, userRateLimiter, userService)
	searchService := application.NewSearchService(appLogger, kendraClient, dbClient)
	suggestionService := application.NewSuggestionService(appLogger, kendraClient)
	bedrockService := application.NewBedrockService(appLogger, *bedrockClient)
	fileManagerService := application.NewFilemanagerService(appLogger, s3Client)

	appLogger.Info("Services initialized")

	// --- Handler Initialization ---
	appLogger.Info("Initializing Handlers...")

	homeHandler := handlers.NewHomeHandler(appLogger, sessionManager)
	searchHandler := handlers.NewSearchHandler(appLogger, searchService, sessionManager)
	authHandler := handlers.NewAuthenticationHandler(appLogger, sessionManager, authenticationService)
	suggestionsHandler := handlers.NewSuggestionsHandler(appLogger, suggestionService)
	uploadHandler := handlers.NewUploadHandler(appLogger, dbClient, bedrockService, fileManagerService, sessionManager)
	userManagementHandler := handlers.NewUserManagementHandler(appLogger, dbClient, sessionManager)
	databaseHandler := handlers.NewDatabaseHandler(appLogger, dbClient)

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
		CookieSameSite: http.SameSiteLaxMode,
		CookieSecure:   cfg.Mode == "prod", // Only set secure cookies in prod
		Skipper: func(c echo.Context) bool {
			path := c.Path()
			if path == "/search/suggestions" {
				return true
			}
			return false
		},
	}))

	// --- Routes Initialization ---
	routes.RegisterAuthenticationRoutes(e, authHandler)
	routes.RegisterDatabaseRoutes(e, databaseHandler, sessionManager)
	routes.RegisterHomeRoutes(e, homeHandler)
	routes.RegisterSearchRoutes(e, searchHandler)
	routes.RegisterSuggestionsRoutes(e, suggestionsHandler)
	routes.RegisterUploadRoutes(e, uploadHandler, sessionManager)
	routes.RegisterUserManagementRoutes(e, userManagementHandler, sessionManager)
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
