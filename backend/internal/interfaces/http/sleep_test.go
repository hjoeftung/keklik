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

type stubBabyAccessChecker struct {
	memberID sleep.FamilyMemberID
	err      error
}

func (c *stubBabyAccessChecker) CheckBabyAccess(_ context.Context, _ auth.AccountID, _ sleep.BabyID) (sleep.FamilyMemberID, error) {
	return c.memberID, c.err
}

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
}

func (r *stubEditableHTTPSleepSessionRepo) Save(_ context.Context, _ sleep.SleepSession) error {
	return r.saveErr
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
func (r *stubEditableHTTPSleepSessionRepo) DeleteByIDAndVersion(_ context.Context, _ sleep.SleepSessionID, _ int) error {
	return r.deleteErr
}
func (r *stubEditableHTTPSleepSessionRepo) FindOverlappingByBabyID(_ context.Context, _ sleep.BabyID, _, _ time.Time, _ *sleep.SleepSessionID) (sleep.SleepSession, error) {
	return sleep.SleepSession{}, apperror.New(apperror.CodeNotFound, "sleep session not found")
}

type stubStartSleepSessionRepo struct {
	saveErr error
}

func (r *stubStartSleepSessionRepo) Save(_ context.Context, _ sleep.SleepSession) error {
	return r.saveErr
}
func (r *stubStartSleepSessionRepo) SaveAll(_ context.Context, _ []sleep.SleepSession) error {
	return nil
}
func (r *stubStartSleepSessionRepo) FindByID(_ context.Context, _ sleep.SleepSessionID) (sleep.SleepSession, error) {
	return sleep.SleepSession{}, nil
}

type stubLogPastSleepSessionRepo struct {
	hasOverlap bool
	overlapErr error
	saveErr    error
}

func (r *stubLogPastSleepSessionRepo) Save(_ context.Context, _ sleep.SleepSession) error {
	return r.saveErr
}
func (r *stubLogPastSleepSessionRepo) SaveAll(_ context.Context, _ []sleep.SleepSession) error {
	return nil
}
func (r *stubLogPastSleepSessionRepo) FindByID(_ context.Context, _ sleep.SleepSessionID) (sleep.SleepSession, error) {
	return sleep.SleepSession{}, nil
}
func (r *stubLogPastSleepSessionRepo) HasOverlappingByBabyID(_ context.Context, _ sleep.BabyID, _, _ time.Time) (bool, error) {
	return r.hasOverlap, r.overlapErr
}

type stubSleepHistoryRepo struct {
	sessions []sleep.SleepSession
	err      error
}

func (r *stubSleepHistoryRepo) FindByBabyIDAndDateRange(_ context.Context, _ sleep.BabyID, _ sleep.DateRange) ([]sleep.SleepSession, error) {
	return r.sessions, r.err
}

type stubNightWindowRepo struct {
	windows []sleep.NightWindow
	err     error
}

func (r *stubNightWindowRepo) Save(_ context.Context, _ sleep.NightWindow) error { return nil }
func (r *stubNightWindowRepo) DeleteByIDs(_ context.Context, _ []sleep.NightWindowID) error {
	return nil
}
func (r *stubNightWindowRepo) FindByBabyID(_ context.Context, _ sleep.BabyID) ([]sleep.NightWindow, error) {
	return r.windows, r.err
}

func mustNightWindow(t *testing.T) sleep.NightWindow {
	t.Helper()

	start, _ := sleep.NewLocalTime(21, 0)
	end, _ := sleep.NewLocalTime(7, 0)
	nw, err := sleep.NewNightWindow("nw-1", "baby-1", start, end, time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), nil)
	if err != nil {
		t.Fatalf("NewNightWindow: %v", err)
	}
	return nw
}

func mustActiveSleepSession(t *testing.T, startedAt time.Time) sleep.SleepSession {
	t.Helper()
	session, err := sleep.NewSleepSession("session-1", "baby-1", "member-1", startedAt)
	if err != nil {
		t.Fatalf("NewSleepSession: %v", err)
	}
	return session
}

