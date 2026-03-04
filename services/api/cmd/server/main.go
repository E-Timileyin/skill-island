package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/E-Timileyin/skill-island/services/api/internal/api"
	"github.com/E-Timileyin/skill-island/services/api/internal/config"
	"github.com/E-Timileyin/skill-island/services/api/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := runMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	h := &api.Handler{}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", h.Health)

	log.Printf("starting server on :%s (env=%s)", cfg.Port, cfg.Env)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// runMigrations applies all pending database migrations from the migrations directory.
func runMigrations(databaseURL string) error {
	// golang-migrate pgx/v5 driver registers the "pgx5" scheme
	migrateURL := databaseURL
	migrateURL = strings.Replace(migrateURL, "postgresql://", "pgx5://", 1)
	migrateURL = strings.Replace(migrateURL, "postgres://", "pgx5://", 1)

	m, err := migrate.New("file://migrations", migrateURL)
	if err != nil {
		return fmt.Errorf("runMigrations: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("runMigrations: %w", err)
	}

	log.Println("migrations applied successfully")
	return nil
}
