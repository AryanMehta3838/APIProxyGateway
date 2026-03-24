package middleware

import (
	"net/http"
	"time"

	"github.com/aryan/apiproxy/internal/telemetry"
)

func Metrics(collector *telemetry.Collector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseCapture{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)

			routeName := "unknown"
			if slot, ok := r.Context().Value(routeNameSlotKey{}).(*string); ok && *slot != "" {
				routeName = *slot
			}

			collector.Observe(routeName, r.Method, rw.status, time.Since(start).Seconds())
		})
	}
}
