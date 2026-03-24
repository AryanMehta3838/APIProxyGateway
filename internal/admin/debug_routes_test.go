package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aryan/apiproxy/internal/config"
)

func TestDebugRoutes_JSON(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Redis: config.RedisConfig{Enabled: true, Addr: "localhost:6379"},
		Routes: []config.Route{
			{
				Name:        "api",
				PathPrefix:  "/api",
				StripPrefix: "/api",
				TimeoutMS:   5000,
				Methods:     []string{http.MethodGet},
				Upstream:    "http://user:must-not-leak@127.0.0.1:9000/hidden/path?x=1",
				RateLimit: config.RateLimitConfig{
					Enabled:       true,
					Requests:      3,
					WindowSeconds: 60,
					KeyStrategy:   "ip",
				},
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/debug/routes", nil)
	rec := httptest.NewRecorder()
	NewRouterWithMetrics(nil, cfg).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("Content-Type %q", ct)
	}

	body := rec.Body.String()
	for _, leak := range []string{"must-not-leak", "user:", "hidden", "?x=1"} {
		if strings.Contains(body, leak) {
			t.Fatalf("response must not contain %q: %s", leak, body)
		}
	}

	var got debugRoutesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !got.Redis.Enabled || got.Redis.Addr != "localhost:6379" {
		t.Fatalf("redis %+v", got.Redis)
	}
	if len(got.Routes) != 1 {
		t.Fatalf("routes len %d", len(got.Routes))
	}
	r0 := got.Routes[0]
	if r0.Name != "api" || r0.PathPrefix != "/api" || r0.StripPrefix != "/api" {
		t.Fatalf("route fields %+v", r0)
	}
	if r0.TimeoutMS != 5000 {
		t.Fatalf("timeout %d", r0.TimeoutMS)
	}
	if r0.UpstreamScheme != "http" || r0.UpstreamHost != "127.0.0.1:9000" {
		t.Fatalf("upstream scheme/host %q %q", r0.UpstreamScheme, r0.UpstreamHost)
	}
	if !r0.RateLimit.Enabled || r0.RateLimit.Requests != 3 || r0.RateLimit.WindowSeconds != 60 || r0.RateLimit.KeyStrategy != "ip" {
		t.Fatalf("rate limit %+v", r0.RateLimit)
	}
}

func TestUpstreamSchemeHost(t *testing.T) {
	t.Parallel()
	scheme, host := upstreamSchemeHost("https://example.com:8443/foo")
	if scheme != "https" || host != "example.com:8443" {
		t.Fatalf("got %q %q", scheme, host)
	}
}

func TestSanitizeRedisAddr_URLWithCredentials(t *testing.T) {
	t.Parallel()
	got := sanitizeRedisAddr("redis://:secret@redis.internal:6379/0")
	if got != "redis.internal:6379" {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(got, "secret") {
		t.Fatal("leaked password")
	}
}
