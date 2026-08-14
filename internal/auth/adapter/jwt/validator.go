package jwt

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/mauriciomendonca/universal-api-gateway/internal/config"
)

const jwksFetchTimeout = 10 * time.Second

var errInvalidToken = errors.New("invalid token")

// Validator validates JWT bearer tokens using HMAC or JWKS-backed keys.
type Validator struct {
	hmacSecret []byte
	keySet     *keySet
	issuer     string
	audience   string
	parser     jwt.Parser
}

// NewValidator builds a JWT authenticator from configuration.
func NewValidator(cfg config.JWTConfig) (*Validator, error) {
	if !cfg.Enabled() {
		return nil, fmt.Errorf("jwt: configuration is disabled")
	}

	v := &Validator{
		issuer:   cfg.Issuer,
		audience: cfg.Audience,
	}

	parserOpts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg(), jwt.SigningMethodRS256.Alg(), jwt.SigningMethodES256.Alg()}),
	}
	if cfg.Issuer != "" {
		parserOpts = append(parserOpts, jwt.WithIssuer(cfg.Issuer))
	}
	if cfg.Audience != "" {
		parserOpts = append(parserOpts, jwt.WithAudience(cfg.Audience))
	}
	v.parser = *jwt.NewParser(parserOpts...)

	switch {
	case cfg.HMACSecret != "":
		v.hmacSecret = []byte(cfg.HMACSecret)
	case cfg.JWKSURL != "":
		ctx, cancel := context.WithTimeout(context.Background(), jwksFetchTimeout)
		defer cancel()

		keys, err := fetchJWKS(ctx, cfg.JWKSURL)
		if err != nil {
			return nil, fmt.Errorf("jwt: fetch JWKS: %w", err)
		}
		v.keySet = keys
	default:
		return nil, fmt.Errorf("jwt: JWKS URL or HMAC secret is required")
	}

	return v, nil
}

// Authenticate validates the token and returns caller identity.
func (v *Validator) Authenticate(_ context.Context, token string) (Identity, error) {
	if strings.TrimSpace(token) == "" {
		return Identity{}, errInvalidToken
	}

	parsed, err := v.parser.Parse(token, v.keyFunc)
	if err != nil {
		return Identity{}, errInvalidToken
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return Identity{}, errInvalidToken
	}

	return mapClaimsToIdentity(claims), nil
}

func (v *Validator) keyFunc(token *jwt.Token) (any, error) {
	switch token.Method.Alg() {
	case jwt.SigningMethodHS256.Alg():
		if len(v.hmacSecret) == 0 {
			return nil, errInvalidToken
		}
		return v.hmacSecret, nil
	case jwt.SigningMethodRS256.Alg(), jwt.SigningMethodES256.Alg():
		if v.keySet == nil {
			return nil, errInvalidToken
		}
		kid, _ := token.Header["kid"].(string)
		key, ok := v.keySet.lookup(kid, token.Method.Alg())
		if !ok {
			return nil, errInvalidToken
		}
		return key, nil
	default:
		return nil, errInvalidToken
	}
}

type jwksDocument struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type keySet struct {
	byKid map[string]crypto.PublicKey
	all   []keyEntry
}

type keyEntry struct {
	alg string
	key crypto.PublicKey
}

func fetchJWKS(ctx context.Context, jwksURL string) (*keySet, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var doc jwksDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}

	set := &keySet{
		byKid: make(map[string]crypto.PublicKey, len(doc.Keys)),
		all:   make([]keyEntry, 0, len(doc.Keys)),
	}

	for _, key := range doc.Keys {
		if key.Use != "" && key.Use != "sig" {
			continue
		}

		pub, alg, err := publicKeyFromJWK(key)
		if err != nil {
			continue
		}

		set.all = append(set.all, keyEntry{alg: alg, key: pub})
		if key.Kid != "" {
			set.byKid[key.Kid] = pub
		}
	}

	if len(set.all) == 0 {
		return nil, errors.New("no usable signing keys in JWKS")
	}

	return set, nil
}

func (s *keySet) lookup(kid, alg string) (crypto.PublicKey, bool) {
	if kid != "" {
		if key, ok := s.byKid[kid]; ok {
			return key, true
		}
	}

	for _, entry := range s.all {
		if entry.alg == alg {
			return entry.key, true
		}
	}

	return nil, false
}

func publicKeyFromJWK(key jwkKey) (crypto.PublicKey, string, error) {
	switch key.Kty {
	case "RSA":
		pub, err := rsaPublicKeyFromJWK(key.N, key.E)
		if err != nil {
			return nil, "", err
		}
		return pub, jwt.SigningMethodRS256.Alg(), nil
	case "EC":
		pub, err := ecPublicKeyFromJWK(key.Crv, key.X, key.Y)
		if err != nil {
			return nil, "", err
		}
		return pub, jwt.SigningMethodES256.Alg(), nil
	default:
		return nil, "", fmt.Errorf("unsupported key type %q", key.Kty)
	}
}

func rsaPublicKeyFromJWK(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, err
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, err
	}

	eInt := 0
	for _, b := range eBytes {
		eInt = eInt<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eInt,
	}, nil
}

func ecPublicKeyFromJWK(crv, xStr, yStr string) (*ecdsa.PublicKey, error) {
	var curve elliptic.Curve
	switch crv {
	case "P-256":
		curve = elliptic.P256()
	default:
		return nil, fmt.Errorf("unsupported curve %q", crv)
	}

	xBytes, err := base64.RawURLEncoding.DecodeString(xStr)
	if err != nil {
		return nil, err
	}

	yBytes, err := base64.RawURLEncoding.DecodeString(yStr)
	if err != nil {
		return nil, err
	}

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}

func mapClaimsToIdentity(claims jwt.MapClaims) Identity {
	subject, _ := claims["sub"].(string)

	stringClaims := make(map[string]string, len(claims))
	for key, value := range claims {
		if key == "sub" {
			continue
		}
		if s := claimString(value); s != "" {
			stringClaims[key] = s
		}
	}

	return Identity{
		Subject: subject,
		Claims:  stringClaims,
	}
}

func claimString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case json.Number:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
