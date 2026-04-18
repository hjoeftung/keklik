package infrastructure

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPPort        = 8080
	defaultShutdownTimeout = 10 * time.Second
	defaultInviteLinkTTL   = 7 * 24 * time.Hour
)

// Config contains process-level settings used to boot the service and inject
// runtime dependencies into application and infrastructure layers.
type Config struct {
	HTTP            HTTPConfig
	Database        DatabaseConfig
	GoogleOAuth     GoogleOAuthConfig
	Auth            AuthConfig
	App             AppConfig
	ShutdownTimeout time.Duration
}

// HTTPConfig contains transport settings.
type HTTPConfig struct {
	Port int
}

// DatabaseConfig contains database connection settings.
type DatabaseConfig struct {
	URL string
}

// GoogleOAuthConfig contains Google OAuth credentials and callback settings.
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// AuthConfig contains authentication feature flags.
type AuthConfig struct {
	EnableTestAuth bool
}

// AppConfig contains application-level URLs and other cross-cutting settings.
type AppConfig struct {
	BaseURL            string
	InviteLinkLifetime time.Duration
	EnableSwaggerUI    bool
}

// LoadConfig loads and validates the runtime environment contract.
func LoadConfig() (Config, error) {
	httpPort, err := readIntEnv("HTTP_PORT", defaultHTTPPort)
	if err != nil {
		return Config{}, err
	}

	enableTestAuth, err := readBoolEnv("ENABLE_TEST_AUTH", false)
	if err != nil {
		return Config{}, err
	}

	inviteLinkLifetime, err := readDurationEnv("FAMILY_INVITE_LINK_EXPIRY", defaultInviteLinkTTL)
	if err != nil {
		return Config{}, err
	}

	enableSwaggerUI, err := readBoolEnv("ENABLE_SWAGGER_UI", false)
	if err != nil {
		return Config{}, err
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	clientID := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"))
	redirectURL := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_REDIRECT_URL"))
	appBaseURL := strings.TrimSpace(os.Getenv("APP_BASE_URL"))

	missing := missingKeys(map[string]string{
		"DATABASE_URL":               databaseURL,
		"GOOGLE_OAUTH_CLIENT_ID":     clientID,
		"GOOGLE_OAUTH_CLIENT_SECRET": clientSecret,
		"GOOGLE_OAUTH_REDIRECT_URL":  redirectURL,
		"APP_BASE_URL":               appBaseURL,
	})
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	if err := requireAbsoluteURL("GOOGLE_OAUTH_REDIRECT_URL", redirectURL); err != nil {
		return Config{}, err
	}

	if err := requireAbsoluteURL("APP_BASE_URL", appBaseURL); err != nil {
		return Config{}, err
	}

	return Config{
		HTTP: HTTPConfig{
			Port: httpPort,
		},
		Database: DatabaseConfig{
			URL: databaseURL,
		},
		GoogleOAuth: GoogleOAuthConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
		},
		Auth: AuthConfig{
			EnableTestAuth: enableTestAuth,
		},
		App: AppConfig{
			BaseURL:            appBaseURL,
			InviteLinkLifetime: inviteLinkLifetime,
			EnableSwaggerUI:    enableSwaggerUI,
		},
		ShutdownTimeout: defaultShutdownTimeout,
	}, nil
}

// Address returns the bind address used by the HTTP server.
func (c Config) Address() string {
	return fmt.Sprintf(":%d", c.HTTP.Port)
}

func readIntEnv(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("invalid %s: must be a positive integer", key)
	}

	return parsed, nil
}

func readBoolEnv(key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid %s: must be a boolean", key)
	}

	return parsed, nil
}

func readDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("invalid %s: must be a positive duration", key)
	}

	return parsed, nil
}

func missingKeys(values map[string]string) []string {
	missing := make([]string, 0, len(values))

	for key, value := range values {
		if value == "" {
			missing = append(missing, key)
		}
	}

	sort.Strings(missing)

	return missing
}

func requireAbsoluteURL(key string, value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid %s: must be an absolute URL", key)
	}

	return nil
}
