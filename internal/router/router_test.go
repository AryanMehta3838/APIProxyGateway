package router

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"

	"github.com/aryan/apiproxy/internal/admin"
	"github.com/aryan/apiproxy/internal/config"
	"github.com/aryan/apiproxy/internal/middleware"
	"github.com/aryan/apiproxy/internal/testkit"
)

func TestNew_ProxiesConfiguredRoute(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(testkit.EchoHandler())
	t.Cleanup(upstream.Close)

	handler, err := New(config.Config{
		Server: config.ServerConfig{Port: 8080},
		Routes: []config.Route{
			{
				Name:       "echo",
				PathPrefix: "/api/echo",
				Methods:    []string{http.MethodGet},
				Upstream:   upstream.URL,
			},
		},
	}, admin.NewRouter())
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/api/echo/hello?name=aryan")
	if err != nil {
		t.Fatalf("get proxied route: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var got testkit.EchoResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Path != "/api/echo/hello" {
		t.Fatalf("expected path %q, got %q", "/api/echo/hello", got.Path)
	}
	if got.Query != "name=aryan" {
		t.Fatalf("expected query %q, got %q", "name=aryan", got.Query)
	}
}

func TestNew_ProxiesConfiguredRouteWithStripPrefix(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(testkit.EchoHandler())
	t.Cleanup(upstream.Close)

	handler, err := New(config.Config{
		Server: config.ServerConfig{Port: 8080},
		Routes: []config.Route{
			{
				Name:        "echo",
				PathPrefix:  "/api/echo",
				StripPrefix: "/api/echo",
				Methods:     []string{http.MethodGet},
				Upstream:    upstream.URL,
			},
		},
	}, admin.NewRouter())
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/api/echo/hello?name=aryan")
	if err != nil {
		t.Fatalf("get proxied route: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var got testkit.EchoResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Path != "/hello" {
		t.Fatalf("expected path %q, got %q", "/hello", got.Path)
	}
	if got.Query != "name=aryan" {
		t.Fatalf("expected query %q, got %q", "name=aryan", got.Query)
	}
}

func TestNew_KeepsAdminEndpointsAccessible(t *testing.T) {
	t.Parallel()

	handler, err := New(config.Config{
		Server: config.ServerConfig{Port: 8080},
		Routes: []config.Route{
			{
				Name:       "echo",
				PathPrefix: "/api/echo",
				Methods:    []string{http.MethodGet},
				Upstream:   "http://localhost:9091",
			},
		},
	}, admin.NewRouter())
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestNew_GeneratesRequestIDAndReturnsIt(t *testing.T) {
	t.Parallel()

	var upstreamRequestID string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamRequestID = r.Header.Get(middleware.RequestIDHeader)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)

	handler, err := New(config.Config{
		Server: config.ServerConfig{Port: 8080},
		Routes: []config.Route{
			{
				Name:       "echo",
				PathPrefix: "/api/echo",
				Methods:    []string{http.MethodGet},
				Upstream:   upstream.URL,
			},
		},
	}, admin.NewRouter())
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/api/echo")
	if err != nil {
		t.Fatalf("get proxied route: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	got := resp.Header.Get(middleware.RequestIDHeader)
	if got == "" {
		t.Fatal("expected response request id header")
	}
	if upstreamRequestID == "" {
		t.Fatal("expected upstream request id header")
	}
	if got != upstreamRequestID {
		t.Fatalf("expected same request id for response and upstream, got response=%q upstream=%q", got, upstreamRequestID)
	}
}

func TestNew_PreservesExistingRequestID(t *testing.T) {
	t.Parallel()

	const want = "provided-request-id"

	var upstreamRequestID string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamRequestID = r.Header.Get(middleware.RequestIDHeader)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)

	handler, err := New(config.Config{
		Server: config.ServerConfig{Port: 8080},
		Routes: []config.Route{
			{
				Name:       "echo",
				PathPrefix: "/api/echo",
				Methods:    []string{http.MethodGet},
				Upstream:   upstream.URL,
			},
		},
	}, admin.NewRouter())
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/echo", nil)
	req.Header.Set(middleware.RequestIDHeader, want)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if got := recorder.Header().Get(middleware.RequestIDHeader); got != want {
		t.Fatalf("expected preserved response request id %q, got %q", want, got)
	}
	if upstreamRequestID != want {
		t.Fatalf("expected preserved upstream request id %q, got %q", want, upstreamRequestID)
	}
}

func TestNew_TimeoutReturns504WithRequestID(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(200 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(upstream.Close)

	handler, err := New(config.Config{
		Server: config.ServerConfig{Port: 8080},
		Routes: []config.Route{
			{
				Name:       "slow",
				PathPrefix: "/api/slow",
				TimeoutMS:  50,
				Methods:    []string{http.MethodGet},
				Upstream:   upstream.URL,
			},
		},
	}, admin.NewRouter())
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/api/slow")
	if err != nil {
		t.Fatalf("get proxied route: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusGatewayTimeout {
		t.Fatalf("expected status %d, got %d", http.StatusGatewayTimeout, resp.StatusCode)
	}
	requestID := resp.Header.Get(middleware.RequestIDHeader)
	if requestID == "" {
		t.Fatal("expected response request id header")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), requestID) {
		t.Fatalf("expected body to include request id %q, got %q", requestID, string(body))
	}
}

func TestNew_RateLimitReturns429(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)

	handler, err := New(config.Config{
		Server: config.ServerConfig{Port: 8080},
		Routes: []config.Route{
			{
				Name:       "echo",
				PathPrefix: "/api/echo",
				Methods:    []string{http.MethodGet},
				Upstream:   upstream.URL,
				RateLimit: config.RateLimitConfig{
					Enabled:       true,
					Requests:      1,
					WindowSeconds: 60,
					KeyStrategy:   "ip",
				},
			},
		},
	}, admin.NewRouter())
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	first, err := http.Get(server.URL + "/api/echo")
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	t.Cleanup(func() { _ = first.Body.Close() })
	if first.StatusCode != http.StatusNoContent {
		t.Fatalf("expected first status %d, got %d", http.StatusNoContent, first.StatusCode)
	}

	second, err := http.Get(server.URL + "/api/echo")
	if err != nil {
		t.Fatalf("second request: %v", err)
	}
	t.Cleanup(func() { _ = second.Body.Close() })
	if second.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected second status %d, got %d", http.StatusTooManyRequests, second.StatusCode)
	}
	requestID := second.Header.Get(middleware.RequestIDHeader)
	if requestID == "" {
		t.Fatal("expected response request id header")
	}
	body, err := io.ReadAll(second.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !strings.Contains(string(body), requestID) {
		t.Fatalf("expected body to include request id %q, got %q", requestID, string(body))
	}
}

func TestNew_HealthzStaysExemptFromRateLimit(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)

	handler, err := New(config.Config{
		Server: config.ServerConfig{Port: 8080},
		Routes: []config.Route{
			{
				Name:       "echo",
				PathPrefix: "/api/echo",
				Methods:    []string{http.MethodGet},
				Upstream:   upstream.URL,
				RateLimit: config.RateLimitConfig{
					Enabled:       true,
					Requests:      1,
					WindowSeconds: 60,
					KeyStrategy:   "ip",
				},
			},
		},
	}, admin.NewRouter())
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.RemoteAddr = "127.0.0.1:9999"
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestNew_UsesRedisRateLimiterWhenEnabled(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)

	handler, err := New(config.Config{
		Server: config.ServerConfig{Port: 8080},
		Redis: config.RedisConfig{
			Enabled: true,
			Addr:    redisServer.Addr(),
		},
		Routes: []config.Route{
			{
				Name:       "echo",
				PathPrefix: "/api/echo",
				Methods:    []string{http.MethodGet},
				Upstream:   upstream.URL,
				RateLimit: config.RateLimitConfig{
					Enabled:       true,
					Requests:      1,
					WindowSeconds: 60,
					KeyStrategy:   "ip",
				},
			},
		},
	}, admin.NewRouter())
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	first, err := http.Get(server.URL + "/api/echo")
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	t.Cleanup(func() { _ = first.Body.Close() })
	if first.StatusCode != http.StatusNoContent {
		t.Fatalf("expected first status %d, got %d", http.StatusNoContent, first.StatusCode)
	}

	second, err := http.Get(server.URL + "/api/echo")
	if err != nil {
		t.Fatalf("second request: %v", err)
	}
	t.Cleanup(func() { _ = second.Body.Close() })
	if second.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected second status %d, got %d", http.StatusTooManyRequests, second.StatusCode)
	}
}

func TestNew_RedisLimiterUnavailableReturns503(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)

	handler, err := New(config.Config{
		Server: config.ServerConfig{Port: 8080},
		Redis: config.RedisConfig{
			Enabled: true,
			Addr:    "127.0.0.1:1",
		},
		Routes: []config.Route{
			{
				Name:       "echo",
				PathPrefix: "/api/echo",
				Methods:    []string{http.MethodGet},
				Upstream:   upstream.URL,
				RateLimit: config.RateLimitConfig{
					Enabled:       true,
					Requests:      1,
					WindowSeconds: 60,
					KeyStrategy:   "ip",
				},
			},
		},
	}, admin.NewRouter())
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/echo", nil)
	req.Header.Set(middleware.RequestIDHeader, "redis-down-test")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}
	if got := recorder.Header().Get(middleware.RequestIDHeader); got != "redis-down-test" {
		t.Fatalf("expected request id header %q, got %q", "redis-down-test", got)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "rate limit backend unavailable") {
		t.Fatalf("expected backend unavailable body, got %q", body)
	}
}

