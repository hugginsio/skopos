// Copyright (c) Kyle Huggins and contributors
// SPDX-License-Identifier: BSD-3-Clause

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/adrg/xdg"
)

// Config holds all configuration for the dnsmon service.
type Config struct {
	Tailscale   TailscaleConfig   // Tailscale configuration
	HealthCheck HealthCheckConfig // Health check configuration
	DNS         DNSConfig         // DNS update configuration
	Development bool              // Development mode
}

// TailscaleConfig holds Tailscale-specific configuration.
type TailscaleConfig struct {
	AuthKey  string // AuthKey is the Tailscale authentication key.
	Hostname string // Hostname is the hostname to use for this tsnet node.
	StateDir string // StateDir is the directory where tsnet state is stored.
}

// HealthCheckConfig holds health check configuration.
type HealthCheckConfig struct {
	Interval           time.Duration // Interval is how often to check device health.
	Timeout            time.Duration // Timeout is how long to wait for a health check to complete.
	UnhealthyThreshold int           // UnhealthyThreshold is how many consecutive failures before marking unhealthy.
}

// DNSConfig holds DNS update configuration.
type DNSConfig struct {
	UpdateTimeout time.Duration // UpdateTimeout is how long to wait for a DNS update to complete.
}

// Load loads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Development: os.Getenv("DEVEL") == "true",
	}

	// Load Tailscale configuration
	cfg.Tailscale.AuthKey = os.Getenv("TS_AUTHKEY")
	cfg.Tailscale.Hostname = getEnvOrDefault("TS_HOSTNAME", "skopos-dnsmon")

	stateDir, err := determineStateDir(cfg.Development)
	if err != nil {
		return nil, fmt.Errorf("failed to determine state directory: %w", err)
	}

	cfg.Tailscale.StateDir = getEnvOrDefault("TS_STATE_DIR", stateDir)

	// Load health check configuration
	cfg.HealthCheck.Interval = getDurationOrDefault("HEALTH_CHECK_INTERVAL", 20*time.Second)
	cfg.HealthCheck.Timeout = getDurationOrDefault("HEALTH_CHECK_TIMEOUT", 5*time.Second)
	cfg.HealthCheck.UnhealthyThreshold = getIntOrDefault("UNHEALTHY_THRESHOLD", 1)

	// Load DNS configuration
	cfg.DNS.UpdateTimeout = getDurationOrDefault("DNS_UPDATE_TIMEOUT", 10*time.Second)

	// Validate required fields
	if cfg.Tailscale.AuthKey == "" {
		return nil, fmt.Errorf("TS_AUTHKEY is required")
	}

	return cfg, nil
}

// determineStateDir determines the appropriate state directory based on the environment.
func determineStateDir(development bool) (string, error) {
	if development {
		return "./tsnet.tmp/", nil
	}

	// Check if running in a container
	if isRunningInContainer() {
		return "/data", nil
	}

	// Use XDG data directory
	return filepath.Join(xdg.DataHome, "skopos", "tsnet"), nil
}

// isRunningInContainer checks if we're running inside a container.
func isRunningInContainer() bool {
	// Check for Docker
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check for generic container environment variable
	if os.Getenv("container") != "" {
		return true
	}

	return false
}

// getEnvOrDefault returns the value of an environment variable or a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

// getDurationOrDefault parses a duration from an environment variable or returns a default.
func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}

	return defaultValue
}

// getIntOrDefault parses an integer from an environment variable or returns a default.
func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}

	return defaultValue
}
