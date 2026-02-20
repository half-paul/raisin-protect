package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// RedisClient wraps the go-redis client.
type RedisClient struct {
	*redis.Client
}

// ConnectRedis establishes a Redis connection.
func ConnectRedis(url string) (*RedisClient, error) {
	if url == "" {
		return nil, fmt.Errorf("redis URL is required")
	}

	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	log.Info().Msg("Connected to Redis")
	return &RedisClient{client}, nil
}

// Close gracefully shuts down the Redis client.
func (r *RedisClient) Close() error {
	return r.Client.Close()
}
