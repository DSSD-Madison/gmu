package application

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	db "github.com/DSSD-Madison/gmu/internal/infra/database/sqlc/generated"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/ratelimiter"
)

type AuthenticationService struct {
	log         logger.Logger
	ipLimiter   ratelimiter.RateLimiter
	userLimiter ratelimiter.RateLimiter
	userService UserManager
}

var (
	ErrRateLimited        = errors.New("rate limited")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func NewLoginService(log logger.Logger, ipRateLimiter ratelimiter.RateLimiter, userRateLimiter ratelimiter.RateLimiter, userService UserManager) *AuthenticationService {
	serviceLogger := log.With("service", "Login")
	return &AuthenticationService{
		log:         serviceLogger,
		ipLimiter:   ipRateLimiter,
		userLimiter: userRateLimiter,
		userService: userService,
	}
}

func (authsrv *AuthenticationService) HandleLogin(ctx context.Context, ip string, username string, password string) (db.User, error) {
	if authsrv.ipLimiter.IsLimited(ip) {
		authsrv.log.WarnContext(ctx, "IP rate limited", "ip", ip, "user", username)
		authsrv.userLimiter.RecordAttempt(username, false) // record failed against user even if ip is blocked
		return db.User{}, ErrRateLimited
	}

	if authsrv.userLimiter.IsLimited(username) {
		authsrv.log.WarnContext(ctx, "User rate limited", "ip", ip, "user", username)
		authsrv.ipLimiter.RecordAttempt(ip, false) // record failed against ip even if user is blocked
		return db.User{}, ErrRateLimited
	}

	user, err := authsrv.userService.GetUser(ctx, username)
	if err != nil {
		// Treat user not found as invalid credentials for security
		// Log the actual error for debugging
		authsrv.log.DebugContext(ctx, "User not found during login attempt", "username", username, "error", err)
		// Record failed attempts since an attempt was made
		authsrv.ipLimiter.RecordAttempt(ip, false)
		authsrv.userLimiter.RecordAttempt(username, false)
		return db.User{}, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	success := err == nil

	authsrv.ipLimiter.RecordAttempt(ip, success)
	authsrv.userLimiter.RecordAttempt(username, success)

	if !success {
		authsrv.log.WarnContext(ctx, "Invalid password attempt", "username", username, "ip", ip)
		return db.User{}, ErrInvalidCredentials
	}

	authsrv.log.InfoContext(ctx, "Login successful", "username", username, "ip", ip)
	return user, nil
}
