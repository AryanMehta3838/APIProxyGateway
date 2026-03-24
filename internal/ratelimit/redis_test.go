package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestRedisAllowWithinWindow(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	limiter, err := NewRedis(redisServer.Addr())
	if err != nil {
		t.Fatalf("new redis limiter: %v", err)
	}
	t.Cleanup(func() {
		_ = limiter.Close()
	})

	policy := Policy{Requests: 2, Window: time.Minute}
	first, err := limiter.Allow(context.Background(), "client-a", policy)
	if err != nil {
		t.Fatalf("first allow: %v", err)
	}
	second, err := limiter.Allow(context.Background(), "client-a", policy)
	if err != nil {
		t.Fatalf("second allow: %v", err)
	}
	third, err := limiter.Allow(context.Background(), "client-a", policy)
	if err != nil {
		t.Fatalf("third allow: %v", err)
	}

	if !first.Allowed || !second.Allowed {
		t.Fatal("expected first two requests to be allowed")
	}
	if third.Allowed {
		t.Fatal("expected third request to be denied")
	}
}

func TestRedisResetsAfterWindow(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	limiter, err := NewRedis(redisServer.Addr())
	if err != nil {
		t.Fatalf("new redis limiter: %v", err)
	}
	t.Cleanup(func() {
		_ = limiter.Close()
	})

	policy := Policy{Requests: 1, Window: time.Second}
	first, err := limiter.Allow(context.Background(), "client-a", policy)
	if err != nil {
		t.Fatalf("first allow: %v", err)
	}
	if !first.Allowed {
		t.Fatal("expected first request to be allowed")
	}

	second, err := limiter.Allow(context.Background(), "client-a", policy)
	if err != nil {
		t.Fatalf("second allow: %v", err)
	}
	if second.Allowed {
		t.Fatal("expected second request in same window to be denied")
	}

	redisServer.FastForward(1100 * time.Millisecond)
	third, err := limiter.Allow(context.Background(), "client-a", policy)
	if err != nil {
		t.Fatalf("third allow: %v", err)
	}
	if !third.Allowed {
		t.Fatal("expected request after window reset to be allowed")
	}
}

func TestRedisUnavailableBackend(t *testing.T) {
	t.Parallel()

	limiter, err := NewRedis("127.0.0.1:1")
	if err != nil {
		t.Fatalf("new redis limiter: %v", err)
	}
	t.Cleanup(func() {
		_ = limiter.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = limiter.Allow(ctx, "client-a", Policy{Requests: 1, Window: time.Minute})
	if err == nil {
		t.Fatal("expected redis limiter error when backend is unavailable")
	}
}

func TestNewRedisRequiresAddress(t *testing.T) {
	t.Parallel()

	if _, err := NewRedis("  "); err == nil {
		t.Fatal("expected error for empty redis address")
	}
}
