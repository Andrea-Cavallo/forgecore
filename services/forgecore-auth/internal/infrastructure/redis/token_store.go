package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	prefixRefresh    = "rt:"
	prefixBlacklist  = "bl:"
	prefixBruteForce = "bf:"
	prefixOneTime    = "ot:"
)

type TokenStore struct {
	client *redis.Client
}

func NewTokenStore(client *redis.Client) *TokenStore {
	return &TokenStore{client: client}
}

func (s *TokenStore) StoreRefreshToken(ctx context.Context, key, token string, ttlSeconds int64) error {
	err := s.client.Set(ctx, prefixRefresh+key, token, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}
	return nil
}

func (s *TokenStore) ValidateRefreshToken(ctx context.Context, key, token string) (bool, error) {
	stored, err := s.client.Get(ctx, prefixRefresh+key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("validate refresh token: %w", err)
	}
	return stored == token, nil
}

func (s *TokenStore) BlacklistJTI(ctx context.Context, jti string, ttlSeconds int64) error {
	err := s.client.Set(ctx, prefixBlacklist+jti, "1", time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("blacklist jti: %w", err)
	}
	return nil
}

func (s *TokenStore) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	exists, err := s.client.Exists(ctx, prefixBlacklist+jti).Result()
	if err != nil {
		return false, fmt.Errorf("check blacklist: %w", err)
	}
	return exists > 0, nil
}

func (s *TokenStore) IncrBruteForce(ctx context.Context, key string) (int64, error) {
	count, err := s.client.Incr(ctx, prefixBruteForce+key).Result()
	if err != nil {
		return 0, fmt.Errorf("incr brute force: %w", err)
	}
	return count, nil
}

func (s *TokenStore) SetBruteForceLockout(ctx context.Context, key string, ttlSeconds int64) error {
	err := s.client.Expire(ctx, prefixBruteForce+key, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("set brute force lockout: %w", err)
	}
	return nil
}

func (s *TokenStore) GetBruteForceCount(ctx context.Context, key string) (int64, error) {
	count, err := s.client.Get(ctx, prefixBruteForce+key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get brute force count: %w", err)
	}
	return count, nil
}

func (s *TokenStore) StoreOneTimeToken(ctx context.Context, key, token string, ttlSeconds int64) error {
	err := s.client.Set(ctx, prefixOneTime+key, token, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("store one-time token: %w", err)
	}
	return nil
}

func (s *TokenStore) PopOneTimeToken(ctx context.Context, key string) (string, error) {
	fullKey := prefixOneTime + key
	token, err := s.client.GetDel(ctx, fullKey).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("token non trovato o scaduto")
	}
	if err != nil {
		return "", fmt.Errorf("pop one-time token: %w", err)
	}
	return token, nil
}
