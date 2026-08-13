package domain

import "context"

// SessionRepository persists and loads sessions.
type SessionRepository interface {
	Save(ctx context.Context, session *Session) error
	Get(ctx context.Context, id SessionID) (*Session, error)
}
