package inbound

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwk"
	log "github.com/sirupsen/logrus"
)

var jwkUrl = "https://login.microsoftonline.com/865638a4-e4fb-4aef-89e1-6824acc3a785/discovery/v2.0/keys"
var jwkSet jwk.Set

func TraceIdMiddleware(ginCtx *gin.Context) {
	uuid := generateTraceId()
	newCtx := context.WithValue(ginCtx.Request.Context(), filmkritiken.Context_TraceId, uuid)
	ginCtx.Request = ginCtx.Request.WithContext(newCtx)
}

func generateTraceId() string {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return ""
	}
	return uuid.String()
}

func NewAuthHandler(allowedRoles []string) func(ginCtx *gin.Context) {
	return func(ginCtx *gin.Context) {
		authHandler(ginCtx, allowedRoles)
	}
}

func authHandler(ginCtx *gin.Context, allowedRoles []string) {
	authHeader := ginCtx.Request.Header["Authorization"]

	if len(authHeader) == 0 {
		log.Warn("received request without auth header")
		ginCtx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// "Bearer ..."
	fields := strings.Fields(authHeader[0])
	if len(fields) <= 1 {
		log.Warn("received request with malformed auth header")
		ginCtx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	tokenString := fields[1]

	token, err := jwt.Parse(tokenString, getKey)
	if err != nil {
		log.Errorf("could not parse jwt token: %v", err)
		ginCtx.AbortWithStatus(http.StatusForbidden)
		return
	}

	if !token.Valid {
		log.Warnf("client tried accessing endpoint with invalid token: %v", err)
		ginCtx.AbortWithStatus(http.StatusForbidden)
		return
	}

	if !hasRole(allowedRoles, token) {
		log.Warnf("client tried accessing endpoint without roles (%s): %v", allowedRoles, err)
		ginCtx.AbortWithStatus(http.StatusForbidden)
		return
	}

	err = setUsername(ginCtx, token)
	if err != nil {
		log.Errorf("could not extract username: %v", err)
		ginCtx.AbortWithStatus(http.StatusForbidden)
		return
	}

}

func getKey(token *jwt.Token) (interface{}, error) {
	// Get JWK Set from JWKS endpoint
	if jwkSet == nil {
		set, err := jwk.Fetch(context.Background(), jwkUrl)
		if err != nil {
			return nil, err
		}
		jwkSet = set
	}

	keyId, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("JWT Header did not include kid")
	}

	if key, ok := jwkSet.LookupKeyID(keyId); ok {
		var rawKey interface{}
		err := key.Raw(&rawKey)
		if err != nil {
			return nil, fmt.Errorf("error getting raw key id from jwkSet: %w", err)
		}

		return rawKey, nil
	}

	return nil, fmt.Errorf("unable to find key %s", keyId)
}

func hasRole(allowedRoles []string, token *jwt.Token) bool {
	claims := token.Claims.(jwt.MapClaims)
	tokenRoles := claims["roles"].([]interface{})

	for _, tokenRole := range tokenRoles {
		for _, allowedRole := range allowedRoles {
			if allowedRole == tokenRole {
				return true
			}
		}
	}

	return false
}

func setUsername(ginCtx *gin.Context, token *jwt.Token) error {
	claims := token.Claims.(jwt.MapClaims)
	name, ok := claims["name"].(string)
	if !ok {
		return fmt.Errorf("could not extract name from claims: %v", claims)
	}

	newCtx := context.WithValue(ginCtx.Request.Context(), filmkritiken.Context_Username, name)
	ginCtx.Request = ginCtx.Request.WithContext(newCtx)

	return nil
}
