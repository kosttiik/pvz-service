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
	if redisClient == nil {
		panic("redis client cannot be nil")
	}
	return &TokenCache{redis: redisClient}
}

func (c *TokenCache) Set(ctx context.Context, userID string, token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be nil")
	}

	key := c.formatKey(userID)
	pipe := c.redis.Pipeline()
	pipe.Set(ctx, key, token, tokenTTL)
	pipe.Set(ctx, c.formatRefreshKey(userID), time.Now().Unix(), tokenTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to cache token: %w", err)
	}

	return nil
}

func (c *TokenCache) Get(ctx context.Context, userID string) (string, error) {
	key := c.formatKey(userID)
	token, err := c.redis.Get(ctx, key).Result()

	if err == redis.Nil {
		return "", fmt.Errorf("token not found in cache")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	if err := c.RefreshTTL(ctx, userID); err != nil {
		fmt.Printf("failed to refresh token TTL: %v\n", err)
	}

	return token, nil
}

func (c *TokenCache) Invalidate(ctx context.Context, userID string) error {
	key := c.formatKey(userID)
	pipe := c.redis.Pipeline()
	pipe.Del(ctx, key)
	pipe.Del(ctx, c.formatRefreshKey(userID))

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to invalidate token: %w", err)
	}

	return nil
}

func (c *TokenCache) RefreshTTL(ctx context.Context, userID string) error {
	key := c.formatKey(userID)
	pipe := c.redis.Pipeline()
	pipe.Expire(ctx, key, tokenTTL)
	pipe.Expire(ctx, c.formatRefreshKey(userID), tokenTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh token TTL: %w", err)
	}

	return nil
}

func (c *TokenCache) Exists(ctx context.Context, userID string) (bool, error) {
	key := c.formatKey(userID)
	exists, err := c.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token existence: %w", err)
	}
	return exists == 1, nil
}

func (c *TokenCache) formatKey(userID string) string {
	return fmt.Sprintf("%s%s", tokenPrefix, userID)
}

func (c *TokenCache) formatRefreshKey(userID string) string {
	return fmt.Sprintf("%s%s:refresh", tokenPrefix, userID)
}
