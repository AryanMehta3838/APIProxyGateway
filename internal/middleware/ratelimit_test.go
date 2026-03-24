package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aryan/apiproxy/internal/ratelimit"
)

func TestRateLimit_AllowsWithinPolicy(t *testing.T) {
	t.Parallel()

	limiter := ratelimit.NewInMemory()
	handler := RateLimit("echo", limiter, RateLimitPolicy(2, 60), "ip", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	first := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/api/echo", nil)
	req1.RemoteAddr = "127.0.0.1:1234"
	handler.ServeHTTP(first, req1)

	second := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/echo", nil)
	req2.RemoteAddr = "127.0.0.1:1235"
	handler.ServeHTTP(second, req2)

	if first.Code != http.StatusNoContent || second.Code != http.StatusNoContent {
		t.Fatalf("expected allowed requests, got %d and %d", first.Code, second.Code)
	}
}

func TestRateLimit_DeniesAfterQuota(t *testing.T) {
	t.Parallel()

	limiter := ratelimit.NewInMemory()
	handler := RateLimit("echo", limiter, RateLimitPolicy(1, 60), "ip", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	firstReq := httptest.NewRequest(http.MethodGet, "/api/echo", nil)
	firstReq.RemoteAddr = "127.0.0.1:1234"
	firstReq.Header.Set(RequestIDHeader, "req-1")
	first := httptest.NewRecorder()
	handler.ServeHTTP(first, firstReq)

	secondReq := httptest.NewRequest(http.MethodGet, "/api/echo", nil)
	secondReq.RemoteAddr = "127.0.0.1:1235"
	secondReq.Header.Set(RequestIDHeader, "req-1")
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, secondReq)

	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, second.Code)
	}
	if got := second.Header().Get(RequestIDHeader); got != "req-1" {
		t.Fatalf("expected request id header %q, got %q", "req-1", got)
	}
	if body := second.Body.String(); !strings.Contains(body, "request_id=req-1") {
		t.Fatalf("expected request id in body, got %q", body)
	}
}
