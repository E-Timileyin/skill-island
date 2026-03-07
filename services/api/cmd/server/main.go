package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/E-Timileyin/skill-island/services/api/internal/api/handlers"
	"github.com/E-Timileyin/skill-island/services/api/internal/api/routes"
	"github.com/E-Timileyin/skill-island/services/api/internal/config"
	"github.com/E-Timileyin/skill-island/services/api/internal/db"
	"github.com/E-Timileyin/skill-island/services/api/internal/ws"
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

	h := &handlers.Handler{
		DB:  pool,
		Cfg: cfg,
	}

	// Start WebSocket hub.
	hub := ws.NewHub(pool)
	go hub.Run()

	wsHandler := &handlers.WSHandler{
		Hub:       hub,
		JWTSecret: cfg.JWTSecret,
	}

	r := routes.SetupRouter(h, wsHandler, cfg)

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
