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

const testBabyID = "00000000-0000-0000-0000-000000000001"

// stubBabyAccessChecker is a babyAccessChecker test double.
type stubBabyAccessChecker struct {
	memberID sleep.FamilyMemberID
	err      error
}

func (c *stubBabyAccessChecker) CheckBabyAccess(_ context.Context, _ string, _ sleep.BabyID) (sleep.FamilyMemberID, error) {
	return c.memberID, c.err
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

func (r *stubStopSleepSessionRepo) SaveAll(_ context.Context, _ []sleep.SleepSession) error {
	return nil
}

func (r *stubStopSleepSessionRepo) FindByID(_ context.Context, _ sleep.SleepSessionID) (sleep.SleepSession, error) {
	return sleep.SleepSession{}, nil
}

func (r *stubStopSleepSessionRepo) FindActiveByBabyID(_ context.Context, _ sleep.BabyID) (sleep.SleepSession, error) {
	return r.active, r.findErr
}

type stubEditableHTTPSleepSessionRepo struct {
	session   sleep.SleepSession
	findErr   error
	saveErr   error
	deleteErr error
	saved     *sleep.SleepSession
}

func (r *stubEditableHTTPSleepSessionRepo) Save(_ context.Context, s sleep.SleepSession) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved = &s
	return nil
}

func (r *stubEditableHTTPSleepSessionRepo) SaveAll(_ context.Context, _ []sleep.SleepSession) error {
	return nil
}

func (r *stubEditableHTTPSleepSessionRepo) FindByID(_ context.Context, _ sleep.SleepSessionID) (sleep.SleepSession, error) {
	return r.session, r.findErr
}

func (r *stubEditableHTTPSleepSessionRepo) DeleteByID(_ context.Context, _ sleep.SleepSessionID) error {
	return r.deleteErr
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
	checker babyAccessChecker,
	sessRepo *stubStopSleepSessionRepo,
	profRepo *stubStopSleepProfileRepo,
) *http.Server {
	account := auth.Account{ID: "test-account-id", GoogleSubjectID: "google-subject-123"}
	stopSleep := sleep.NewStopSleepHandler(sessRepo, profRepo)

	return NewServer(
		infrastructure.Config{HTTP: infrastructure.HTTPConfig{Port: 8080}},
		Dependencies{
			Accounts:   &stubAccountRepository{account: account},
			Validator:  &stubTokenValidator{validToken: testSessionToken, identity: auth.Identity{AccountID: account.ID}},
			BabyAccess: checker,
			StopSleep:  stopSleep,
		},
	)
}

func newEditDeleteSleepTestServer(
	checker babyAccessChecker,
	sessRepo *stubEditableHTTPSleepSessionRepo,
	profRepo *stubStopSleepProfileRepo,
) *http.Server {
	account := auth.Account{ID: "test-account-id", GoogleSubjectID: "google-subject-123"}
	editSleep := sleep.NewEditSleepSessionHandler(sessRepo, profRepo)
	deleteSleep := sleep.NewDeleteSleepSessionHandler(sessRepo)

	return NewServer(
		infrastructure.Config{HTTP: infrastructure.HTTPConfig{Port: 8080}},
		Dependencies{
			Accounts:           &stubAccountRepository{account: account},
			Validator:          &stubTokenValidator{validToken: testSessionToken, identity: auth.Identity{AccountID: account.ID}},
			BabyAccess:         checker,
			EditSleepSession:   editSleep,
			DeleteSleepSession: deleteSleep,
		},
	)
}

// stubStartSleepSessionRepo implements SleepSessionRepository for start-sleep tests.
type stubStartSleepSessionRepo struct {
	saveErr error
	saved   *sleep.SleepSession
}

func (r *stubStartSleepSessionRepo) Save(_ context.Context, s sleep.SleepSession) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved = &s
	return nil
}

func (r *stubStartSleepSessionRepo) SaveAll(_ context.Context, _ []sleep.SleepSession) error {
	return nil
}

func (r *stubStartSleepSessionRepo) FindByID(_ context.Context, _ sleep.SleepSessionID) (sleep.SleepSession, error) {
	return sleep.SleepSession{}, nil
}

// stubSleepHistoryRepo implements SleepSessionHistoryRepository for history tests.
type stubSleepHistoryRepo struct {
	sessions []sleep.SleepSession
	err      error
}

func (r *stubSleepHistoryRepo) FindByBabyIDAndDateRange(_ context.Context, _ sleep.BabyID, _ sleep.DateRange) ([]sleep.SleepSession, error) {
	return r.sessions, r.err
}

