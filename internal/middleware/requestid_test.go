package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	t.Parallel()

	var gotHeader string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get(RequestIDHeader)
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if gotHeader == "" {
		t.Fatal("expected request id to be added to request header")
	}
	if got := recorder.Header().Get(RequestIDHeader); got == "" {
		t.Fatal("expected request id response header")
	} else if got != gotHeader {
		t.Fatalf("expected same request id on request and response, got request=%q response=%q", gotHeader, got)
	}
}

func TestRequestID_PreservesExistingValue(t *testing.T) {
	t.Parallel()

	const want = "existing-request-id"

	var gotHeader string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get(RequestIDHeader)
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, want)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if gotHeader != want {
		t.Fatalf("expected preserved request id %q, got %q", want, gotHeader)
	}
	if got := recorder.Header().Get(RequestIDHeader); got != want {
		t.Fatalf("expected preserved response header %q, got %q", want, got)
	}
}
