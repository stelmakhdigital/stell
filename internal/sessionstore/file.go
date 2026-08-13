package sessionstore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/budaev/agent/internal/domain"
)

// FileStore persists sessions as JSON files (shared volume for Brain replicas).
type FileStore struct {
	dir string
	mu  sync.Mutex
}

// NewFileStore creates a directory-backed session repository.
func NewFileStore(dir string) (*FileStore, error) {
	if dir == "" {
		return nil, fmt.Errorf("session store dir is required")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &FileStore{dir: dir}, nil
}

func (s *FileStore) path(id domain.SessionID) string {
	return filepath.Join(s.dir, string(id)+".json")
}

// Save writes a session atomically.
func (s *FileStore) Save(_ context.Context, session *domain.Session) error {
	if session == nil {
		return fmt.Errorf("nil session")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path(session.ID) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path(session.ID))
}

// Get loads a session.
func (s *FileStore) Get(_ context.Context, id domain.SessionID) (*domain.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	var sess domain.Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}
