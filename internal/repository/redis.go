package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/maojcn/shortlink/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisRepo encapsulates Redis connection and operations
type RedisRepo struct {
	client *redis.Client
	logger *zap.Logger
	ctx    context.Context
}

// NewRedisRepo creates and initializes a Redis repository
func NewRedisRepo(ctx context.Context, cfg config.RedisConfig, logger *zap.Logger) (*RedisRepo, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis connection test failed: %w", err)
	}

	logger.Info("Successfully connected to Redis")

	return &RedisRepo{
		client: client,
		logger: logger,
		ctx:    ctx,
	}, nil
}

// Close closes the Redis connection
func (r *RedisRepo) Close() error {
	return r.client.Close()
}

// SetCache sets a cache with an expiration time
func (r *RedisRepo) SetCache(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("Failed to serialize data: %w", err)
	}

	err = r.client.Set(r.ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("Failed to set cache: %w", err)
	}

	return nil
}

// GetCache retrieves a cache and deserializes it into the provided object
func (r *RedisRepo) GetCache(key string, dest interface{}) (bool, error) {
	data, err := r.client.Get(r.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil // Key does not exist, not an error
		}
		return false, fmt.Errorf("Failed to get cache: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return false, fmt.Errorf("Failed to deserialize data: %w", err)
	}

	return true, nil
}

// DeleteCache deletes a cache
func (r *RedisRepo) DeleteCache(key string) error {
	err := r.client.Del(r.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("Failed to delete cache: %w", err)
	}

	return nil
}

// Incr atomically increments a counter
func (r *RedisRepo) Incr(key string) (int64, error) {
	val, err := r.client.Incr(r.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("Failed to increment counter: %w", err)
	}
	return val, nil
}

// SetWithExpiration sets a key-value pair with an expiration time
func (r *RedisRepo) SetWithExpiration(key string, value string, expiration time.Duration) error {
	err := r.client.Set(r.ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("Failed to set key-value pair with expiration: %w", err)
	}
	return nil
}

// GetUserCacheKey generates a user cache key
func GetUserCacheKey(id int64) string {
	return fmt.Sprintf("user:%d", id)
}
