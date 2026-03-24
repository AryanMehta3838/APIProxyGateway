package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type routeNameSlotKey struct{}

// NamedRoute records the logical route name for AccessLog by filling a *string slot
// placed on the request context by AccessLog (same *http.Request is propagated).
func NamedRoute(name string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if slot, ok := r.Context().Value(routeNameSlotKey{}).(*string); ok {
			*slot = name
		}
		next.ServeHTTP(w, r)
	})
}

// accessLogWriter is swapped in tests to capture JSON lines.
var (
	accessLogMu  sync.Mutex
	accessLogOut io.Writer = os.Stdout
)

// AccessLog emits one JSON line per request with status, latency, route name, and request ID.
// Request headers are not logged (avoids leaking sensitive values).
func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseCapture{ResponseWriter: w, status: http.StatusOK}
		var routeName string
		ctx := context.WithValue(r.Context(), routeNameSlotKey{}, &routeName)
		next.ServeHTTP(rw, r.WithContext(ctx))

		name := routeName
		if name == "" {
			name = "unknown"
		}

		entry := accessLogEntry{
			Msg:        "access",
			Time:       time.Now().UTC().Format(time.RFC3339Nano),
			RequestID:  r.Header.Get(RequestIDHeader),
			RouteName:  name,
			Method:     r.Method,
			Path:       r.URL.Path,
			Status:     rw.status,
			DurationMs: float64(time.Since(start).Microseconds()) / 1000,
		}

		accessLogMu.Lock()
		enc := json.NewEncoder(accessLogOut)
		enc.SetEscapeHTML(true)
		_ = enc.Encode(entry)
		accessLogMu.Unlock()
	})
}

type accessLogEntry struct {
	Msg        string  `json:"msg"`
	Time       string  `json:"ts"`
	RequestID  string  `json:"request_id"`
	RouteName  string  `json:"route_name"`
	Method     string  `json:"method"`
	Path       string  `json:"path"`
	Status     int     `json:"status"`
	DurationMs float64 `json:"duration_ms"`
}

type responseCapture struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *responseCapture) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseCapture) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.status = http.StatusOK
		w.wroteHeader = true
	}
	return w.ResponseWriter.Write(b)
}
