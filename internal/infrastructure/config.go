package infrastructure

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultHTTPPort        = 8080
	defaultShutdownTimeout = 10 * time.Second
)

// Config contains process-level settings used to boot the service.
type Config struct {
	HTTPPort        int
	ShutdownTimeout time.Duration
}

// LoadConfig provides a small bootstrap configuration surface for the initial service skeleton.
func LoadConfig() Config {
	return Config{
		HTTPPort:        readIntEnv("HTTP_PORT", defaultHTTPPort),
		ShutdownTimeout: defaultShutdownTimeout,
	}
}

// Address returns the bind address used by the HTTP server.
func (c Config) Address() string {
	return fmt.Sprintf(":%d", c.HTTPPort)
}

func readIntEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}
