package testkit

import (
	"encoding/json"
	"io"
	"net/http"
)

// maxEchoBody caps how much of the request body is echoed to avoid unbounded reads.
const maxEchoBody = 1 << 20 // 1 MiB

// EchoResponse is the JSON shape returned by EchoHandler for demos and tests.
type EchoResponse struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Query  string `json:"query,omitempty"`
	Body   string `json:"body,omitempty"`
}

// EchoHandler returns an HTTP handler that responds with JSON describing the request.
func EchoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body string
		if r.Body != nil {
			limited := http.MaxBytesReader(w, r.Body, maxEchoBody)
			b, err := io.ReadAll(limited)
			if err != nil {
				http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
				return
			}
			body = string(b)
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(true)
		_ = enc.Encode(EchoResponse{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  r.URL.RawQuery,
			Body:   body,
		})
	})
}
