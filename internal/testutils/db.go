package testutils

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kosttiik/pvz-service/pkg/logger"
)

func SetupTestDB(t *testing.T) *pgxpool.Pool {
	if err := logger.Init(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgresql://postgres:postgres@localhost:5432/pvz_db?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Failed to ping db: %v", err)
	}

	return pool
}