func newStartSleepTestServer(
	checker babyAccessChecker,
	sessRepo *stubStartSleepSessionRepo,
) *http.Server {
	account := auth.Account{ID: "test-account-id", GoogleSubjectID: "google-subject-123"}
	startSleep := sleep.NewStartSleepHandler(sessRepo)

	return NewServer(
		infrastructure.Config{HTTP: infrastructure.HTTPConfig{Port: 8080}},
		Dependencies{
			Accounts:   &stubAccountRepository{account: account},
			Validator:  &stubTokenValidator{validToken: testSessionToken, identity: auth.Identity{AccountID: account.ID}},
			BabyAccess: checker,
			StartSleep: startSleep,
		},
	)
}

func newGetSleepHistoryTestServer(
	checker babyAccessChecker,
	histRepo *stubSleepHistoryRepo,
	profRepo *stubStopSleepProfileRepo,
) *http.Server {
	account := auth.Account{ID: "test-account-id", GoogleSubjectID: "google-subject-123"}
	getSleepHistory := sleep.NewGetSleepHistoryHandler(histRepo, profRepo)

	return NewServer(
		infrastructure.Config{HTTP: infrastructure.HTTPConfig{Port: 8080}},
		Dependencies{
			Accounts:        &stubAccountRepository{account: account},
			Validator:       &stubTokenValidator{validToken: testSessionToken, identity: auth.Identity{AccountID: account.ID}},
			BabyAccess:      checker,
			GetSleepHistory: getSleepHistory,
		},
	)
}

func getJSON(t *testing.T, server *http.Server, path string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+testSessionToken)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)
	return rec
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

