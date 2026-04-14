package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/infrastructure"
	"github.com/hjoeftung/keklik/internal/sleep"
)

// stubSleepContextResolver is a sleepContextResolver test double.
type stubSleepContextResolver struct {
	babyID   sleep.BabyID
	memberID sleep.FamilyMemberID
	err      error
}

func (r *stubSleepContextResolver) ResolveSleepContext(_ context.Context, _ string) (sleep.BabyID, sleep.FamilyMemberID, error) {
	return r.babyID, r.memberID, r.err
}

// stubStopSleepSessionRepo implements stopSleepSessionRepository for tests.
type stubStopSleepSessionRepo struct {
	active  sleep.SleepSession
	findErr error
	saveErr error
}

func (r *stubStopSleepSessionRepo) Save(_ context.Context, _ sleep.SleepSession) error {
	return r.saveErr
}

func (r *stubStopSleepSessionRepo) FindByID(_ context.Context, _ sleep.SleepSessionID) (sleep.SleepSession, error) {
	return sleep.SleepSession{}, nil
}

func (r *stubStopSleepSessionRepo) FindActiveByBabyID(_ context.Context, _ sleep.BabyID) (sleep.SleepSession, error) {
	return r.active, r.findErr
}

// stubStopSleepProfileRepo implements SleepProfileRepository for tests.
type stubStopSleepProfileRepo struct {
	profile sleep.SleepProfile
	err     error
}

func (r *stubStopSleepProfileRepo) Save(_ context.Context, _ sleep.SleepProfile) error {
	return nil
}

func (r *stubStopSleepProfileRepo) FindByBabyID(_ context.Context, _ sleep.BabyID) (sleep.SleepProfile, error) {
	return r.profile, r.err
}

func mustSleepProfile(t *testing.T) sleep.SleepProfile {
	t.Helper()

	nightStart, err := sleep.NewLocalTime(21, 0)
	if err != nil {
		t.Fatalf("NewLocalTime: %v", err)
	}

	nightEnd, err := sleep.NewLocalTime(7, 0)
	if err != nil {
		t.Fatalf("NewLocalTime: %v", err)
	}

	nw, err := sleep.NewNightWindow(nightStart, nightEnd)
	if err != nil {
		t.Fatalf("NewNightWindow: %v", err)
	}

	profile, err := sleep.NewSleepProfile(sleep.BabyID("baby-1"), "UTC", nw)
	if err != nil {
		t.Fatalf("NewSleepProfile: %v", err)
	}

	return profile
}

func mustActiveSleepSession(t *testing.T, startedAt time.Time) sleep.SleepSession {
	t.Helper()

	session, err := sleep.NewSleepSession(
		sleep.SleepSessionID("session-1"),
		sleep.BabyID("baby-1"),
		sleep.FamilyMemberID("member-1"),
		startedAt,
	)
	if err != nil {
		t.Fatalf("NewSleepSession: %v", err)
	}

	return session
}

func newStopSleepTestServer(
	resolver sleepContextResolver,
	sessRepo *stubStopSleepSessionRepo,
	profRepo *stubStopSleepProfileRepo,
) *http.Server {
	account := auth.Account{ID: "test-account-id", GoogleSubjectID: "google-subject-123"}
	session := auth.Session{
		Token:     testSessionToken,
		AccountID: "test-account-id",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	stopSleep := sleep.NewStopSleepHandler(sessRepo, profRepo)

	return NewServer(
		infrastructure.Config{HTTP: infrastructure.HTTPConfig{Port: 8080}},
		&stubAccountRepository{account: account},
		&stubSessionRepository{session: session},
		nil,
		nil,
		resolver,
		nil,
		nil,
		stopSleep,
	)
}

func deleteJSON(t *testing.T, server *http.Server, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testSessionToken)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)
	return rec
}

func TestStopSleepReturns200WithResult(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 22, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(8 * time.Hour)

	resolver := &stubSleepContextResolver{babyID: "baby-1", memberID: "member-1"}
	sessRepo := &stubStopSleepSessionRepo{active: mustActiveSleepSession(t, startedAt)}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newStopSleepTestServer(resolver, sessRepo, profRepo)
	rec := deleteJSON(t, server, "/sleep-sessions/active", map[string]any{
		"stopped_at": stoppedAt,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp stopSleepResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.ID == "" {
		t.Error("expected non-empty id")
	}

	if resp.Classification == "" {
		t.Error("expected non-empty classification")
	}
}

func TestStopSleepReturns401WhenUnauthenticated(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	resolver := &stubSleepContextResolver{babyID: "baby-1", memberID: "member-1"}
	sessRepo := &stubStopSleepSessionRepo{active: mustActiveSleepSession(t, startedAt)}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newStopSleepTestServer(resolver, sessRepo, profRepo)

	req := httptest.NewRequest(http.MethodDelete, "/sleep-sessions/active", nil)
	// No Authorization header.
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestStopSleepReturns404WhenNoActiveSession(t *testing.T) {
	t.Parallel()

	resolver := &stubSleepContextResolver{babyID: "baby-1", memberID: "member-1"}
	sessRepo := &stubStopSleepSessionRepo{
		findErr: apperror.New(apperror.CodeNotFound, "sleep session not found"),
	}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newStopSleepTestServer(resolver, sessRepo, profRepo)
	rec := deleteJSON(t, server, "/sleep-sessions/active", map[string]any{})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStopSleepReturns400WhenStopBeforeStart(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(-time.Minute)

	resolver := &stubSleepContextResolver{babyID: "baby-1", memberID: "member-1"}
	sessRepo := &stubStopSleepSessionRepo{active: mustActiveSleepSession(t, startedAt)}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newStopSleepTestServer(resolver, sessRepo, profRepo)
	rec := deleteJSON(t, server, "/sleep-sessions/active", map[string]any{
		"stopped_at": stoppedAt,
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}
