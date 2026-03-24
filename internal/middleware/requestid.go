package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const RequestIDHeader = "X-Request-ID"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = newRequestID()
			r.Header.Set(RequestIDHeader, requestID)
		}

		w.Header().Set(RequestIDHeader, requestID)
		next.ServeHTTP(w, r)
	})
}

func newRequestID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "request-id-unavailable"
	}
	return hex.EncodeToString(buf[:])
}
