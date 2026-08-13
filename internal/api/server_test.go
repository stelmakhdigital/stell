package api_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/budaev/stell/internal/agent"
	pubapi "github.com/budaev/stell/internal/api"
	"github.com/budaev/stell/internal/eventbus"
	"github.com/budaev/stell/internal/hooks"
	"github.com/budaev/stell/internal/llm"
	"github.com/budaev/stell/pkg/protocol"
)

func spawnFake(inner llm.Provider) pubapi.Spawner {
	return func(hooks.Approver) (*agent.Agent, *eventbus.Bus) {
		bus := eventbus.New()
		a := agent.New(agent.WithProvider(inner), agent.WithEventBus(bus))
		return a, bus
	}
}

func TestCreateStreamAndAuth(t *testing.T) {
	srv := pubapi.New("tok", spawnFake(llm.NewFake(llm.Response{
		Message: llm.Message{Role: llm.RoleAssistant, Content: "hello from agent"},
	})))
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	resp, err := http.Post(ts.URL+"/v1/sessions", "application/json", bytes.NewReader([]byte(`{"message":"hi"}`)))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/sessions", bytes.NewReader([]byte(`{"message":"hi"}`)))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("create=%d", resp.StatusCode)
	}
	var created protocol.CreateSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	evReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL+"/v1/sessions/"+created.SessionID+"/events", nil)
	evReq.Header.Set("Authorization", "Bearer tok")
	evResp, err := http.DefaultClient.Do(evReq)
	if err != nil {
		t.Fatal(err)
	}
	defer evResp.Body.Close()
	br := bufio.NewReader(evResp.Body)
	var blob strings.Builder
	for {
		line, err := br.ReadString('\n')
		blob.WriteString(line)
		if strings.Contains(blob.String(), "session_start") || strings.Contains(blob.String(), "session_end") || strings.Contains(blob.String(), "model_response") {
			return
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
	}
	t.Fatalf("expected streamed events, got %q", blob.String())
}

func TestCancelInterrupts(t *testing.T) {
	block := make(chan struct{})
	p := &waitProvider{block: block}
	srv := pubapi.New("tok", spawnFake(p))
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/sessions", bytes.NewReader([]byte(`{"message":"slow"}`)))
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var created protocol.CreateSessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()

	time.Sleep(50 * time.Millisecond)
	cReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/sessions/"+created.SessionID+"/cancel", nil)
	cReq.Header.Set("Authorization", "Bearer tok")
	cResp, err := http.DefaultClient.Do(cReq)
	if err != nil {
		t.Fatal(err)
	}
	cResp.Body.Close()
	if cResp.StatusCode != http.StatusAccepted {
		t.Fatalf("cancel=%d", cResp.StatusCode)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		gReq, _ := http.NewRequest(http.MethodGet, ts.URL+"/v1/sessions/"+created.SessionID, nil)
		gReq.Header.Set("Authorization", "Bearer tok")
		gResp, err := http.DefaultClient.Do(gReq)
		if err != nil {
			t.Fatal(err)
		}
		var st protocol.SessionStatus
		_ = json.NewDecoder(gResp.Body).Decode(&st)
		gResp.Body.Close()
		if !st.Running {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("session still running after cancel")
}

type waitProvider struct{ block chan struct{} }

func (p *waitProvider) Name() string { return "wait" }
func (p *waitProvider) Generate(ctx context.Context, _ llm.Request) (llm.Response, error) {
	select {
	case <-ctx.Done():
		return llm.Response{}, ctx.Err()
	case <-p.block:
		return llm.Response{Message: llm.Message{Role: llm.RoleAssistant, Content: "late"}}, nil
	}
}
