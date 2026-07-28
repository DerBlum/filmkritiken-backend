package session

import (
	"context"
	"time"
)

//go:generate mockgen -source=SessionRepository.go -destination=../../mocks/SessionRepository.go -package mocks
type SessionRepository interface {
	SaveSession(ctx context.Context, session *Session) error
	FindSession(ctx context.Context, sessionID string) (*Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
	RefreshSession(ctx context.Context, sessionID string, duration time.Duration) error
}
