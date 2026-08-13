package domain

import (
	"context"
	"sync"
)

// MemorySessionRepository is an in-memory SessionRepository for MVP.
type MemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[SessionID]*Session
}

// NewMemorySessionRepository creates an empty in-memory store.
func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{
		sessions: make(map[SessionID]*Session),
	}
}

// Save stores a session.
func (r *MemorySessionRepository) Save(_ context.Context, session *Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *session
	turns := make([]Turn, len(session.Turns))
	copy(turns, session.Turns)
	cp.Turns = turns
	r.sessions[session.ID] = &cp
	return nil
}

// Get loads a session by id.
func (r *MemorySessionRepository) Get(_ context.Context, id SessionID) (*Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *s
	return &cp, nil
}
