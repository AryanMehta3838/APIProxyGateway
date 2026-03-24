package ratelimit

import (
	"context"
	"sync"
	"time"
)

type Policy struct {
	Requests int
	Window   time.Duration
}

type Decision struct {
	Allowed bool
}

type Limiter interface {
	Allow(ctx context.Context, key string, policy Policy) (Decision, error)
}

type InMemory struct {
	mu      sync.Mutex
	buckets map[string]bucket
}

type bucket struct {
	count       int
	windowStart time.Time
}

func NewInMemory() *InMemory {
	return &InMemory{buckets: make(map[string]bucket)}
}

func (l *InMemory) Allow(_ context.Context, key string, policy Policy) (Decision, error) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	current, ok := l.buckets[key]
	if !ok || now.Sub(current.windowStart) >= policy.Window {
		l.buckets[key] = bucket{
			count:       1,
			windowStart: now,
		}
		return Decision{Allowed: true}, nil
	}

	if current.count >= policy.Requests {
		return Decision{Allowed: false}, nil
	}

	current.count++
	l.buckets[key] = current
	return Decision{Allowed: true}, nil
}
