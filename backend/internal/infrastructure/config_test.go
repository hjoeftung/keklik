package infrastructure

import (
	"testing"
	"time"
)

func TestLoadConfigUsesDefaultHTTPPortAndLoadsRequiredEnvironment(t *testing.T) {
	setRequiredEnv(t)

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if config.HTTP.Port != defaultHTTPPort {
		t.Fatalf("expected default HTTP port %d, got %d", defaultHTTPPort, config.HTTP.Port)
	}

	if config.Database.URL != "postgres://keklik:keklik@localhost:5432/keklik?sslmode=disable" {
		t.Fatalf("unexpected database URL: %q", config.Database.URL)
	}

	if config.GoogleOAuth.ClientID != "client-id" {
		t.Fatalf("unexpected google oauth client id: %q", config.GoogleOAuth.ClientID)
	}

	if config.GoogleOAuth.ClientSecret != "client-secret" {
		t.Fatalf("unexpected google oauth client secret: %q", config.GoogleOAuth.ClientSecret)
	}

	if config.GoogleOAuth.RedirectURL != "http://localhost:8080/auth/google/callback" {
		t.Fatalf("unexpected google oauth redirect url: %q", config.GoogleOAuth.RedirectURL)
	}

	if config.App.BaseURL != "http://localhost:8080" {
		t.Fatalf("unexpected app base url: %q", config.App.BaseURL)
	}

	if config.App.InviteLinkLifetime != defaultInviteLinkTTL {
		t.Fatalf("unexpected invite link lifetime: %s", config.App.InviteLinkLifetime)
	}

	if config.Auth.EnableTestAuth {
		t.Fatal("expected test auth to be disabled by default")
	}

	if config.Address() != ":8080" {
		t.Fatalf("unexpected server address: %q", config.Address())
	}
}

func TestLoadConfigEnablesTestAuthWhenConfigured(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("ENABLE_TEST_AUTH", "true")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if !config.Auth.EnableTestAuth {
		t.Fatal("expected test auth to be enabled")
	}
}

func TestLoadConfigUsesCustomInviteLinkLifetime(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("FAMILY_INVITE_LINK_EXPIRY", "48h")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if config.App.InviteLinkLifetime != 48*time.Hour {
		t.Fatalf("expected 48h invite link lifetime, got %s", config.App.InviteLinkLifetime)
	}
}

func TestLoadConfigFailsOnMissingRequiredEnvironment(t *testing.T) {
	t.Setenv("HTTP_PORT", "8081")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for missing required environment variables")
	}

	got := err.Error()
	expected := "missing required environment variables: APP_BASE_URL, DATABASE_URL, GOOGLE_OAUTH_CLIENT_ID, GOOGLE_OAUTH_CLIENT_SECRET, GOOGLE_OAUTH_REDIRECT_URL"
	if got != expected {
		t.Fatalf("expected error %q, got %q", expected, got)
	}
}

func TestLoadConfigFailsOnInvalidHTTPPort(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("HTTP_PORT", "invalid")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for invalid HTTP_PORT")
	}

	if got := err.Error(); got != "invalid HTTP_PORT: must be a positive integer" {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestLoadConfigFailsOnInvalidURLs(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("APP_BASE_URL", "localhost:8080")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for invalid APP_BASE_URL")
	}

	if got := err.Error(); got != "invalid APP_BASE_URL: must be an absolute URL" {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestLoadConfigFailsOnInvalidEnableTestAuth(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("ENABLE_TEST_AUTH", "not-a-bool")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for invalid ENABLE_TEST_AUTH")
	}

	if got := err.Error(); got != "invalid ENABLE_TEST_AUTH: must be a boolean" {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestLoadConfigFailsOnInvalidInviteLinkLifetime(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("FAMILY_INVITE_LINK_EXPIRY", "later")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for invalid FAMILY_INVITE_LINK_EXPIRY")
	}

	if got := err.Error(); got != "invalid FAMILY_INVITE_LINK_EXPIRY: must be a positive duration" {
		t.Fatalf("unexpected error: %q", got)
	}
}

func setRequiredEnv(t *testing.T) {
	t.Helper()

	t.Setenv("DATABASE_URL", "postgres://keklik:keklik@localhost:5432/keklik?sslmode=disable")
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "client-id")
	t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "client-secret")
	t.Setenv("GOOGLE_OAUTH_REDIRECT_URL", "http://localhost:8080/auth/google/callback")
	t.Setenv("APP_BASE_URL", "http://localhost:8080")
}
