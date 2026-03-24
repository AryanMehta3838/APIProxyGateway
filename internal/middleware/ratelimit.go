package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/aryan/apiproxy/internal/ratelimit"
)

func RateLimit(routeName string, limiter ratelimit.Limiter, policy ratelimit.Policy, keyStrategy string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := rateLimitKey(routeName, keyStrategy, r)
		decision, err := limiter.Allow(r.Context(), key, policy)
		if err != nil {
			writeRateLimitError(w, r, http.StatusServiceUnavailable, "rate limit backend unavailable")
			return
		}
		if !decision.Allowed {
			writeRateLimitError(w, r, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeRateLimitError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	requestID := r.Header.Get(RequestIDHeader)
	if requestID != "" {
		w.Header().Set(RequestIDHeader, requestID)
		msg = fmt.Sprintf("%s (request_id=%s)", msg, requestID)
	}
	http.Error(w, msg, status)
}

func rateLimitKey(routeName string, keyStrategy string, r *http.Request) string {
	strategy := strings.ToLower(keyStrategy)
	if strategy == "" {
		strategy = "ip"
	}

	switch strategy {
	case "ip":
		return fmt.Sprintf("%s:%s", routeName, clientIP(r))
	default:
		return fmt.Sprintf("%s:%s", routeName, clientIP(r))
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}

func RateLimitPolicy(requests int, windowSeconds int) ratelimit.Policy {
	return ratelimit.Policy{
		Requests: requests,
		Window:   time.Duration(windowSeconds) * time.Second,
	}
}

func AllowWithContext(ctx context.Context, limiter ratelimit.Limiter, key string, policy ratelimit.Policy) (ratelimit.Decision, error) {
	return limiter.Allow(ctx, key, policy)
}
