package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

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
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	sessionCookieName = "gmu_session"
	ipMaxAttempts     = 10
	ipBlockDuration   = 5 * time.Minute
	ipWindow          = 1 * time.Minute
	userMaxAttempts   = 5
	userBlockDuration = 15 * time.Minute
	userWindow        = 5 * time.Minute
	shutdownTimeout   = 15 * time.Second
)

type Server struct {
	config  *config.Config
	log     logger.Logger
	db      *sql.DB
	queries *db.Queries
	echo    *echo.Echo

	// Core Infra
	kendraClient    kendra.Client
	bedrockClient   *bedrock.BedrockClient
	s3Client        *s3.S3Client
	cookieStore     sessions.Store
	ipRateLimiter   ratelimiter.RateLimiter
	userRateLimiter ratelimiter.RateLimiter
	sessionManager  application.SessionManager

	// services
	authService    application.AuthenticationManager
	searchService  application.Searcher
	suggestService application.Suggester
	bedrockService *application.BedrockService
	fileManService *application.FilemanagerService
	userService    application.UserManager

	// Handlers
	HomeHandler           *handlers.HomeHandler
	SearchHandler         *handlers.SearchHandler
	AuthHandler           *handlers.AuthenticationHandler
	SuggestionsHandler    *handlers.SuggestionsHandler
	UploadHandler         *handlers.UploadHandler
	UserManagementHandler *handlers.UserManagementHandler
	DatabaseHandler       *handlers.DatabaseHandler
}

func NewServer(cfg *config.Config) (*Server, error) {
	srv := &Server{
		config: cfg,
	}

	// --- Logger ---
	if err := srv.setupLogger(); err != nil {
		return nil, fmt.Errorf("setupLogger failed: %w", err)
	}
	srv.log.Info("Logger initialized", "mode", cfg.Mode, "level", cfg.LogLevel)

	// --- Database ---
	if err := srv.setupDatabase(); err != nil {
		return nil, fmt.Errorf("setupDatabase failed: %w", err)
	}
	srv.log.Info("Database connection established")

	// --- AWS Clients ---
	if err := srv.setupAWSClients(); err != nil {
		return nil, fmt.Errorf("setupAWSClients failed: %w", err)
	}
	srv.log.Info("AWS Clients initialized (Kendra, Bedrock, S3)")

	// --- Session Store ---
	if err := srv.setupSessionStore(); err != nil {
		return nil, fmt.Errorf("setupSessionStore failed: %w", err)
	}
	srv.log.Info("Session Store initialized")

	// --- Rate Limiters ---
	srv.setupRateLimiters()
	srv.log.Info("Rate Limiters initialized")

	if err := srv.setupServices(); err != nil {
		return nil, fmt.Errorf("setupServices failed: %w", err)
	}
	srv.log.Info("Services initialized")

	srv.setupHandlers()
	srv.log.Info("Handlers initialized")

	srv.echo = echo.New()

	return srv, nil
}

func (s *Server) LogEvent(message string, fields ...any) {
	s.log.Warn(message, fields)
}

func (s *Server) setupLogger() error {
	var level slog.Level
	switch s.config.LogLevel {
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
		Mode:      s.config.Mode,
		Level:     level,
		AddSource: s.config.Mode == "dev",
	}
	s.log = logger.New(&loggerOpts)
	return nil
}

func (s *Server) setupDatabase() error {
	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		s.config.Database.User, s.config.Database.Password, s.config.Database.Host, s.config.Database.Name,
	)

	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		s.log.Error("Unable to initialize sql.DB", "error", err)
		return err
	}
	defer func(sqlDB *sql.DB) {
		if err := sqlDB.Close(); err != nil {
			s.log.Error("Failed to close sql.DB", "error", err)
		}
	}(sqlDB)

	if err := sqlDB.PingContext(context.Background()); err != nil {
		s.log.Error("Unable to ping database", "error", err)
		return err
	}
	s.log.Info("Database connection established")

	s.queries = db.New(sqlDB)
	s.db = sqlDB

	return nil
}

func (s *Server) setupAWSClients() error {
	kendraClient, err := kendra.NewClient(*s.config, s.log)
	if err != nil {
		s.log.Error("Could not initialize kendra client", "error", err)
		return err
	}

	// TODO: Add DI and make an interface
	bedrockClient, err := bedrock.NewBedrockClient(*s.config)
	if err != nil {
		s.log.Error("Could not initialize bedrock client", "error", err)
		return err
	}

	s3Client, err := s3.NewS3Client(*s.config)
	if err != nil {
		s.log.Error("Failed to initialize s3 client", "error", err)
		return err
	}

	s.bedrockClient = bedrockClient
	s.kendraClient = kendraClient
	s.s3Client = s3Client

	return nil
}

func (s *Server) setupSessionStore() error {
	s.log.Info("Initializing Session Store...")
	sessionSecretKey := os.Getenv("SESSION_SECRET_KEY")
	if sessionSecretKey == "" {
		if s.config.Mode == "prod" {
			s.log.Error("SESSION_SECRET_KEY environment variable not set in prod. Exiting now.")
			return fmt.Errorf("Invalid Session Secret Key")
		}
		s.log.Warn("SESSION_SECRET_KEY environment variable not set. Using insecure default (dev only).")
		sessionSecretKey = "insecure-default-key-for-dev-only-change-me"
	}

	cookieStore := sessions.NewCookieStore([]byte(sessionSecretKey))
	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   s.config.Mode == "prod",
		SameSite: http.SameSiteLaxMode,
	}
	s.log.Info("Session Store initialized", "name", sessionCookieName, "secure", cookieStore.Options.Secure)
	s.cookieStore = cookieStore

	return nil
}

