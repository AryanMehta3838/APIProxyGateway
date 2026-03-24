package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/aryan/apiproxy/internal/config"
	"github.com/aryan/apiproxy/internal/middleware"
)

func NewRouter() http.Handler {
	return NewRouterWithMetrics(nil, config.Config{})
}

func NewRouterWithMetrics(metricsHandler http.Handler, cfg config.Config) http.Handler {
	router := chi.NewRouter()
	router.Method(http.MethodGet, "/healthz", middleware.NamedRoute("healthz", http.HandlerFunc(healthz)))
	router.Method(http.MethodGet, "/readyz", middleware.NamedRoute("readyz", http.HandlerFunc(readyz)))
	router.Method(http.MethodGet, "/debug/routes", middleware.NamedRoute("debug_routes", debugRoutesHandler(cfg)))
	if metricsHandler != nil {
		router.Method(http.MethodGet, "/metrics", middleware.NamedRoute("metrics", metricsHandler))
	}
	return router
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func readyz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready\n"))
}
