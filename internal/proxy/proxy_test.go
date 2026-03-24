package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aryan/apiproxy/internal/config"
	"github.com/aryan/apiproxy/internal/middleware"
	"github.com/aryan/apiproxy/internal/testkit"
)

func TestNew_ForwardsRequest(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(testkit.EchoHandler())
	t.Cleanup(upstream.Close)

	handler, err := New(config.Route{
		Name:       "echo",
		PathPrefix: "/api/echo",
		Methods:    []string{http.MethodGet},
		Upstream:   upstream.URL,
	})
	if err != nil {
		t.Fatalf("new proxy: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/echo/hello?name=aryan", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var got testkit.EchoResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Path != "/api/echo/hello" {
		t.Fatalf("expected path %q, got %q", "/api/echo/hello", got.Path)
	}
	if got.Query != "name=aryan" {
		t.Fatalf("expected query %q, got %q", "name=aryan", got.Query)
	}
}

func TestNew_StripsConfiguredPrefix(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(testkit.EchoHandler())
	t.Cleanup(upstream.Close)

	handler, err := New(config.Route{
		Name:        "echo",
		PathPrefix:  "/api/echo",
		StripPrefix: "/api/echo",
		Methods:     []string{http.MethodGet},
		Upstream:    upstream.URL,
	})
	if err != nil {
		t.Fatalf("new proxy: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/echo/hello?name=aryan", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var got testkit.EchoResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Path != "/hello" {
		t.Fatalf("expected path %q, got %q", "/hello", got.Path)
	}
	if got.Query != "name=aryan" {
		t.Fatalf("expected query %q, got %q", "name=aryan", got.Query)
	}
}

func TestRewritePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		path        string
		stripPrefix string
		want        string
	}{
		{name: "no strip prefix", path: "/api/echo/hello", stripPrefix: "", want: "/api/echo/hello"},
		{name: "strip nested path", path: "/api/echo/hello", stripPrefix: "/api/echo", want: "/hello"},
		{name: "strip exact prefix", path: "/api/echo", stripPrefix: "/api/echo", want: "/"},
		{name: "prefix does not match", path: "/api/other", stripPrefix: "/api/echo", want: "/api/other"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := rewritePath(tc.path, tc.stripPrefix); got != tc.want {
				t.Fatalf("rewritePath(%q, %q) = %q, want %q", tc.path, tc.stripPrefix, got, tc.want)
			}
		})
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

	handler, err := New(config.Route{
		Name:       "slow",
		PathPrefix: "/api/slow",
		TimeoutMS:  50,
		Methods:    []string{http.MethodGet},
		Upstream:   upstream.URL,
	})
	if err != nil {
		t.Fatalf("new proxy: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/slow", nil)
	req = req.WithContext(context.Background())
	req.Header.Set(middleware.RequestIDHeader, "timeout-request-id")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status %d, got %d", http.StatusGatewayTimeout, recorder.Code)
	}
	if got := recorder.Header().Get(middleware.RequestIDHeader); got != "timeout-request-id" {
		t.Fatalf("expected request id header %q, got %q", "timeout-request-id", got)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "request_id=timeout-request-id") {
		t.Fatalf("expected request id in body, got %q", body)
	}
}

func TestNew_UnreachableUpstreamReturns502WithRequestID(t *testing.T) {
	t.Parallel()

	handler, err := New(config.Route{
		Name:       "dead",
		PathPrefix: "/api/dead",
		Methods:    []string{http.MethodGet},
		Upstream:   "http://127.0.0.1:1",
	})
	if err != nil {
		t.Fatalf("new proxy: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/dead", nil)
	req.Header.Set(middleware.RequestIDHeader, "bad-gateway-request-id")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, recorder.Code)
	}
	if got := recorder.Header().Get(middleware.RequestIDHeader); got != "bad-gateway-request-id" {
		t.Fatalf("expected request id header %q, got %q", "bad-gateway-request-id", got)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "request_id=bad-gateway-request-id") {
		t.Fatalf("expected request id in body, got %q", body)
	}
}
