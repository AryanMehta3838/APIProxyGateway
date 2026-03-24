package admin

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/aryan/apiproxy/internal/config"
)

// debugRoutesResponse is the JSON shape for GET /debug/routes.
// Upstream targets expose only scheme and host (no path, query, or userinfo).
type debugRoutesResponse struct {
	Redis  redisDebugInfo   `json:"redis"`
	Routes []routeDebugInfo `json:"routes"`
}

type redisDebugInfo struct {
	Enabled bool   `json:"enabled"`
	Addr    string `json:"addr,omitempty"`
}

type routeDebugInfo struct {
	Name           string              `json:"name"`
	PathPrefix     string              `json:"path_prefix"`
	StripPrefix    string              `json:"strip_prefix,omitempty"`
	Methods        []string            `json:"methods"`
	TimeoutMS      int                 `json:"timeout_ms,omitempty"`
	UpstreamScheme string              `json:"upstream_scheme,omitempty"`
	UpstreamHost   string              `json:"upstream_host,omitempty"`
	RateLimit      rateLimitDebugInfo  `json:"rate_limit"`
}

type rateLimitDebugInfo struct {
	Enabled       bool   `json:"enabled"`
	Requests      int    `json:"requests,omitempty"`
	WindowSeconds int    `json:"window_seconds,omitempty"`
	KeyStrategy   string `json:"key_strategy,omitempty"`
}

func debugRoutesHandler(cfg config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		out := debugRoutesResponse{
			Redis: redisDebugInfo{
				Enabled: cfg.Redis.Enabled,
			},
			Routes: make([]routeDebugInfo, 0, len(cfg.Routes)),
		}
		if cfg.Redis.Enabled {
			out.Redis.Addr = sanitizeRedisAddr(cfg.Redis.Addr)
		}

		for _, rt := range cfg.Routes {
			scheme, host := upstreamSchemeHost(rt.Upstream)
			rl := rateLimitDebugInfo{
				Enabled:     rt.RateLimit.Enabled,
				KeyStrategy: rt.RateLimit.KeyStrategy,
			}
			if rt.RateLimit.Enabled {
				rl.Requests = rt.RateLimit.Requests
				rl.WindowSeconds = rt.RateLimit.WindowSeconds
			}

			out.Routes = append(out.Routes, routeDebugInfo{
				Name:           rt.Name,
				PathPrefix:     rt.PathPrefix,
				StripPrefix:    rt.StripPrefix,
				Methods:        append([]string(nil), rt.Methods...),
				TimeoutMS:      rt.TimeoutMS,
				UpstreamScheme: scheme,
				UpstreamHost:   host,
				RateLimit:      rl,
			})
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(true)
		_ = enc.Encode(out)
	})
}

func upstreamSchemeHost(raw string) (scheme, host string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", ""
	}
	return u.Scheme, u.Host
}

func sanitizeRedisAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	if !strings.Contains(addr, "://") {
		return addr
	}
	u, err := url.Parse(addr)
	if err != nil || u.Host == "" {
		return ""
	}
	return u.Host
}
