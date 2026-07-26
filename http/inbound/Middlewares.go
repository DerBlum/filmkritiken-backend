package inbound

import (
	"context"
	"net/http"
	"time"

	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	"github.com/DerBlum/filmkritiken-backend/domain/session"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func TraceIdMiddleware(ginCtx *gin.Context) {
	id := generateTraceId()
	newCtx := context.WithValue(ginCtx.Request.Context(), filmkritiken.Context_TraceId, id)
	ginCtx.Request = ginCtx.Request.WithContext(newCtx)
}

func generateTraceId() string {
	id, err := uuid.NewUUID()
	if err != nil {
		return ""
	}
	return id.String()
}

func NewEmptyHandler() func(ginCtx *gin.Context) {
	return func(ginCtx *gin.Context) {
		// do nothing
	}
}

func NewAuthHandler(sessionRepo session.SessionRepository, allowedRoles []string) func(ginCtx *gin.Context) {
	return func(ginCtx *gin.Context) {
		authHandler(ginCtx, sessionRepo, allowedRoles)
	}
}

func authHandler(ginCtx *gin.Context, sessionRepo session.SessionRepository, allowedRoles []string) {
	if sessionRepo == nil {
		log.Warn("sessionRepo not configured in authHandler")
		ginCtx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	sessionID, err := ginCtx.Cookie(SessionCookieName)
	if err != nil || sessionID == "" {
		log.Warn("received request without valid session cookie")
		ginCtx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	sess, err := sessionRepo.FindSession(ginCtx.Request.Context(), sessionID)
	if err != nil || sess == nil || sess.ExpiresAt.Before(time.Now()) {
		log.Warnf("session invalid or expired: %v", err)
		ginCtx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if !hasSessionRole(allowedRoles, sess.Permissions) {
		log.Warnf("session user %s lacks required roles %v", sess.Name, allowedRoles)
		ginCtx.AbortWithStatus(http.StatusForbidden)
		return
	}

	newCtx := context.WithValue(ginCtx.Request.Context(), filmkritiken.Context_Username, sess.Name)
	ginCtx.Request = ginCtx.Request.WithContext(newCtx)
}

func hasSessionRole(allowedRoles []string, userPermissions []string) bool {
	if len(allowedRoles) == 0 {
		return true
	}
	for _, role := range userPermissions {
		for _, allowed := range allowedRoles {
			if role == allowed {
				return true
			}
		}
	}
	return false
}