func newServerDeps() (auth.Account, *stubTokenValidator) {
	account := auth.Account{ID: "test-account-id", GoogleSubjectID: "google-subject-123"}
	validator := &stubTokenValidator{validToken: testSessionToken, identity: auth.Identity{AccountID: account.ID}}
	return account, validator
}

func newStopSleepTestServer(checker babyAccessChecker, sessRepo *stubStopSleepSessionRepo) *http.Server {
	account, validator := newServerDeps()
	return NewServer(infrastructure.Config{HTTP: infrastructure.HTTPConfig{Port: 8080}}, Dependencies{
		Accounts:   &stubAccountRepository{account: account},
		Validator:  validator,
		BabyAccess: checker,
		StopSleep:  sleep.NewStopSleepHandler(sessRepo),
	})
}

func newEditDeleteSleepTestServer(checker babyAccessChecker, sessRepo *stubEditableHTTPSleepSessionRepo) *http.Server {
	account, validator := newServerDeps()
	return NewServer(infrastructure.Config{HTTP: infrastructure.HTTPConfig{Port: 8080}}, Dependencies{
		Accounts:           &stubAccountRepository{account: account},
		Validator:          validator,
		BabyAccess:         checker,
		EditSleepSession:   sleep.NewEditSleepSessionHandler(sessRepo),
		DeleteSleepSession: sleep.NewDeleteSleepSessionHandler(sessRepo),
	})
}

func newStartSleepTestServer(checker babyAccessChecker, sessRepo *stubStartSleepSessionRepo) *http.Server {
	account, validator := newServerDeps()
	return NewServer(infrastructure.Config{HTTP: infrastructure.HTTPConfig{Port: 8080}}, Dependencies{
		Accounts:   &stubAccountRepository{account: account},
		Validator:  validator,
		BabyAccess: checker,
		StartSleep: sleep.NewStartSleepHandler(sessRepo),
	})
}

func newLogPastSleepTestServer(checker babyAccessChecker, sessRepo *stubLogPastSleepSessionRepo) *http.Server {
	account, validator := newServerDeps()
	return NewServer(infrastructure.Config{HTTP: infrastructure.HTTPConfig{Port: 8080}}, Dependencies{
		Accounts:     &stubAccountRepository{account: account},
		Validator:    validator,
		BabyAccess:   checker,
		LogPastSleep: sleep.NewLogPastSleepHandler(sessRepo),
	})
}

func newGetSleepHistoryTestServer(checker babyAccessChecker, histRepo *stubSleepHistoryRepo, nwRepo *stubNightWindowRepo) *http.Server {
	account, validator := newServerDeps()
	return NewServer(infrastructure.Config{HTTP: infrastructure.HTTPConfig{Port: 8080}}, Dependencies{
		Accounts:        &stubAccountRepository{account: account},
		Validator:       validator,
		BabyAccess:      checker,
		GetSleepHistory: sleep.NewGetSleepHistoryHandler(histRepo, nwRepo),
	})
}

func requestJSON(t *testing.T, server *http.Server, method, path string, body any, authz bool) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reader = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authz {
		req.Header.Set("Authorization", "Bearer "+testSessionToken)
	}
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)
	return rec
}