func TestNew_RedisRateLimitSharedAcrossRouterInstances(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)

	cfg := config.Config{
		Server: config.ServerConfig{Port: 8080},
		Redis: config.RedisConfig{
			Enabled: true,
			Addr:    redisServer.Addr(),
		},
		Routes: []config.Route{
			{
				Name:       "echo",
				PathPrefix: "/api/echo",
				Methods:    []string{http.MethodGet},
				Upstream:   upstream.URL,
				RateLimit: config.RateLimitConfig{
					Enabled:       true,
					Requests:      1,
					WindowSeconds: 60,
					KeyStrategy:   "ip",
				},
			},
		},
	}

	handlerA, err := New(cfg, admin.NewRouter())
	if err != nil {
		t.Fatalf("new router A: %v", err)
	}
	handlerB, err := New(cfg, admin.NewRouter())
	if err != nil {
		t.Fatalf("new router B: %v", err)
	}

	serverA := httptest.NewServer(handlerA)
	t.Cleanup(serverA.Close)
	serverB := httptest.NewServer(handlerB)
	t.Cleanup(serverB.Close)

	first, err := http.Get(serverA.URL + "/api/echo")
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	t.Cleanup(func() { _ = first.Body.Close() })
	if first.StatusCode != http.StatusNoContent {
		t.Fatalf("expected first status %d, got %d", http.StatusNoContent, first.StatusCode)
	}

	second, err := http.Get(serverB.URL + "/api/echo")
	if err != nil {
		t.Fatalf("second request: %v", err)
	}
	t.Cleanup(func() { _ = second.Body.Close() })
	if second.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected second status %d, got %d", http.StatusTooManyRequests, second.StatusCode)
	}
}
