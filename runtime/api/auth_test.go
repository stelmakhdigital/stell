package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/budaev/stell/pkg/hmacauth"
	"github.com/budaev/stell/runtime/api"
	"github.com/budaev/stell/runtime/executor"
	"github.com/budaev/stell/runtime/sandbox"
)

func TestExecuteRequiresHMAC(t *testing.T) {
	srv := &api.Server{
		Exec:        executor.New(sandbox.NewDocker(sandbox.DefaultPolicy())),
		HMACKey:     []byte("k"),
		RequireHMAC: true,
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	resp, err := http.Post(ts.URL+"/v1/execute", "application/json", bytes.NewReader([]byte(`{"tool":"glob","args":{"pattern":"*"}}`)))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status=%d", resp.StatusCode)
	}

	body, _ := json.Marshal(map[string]any{"tool": "glob", "args": map[string]any{"pattern": "*"}})
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	hmacauth.SignRequest(req, []byte("k"), body, time.Now(), "n1")
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode == http.StatusUnauthorized {
		t.Fatal("valid hmac rejected")
	}
}