func TestStopSleepReturns200WithResult(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 22, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(8 * time.Hour)
	server := newStopSleepTestServer(&stubBabyAccessChecker{memberID: "member-1"}, &stubStopSleepSessionRepo{
		active: mustActiveSleepSession(t, startedAt),
	})

	rec := requestJSON(t, server, http.MethodDelete, "/babies/"+testBabyID+"/sleep-sessions/active", map[string]any{"stopped_at": stoppedAt}, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp stopSleepResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID == "" {
		t.Fatal("expected non-empty id")
	}
}

func TestStopSleepReturns401WhenUnauthenticated(t *testing.T) {
	t.Parallel()

	server := newStopSleepTestServer(&stubBabyAccessChecker{memberID: "member-1"}, &stubStopSleepSessionRepo{})
	rec := requestJSON(t, server, http.MethodDelete, "/babies/"+testBabyID+"/sleep-sessions/active", nil, false)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestStopSleepReturns404WhenNoActiveSession(t *testing.T) {
	t.Parallel()

	server := newStopSleepTestServer(&stubBabyAccessChecker{memberID: "member-1"}, &stubStopSleepSessionRepo{
		findErr: apperror.New(apperror.CodeNotFound, "sleep session not found"),
	})
	rec := requestJSON(t, server, http.MethodDelete, "/babies/"+testBabyID+"/sleep-sessions/active", map[string]any{}, true)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestEditSleepSessionReturns200WithUpdatedSession(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 10, 0, 0, 0, time.UTC)
	server := newEditDeleteSleepTestServer(&stubBabyAccessChecker{memberID: "member-1"}, &stubEditableHTTPSleepSessionRepo{
		session: mustActiveSleepSession(t, startedAt),
	})

	rec := requestJSON(t, server, http.MethodPatch, "/babies/"+testBabyID+"/sleep-sessions/session-1", map[string]any{
		"started_at": startedAt.Add(-15 * time.Minute),
		"version":    0,
	}, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteSleepSessionReturns204(t *testing.T) {
	t.Parallel()

	server := newEditDeleteSleepTestServer(&stubBabyAccessChecker{memberID: "member-1"}, &stubEditableHTTPSleepSessionRepo{})
	rec := requestJSON(t, server, http.MethodDelete, "/babies/"+testBabyID+"/sleep-sessions/session-1", map[string]any{"version": 0}, true)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStartSleepReturns201WithResult(t *testing.T) {
	t.Parallel()

	server := newStartSleepTestServer(&stubBabyAccessChecker{memberID: "member-1"}, &stubStartSleepSessionRepo{})
	rec := requestJSON(t, server, http.MethodPost, "/babies/"+testBabyID+"/sleep-sessions/active", map[string]any{
		"started_at": time.Date(2026, time.April, 16, 10, 0, 0, 0, time.UTC),
	}, true)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLogPastSleepReturns201WithResult(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 22, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(8 * time.Hour)
	server := newLogPastSleepTestServer(&stubBabyAccessChecker{memberID: "member-1"}, &stubLogPastSleepSessionRepo{})

	rec := requestJSON(t, server, http.MethodPost, "/babies/"+testBabyID+"/sleep-sessions", map[string]any{
		"started_at": startedAt,
		"stopped_at": stoppedAt,
	}, true)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp logPastSleepResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID == "" {
		t.Fatal("expected non-empty id")
	}
}

func TestLogPastSleepReturns401WhenUnauthenticated(t *testing.T) {
	t.Parallel()

	server := newLogPastSleepTestServer(&stubBabyAccessChecker{memberID: "member-1"}, &stubLogPastSleepSessionRepo{})
	rec := requestJSON(t, server, http.MethodPost, "/babies/"+testBabyID+"/sleep-sessions", nil, false)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestLogPastSleepReturns409WhenOverlapping(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 22, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(8 * time.Hour)
	server := newLogPastSleepTestServer(&stubBabyAccessChecker{memberID: "member-1"}, &stubLogPastSleepSessionRepo{hasOverlap: true})

	rec := requestJSON(t, server, http.MethodPost, "/babies/"+testBabyID+"/sleep-sessions", map[string]any{
		"started_at": startedAt,
		"stopped_at": stoppedAt,
	}, true)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLogPastSleepReturns400WhenStopBeforeStart(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 14, 22, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(-1 * time.Hour)
	server := newLogPastSleepTestServer(&stubBabyAccessChecker{memberID: "member-1"}, &stubLogPastSleepSessionRepo{})

	rec := requestJSON(t, server, http.MethodPost, "/babies/"+testBabyID+"/sleep-sessions", map[string]any{
		"started_at": startedAt,
		"stopped_at": stoppedAt,
	}, true)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func newGetSleepStatsTestServer(checker babyAccessChecker, sessRepo *stubSleepHistoryRepo, nwRepo *stubNightWindowRepo) *http.Server {
	account, validator := newServerDeps()
	return NewServer(infrastructure.Config{HTTP: infrastructure.HTTPConfig{Port: 8080}}, Dependencies{
		Accounts:      &stubAccountRepository{account: account},
		Validator:     validator,
		BabyAccess:    checker,
		GetSleepStats: sleep.NewGetSleepStatsHandler(sessRepo, nwRepo),
	})
}

func TestGetSleepStatsRequiresTimezone(t *testing.T) {
	t.Parallel()

	server := newGetSleepStatsTestServer(&stubBabyAccessChecker{memberID: "member-1"}, &stubSleepHistoryRepo{}, &stubNightWindowRepo{})
	rec := requestJSON(t, server, http.MethodGet, "/babies/"+testBabyID+"/sleep-stats", nil, true)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetSleepStatsReturns401WhenUnauthenticated(t *testing.T) {
	t.Parallel()

	server := newGetSleepStatsTestServer(&stubBabyAccessChecker{memberID: "member-1"}, &stubSleepHistoryRepo{}, &stubNightWindowRepo{})
	rec := requestJSON(t, server, http.MethodGet, "/babies/"+testBabyID+"/sleep-stats?timezone=UTC", nil, false)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestGetSleepStatsReturns200(t *testing.T) {
	t.Parallel()

	nap, err := sleep.NewCompletedSleepSession("nap-1", "baby-1", "member-1",
		time.Date(2026, time.April, 28, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 28, 11, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("NewCompletedSleepSession: %v", err)
	}

	server := newGetSleepStatsTestServer(
		&stubBabyAccessChecker{memberID: "member-1"},
		&stubSleepHistoryRepo{sessions: []sleep.SleepSession{nap}},
		&stubNightWindowRepo{windows: []sleep.NightWindow{mustNightWindow(t)}},
	)
	rec := requestJSON(t, server, http.MethodGet, "/babies/"+testBabyID+"/sleep-stats?timezone=UTC", nil, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		DiaryWindow struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"diary_window"`
		Summary map[string]struct {
			AvgSleepSeconds float64 `json:"avg_sleep_seconds"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.DiaryWindow.Start == "" || resp.DiaryWindow.End == "" {
		t.Fatalf("expected non-empty diary_window, got %+v", resp.DiaryWindow)
	}
	for _, key := range []string{"7d", "14d", "30d", "90d"} {
		if _, ok := resp.Summary[key]; !ok {
			t.Fatalf("missing summary key %q", key)
		}
	}
}

func TestGetSleepHistoryRequiresTimezone(t *testing.T) {
	t.Parallel()

	server := newGetSleepHistoryTestServer(&stubBabyAccessChecker{memberID: "member-1"}, &stubSleepHistoryRepo{}, &stubNightWindowRepo{})
	rec := requestJSON(t, server, http.MethodGet, "/babies/"+testBabyID+"/sleep-sessions", nil, true)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetSleepHistoryReturns200(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.April, 14, 22, 0, 0, 0, time.UTC)
	stop := start.Add(8 * time.Hour)
	session, err := sleep.NewCompletedSleepSession("session-1", "baby-1", "member-1", start, stop)
	if err != nil {
		t.Fatalf("NewCompletedSleepSession: %v", err)
	}

	server := newGetSleepHistoryTestServer(
		&stubBabyAccessChecker{memberID: "member-1"},
		&stubSleepHistoryRepo{sessions: []sleep.SleepSession{session}},
		&stubNightWindowRepo{windows: []sleep.NightWindow{mustNightWindow(t)}},
	)
	rec := requestJSON(t, server, http.MethodGet, "/babies/"+testBabyID+"/sleep-sessions?timezone=UTC", nil, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp []sleepSessionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 || resp[0].Classification != string(sleep.SleepClassificationNight) {
		t.Fatalf("expected classified response, got %+v", resp)
	}
}
