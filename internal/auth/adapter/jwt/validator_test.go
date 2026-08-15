package jwt

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
)

const testHMACSecret = "dev-secret-at-least-32-chars-long"

func TestValidatorAcceptsValidHS256Token(t *testing.T) {
	t.Parallel()

	validator := mustHMACValidator(t, config.JWTConfig{HMACSecret: testHMACSecret})

	token := signHS256Token(t, testHMACSecret, jwt.MapClaims{
		"sub":  "user-123",
		"role": "admin",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})

	identity, err := validator.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if identity.Subject != "user-123" {
		t.Fatalf("subject = %q, want user-123", identity.Subject)
	}

	if identity.Claims["role"] != "admin" {
		t.Fatalf("claims = %+v", identity.Claims)
	}
}

func TestValidatorRejectsExpiredToken(t *testing.T) {
	t.Parallel()

	validator := mustHMACValidator(t, config.JWTConfig{HMACSecret: testHMACSecret})

	token := signHS256Token(t, testHMACSecret, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})

	if _, err := validator.Authenticate(context.Background(), token); err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestValidatorRejectsInvalidSignature(t *testing.T) {
	t.Parallel()

	validator := mustHMACValidator(t, config.JWTConfig{HMACSecret: testHMACSecret})

	token := signHS256Token(t, "other-secret-at-least-32-characters", jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	if _, err := validator.Authenticate(context.Background(), token); err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestValidatorRejectsWrongIssuer(t *testing.T) {
	t.Parallel()

	validator := mustHMACValidator(t, config.JWTConfig{
		HMACSecret: testHMACSecret,
		Issuer:     "https://issuer.example.com",
	})

	token := signHS256Token(t, testHMACSecret, jwt.MapClaims{
		"sub": "user-123",
		"iss": "https://wrong.example.com",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	if _, err := validator.Authenticate(context.Background(), token); err == nil {
		t.Fatal("expected error for wrong issuer")
	}
}

func TestValidatorRejectsWrongAudience(t *testing.T) {
	t.Parallel()

	validator := mustHMACValidator(t, config.JWTConfig{
		HMACSecret: testHMACSecret,
		Audience:   "my-api",
	})

	token := signHS256Token(t, testHMACSecret, jwt.MapClaims{
		"sub": "user-123",
		"aud": "other-api",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	if _, err := validator.Authenticate(context.Background(), token); err == nil {
		t.Fatal("expected error for wrong audience")
	}
}

func TestValidatorRejectsEmptyToken(t *testing.T) {
	t.Parallel()

	validator := mustHMACValidator(t, config.JWTConfig{HMACSecret: testHMACSecret})

	for _, token := range []string{"", "   "} {
		if _, err := validator.Authenticate(context.Background(), token); err == nil {
			t.Fatalf("expected error for token %q", token)
		}
	}
}

func TestValidatorAcceptsValidRS256TokenFromJWKS(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	jwks := rsaJWKS(t, privateKey, "test-kid")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	t.Cleanup(server.Close)

	validator, err := NewValidator(config.JWTConfig{JWKSURL: server.URL})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	token := signRS256Token(t, privateKey, "test-kid", jwt.MapClaims{
		"sub": "jwks-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	identity, err := validator.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if identity.Subject != "jwks-user" {
		t.Fatalf("subject = %q, want jwks-user", identity.Subject)
	}
}

func TestNewValidatorRejectsDisabledConfig(t *testing.T) {
	t.Parallel()

	if _, err := NewValidator(config.JWTConfig{}); err == nil {
		t.Fatal("expected error for disabled config")
	}
}

func TestFetchJWKSRejectsEmptyKeySet(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"keys":[]}`))
	}))
	t.Cleanup(server.Close)

	_, err := fetchJWKS(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected error for empty JWKS")
	}
}

func TestMapClaimsToIdentity(t *testing.T) {
	t.Parallel()

	identity := mapClaimsToIdentity(jwt.MapClaims{
		"sub":   "user-1",
		"role":  "admin",
		"count": float64(3),
	})

	if identity.Subject != "user-1" || identity.Claims["role"] != "admin" || identity.Claims["count"] != "3" {
		t.Fatalf("identity = %+v", identity)
	}
}

func TestValidatorAcceptsES256TokenFromJWKS(t *testing.T) {
	t.Parallel()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	jwks := ecJWKS(t, privateKey, "ec-kid")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	t.Cleanup(server.Close)

	validator, err := NewValidator(config.JWTConfig{JWKSURL: server.URL})
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	token := signES256Token(t, privateKey, "ec-kid", jwt.MapClaims{
		"sub": "ec-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	identity, err := validator.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if identity.Subject != "ec-user" {
		t.Fatalf("subject = %q, want ec-user", identity.Subject)
	}
}

func TestFetchJWKSRejectsNonOKStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	_, err := fetchJWKS(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected error for non-200 JWKS response")
	}
}

func TestMapClaimsToIdentityCoercesTypes(t *testing.T) {
	t.Parallel()

	identity := mapClaimsToIdentity(jwt.MapClaims{
		"sub":   "user-1",
		"admin": true,
		"num":   json.Number("42"),
	})

	if identity.Claims["admin"] != "true" || identity.Claims["num"] != "42" {
		t.Fatalf("claims = %+v", identity.Claims)
	}
}

func mustHMACValidator(t *testing.T, cfg config.JWTConfig) *Validator {
	t.Helper()

	validator, err := NewValidator(cfg)
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	return validator
}

func signHS256Token(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	return signed
}

func signRS256Token(t *testing.T, key *rsa.PrivateKey, kid string, claims jwt.MapClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid

	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	return signed
}

func signES256Token(t *testing.T, key *ecdsa.PrivateKey, kid string, claims jwt.MapClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = kid

	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	return signed
}

func rsaJWKS(t *testing.T, key *rsa.PrivateKey, kid string) jwksDocument {
	t.Helper()

	return jwksDocument{
		Keys: []jwkKey{{
			Kty: "RSA",
			Kid: kid,
			Use: "sig",
			Alg: "RS256",
			N:   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			E:   base64.RawURLEncoding.EncodeToString(bigIntToBytes(key.PublicKey.E)),
		}},
	}
}

func ecJWKS(t *testing.T, key *ecdsa.PrivateKey, kid string) jwksDocument {
	t.Helper()

	return jwksDocument{
		Keys: []jwkKey{{
			Kty: "EC",
			Kid: kid,
			Use: "sig",
			Alg: "ES256",
			Crv: "P-256",
			X:   base64.RawURLEncoding.EncodeToString(key.PublicKey.X.Bytes()),
			Y:   base64.RawURLEncoding.EncodeToString(key.PublicKey.Y.Bytes()),
		}},
	}
}

func bigIntToBytes(value int) []byte {
	if value == 0 {
		return []byte{0}
	}

	var out []byte
	for value > 0 {
		out = append([]byte{byte(value & 0xff)}, out...)
		value >>= 8
	}

	return out
}
