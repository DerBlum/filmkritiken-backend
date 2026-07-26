package inbound

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/DerBlum/filmkritiken-backend/domain/session"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const (
	SessionCookieName    = "session_id"
	StateCookieName      = "oauth_state"
	VerifierCookieName   = "oauth_verifier"
	RedirectCookieName   = "oauth_redirect"
	SecondsPerDay        = 86400
	OAuthStateTTLSeconds = 600
	StateTokenBytes      = 16
)

type AuthConfig struct {
	EntraTenantID       string `env:"ENTRA_TENANT_ID"`
	EntraClientID       string `env:"ENTRA_CLIENT_ID"`
	EntraClientSecret   string `env:"ENTRA_CLIENT_SECRET,unset"`
	EntraRedirectURI    string `env:"ENTRA_REDIRECT_URI"`
	FrontendURL         string `env:"FRONTEND_URL" envDefault:"http://localhost:5173"`
	SessionDurationDays int    `env:"SESSION_DURATION_DAYS" envDefault:"7"`
}

type BffAuthHandler struct {
	config      *AuthConfig
	sessionRepo session.SessionRepository
	oauthConfig *oauth2.Config
}

func NewBffAuthHandler(config *AuthConfig, sessionRepo session.SessionRepository) *BffAuthHandler {
	clientSecret := config.EntraClientSecret
	if len(clientSecret) > 0 && (clientSecret[0] == '<' || clientSecret == "unset") {
		clientSecret = ""
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.EntraClientID,
		ClientSecret: clientSecret,
		RedirectURL:  config.EntraRedirectURI,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/" + config.EntraTenantID + "/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/" + config.EntraTenantID + "/oauth2/v2.0/token",
		},
		Scopes: []string{"openid", "profile", "offline_access"},
	}

	return &BffAuthHandler{
		config:      config,
		sessionRepo: sessionRepo,
		oauthConfig: oauthConfig,
	}
}

func (h *BffAuthHandler) setCookie(c *gin.Context, name, value string, maxAge int) {
	isHTTPS := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	if isHTTPS {
		c.SetSameSite(http.SameSiteNoneMode)
		c.SetCookie(name, value, maxAge, "/", "", true, true)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(name, value, maxAge, "/", "", false, true)
	}
}

func (h *BffAuthHandler) handleLogin(c *gin.Context) {
	b := make([]byte, StateTokenBytes)
	_, _ = rand.Read(b)
	state := hex.EncodeToString(b)

	verifier := oauth2.GenerateVerifier()

	h.setCookie(c, StateCookieName, state, OAuthStateTTLSeconds)
	h.setCookie(c, VerifierCookieName, verifier, OAuthStateTTLSeconds)

	redirectPath := c.Query("redirect")
	if redirectPath == "" {
		redirectPath = c.Query("returnUrl")
	}
	if redirectPath != "" && strings.HasPrefix(redirectPath, "/") && !strings.HasPrefix(redirectPath, "//") {
		h.setCookie(c, RedirectCookieName, redirectPath, OAuthStateTTLSeconds)
	}

	authURL := h.oauthConfig.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
	c.Redirect(http.StatusFound, authURL)
}

