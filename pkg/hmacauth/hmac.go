package hmacauth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	HeaderTimestamp = "X-Agent-Timestamp"
	HeaderNonce     = "X-Agent-Nonce"
	HeaderSignature = "X-Agent-Signature"
	DefaultSkew     = 30 * time.Second
)

// Sign returns hex(hmac-sha256(key, ts + "\n" + nonce + "\n" + sha256(body))).
func Sign(key []byte, ts int64, nonce string, body []byte) string {
	sum := sha256.Sum256(body)
	mac := hmac.New(sha256.New, key)
	fmt.Fprintf(mac, "%d\n%s\n%x", ts, nonce, sum)
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify checks signature, skew, and nonce replay.
func Verify(key []byte, ts int64, nonce, sig string, body []byte, now time.Time, skew time.Duration, seen *NonceCache) error {
	if len(key) == 0 {
		return fmt.Errorf("hmac key is empty")
	}
	if nonce == "" {
		return fmt.Errorf("missing nonce")
	}
	if skew <= 0 {
		skew = DefaultSkew
	}
	t := time.Unix(ts, 0)
	if now.Sub(t) > skew || t.Sub(now) > skew {
		return fmt.Errorf("timestamp outside skew window")
	}
	want, err := hex.DecodeString(sig)
	if err != nil {
		return fmt.Errorf("invalid signature encoding")
	}
	got, err := hex.DecodeString(Sign(key, ts, nonce, body))
	if err != nil {
		return err
	}
	if !hmac.Equal(want, got) {
		return fmt.Errorf("invalid hmac signature")
	}
	if seen != nil && !seen.Accept(nonce, ts) {
		return fmt.Errorf("replayed nonce")
	}
	return nil
}

// SignRequest adds HMAC headers to an outgoing request.
func SignRequest(req *http.Request, key []byte, body []byte, now time.Time, nonce string) {
	ts := now.Unix()
	req.Header.Set(HeaderTimestamp, strconv.FormatInt(ts, 10))
	req.Header.Set(HeaderNonce, nonce)
	req.Header.Set(HeaderSignature, Sign(key, ts, nonce, body))
}

// Middleware protects /v1/execute. /healthz is left open.
func Middleware(key []byte, next http.Handler) http.Handler {
	seen := NewNonceCache(2 * DefaultSkew)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		ts, _ := strconv.ParseInt(r.Header.Get(HeaderTimestamp), 10, 64)
		err = Verify(key, ts, r.Header.Get(HeaderNonce), r.Header.Get(HeaderSignature), body, time.Now(), DefaultSkew, seen)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NonceCache rejects reused nonces for a TTL.
type NonceCache struct {
	mu   sync.Mutex
	ttl  time.Duration
	seen map[string]int64
}

// NewNonceCache creates a replay cache.
func NewNonceCache(ttl time.Duration) *NonceCache {
	if ttl <= 0 {
		ttl = time.Minute
	}
	return &NonceCache{ttl: ttl, seen: make(map[string]int64)}
}

// Accept returns false if nonce was already used.
func (c *NonceCache) Accept(nonce string, ts int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	cutoff := time.Now().Unix() - int64(c.ttl.Seconds())
	for k, v := range c.seen {
		if v < cutoff {
			delete(c.seen, k)
		}
	}
	if _, ok := c.seen[nonce]; ok {
		return false
	}
	c.seen[nonce] = ts
	return true
}
