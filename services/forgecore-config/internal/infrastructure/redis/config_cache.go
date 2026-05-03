package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type ConfigCache struct {
	client *redis.Client
}

func NewConfigCache(client *redis.Client) *ConfigCache {
	return &ConfigCache{client: client}
}

func cacheKey(tenantID uuid.UUID, key string) string {
	return fmt.Sprintf("cfg:%s:%s", tenantID, key)
}

func (c *ConfigCache) Get(ctx context.Context, tenantID uuid.UUID, key string) (string, error) {
	val, err := c.client.Get(ctx, cacheKey(tenantID, key)).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("cache miss")
	}
	if err != nil {
		return "", fmt.Errorf("cache get: %w", err)
	}
	return val, nil
}

func (c *ConfigCache) Set(ctx context.Context, tenantID uuid.UUID, key, value string, ttlSeconds int64) error {
	err := c.client.Set(ctx, cacheKey(tenantID, key), value, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("cache set: %w", err)
	}
	return nil
}

func (c *ConfigCache) Invalidate(ctx context.Context, tenantID uuid.UUID, key string) error {
	err := c.client.Del(ctx, cacheKey(tenantID, key)).Err()
	if err != nil {
		return fmt.Errorf("cache invalidate: %w", err)
	}
	return nil
}
