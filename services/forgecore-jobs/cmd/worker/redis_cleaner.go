package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// redisTokenCleaner implements jobs.TokenCleaner using Redis.
// It removes token-blacklist keys whose names match the tenant pattern.
type redisTokenCleaner struct {
	rdb *redis.Client
}

// CleanExpiredTokens scans for keys matching "token:<tenantID>:*" and deletes them.
// Redis TTL-based expiry handles most token cleanup automatically; this covers edge cases.
func (c *redisTokenCleaner) CleanExpiredTokens(ctx context.Context, tenantID string) (int64, error) {
	pattern := fmt.Sprintf("token:%s:*", tenantID)
	var deleted int64
	iter := c.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := c.rdb.Del(ctx, iter.Val()).Err(); err != nil {
			return deleted, fmt.Errorf("eliminazione token: %w", err)
		}
		deleted++
	}
	if err := iter.Err(); err != nil {
		return deleted, fmt.Errorf("scan token: %w", err)
	}
	return deleted, nil
}
