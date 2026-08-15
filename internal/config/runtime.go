package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	RuntimeServer = "server"
	RuntimeLambda = "lambda"

	defaultLambdaTimeout = 30 * time.Second
)

// RuntimeConfig describes how the gateway process is hosted.
type RuntimeConfig struct {
	Mode          string
	LambdaTimeout time.Duration
}

// IsLambda reports whether the gateway runs under AWS Lambda.
func (c RuntimeConfig) IsLambda() bool {
	return c.Mode == RuntimeLambda
}

func loadRuntime() (RuntimeConfig, error) {
	mode := strings.ToLower(strings.TrimSpace(envString("GATEWAY_RUNTIME", RuntimeServer)))
	switch mode {
	case RuntimeServer, RuntimeLambda:
	default:
		return RuntimeConfig{}, fmt.Errorf("invalid GATEWAY_RUNTIME %q: want server or lambda", mode)
	}

	cfg := RuntimeConfig{Mode: mode}
	if mode != RuntimeLambda {
		return cfg, nil
	}

	timeoutRaw := strings.TrimSpace(os.Getenv("AWS_LAMBDA_FUNCTION_TIMEOUT"))
	if timeoutRaw == "" {
		cfg.LambdaTimeout = defaultLambdaTimeout
		return cfg, nil
	}

	seconds, err := strconv.Atoi(timeoutRaw)
	if err != nil || seconds <= 0 {
		return RuntimeConfig{}, fmt.Errorf("invalid AWS_LAMBDA_FUNCTION_TIMEOUT: must be a positive integer (seconds)")
	}

	cfg.LambdaTimeout = time.Duration(seconds) * time.Second
	return cfg, nil
}
