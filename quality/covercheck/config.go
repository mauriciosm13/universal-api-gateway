package covercheck

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Config is the coverage gate configuration from quality/config/coverage.yaml.
type Config struct {
	GlobalMinimum         float64
	ChangedCodeMinimum    float64
	FailOnRegression      bool
	AllowedDropPercentage float64
}

// Baseline is the stored coverage measurement from quality/baselines/quality-baseline.json.
type Baseline struct {
	MeasuredGlobal float64
}

// ParseConfigYAML reads the small coverage.yaml subset used by the gate.
func ParseConfigYAML(r io.Reader) (Config, error) {
	var cfg Config
	seen := map[string]bool{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || line == "coverage:" {
			continue
		}
		key, raw, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		raw = strings.TrimSpace(raw)
		seen[key] = true

		switch key {
		case "global_minimum":
			v, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				return Config{}, fmt.Errorf("coverage.yaml global_minimum: %w", err)
			}
			cfg.GlobalMinimum = v
		case "changed_code_minimum":
			v, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				return Config{}, fmt.Errorf("coverage.yaml changed_code_minimum: %w", err)
			}
			cfg.ChangedCodeMinimum = v
		case "fail_on_regression":
			v, err := parseBool(raw)
			if err != nil {
				return Config{}, fmt.Errorf("coverage.yaml fail_on_regression: %w", err)
			}
			cfg.FailOnRegression = v
		case "allowed_drop_percentage":
			v, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				return Config{}, fmt.Errorf("coverage.yaml allowed_drop_percentage: %w", err)
			}
			cfg.AllowedDropPercentage = v
		}
	}
	if err := scanner.Err(); err != nil {
		return Config{}, err
	}

	for _, key := range []string{"global_minimum", "changed_code_minimum", "fail_on_regression", "allowed_drop_percentage"} {
		if !seen[key] {
			return Config{}, fmt.Errorf("coverage.yaml missing %s", key)
		}
	}
	return cfg, nil
}

func parseBool(raw string) (bool, error) {
	switch strings.ToLower(raw) {
	case "true", "yes", "1":
		return true, nil
	case "false", "no", "0":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool %q", raw)
	}
}

type baselineFile struct {
	Coverage struct {
		MeasuredGlobal float64 `json:"measured_global"`
	} `json:"coverage"`
}

// ParseBaselineJSON reads measured global coverage from the quality baseline file.
func ParseBaselineJSON(r io.Reader) (Baseline, error) {
	var file baselineFile
	if err := json.NewDecoder(r).Decode(&file); err != nil {
		return Baseline{}, fmt.Errorf("baseline json: %w", err)
	}
	return Baseline{MeasuredGlobal: file.Coverage.MeasuredGlobal}, nil
}
