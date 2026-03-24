package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var allowFixedWindowScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
if current > tonumber(ARGV[1]) then
  return 0
end
return 1
`)

type Redis struct {
	client *redis.Client
}

func NewRedis(addr string) (*Redis, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, errors.New("redis limiter requires a non-empty address")
	}

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  250 * time.Millisecond,
		ReadTimeout:  250 * time.Millisecond,
		WriteTimeout: 250 * time.Millisecond,
		PoolTimeout:  250 * time.Millisecond,
	})

	return &Redis{client: client}, nil
}

func (r *Redis) Allow(ctx context.Context, key string, policy Policy) (Decision, error) {
	windowMS := policy.Window.Milliseconds()
	if windowMS <= 0 {
		return Decision{}, fmt.Errorf("rate limit window must be positive: %s", policy.Window)
	}
	if policy.Requests <= 0 {
		return Decision{}, fmt.Errorf("rate limit requests must be positive: %d", policy.Requests)
	}

	allowed, err := allowFixedWindowScript.Run(ctx, r.client, []string{key}, policy.Requests, windowMS).Int()
	if err != nil {
		return Decision{}, err
	}

	return Decision{Allowed: allowed == 1}, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}
