package hmacauth_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/budaev/agent/pkg/hmacauth"
)

func TestMiddlewareRejectsMissingAndReplay(t *testing.T) {
	key := []byte("test-secret-key")
	h := hmacauth.Middleware(key, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/execute", bytes.NewReader([]byte(`{}`)))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("missing hmac code=%d", rr.Code)
	}

	body := []byte(`{"tool":"glob"}`)
	ok := httptest.NewRequest(http.MethodPost, "/v1/execute", bytes.NewReader(body))
	hmacauth.SignRequest(ok, key, body, time.Now(), "nonce-1")
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, ok)
	if rr.Code != http.StatusOK {
		t.Fatalf("valid hmac code=%d body=%s", rr.Code, rr.Body.String())
	}

	replay := httptest.NewRequest(http.MethodPost, "/v1/execute", bytes.NewReader(body))
	hmacauth.SignRequest(replay, key, body, time.Now(), "nonce-1")
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, replay)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("replay code=%d", rr.Code)
	}
}

func TestRejectsOldTimestamp(t *testing.T) {
	key := []byte("test-secret-key")
	h := hmacauth.Middleware(key, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	body := []byte(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/execute", bytes.NewReader(body))
	hmacauth.SignRequest(req, key, body, time.Now().Add(-2*time.Minute), "nonce-old")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("old ts code=%d", rr.Code)
	}
}
