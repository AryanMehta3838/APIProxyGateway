package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAccessLog_JSONShape(t *testing.T) {
	var buf bytes.Buffer
	accessLogMu.Lock()
	orig := accessLogOut
	accessLogOut = &buf
	accessLogMu.Unlock()
	t.Cleanup(func() {
		accessLogMu.Lock()
		accessLogOut = orig
		accessLogMu.Unlock()
	})

	chain := AccessLog(RequestID(NamedRoute("echo", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))))

	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("expected access log line")
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatalf("unmarshal log: %v", err)
	}

	if got["msg"] != "access" {
		t.Fatalf("msg %v", got["msg"])
	}
	if got["route_name"] != "echo" {
		t.Fatalf("route_name %v", got["route_name"])
	}
	if got["method"] != http.MethodGet {
		t.Fatalf("method %v", got["method"])
	}
	if got["path"] != "/api/x" {
		t.Fatalf("path %v", got["path"])
	}
	if got["status"] != float64(http.StatusCreated) {
		t.Fatalf("status %v", got["status"])
	}
	if got["request_id"] == nil || got["request_id"] == "" {
		t.Fatal("expected request_id")
	}
	if got["ts"] == nil || got["ts"] == "" {
		t.Fatal("expected ts")
	}
	if _, ok := got["duration_ms"].(float64); !ok {
		t.Fatalf("duration_ms type %T", got["duration_ms"])
	}
	if strings.Contains(line, "Authorization") {
		t.Fatal("log should not contain Authorization")
	}
}

func TestAccessLog_DoesNotDumpAuthorizationHeader(t *testing.T) {
	var buf bytes.Buffer
	accessLogMu.Lock()
	orig := accessLogOut
	accessLogOut = &buf
	accessLogMu.Unlock()
	t.Cleanup(func() {
		accessLogMu.Lock()
		accessLogOut = orig
		accessLogMu.Unlock()
	})

	chain := AccessLog(RequestID(NamedRoute("healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	line := buf.String()
	if strings.Contains(line, "secret-token") || strings.Contains(line, "Bearer") {
		t.Fatalf("log leaked auth header: %s", line)
	}
}

func TestAccessLog_UnknownRouteName(t *testing.T) {
	var buf bytes.Buffer
	accessLogMu.Lock()
	orig := accessLogOut
	accessLogOut = &buf
	accessLogMu.Unlock()
	t.Cleanup(func() {
		accessLogMu.Lock()
		accessLogOut = orig
		accessLogMu.Unlock()
	})

	chain := AccessLog(RequestID(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})))

	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["route_name"] != "unknown" {
		t.Fatalf("route_name %v", got["route_name"])
	}
}
