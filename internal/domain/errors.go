package domain

import "errors"

var (
	// ErrNotFound is returned when an entity is missing.
	ErrNotFound = errors.New("not found")
	// ErrMaxDepthExceeded is returned when the agent loop hits max depth.
	ErrMaxDepthExceeded = errors.New("max loop depth exceeded")
	// ErrSessionCancelled is returned when the session context is cancelled.
	ErrSessionCancelled = errors.New("session cancelled")
)
