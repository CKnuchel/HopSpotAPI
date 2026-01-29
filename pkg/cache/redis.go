package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"hopSpotAPI/internal/config"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(cfg config.Config) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		// Redis not available -- return nil to indicate no caching
		return nil
	}

	return &RedisClient{client: client}
}

func (r *RedisClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string, target any) (bool, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil // Key not found
		}
		return false, err // other error
	}

	if err := json.Unmarshal(data, target); err != nil {
		return false, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return true, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	// Incr creates key with 1 if not exists else +1
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Define ttl only on first request
	if count == 1 {
		r.client.Expire(ctx, key, ttl)
	}

	return count, nil
}
