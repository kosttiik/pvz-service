package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	tokenPrefix = "token:"
	tokenTTL    = 24 * time.Hour
)

type TokenCache struct {
	redis *redis.Client
}

func NewTokenCache(redisClient *redis.Client) *TokenCache {
	return &TokenCache{redis: redisClient}
}

func (c *TokenCache) Set(ctx context.Context, userID string, token string) error {
	key := fmt.Sprintf("%s%s", tokenPrefix, userID)
	return c.redis.Set(ctx, key, token, tokenTTL).Err()
}

func (c *TokenCache) Get(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("%s%s", tokenPrefix, userID)
	return c.redis.Get(ctx, key).Result()
}

func (c *TokenCache) Invalidate(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s%s", tokenPrefix, userID)
	return c.redis.Del(ctx, key).Err()
}
