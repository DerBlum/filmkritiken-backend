package inbound

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DerBlum/filmkritiken-backend/domain/filmkritiken"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v3/jwk"
)

const testKid = "test-key-id"

// newTestJwkSet generates an RSA key pair, wraps the public key in a JWK set
// (with kid=testKid), injects it into the package-level jwkSet, and returns
// the private key for signing test tokens.
func newTestJwkSet(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	jwkKey, err := jwk.Import(privKey.Public())
	if err != nil {
		t.Fatalf("failed to import public key as JWK: %v", err)
	}
	if err := jwkKey.Set(jwk.KeyIDKey, testKid); err != nil {
		t.Fatalf("failed to set kid on JWK: %v", err)
	}

	set := jwk.NewSet()
	if err := set.AddKey(jwkKey); err != nil {
		t.Fatalf("failed to add key to JWK set: %v", err)
	}

	// Inject into package-level variable so getKey() skips the network fetch.
	jwkSet = set

	return privKey
}

// signToken creates a signed JWT using the provided private key and claims.
func signToken(t *testing.T, privKey *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = testKid

	signed, err := token.SignedString(privKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}

// validClaims returns a minimal set of claims that will pass all validation
// and satisfy the role/name requirements of authHandler.
func validClaims(roles []string) jwt.MapClaims {
	rolesIface := make([]interface{}, len(roles))
	for i, r := range roles {
		rolesIface[i] = r
	}
	return jwt.MapClaims{
		"roles": rolesIface,
		"name":  "Test User",
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
}

func TestAuthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	privKey := newTestJwkSet(t)

	validToken := signToken(t, privKey, validClaims([]string{"film.add"}))
	wrongRoleToken := signToken(t, privKey, validClaims([]string{"other.role"}))
	expiredToken := signToken(t, privKey, jwt.MapClaims{
		"roles": []interface{}{"film.add"},
		"name":  "Test User",
		"exp":   time.Now().Add(-time.Hour).Unix(),
	})

	allowedRoles := []string{"film.add"}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "no Authorization header returns 401",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "raw token without Bearer prefix returns 401",
			authHeader:     validToken,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "wrong scheme (Basic) returns 401",
			authHeader:     "Basic somebase64value",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired token returns 403",
			authHeader:     fmt.Sprintf("Bearer %s", expiredToken),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "valid token without required role returns 403",
			authHeader:     fmt.Sprintf("Bearer %s", wrongRoleToken),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "garbage token string returns 403",
			authHeader:     "Bearer thisisnotavalidjwt",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "valid token with correct role returns 200",
			authHeader:     fmt.Sprintf("Bearer %s", validToken),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			r.GET("/test", NewAuthHandler(allowedRoles), func(ctx *gin.Context) {
				ctx.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			r.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_SetsUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)
	privKey := newTestJwkSet(t)

	const expectedName = "Max Mustermann"
	token := signToken(t, privKey, jwt.MapClaims{
		"roles": []interface{}{"film.add"},
		"name":  expectedName,
		"exp":   time.Now().Add(time.Hour).Unix(),
	})

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	var capturedName string
	r.GET("/test", NewAuthHandler([]string{"film.add"}), func(ctx *gin.Context) {
		capturedName, _ = ctx.Request.Context().Value(filmkritiken.Context_Username).(string)
		ctx.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if capturedName != expectedName {
		t.Errorf("expected username %q in context, got %q", expectedName, capturedName)
	}
}
