package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pkg/errors"

	"github.com/100bench/subscription_aggregator/internal/adapters/storage/postgres"
	"github.com/100bench/subscription_aggregator/internal/cases"
	"github.com/100bench/subscription_aggregator/internal/ports/http/public"
)

func main() {
	ctx := context.Background()

	// Initialize PostgreSQL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Apply database migrations
	m, err := migrate.New(
		"file://migrations", // Path to your migration files
		dbURL)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", errors.Wrap(err, "migrate.New"))
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to apply migrations: %v", errors.Wrap(err, "migrate.Up"))
	}
	log.Println("Database migrations applied successfully!")

	storage, err := postgres.NewPgxClient(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", errors.Wrap(err, "postgres.NewPgxClient"))
	}
	defer storage.Close()

	// Initialize Subscription Service
	subscriptionService := cases.NewSubscriptionService(storage)

	// Set up HTTP server
	httpServer := public.NewHttpServer(subscriptionService)

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("Server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, httpServer.GetRouter()))
}
