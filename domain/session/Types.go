//go:generate mockgen -source=Session.go -destination=../../mocks/SessionRepository.go -package mocks

package session

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Session struct {
	ID          string    `json:"id" bson:"_id"`
	Name        string    `json:"name" bson:"name"`
	Permissions []string  `json:"permissions" bson:"permissions"`
	ExpiresAt   time.Time `json:"expiresAt" bson:"expiresAt"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt"`
}

func HashSessionID(sessionID string) string {
	if sessionID == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(sessionID))
	return hex.EncodeToString(hash[:])
}

type SessionRepository interface {
	SaveSession(ctx context.Context, session *Session) error
	FindSession(ctx context.Context, sessionID string) (*Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
	RefreshSession(ctx context.Context, sessionID string, duration time.Duration) error
}
