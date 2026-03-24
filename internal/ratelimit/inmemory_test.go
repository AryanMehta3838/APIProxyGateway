package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryAllowWithinWindow(t *testing.T) {
	t.Parallel()

	limiter := NewInMemory()
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

func TestInMemoryResetsAfterWindow(t *testing.T) {
	t.Parallel()

	limiter := NewInMemory()
	policy := Policy{Requests: 1, Window: 20 * time.Millisecond}

	first, err := limiter.Allow(context.Background(), "client-a", policy)
	if err != nil {
		t.Fatalf("first allow: %v", err)
	}
	if !first.Allowed {
		t.Fatal("expected first request to be allowed")
	}

	time.Sleep(25 * time.Millisecond)

	second, err := limiter.Allow(context.Background(), "client-a", policy)
	if err != nil {
		t.Fatalf("second allow: %v", err)
	}
	if !second.Allowed {
		t.Fatal("expected second request after window reset to be allowed")
	}
}
