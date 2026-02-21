// Package config resolves the Google PageSpeed Insights API key from multiple sources.
// Priority order: CLI flag > GOOGLE_PSI_API_KEY environment variable > .env file.
package config

import (
	"bufio"
	"log/slog"
	"os"
	"strings"
)

const (
	envVarName  = "GOOGLE_PSI_API_KEY"
	dotEnvFile  = ".env"
	dotEnvKeyEq = envVarName + "="
)

// Config holds resolved configuration values.
type Config struct {
	// APIKey is the Google PageSpeed Insights API key.
	APIKey string
}

// Resolve returns a Config with the API key loaded from the highest-priority source.
// Sources in descending priority: flag argument, environment variable, .env file.
func Resolve(flagValue string) Config {
	if flagValue != "" {
		slog.Debug("api key loaded from CLI flag")
		return Config{APIKey: flagValue}
	}

	if v := os.Getenv(envVarName); v != "" {
		slog.Debug("api key loaded from environment variable")
		return Config{APIKey: v}
	}

	if v := loadFromDotEnv(); v != "" {
		slog.Debug("api key loaded from .env file")
		return Config{APIKey: v}
	}

	return Config{}
}

// loadFromDotEnv reads GOOGLE_PSI_API_KEY from a .env file in the current directory.
// Returns an empty string if the file does not exist or the key is not found.
func loadFromDotEnv() string {
	f, err := os.Open(dotEnvFile)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if after, ok := strings.CutPrefix(line, dotEnvKeyEq); ok {
			return strings.Trim(after, `"'`)
		}
	}
	return ""
}
