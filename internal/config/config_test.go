package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	path := writeTempConfig(t, `
server:
  port: 8080
redis:
  enabled: true
  addr: localhost:6379
routes:
  - name: echo
    path_prefix: /api/echo
    strip_prefix: /api/echo
    timeout_ms: 3000
    methods: [GET]
    upstream: http://localhost:9091
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Fatalf("expected port 8080, got %d", cfg.Server.Port)
	}

	if len(cfg.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(cfg.Routes))
	}
}

func TestLoadInvalidConfig(t *testing.T) {
	path := writeTempConfig(t, `
server:
  port: 0
routes:
  - name: ""
    path_prefix: api/echo
    upstream: ftp://localhost:9091
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected invalid config error")
	}

	if !strings.Contains(err.Error(), "server.port") {
		t.Fatalf("expected server.port validation error, got %v", err)
	}
}

func TestValidateRoutePathPrefix(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 8080},
		Routes: []Route{
			{
				Name:       "echo",
				PathPrefix: "api/echo",
				Upstream:   "http://localhost:9091",
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "routes[0].path_prefix") {
		t.Fatalf("expected path_prefix validation error, got %v", err)
	}
}

func TestValidateRouteUpstream(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 8080},
		Routes: []Route{
			{
				Name:       "echo",
				PathPrefix: "/api/echo",
				Methods:    []string{"GET"},
				Upstream:   "ftp://localhost:9091",
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "routes[0].upstream") {
		t.Fatalf("expected upstream validation error, got %v", err)
	}
}

func TestValidateRouteMethods(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 8080},
		Routes: []Route{
			{
				Name:       "echo",
				PathPrefix: "/api/echo",
				Methods:    []string{"NOPE"},
				Upstream:   "http://localhost:9091",
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "routes[0].methods[0]") {
		t.Fatalf("expected methods validation error, got %v", err)
	}
}

func TestValidateRouteStripPrefix(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 8080},
		Routes: []Route{
			{
				Name:        "echo",
				PathPrefix:  "/api/echo",
				StripPrefix: "api/echo",
				Methods:     []string{"GET"},
				Upstream:    "http://localhost:9091",
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "routes[0].strip_prefix") {
		t.Fatalf("expected strip_prefix validation error, got %v", err)
	}
}

func TestValidateRouteTimeoutMS(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 8080},
		Routes: []Route{
			{
				Name:       "echo",
				PathPrefix: "/api/echo",
				TimeoutMS:  -1,
				Methods:    []string{"GET"},
				Upstream:   "http://localhost:9091",
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "routes[0].timeout_ms") {
		t.Fatalf("expected timeout_ms validation error, got %v", err)
	}
}

func TestValidateRouteRateLimitRequests(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 8080},
		Routes: []Route{
			{
				Name:       "echo",
				PathPrefix: "/api/echo",
				Methods:    []string{"GET"},
				Upstream:   "http://localhost:9091",
				RateLimit: RateLimitConfig{
					Enabled:       true,
					Requests:      0,
					WindowSeconds: 60,
					KeyStrategy:   "ip",
				},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "routes[0].rate_limit.requests") {
		t.Fatalf("expected rate_limit.requests validation error, got %v", err)
	}
}

func TestValidateRouteRateLimitKeyStrategy(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 8080},
		Routes: []Route{
			{
				Name:       "echo",
				PathPrefix: "/api/echo",
				Methods:    []string{"GET"},
				Upstream:   "http://localhost:9091",
				RateLimit: RateLimitConfig{
					Enabled:       true,
					Requests:      2,
					WindowSeconds: 60,
					KeyStrategy:   "header",
				},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "routes[0].rate_limit.key_strategy") {
		t.Fatalf("expected rate_limit.key_strategy validation error, got %v", err)
	}
}

func TestValidateRedisEnabledRequiresAddr(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Port: 8080},
		Redis: RedisConfig{
			Enabled: true,
			Addr:    " ",
		},
		Routes: []Route{
			{
				Name:       "echo",
				PathPrefix: "/api/echo",
				Methods:    []string{"GET"},
				Upstream:   "http://localhost:9091",
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "redis.addr") {
		t.Fatalf("expected redis.addr validation error, got %v", err)
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "gateway.yaml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	return path
}
