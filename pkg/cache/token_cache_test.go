package cache

import (
	"context"
	"os"
	"testing"

	"github.com/kosttiik/pvz-service/pkg/redis"
)

func TestMain(m *testing.M) {
	os.Setenv("REDIS_HOST", "localhost")
	if err := redis.Connect(); err != nil {
		panic(err)
	}
	code := m.Run()
	redis.Close()
	os.Exit(code)
}

func TestTokenCache(t *testing.T) {
	cache := NewTokenCache(redis.Client)
	ctx := context.Background()
	userID := "test-user"
	token := "test-token"

	t.Run("Set and Get", func(t *testing.T) {
		if err := cache.Set(ctx, userID, token); err != nil {
			t.Fatalf("failed to set token: %v", err)
		}

		got, err := cache.Get(ctx, userID)
		if err != nil {
			t.Fatalf("failed to get token: %v", err)
		}
		if got != token {
			t.Errorf("got token %s, want %s", got, token)
		}
	})

	t.Run("Invalidate", func(t *testing.T) {
		if err := cache.Invalidate(ctx, userID); err != nil {
			t.Fatalf("failed to invalidate token: %v", err)
		}

		exists, err := cache.Exists(ctx, userID)
		if err != nil {
			t.Fatalf("failed to check token existence: %v", err)
		}
		if exists {
			t.Error("token still exists after invalidation")
		}
	})
}