func patchJSON(t *testing.T, server *http.Server, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(b))
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

	checker := &stubBabyAccessChecker{memberID: "member-1"}
	sessRepo := &stubStopSleepSessionRepo{active: mustActiveSleepSession(t, startedAt)}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newStopSleepTestServer(checker, sessRepo, profRepo)
	rec := deleteJSON(t, server, "/babies/"+testBabyID+"/sleep-sessions/active", map[string]any{
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
	checker := &stubBabyAccessChecker{memberID: "member-1"}
	sessRepo := &stubStopSleepSessionRepo{active: mustActiveSleepSession(t, startedAt)}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newStopSleepTestServer(checker, sessRepo, profRepo)

	req := httptest.NewRequest(http.MethodDelete, "/babies/"+testBabyID+"/sleep-sessions/active", nil)
	// No Authorization header.
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestStopSleepReturns404WhenNoActiveSession(t *testing.T) {
	t.Parallel()

	checker := &stubBabyAccessChecker{memberID: "member-1"}
	sessRepo := &stubStopSleepSessionRepo{
		findErr: apperror.New(apperror.CodeNotFound, "sleep session not found"),
	}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newStopSleepTestServer(checker, sessRepo, profRepo)
	rec := deleteJSON(t, server, "/babies/"+testBabyID+"/sleep-sessions/active", map[string]any{})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStopSleepReturns400WhenStopBeforeStart(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(-time.Minute)

	checker := &stubBabyAccessChecker{memberID: "member-1"}
	sessRepo := &stubStopSleepSessionRepo{active: mustActiveSleepSession(t, startedAt)}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newStopSleepTestServer(checker, sessRepo, profRepo)
	rec := deleteJSON(t, server, "/babies/"+testBabyID+"/sleep-sessions/active", map[string]any{
		"stopped_at": stoppedAt,
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStopSleepReturns403WhenNotFamilyMember(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	checker := &stubBabyAccessChecker{err: apperror.New(apperror.CodeForbidden, "access to this baby is not allowed")}
	sessRepo := &stubStopSleepSessionRepo{active: mustActiveSleepSession(t, startedAt)}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newStopSleepTestServer(checker, sessRepo, profRepo)
	rec := deleteJSON(t, server, "/babies/"+testBabyID+"/sleep-sessions/active", map[string]any{})

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStopSleepReturns404WhenBabyNotFound(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	checker := &stubBabyAccessChecker{err: apperror.New(apperror.CodeNotFound, "baby not found")}
	sessRepo := &stubStopSleepSessionRepo{active: mustActiveSleepSession(t, startedAt)}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newStopSleepTestServer(checker, sessRepo, profRepo)
	rec := deleteJSON(t, server, "/babies/"+testBabyID+"/sleep-sessions/active", map[string]any{})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestEditSleepSessionReturns200WithUpdatedSession(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	editedStart := startedAt.Add(-15 * time.Minute)

	checker := &stubBabyAccessChecker{memberID: "member-1"}
	sessRepo := &stubEditableHTTPSleepSessionRepo{session: mustActiveSleepSession(t, startedAt)}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newEditDeleteSleepTestServer(checker, sessRepo, profRepo)
	rec := patchJSON(t, server, "/babies/"+testBabyID+"/sleep-sessions/session-1", map[string]any{
		"started_at": editedStart,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp sleepSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.ID != "session-1" {
		t.Fatalf("expected session-1, got %q", resp.ID)
	}
	if resp.StoppedAt != nil {
		t.Fatal("expected edited active session to remain active")
	}
}

func TestEditSleepSessionReturns400ForInvalidInterval(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(-time.Minute)

	checker := &stubBabyAccessChecker{memberID: "member-1"}
	sessRepo := &stubEditableHTTPSleepSessionRepo{session: mustActiveSleepSession(t, startedAt)}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newEditDeleteSleepTestServer(checker, sessRepo, profRepo)
	rec := patchJSON(t, server, "/babies/"+testBabyID+"/sleep-sessions/session-1", map[string]any{
		"stopped_at": stoppedAt,
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteSleepSessionReturns204(t *testing.T) {
	t.Parallel()

	checker := &stubBabyAccessChecker{memberID: "member-1"}
	sessRepo := &stubEditableHTTPSleepSessionRepo{}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newEditDeleteSleepTestServer(checker, sessRepo, profRepo)
	rec := deleteJSON(t, server, "/babies/"+testBabyID+"/sleep-sessions/session-1", map[string]any{})

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteSleepSessionReturns404WhenMissing(t *testing.T) {
	t.Parallel()

	checker := &stubBabyAccessChecker{memberID: "member-1"}
	sessRepo := &stubEditableHTTPSleepSessionRepo{
		deleteErr: apperror.New(apperror.CodeNotFound, "sleep session not found"),
	}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newEditDeleteSleepTestServer(checker, sessRepo, profRepo)
	rec := deleteJSON(t, server, "/babies/"+testBabyID+"/sleep-sessions/missing", map[string]any{})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- startSleepHandler tests ---

func TestStartSleepReturns201WithResult(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 16, 10, 0, 0, 0, time.UTC)
	checker := &stubBabyAccessChecker{memberID: "member-1"}
	sessRepo := &stubStartSleepSessionRepo{}

	server := newStartSleepTestServer(checker, sessRepo)
	rec := postJSON(t, server, "/babies/"+testBabyID+"/sleep-sessions/active", map[string]any{
		"started_at": startedAt,
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp startSleepResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID == "" {
		t.Error("expected non-empty id")
	}
}

func TestStartSleepReturns401WhenUnauthenticated(t *testing.T) {
	t.Parallel()

	checker := &stubBabyAccessChecker{memberID: "member-1"}
	sessRepo := &stubStartSleepSessionRepo{}

	server := newStartSleepTestServer(checker, sessRepo)

	req := httptest.NewRequest(http.MethodPost, "/babies/"+testBabyID+"/sleep-sessions/active", nil)
	// No Authorization header.
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestStartSleepReturns409WhenActiveSessionExists(t *testing.T) {
	t.Parallel()

	checker := &stubBabyAccessChecker{memberID: "member-1"}
	sessRepo := &stubStartSleepSessionRepo{
		saveErr: apperror.New(apperror.CodeActiveSleepExists, "active session exists"),
	}

	server := newStartSleepTestServer(checker, sessRepo)
	rec := postJSON(t, server, "/babies/"+testBabyID+"/sleep-sessions/active", map[string]any{})

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- getSleepHistoryHandler tests ---

func TestGetSleepHistoryReturns200WithDefaultPeriod(t *testing.T) {
	t.Parallel()

	checker := &stubBabyAccessChecker{memberID: "member-1"}
	histRepo := &stubSleepHistoryRepo{sessions: []sleep.SleepSession{}}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newGetSleepHistoryTestServer(checker, histRepo, profRepo)
	rec := getJSON(t, server, "/babies/"+testBabyID+"/sleep-sessions")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp []sleepSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil array")
	}
}

func TestGetSleepHistoryReturns400ForInvalidPeriod(t *testing.T) {
	t.Parallel()

	checker := &stubBabyAccessChecker{memberID: "member-1"}
	histRepo := &stubSleepHistoryRepo{}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newGetSleepHistoryTestServer(checker, histRepo, profRepo)
	rec := getJSON(t, server, "/babies/"+testBabyID+"/sleep-sessions?period=30d")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetSleepHistoryReturns401WhenUnauthenticated(t *testing.T) {
	t.Parallel()

	checker := &stubBabyAccessChecker{memberID: "member-1"}
	histRepo := &stubSleepHistoryRepo{}
	profRepo := &stubStopSleepProfileRepo{profile: mustSleepProfile(t)}

	server := newGetSleepHistoryTestServer(checker, histRepo, profRepo)

	req := httptest.NewRequest(http.MethodGet, "/babies/"+testBabyID+"/sleep-sessions", nil)
	// No Authorization header.
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
