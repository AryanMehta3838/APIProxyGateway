package router

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/aryan/apiproxy/internal/config"
	"github.com/aryan/apiproxy/internal/middleware"
	"github.com/aryan/apiproxy/internal/proxy"
	"github.com/aryan/apiproxy/internal/ratelimit"
)

func New(cfg config.Config, adminHandler http.Handler) (http.Handler, error) {
	router := chi.NewRouter()
	router.Use(middleware.AccessLog)
	router.Use(middleware.RequestID)
	router.Mount("/", adminHandler)
	limiter, err := buildLimiter(cfg)
	if err != nil {
		return nil, err
	}

	for _, route := range cfg.Routes {
		handler, err := proxy.New(route)
		if err != nil {
			return nil, fmt.Errorf("build proxy for route %q: %w", route.Name, err)
		}
		if route.RateLimit.Enabled {
			handler = middleware.RateLimit(
				route.Name,
				limiter,
				middleware.RateLimitPolicy(route.RateLimit.Requests, route.RateLimit.WindowSeconds),
				route.RateLimit.KeyStrategy,
				handler,
			)
		}
		handler = middleware.NamedRoute(route.Name, handler)

		for _, method := range route.Methods {
			router.Method(method, route.PathPrefix, handler)
			router.Method(method, route.PathPrefix+"/*", handler)
		}
	}

	return router, nil
}

func buildLimiter(cfg config.Config) (ratelimit.Limiter, error) {
	needsLimiter := false
	for _, route := range cfg.Routes {
		if route.RateLimit.Enabled {
			needsLimiter = true
			break
		}
	}
	if !needsLimiter {
		return nil, nil
	}

	if cfg.Redis.Enabled {
		limiter, err := ratelimit.NewRedis(cfg.Redis.Addr)
		if err != nil {
			return nil, fmt.Errorf("build redis limiter: %w", err)
		}
		return limiter, nil
	}

	return ratelimit.NewInMemory(), nil
}
