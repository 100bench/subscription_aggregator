// @Subscription Aggregator API
// @version 1.0
// @description REST API for aggregating user subscriptions
// @host localhost:8080
// @BasePath /
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pkg/errors"

	"github.com/100bench/subscription_aggregator/internal/adapters/storage/postgres"
	"github.com/100bench/subscription_aggregator/internal/cases"
	"github.com/100bench/subscription_aggregator/internal/pkg/logger"
	"github.com/100bench/subscription_aggregator/internal/ports/http/public"

	_ "github.com/100bench/subscription_aggregator/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	// Initialize logger
	logger.Init()

	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Log.Error("DATABASE_URL environment variable is not set")
		os.Exit(1)
	}

	m, err := migrate.New(
		"file://migrations",
		dbURL)
	if err != nil {
		logger.Log.Error("Failed to create migrate instance", "error", errors.Wrap(err, "migrate.New"))
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Log.Error("Failed to apply migrations", "error", errors.Wrap(err, "migrate.Up"))
		os.Exit(1)
	}
	logger.Log.Info("Database migrations applied successfully")

	storage, err := postgres.NewPgxClient(ctx, dbURL)
	if err != nil {
		logger.Log.Error("Failed to connect to postgres", "error", errors.Wrap(err, "postgres.NewPgxClient"))
		os.Exit(1)
	}
	defer storage.Close()

	subscriptionService, err := cases.NewServiceProvider(storage)
	if err != nil {
		logger.Log.Error("Failed to create subscription service", "error", err)
		os.Exit(1)
	}

	httpServer, err := public.NewServer(subscriptionService)
	if err != nil {
		logger.Log.Error("Failed to init http server", "error", err)
		os.Exit(1)
	}
	r := httpServer.GetRouter()

	r.Handle("/swagger/*", httpSwagger.WrapHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	// Create HTTP server
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Create channel for shutdown signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		logger.Log.Info("Server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	logger.Log.Info("Shutdown signal received, gracefully shutting down...")

	// Graceful shutdown with 30 second timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	storage.Close()
	logger.Log.Info("Server stopped")
}