func (h *BffAuthHandler) handleCallback(c *gin.Context) {
	errorQuery := c.Query("error")
	if errorQuery != "" {
		errorDesc := c.Query("error_description")
		log.Warnf("EntraID authorization error: %s - %s", errorQuery, errorDesc)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             errorQuery,
			"error_description": errorDesc,
		})
		return
	}

	stateQuery := c.Query("state")
	codeQuery := c.Query("code")

	stateCookie, err := c.Cookie(StateCookieName)
	if err != nil || stateCookie == "" || stateCookie != stateQuery {
		log.Warnf("State cookie mismatch or missing: stateQuery=%s, stateCookie=%s, err=%v", stateQuery, stateCookie, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OAuth state parameter"})
		return
	}

	verifierCookie, _ := c.Cookie(VerifierCookieName)

	h.setCookie(c, StateCookieName, "", -1)
	h.setCookie(c, VerifierCookieName, "", -1)

	if codeQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}

	var opts []oauth2.AuthCodeOption
	if verifierCookie != "" {
		opts = append(opts, oauth2.VerifierOption(verifierCookie))
	}

	token, err := h.oauthConfig.Exchange(c.Request.Context(), codeQuery, opts...)
	if err != nil && h.oauthConfig.ClientSecret != "" {
		log.Warnf("Primary OAuth exchange failed (%v), retrying without client_secret for Public Client App Registration...", err)
		fallbackConfig := *h.oauthConfig
		fallbackConfig.ClientSecret = ""
		token, err = fallbackConfig.Exchange(c.Request.Context(), codeQuery, opts...)
	}

	if err != nil {
		log.Errorf("OAuth token exchange failed: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "Failed to exchange authorization code",
			"details": err.Error(),
		})
		return
	}

	idTokenRaw, _ := token.Extra("id_token").(string)
	name := "Filmtreff-Mitglied"
	var roles []string

	if idTokenRaw != "" {
		parser := jwt.NewParser()
		parsedToken, _, err := parser.ParseUnverified(idTokenRaw, jwt.MapClaims{})
		if err == nil {
			if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
				if n, ok := claims["name"].(string); ok && n != "" {
					name = n
				} else if u, ok := claims["preferred_username"].(string); ok && u != "" {
					name = u
				}

				if rolesRaw, ok := claims["roles"].([]interface{}); ok {
					for _, r := range rolesRaw {
						if rStr, ok := r.(string); ok {
							roles = append(roles, rStr)
						}
					}
				}
			}
		}
	}

	if roles == nil {
		roles = make([]string, 0)
	}

	sessionDuration := time.Duration(h.config.SessionDurationDays) * 24 * time.Hour
	newSession := &session.Session{
		ID:          uuid.NewString(),
		Name:        name,
		Permissions: roles,
		ExpiresAt:   time.Now().Add(sessionDuration),
		CreatedAt:   time.Now(),
	}

	if err := h.sessionRepo.SaveSession(c.Request.Context(), newSession); err != nil {
		log.Errorf("Failed to save session to DB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	maxAge := h.config.SessionDurationDays * SecondsPerDay
	h.setCookie(c, SessionCookieName, newSession.ID, maxAge)

	redirectCookie, _ := c.Cookie(RedirectCookieName)
	h.setCookie(c, RedirectCookieName, "", -1)

	frontendURL := strings.TrimRight(h.config.FrontendURL, "/")
	targetPath := "/"
	if redirectCookie != "" && strings.HasPrefix(redirectCookie, "/") && !strings.HasPrefix(redirectCookie, "//") {
		targetPath = redirectCookie
	}

	c.Redirect(http.StatusFound, frontendURL+targetPath)
}

func (h *BffAuthHandler) handleMe(c *gin.Context) {
	sessionID, err := c.Cookie(SessionCookieName)
	if err != nil || sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthenticated"})
		return
	}

	sess, err := h.sessionRepo.FindSession(c.Request.Context(), sessionID)
	if err != nil || sess == nil {
		h.setCookie(c, SessionCookieName, "", -1)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired or invalid"})
		return
	}

	sessionDuration := time.Duration(h.config.SessionDurationDays) * 24 * time.Hour
	_ = h.sessionRepo.RefreshSession(c.Request.Context(), sessionID, sessionDuration)

	maxAge := h.config.SessionDurationDays * SecondsPerDay
	h.setCookie(c, SessionCookieName, sessionID, maxAge)

	permissions := sess.Permissions
	if permissions == nil {
		permissions = make([]string, 0)
	}

	c.JSON(http.StatusOK, gin.H{
		"name":        sess.Name,
		"permissions": permissions,
	})
}

func (h *BffAuthHandler) handleLogout(c *gin.Context) {
	sessionID, _ := c.Cookie(SessionCookieName)
	if sessionID != "" {
		_ = h.sessionRepo.DeleteSession(c.Request.Context(), sessionID)
	}

	h.setCookie(c, SessionCookieName, "", -1)
	c.Status(http.StatusNoContent)
}
