package testkit

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEchoHandler_GET(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(EchoHandler())
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/hello?name=aryan")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type %q", ct)
	}

	var got EchoResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Method != http.MethodGet {
		t.Fatalf("method %q", got.Method)
	}
	if got.Path != "/hello" {
		t.Fatalf("path %q", got.Path)
	}
	if got.Query != "name=aryan" {
		t.Fatalf("query %q", got.Query)
	}
}

func TestEchoHandler_POSTBody(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(EchoHandler())
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/v1/ping", strings.NewReader(`{"ping":true}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	var got EchoResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Method != http.MethodPost {
		t.Fatalf("method %q", got.Method)
	}
	if got.Path != "/v1/ping" {
		t.Fatalf("path %q", got.Path)
	}
	if got.Body != `{"ping":true}` {
		t.Fatalf("body %q", got.Body)
	}
}

func TestEchoHandler_bodyTooLarge(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(EchoHandler())
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/", strings.NewReader(strings.Repeat("a", maxEchoBody+1)))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status %d", resp.StatusCode)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
}
