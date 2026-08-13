package runtimeclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/budaev/stell/internal/runtimeclient"
	"github.com/budaev/stell/pkg/hmacauth"
)

func TestFailoverSkipsUnhealthy(t *testing.T) {
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			http.Error(w, "down", http.StatusServiceUnavailable)
			return
		}
		http.Error(w, "no", http.StatusInternalServerError)
	}))
	t.Cleanup(dead.Close)

	live := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"content": "ok"})
	}))
	t.Cleanup(live.Close)

	f, err := runtimeclient.NewFailover([]string{dead.URL, live.URL}, "", false)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := f.Execute(context.Background(), runtimeclient.ExecuteRequest{Tool: "glob", Args: map[string]any{"pattern": "*"}})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "ok" {
		t.Fatalf("content=%q", resp.Content)
	}
}

func TestHTTPClientStillSigns(t *testing.T) {
	var sawHMAC bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.WriteHeader(200)
			return
		}
		if r.Header.Get(hmacauth.HeaderSignature) != "" {
			sawHMAC = true
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"content": "x"})
	}))
	t.Cleanup(srv.Close)
	c := runtimeclient.NewHTTP(srv.URL).WithHMACKey("secret")
	_, err := c.Execute(context.Background(), runtimeclient.ExecuteRequest{Tool: "glob"})
	if err != nil {
		t.Fatal(err)
	}
	if !sawHMAC {
		t.Fatal("HMAC missing — auth must not be weakened")
	}
}
