package api

import (
	"encoding/json"
	"net/http"

	"github.com/budaev/agent/pkg/hmacauth"
	"github.com/budaev/agent/runtime/executor"
)

// Server exposes Hands over HTTP.
type Server struct {
	Exec        *executor.Executor
	HMACKey     []byte
	RequireHMAC bool
}

type executeRequest struct {
	Tool      string         `json:"tool"`
	Args      map[string]any `json:"args"`
	Workspace string         `json:"workspace"`
	TimeoutMs int            `json:"timeout_ms"`
}

type executeResponse struct {
	Content   string         `json:"content"`
	Truncated bool           `json:"truncated,omitempty"`
	Error     string         `json:"error,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Handler returns the HTTP mux.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/v1/execute", s.handleExecute)
	if s.RequireHMAC || len(s.HMACKey) > 0 {
		return hmacauth.Middleware(s.HMACKey, mux)
	}
	return mux
}

func (s *Server) handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req executeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := s.Exec.Execute(r.Context(), executor.Request{
		Tool:      req.Tool,
		Args:      req.Args,
		Workspace: req.Workspace,
		TimeoutMs: req.TimeoutMs,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(executeResponse{
		Content:   res.Content,
		Truncated: res.Truncated,
		Error:     res.Error,
		Metadata:  res.Metadata,
	})
}
