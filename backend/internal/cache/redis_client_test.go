package cache

import (
	"context"
	"os"
	"testing"
	"time"

	_ "github.com/redis/go-redis/v9"
)

// This test runs only when REDIS_ADDR is provided in the environment. It
// verifies BlacklistToken and IsTokenBlacklisted against a real Redis instance.
func TestBlacklistAndCheck(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR not set; skipping redis integration test")
	}
	pwd := os.Getenv("REDIS_PASSWORD")
	rdb := NewRedisClient(addr, pwd)
	if rdb == nil {
		t.Fatal("failed to create redis client")
	}
	ctx := context.Background()
	// Use a unique token key based on time
	token := "test-token-" + time.Now().Format("20060102150405")
	ttl := 5 * time.Second

	// Ensure key doesn't exist
	_ = rdb.Del(ctx, "bl:"+token).Err()

	if err := BlacklistToken(ctx, rdb, token, ttl); err != nil {
		t.Fatalf("BlacklistToken error: %v", err)
	}
	ok, err := IsTokenBlacklisted(ctx, rdb, token)
	if err != nil {
		t.Fatalf("IsTokenBlacklisted error: %v", err)
	}
	if !ok {
		t.Fatalf("expected token to be blacklisted")
	}

	// Wait until ttl expires and ensure key eventually disappears
	time.Sleep(ttl + 1*time.Second)
	ok, err = IsTokenBlacklisted(ctx, rdb, token)
	if err != nil {
		t.Fatalf("IsTokenBlacklisted error after ttl: %v", err)
	}
	if ok {
		t.Fatalf("expected token to have expired from blacklist after ttl")
	}
	// cleanup
	_ = rdb.Del(ctx, "bl:"+token).Err()
	// close client
	_ = rdb.Close()
}
