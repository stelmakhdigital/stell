package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/budaev/stell/pkg/observability"
)

const rotateBytes = 32 * 1024 * 1024

// Record is one DCL audit line.
type Record struct {
	At           time.Time      `json:"at"`
	SessionID    string         `json:"session_id"`
	TurnID       string         `json:"turn_id,omitempty"`
	Tool         string         `json:"tool"`
	ArgsRedacted map[string]any `json:"args_redacted,omitempty"`
	ArgsHash     string         `json:"args_hash,omitempty"`
	Blocked      bool           `json:"blocked,omitempty"`
	Error        bool           `json:"error,omitempty"`
	Kind         string         `json:"kind"`
}

// Store appends JSONL audit records. Writes are serialized.
type Store struct {
	mu   sync.Mutex
	path string
}

// NewStore creates an append-only JSONL store.
func NewStore(path string) *Store {
	return &Store{path: path}
}

// Append writes one record. Empty path is a no-op.
func (s *Store) Append(rec Record) error {
	if s == nil || s.path == "" {
		return nil
	}
	if rec.At.IsZero() {
		rec.At = time.Now().UTC()
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	if err := rotateIfNeeded(s.path); err != nil {
		return err
	}
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(data, '\n'))
	return err
}

// HashArgs hashes canonical JSON of args (unredacted) for correlation.
func HashArgs(args map[string]any) string {
	if args == nil {
		return ""
	}
	data, err := json.Marshal(args)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// RedactArgs returns a copy safe for logs.
func RedactArgs(args map[string]any) map[string]any {
	return observability.RedactMap(args)
}

func rotateIfNeeded(path string) error {
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if st.Size() < rotateBytes {
		return nil
	}
	bak := path + ".1"
	_ = os.Remove(bak)
	return os.Rename(path, bak)
}
