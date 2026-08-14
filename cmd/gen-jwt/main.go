// Command gen-jwt prints a signed HS256 JWT for local gateway testing.
//
// Usage:
//
//	export GATEWAY_JWT_HMAC_SECRET="dev-secret-at-least-32-chars-long"
//	go run ./cmd/gen-jwt
//
// Optional:
//
//	export GATEWAY_JWT_ISSUER="https://issuer.example.com"
//	export GATEWAY_JWT_AUDIENCE="my-api"
//	go run ./cmd/gen-jwt -sub user-123
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultDevSecret = "dev-secret-at-least-32-chars-long"

func main() {
	subject := flag.String("sub", "user-123", "JWT sub claim")
	ttl := flag.Duration("ttl", time.Hour, "token lifetime")
	flag.Parse()

	secret := os.Getenv("GATEWAY_JWT_HMAC_SECRET")
	if secret == "" {
		secret = defaultDevSecret
		fmt.Fprintln(os.Stderr, "warning: GATEWAY_JWT_HMAC_SECRET unset, using dev default")
	}

	claims := jwt.MapClaims{
		"sub": *subject,
		"exp": time.Now().Add(*ttl).Unix(),
	}

	if iss := os.Getenv("GATEWAY_JWT_ISSUER"); iss != "" {
		claims["iss"] = iss
	}

	if aud := os.Getenv("GATEWAY_JWT_AUDIENCE"); aud != "" {
		claims["aud"] = aud
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Fprintf(os.Stderr, "sign token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(signed)
}