func (s *Server) setupRateLimiters() error {
	ipRateLimiter := ratelimiter.NewInMemoryRateLimiter(s.log, ipMaxAttempts, ipBlockDuration, ipWindow)
	userRateLimiter := ratelimiter.NewInMemoryRateLimiter(s.log, userMaxAttempts, userBlockDuration, userWindow)
	s.userRateLimiter = userRateLimiter
	s.ipRateLimiter = ipRateLimiter

	return nil
}

func (s *Server) setupServices() error {
	sessionManager, err := application.NewGorillaSessionManager(s.cookieStore, sessionCookieName, s.log, s.queries)
	if err != nil {
		s.log.Error("Failed to create session manager", "error", err)
		return err
	}

	userService := application.NewUserService(s.log, s.queries)
	authenticationService := application.NewLoginService(s.log, s.ipRateLimiter, s.userRateLimiter, userService)
	searchService := application.NewSearchService(s.log, s.kendraClient, s.queries)
	suggestionService := application.NewSuggestionService(s.log, s.kendraClient)
	bedrockService := application.NewBedrockService(s.log, *s.bedrockClient)
	fileManagerService := application.NewFilemanagerService(s.log, s.s3Client)

	s.sessionManager = sessionManager
	s.userService = userService
	s.authService = authenticationService
	s.searchService = searchService
	s.suggestService = suggestionService
	s.bedrockService = bedrockService
	s.fileManService = fileManagerService

	return nil
}

func (s *Server) setupHandlers() error {
	homeHandler := handlers.NewHomeHandler(s.log, s.sessionManager)
	searchHandler := handlers.NewSearchHandler(s.log, s.searchService, s.sessionManager)
	authHandler := handlers.NewAuthenticationHandler(s.log, s.sessionManager, s.authService)
	suggestionsHandler := handlers.NewSuggestionsHandler(s.log, s.suggestService)
	uploadHandler := handlers.NewUploadHandler(s.log, s.queries, s.bedrockService, s.fileManService, s.sessionManager)
	userManagementHandler := handlers.NewUserManagementHandler(s.log, s.queries, s.sessionManager)
	databaseHandler := handlers.NewDatabaseHandler(s.log, s.queries)

	s.HomeHandler = homeHandler
	s.SearchHandler = searchHandler
	s.AuthHandler = authHandler
	s.SuggestionsHandler = suggestionsHandler
	s.UploadHandler = uploadHandler
	s.UserManagementHandler = userManagementHandler
	s.DatabaseHandler = databaseHandler

	return nil
}

func (s *Server) configureMiddleware() {
	s.echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogRemoteIP: true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				s.log.InfoContext(c.Request().Context(), "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("ip", v.RemoteIP),
				)
			} else {
				s.log.ErrorContext(c.Request().Context(), "REQUEST ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
					slog.String("ip", v.RemoteIP),
				)
			}
			return nil
		},
	}))

	s.echo.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "form:_csrf",
		CookieName:     "csrf",
		CookiePath:     "/",
		CookieDomain:   "",
		ContextKey:     "csrf",
		CookieSameSite: http.SameSiteLaxMode,
		CookieSecure:   s.config.Mode == "prod", // Only set secure cookies in prod
		Skipper: func(c echo.Context) bool {
			path := c.Path()
			if path == "/search/suggestions" {
				return true
			}
			return false
		},
	}))

}

func (s *Server) registerRoutes() {
	routes.RegisterAuthenticationRoutes(s.echo, s.AuthHandler)
	routes.RegisterDatabaseRoutes(s.echo, s.DatabaseHandler, s.sessionManager)
	routes.RegisterHomeRoutes(s.echo, s.HomeHandler)
	routes.RegisterSearchRoutes(s.echo, s.SearchHandler)
	routes.RegisterSuggestionsRoutes(s.echo, s.SuggestionsHandler)
	routes.RegisterUploadRoutes(s.echo, s.UploadHandler, s.sessionManager)
	routes.RegisterUserManagementRoutes(s.echo, s.UserManagementHandler, s.sessionManager)
	s.log.Info("Routes initialized")

	s.echo.Static("/images", "web/assets/images")
	s.echo.Static("/css", "web/assets/css")
	s.echo.Static("/svg", "web/assets/svg")
	s.echo.Static("/js", "web/assets/js")
	s.echo.Static("/favicon", "web/assets/favicon")
}

func (s *Server) Start(address string) error {
	s.configureMiddleware()
	s.registerRoutes()

	s.log.Info("Starting Server", "address", address)
	return s.echo.Start(address)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Warn("Initiating shutdown...")

	if err := s.echo.Shutdown(ctx); err != nil {
		s.log.Error("Error during HTTP server shutdown", "error", err)
	} else {
		s.log.Info("HTTP server shutdown complete")
	}

	s.ipRateLimiter.Shutdown()
	s.userRateLimiter.Shutdown()
	s.log.Info("Rate limiters shutdown")

	if shutdownable, ok := s.kendraClient.(interface {
		Shutdown(ctx context.Context) error
	}); ok {
		s.log.Info("Shutting down Kendra client")
		if err := shutdownable.Shutdown(ctx); err != nil {
			s.log.Error("Error during Kendra client shutdown", "error", err)
		} else {
			s.log.Info("Kendra Client shutdown complete")
		}
	} else {
		s.log.Warn("Kendra Client does not support graceful shutdown")
	}

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			s.log.Error("Failed to close database connection", "error", err)
		} else {
			s.log.Info("Database connection closed")
		}
	}

	s.log.Warn("Application shutdown finished")
	return nil
}
