package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(addr, password string) *redis.Client {
	if addr == "" {
		return nil
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return rdb
}

func BlacklistToken(ctx context.Context, rdb *redis.Client, token string, ttl time.Duration) error {
	if rdb == nil {
		return nil
	}
	key := "bl:" + token
	return rdb.Set(ctx, key, "1", ttl).Err()
}

func IsTokenBlacklisted(ctx context.Context, rdb *redis.Client, token string) (bool, error) {
	if rdb == nil {
		return false, nil
	}
	key := "bl:" + token
	v, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return v > 0, nil
}
