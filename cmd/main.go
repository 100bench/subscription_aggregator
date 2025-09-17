// @Subscription Aggregator API
// @version 1.0
// @description REST API for aggregating user subscriptions
// @host localhost:8080
// @BasePath /
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

	_ "github.com/100bench/subscription_aggregator/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	m, err := migrate.New(
		"file://migrations",
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

	subscriptionService := cases.NewSubscriptionService(storage)

	httpServer, err := public.NewServer(subscriptionService)
	if err != nil {
		log.Fatalf("failed to init http server: %v", err)
	}
	r := httpServer.GetRouter()

	r.Handle("/swagger/*", httpSwagger.WrapHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("Server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
