package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/DSSD-Madison/gmu/internal/infra/http/server"
	"github.com/DSSD-Madison/gmu/pkg/config"
)

const (
	sessionCookieName = "gmu_session"

	ipMaxAttempts   = 10
	ipBlockDuration = 5 * time.Minute
	ipWindow        = 1 * time.Minute

	userMaxAttempts   = 5
	userBlockDuration = 15 * time.Minute
	userWindow        = 5 * time.Minute

	shutdownTimeout = 15 * time.Second
)

func main() {
	// --- Configuration ---
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading app config: %v", err)
	}

	server, err := server.NewServer(cfg)

	// --- Start Server and Handle Graceful Shutdown ---
	go func() {
		address := ":8080"
		if err := server.Start(address); err != nil && !errors.Is(err, http.ErrServerClosed) {
			os.Exit(1)
		}
	}()

	// --- Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	recvdSignal := <-quit

	server.LogEvent("Received OS signal, initiating shutdown...", "signal", recvdSignal.String())

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		server.LogEvent("Error during graceful shutdown", "error", err)
		os.Exit(1)
	}

	server.LogEvent("Shutdown complete")
	os.Exit(0)
}
