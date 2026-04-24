package infrastructure

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPPort             = 8080
	defaultShutdownTimeout      = 10 * time.Second
	defaultInviteLinkTTL        = 7 * 24 * time.Hour
	defaultAccessTokenDuration  = 15 * time.Minute
	defaultRefreshTokenDuration = 30 * 24 * time.Hour
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

// AuthConfig contains authentication feature flags and signing material.
type AuthConfig struct {
	EnableTestAuth       bool
	JWTSigningKey        string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

// AppConfig contains application-level URLs and other cross-cutting settings.
type AppConfig struct {
	BaseURL            string
	FrontendURL        string
	InviteLinkLifetime time.Duration
	EnableSwaggerUI    bool
	IsDev              bool
}

// LoadConfig loads and validates the runtime environment contract.
func LoadConfig(getenv func(string) string) (Config, error) {
	httpPort, err := readIntEnv(getenv, "HTTP_PORT", defaultHTTPPort)
	if err != nil {
		return Config{}, err
	}

	enableTestAuth, err := readBoolEnv(getenv, "ENABLE_TEST_AUTH", false)
	if err != nil {
		return Config{}, err
	}

	inviteLinkLifetime, err := readDurationEnv(getenv, "FAMILY_INVITE_LINK_EXPIRY", defaultInviteLinkTTL)
	if err != nil {
		return Config{}, err
	}

	accessTokenDuration, err := readDurationEnv(getenv, "ACCESS_TOKEN_DURATION", defaultAccessTokenDuration)
	if err != nil {
		return Config{}, err
	}

	refreshTokenDuration, err := readDurationEnv(getenv, "REFRESH_TOKEN_DURATION", defaultRefreshTokenDuration)
	if err != nil {
		return Config{}, err
	}

	enableSwaggerUI, err := readBoolEnv(getenv, "ENABLE_SWAGGER_UI", false)
	if err != nil {
		return Config{}, err
	}

	isDev := strings.TrimSpace(getenv("ENVIRONMENT")) != "production"

	databaseURL := strings.TrimSpace(getenv("DATABASE_URL"))
	clientID := strings.TrimSpace(getenv("GOOGLE_OAUTH_CLIENT_ID"))
	clientSecret := strings.TrimSpace(getenv("GOOGLE_OAUTH_CLIENT_SECRET"))
	redirectURL := strings.TrimSpace(getenv("GOOGLE_OAUTH_REDIRECT_URL"))
	appBaseURL := strings.TrimSpace(getenv("APP_BASE_URL"))
	frontendURL := strings.TrimSpace(getenv("FRONTEND_URL"))
	jwtSigningKey := strings.TrimSpace(getenv("JWT_SIGNING_KEY"))

	missing := missingKeys(map[string]string{
		"DATABASE_URL":               databaseURL,
		"FRONTEND_URL":               frontendURL,
		"GOOGLE_OAUTH_CLIENT_ID":     clientID,
		"GOOGLE_OAUTH_CLIENT_SECRET": clientSecret,
		"GOOGLE_OAUTH_REDIRECT_URL":  redirectURL,
		"APP_BASE_URL":               appBaseURL,
		"JWT_SIGNING_KEY":            jwtSigningKey,
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

	if err := requireAbsoluteURL("FRONTEND_URL", frontendURL); err != nil {
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
			EnableTestAuth:       enableTestAuth,
			JWTSigningKey:        jwtSigningKey,
			AccessTokenDuration:  accessTokenDuration,
			RefreshTokenDuration: refreshTokenDuration,
		},
		App: AppConfig{
			BaseURL:            appBaseURL,
			FrontendURL:        frontendURL,
			InviteLinkLifetime: inviteLinkLifetime,
			EnableSwaggerUI:    enableSwaggerUI,
			IsDev:              isDev,
		},
		ShutdownTimeout: defaultShutdownTimeout,
	}, nil
}

// Address returns the bind address used by the HTTP server.
func (c Config) Address() string {
	return fmt.Sprintf(":%d", c.HTTP.Port)
}

func readIntEnv(getenv func(string) string, key string, fallback int) (int, error) {
	value := getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("invalid %s: must be a positive integer", key)
	}

	return parsed, nil
}

func readBoolEnv(getenv func(string) string, key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid %s: must be a boolean", key)
	}

	return parsed, nil
}

func readDurationEnv(getenv func(string) string, key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(getenv(key))
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
