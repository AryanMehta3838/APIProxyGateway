package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aryan/apiproxy/internal/telemetry"
)

func TestMetricsMiddlewareRecordsStatusAndRateLimit(t *testing.T) {
	t.Parallel()

	collector := telemetry.New()
	chain := Metrics(collector)(NamedRoute("echo", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/echo", nil)
	req = withRouteSlot(req)
	rec := httptest.NewRecorder()

	chain.ServeHTTP(rec, req)

	metrics := collector.Render()
	if !strings.Contains(metrics, `apiproxy_http_requests_total{route="echo",method="GET",status="429"} 1`) {
		t.Fatalf("expected 429 request metric, got:\n%s", metrics)
	}
	if !strings.Contains(metrics, `apiproxy_rate_limit_denied_total{route="echo"} 1`) {
		t.Fatalf("expected rate limit metric, got:\n%s", metrics)
	}
}

func withRouteSlot(req *http.Request) *http.Request {
	var routeName string
	ctx := context.WithValue(req.Context(), routeNameSlotKey{}, &routeName)
	return req.WithContext(ctx)
}
