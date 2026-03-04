package db_test

import (
	"context"
	"testing"

	"github.com/E-Timileyin/skill-island/services/api/internal/db"
)

func TestConnect_InvalidURL(t *testing.T) {
	ctx := context.Background()
	pool, err := db.Connect(ctx, "postgres://invalid:invalid@localhost:1/nonexistent")
	if err == nil {
		pool.Close()
		t.Fatal("expected error for invalid database URL, got nil")
	}
}
