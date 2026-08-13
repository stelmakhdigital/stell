package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/budaev/stell/internal/agent"
	"github.com/budaev/stell/internal/eventbus"
	"github.com/budaev/stell/internal/hooks"
	"github.com/budaev/stell/pkg/protocol"
	"github.com/google/uuid"
)

// Spawner builds an isolated agent+bus for one public session.
type Spawner func(approver hooks.Approver) (*agent.Agent, *eventbus.Bus)

// Server is the public HTTP gateway (token auth, not HMAC).
type Server struct {
	Token    string
	Spawn    Spawner
	mu       sync.Mutex
	sessions map[string]*run
}

type run struct {
	mu      sync.Mutex
	cancel  context.CancelFunc
	events  []protocol.Event
	subs    []chan protocol.Event
	hitl    *hooks.ChanApprover
	running bool
	final   string
	err     string
}

// New creates a gateway.
func New(token string, spawn Spawner) *Server {
	return &Server{Token: token, Spawn: spawn, sessions: make(map[string]*run)}
}

// Handler returns the mux.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("POST /v1/sessions", s.handleCreate)
	mux.HandleFunc("GET /v1/sessions/{id}", s.handleGet)
	mux.HandleFunc("GET /v1/sessions/{id}/events", s.handleEvents)
	mux.HandleFunc("POST /v1/sessions/{id}/cancel", s.handleCancel)
	mux.HandleFunc("POST /v1/sessions/{id}/hitl", s.handleHITL)
	mux.HandleFunc("POST /v1/sessions/{id}/messages", s.handleCreate)
	return s.cors(s.auth(mux))
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-API-Token")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		if s.Token == "" {
			http.Error(w, "api token not configured", http.StatusUnauthorized)
			return
		}
		got := r.Header.Get("Authorization")
		got = strings.TrimPrefix(got, "Bearer ")
		if got == "" {
			got = r.Header.Get("X-API-Token")
		}
		if got != s.Token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	if s.Spawn == nil {
		http.Error(w, "spawner not configured", http.StatusInternalServerError)
		return
	}
	var req protocol.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Message) == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}
	hitl := hooks.NewChanApprover()
	a, bus := s.Spawn(hitl)
	id := uuid.NewString()
	if r.PathValue("id") != "" {
		id = r.PathValue("id")
	}
	ctx, cancel := context.WithCancel(context.Background())
	rn := &run{cancel: cancel, hitl: hitl, running: true}
	s.subscribe(bus, id, rn)
	s.mu.Lock()
	s.sessions[id] = rn
	s.mu.Unlock()

	go func() {
		res := a.Run(ctx, req.Message)
		rn.mu.Lock()
		rn.running = false
		rn.final = res.FinalText
		if res.Err != nil {
			rn.err = res.Err.Error()
		}
		subs := append([]chan protocol.Event(nil), rn.subs...)
		rn.mu.Unlock()
		for _, ch := range subs {
			close(ch)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(protocol.CreateSessionResponse{SessionID: id})
}

func (s *Server) subscribe(bus *eventbus.Bus, id string, rn *run) {
	types := []eventbus.EventType{
		eventbus.EventSessionStart, eventbus.EventSessionEnd,
		eventbus.EventTurnStart, eventbus.EventTurnEnd,
		eventbus.EventToolCall, eventbus.EventToolResult,
		eventbus.EventModelRequest, eventbus.EventModelResponse,
		eventbus.EventAgentError, eventbus.EventGuardrailBlock,
		eventbus.EventHITLRequest, eventbus.EventHITLDecision,
		eventbus.EventContextCompact,
	}
	for _, t := range types {
		t := t
		bus.Subscribe(t, func(e *eventbus.Event) (*eventbus.EventResult, error) {
			pe := protocol.FromBus(e)
			pe.SessionID = id
			rn.mu.Lock()
			rn.events = append(rn.events, pe)
			subs := append([]chan protocol.Event(nil), rn.subs...)
			rn.mu.Unlock()
			for _, ch := range subs {
				select {
				case ch <- pe:
				default:
				}
			}
			return nil, nil
		})
	}
}

func (s *Server) get(id string) *run {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessions[id]
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	rn := s.get(r.PathValue("id"))
	if rn == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	rn.mu.Lock()
	st := protocol.SessionStatus{SessionID: r.PathValue("id"), Running: rn.running, FinalText: rn.final, Error: rn.err}
	rn.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(st)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	rn := s.get(r.PathValue("id"))
	if rn == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	ch := make(chan protocol.Event, 32)
	rn.mu.Lock()
	past := append([]protocol.Event(nil), rn.events...)
	done := !rn.running
	if !done {
		rn.subs = append(rn.subs, ch)
	}
	rn.mu.Unlock()
	for _, ev := range past {
		writeSSE(w, ev)
	}
	flusher.Flush()
	if done {
		return
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			writeSSE(w, ev)
			flusher.Flush()
		case <-time.After(25 * time.Second):
			_, _ = w.Write([]byte(": ping\n\n"))
			flusher.Flush()
		}
	}
}

func writeSSE(w http.ResponseWriter, ev protocol.Event) {
	data, _ := json.Marshal(ev)
	_, _ = w.Write([]byte("event: " + ev.Type + "\n"))
	_, _ = w.Write([]byte("data: "))
	_, _ = w.Write(data)
	_, _ = w.Write([]byte("\n\n"))
}

func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request) {
	rn := s.get(r.PathValue("id"))
	if rn == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	rn.cancel()
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) handleHITL(w http.ResponseWriter, r *http.Request) {
	rn := s.get(r.PathValue("id"))
	if rn == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var req protocol.HITLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	select {
	case rn.hitl.C <- req.Decision:
		w.WriteHeader(http.StatusAccepted)
	default:
		http.Error(w, "no pending hitl", http.StatusConflict)
	}
}
