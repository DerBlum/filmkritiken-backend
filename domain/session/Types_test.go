package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/DerBlum/filmkritiken-backend/domain/session"
)

func TestHashSessionID(t *testing.T) {
	t.Run("returns empty string for empty input", func(t *testing.T) {
		if hash := session.HashSessionID(""); hash != "" {
			t.Errorf("expected empty string, got %s", hash)
		}
	})

	t.Run("returns correct SHA256 hex hash for session ID", func(t *testing.T) {
		sessionID := "test-session-uuid-12345"
		expectedHashRaw := sha256.Sum256([]byte(sessionID))
		expectedHash := hex.EncodeToString(expectedHashRaw[:])

		hash := session.HashSessionID(sessionID)
		if hash != expectedHash {
			t.Errorf("expected hash %s, got %s", expectedHash, hash)
		}
	})
}
