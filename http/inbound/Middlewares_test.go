package inbound

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	"github.com/DerBlum/filmkritiken-backend/domain/session"
	"github.com/DerBlum/filmkritiken-backend/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
)

func TestAuthHandler_BffSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validSessID := "valid-session-123"
	expiredSessID := "expired-session-456"

	validSession := &session.Session{
		ID:          validSessID,
		Name:        "Stefan Blum",
		Permissions: []string{"film.add", "bewertung.add"},
		ExpiresAt:   time.Now().Add(time.Hour),
	}

	expiredSession := &session.Session{
		ID:          expiredSessID,
		Name:        "Expired User",
		Permissions: []string{"film.add"},
		ExpiresAt:   time.Now().Add(-time.Hour),
	}

	allowedRoles := []string{"bewertung.add"}

	t.Run("missing cookie returns 401", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockSessionRepository(ctrl)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.GET("/test", NewAuthHandler(mockRepo, allowedRoles), func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("expired session returns 401", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockSessionRepository(ctrl)
		mockRepo.EXPECT().FindSession(gomock.Any(), expiredSessID).Return(expiredSession, nil)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.GET("/test", NewAuthHandler(mockRepo, allowedRoles), func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: expiredSessID})
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("valid session with role returns 200 and sets username in context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockSessionRepository(ctrl)
		mockRepo.EXPECT().FindSession(gomock.Any(), validSessID).Return(validSession, nil)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)

		var capturedName string
		r.GET("/test", NewAuthHandler(mockRepo, allowedRoles), func(ctx *gin.Context) {
			capturedName, _ = ctx.Request.Context().Value(filmkritiken.Context_Username).(string)
			ctx.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: validSessID})
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if capturedName != "Stefan Blum" {
			t.Errorf("expected username 'Stefan Blum', got %q", capturedName)
		}
	})

	t.Run("valid session lacking required role returns 403", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRepo := mocks.NewMockSessionRepository(ctrl)
		mockRepo.EXPECT().FindSession(gomock.Any(), validSessID).Return(validSession, nil)

		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.GET("/test", NewAuthHandler(mockRepo, []string{"admin.only"}), func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: validSessID})
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", w.Code)
		}
	})
}
