package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
)

func Connect() error {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	fmt.Println("Connected to the Redis")

	return nil
}

func Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return Client.Set(ctx, key, value, expiration).Err()
}

func Get(ctx context.Context, key string) (string, error) {
	return Client.Get(ctx, key).Result()
}

func Delete(ctx context.Context, key string) error {
	return Client.Del(ctx, key).Err()
}
