package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/family"
	"github.com/hjoeftung/keklik/internal/infrastructure"
	"github.com/hjoeftung/keklik/internal/sleep"
)

const testSessionToken = "test-session-token"

// stubFamilyRepository is a minimal FamilyRepository test double.
type stubFamilyRepository struct {
	err error
}

func (r *stubFamilyRepository) Save(_ context.Context, _ family.Family) error {
	return r.err
}

func (r *stubFamilyRepository) FindByID(_ context.Context, _ family.FamilyID) (family.Family, error) {
	return family.Family{}, errors.New("not implemented")
}

func (r *stubFamilyRepository) FindByMemberID(_ context.Context, _ family.FamilyMemberID) (family.Family, error) {
	return family.Family{}, errors.New("not implemented")
}

func (r *stubFamilyRepository) FindByInviteToken(_ context.Context, _ family.InviteToken) (family.Family, error) {
	return family.Family{}, errors.New("not implemented")
}

// stubAccountRepository is a minimal AccountRepository test double.
type stubAccountRepository struct {
	account auth.Account
}

func (r *stubAccountRepository) Save(_ context.Context, _ auth.Account) error { return nil }
func (r *stubAccountRepository) FindByID(_ context.Context, _ auth.AccountID) (auth.Account, error) {
	return r.account, nil
}
func (r *stubAccountRepository) FindByGoogleSubjectID(_ context.Context, _ string) (auth.Account, error) {
	return r.account, nil
}

// stubSessionRepository is a minimal SessionRepository test double.
type stubSessionRepository struct {
	session auth.Session
	err     error
}

func (r *stubSessionRepository) Save(_ context.Context, _ auth.Session) error { return nil }
func (r *stubSessionRepository) FindByToken(_ context.Context, _ auth.SessionToken) (auth.Session, error) {
	if r.err != nil {
		return auth.Session{}, r.err
	}
	return r.session, nil
}

// stubSleepProfileRepository is a minimal SleepProfileRepository test double.
type stubSleepProfileRepository struct {
	err error
}

func (r *stubSleepProfileRepository) Save(_ context.Context, _ sleep.SleepProfile) error {
	return r.err
}

func (r *stubSleepProfileRepository) FindByBabyID(_ context.Context, _ sleep.BabyID) (sleep.SleepProfile, error) {
	return sleep.SleepProfile{}, errors.New("not implemented")
}

func newTestServer(familyRepo family.FamilyRepository) *http.Server {
	createFamily := family.NewCreateFamilyHandler(familyRepo)
	createSleepProfile := sleep.NewCreateSleepProfileHandler(&stubSleepProfileRepository{})
	account := auth.Account{ID: "test-account-id", GoogleSubjectID: "google-subject-123"}
	session := auth.Session{
		Token:     testSessionToken,
		AccountID: "test-account-id",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	return NewServer(
		infrastructure.Config{HTTP: infrastructure.HTTPConfig{Port: 8080}},
		Dependencies{
			Accounts:           &stubAccountRepository{account: account},
			Sessions:           &stubSessionRepository{session: session},
			CreateFamily:       createFamily,
			CreateSleepProfile: createSleepProfile,
		},
	)
}

func validCreateFamilyBody() map[string]any {
	return map[string]any{
		"family_name":  "Smith Family",
		"baby_name":    "Emma",
		"creator_name": "Alice",
	}
}

func postJSON(t *testing.T, server *http.Server, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testSessionToken)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)
	return rec
}

func TestCreateFamilyReturns201WithIDs(t *testing.T) {
	t.Parallel()

	server := newTestServer(&stubFamilyRepository{})
	rec := postJSON(t, server, "/families", validCreateFamilyBody())

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp createFamilyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.FamilyID == "" {
		t.Error("expected non-empty family_id")
	}
	if resp.MemberID == "" {
		t.Error("expected non-empty member_id")
	}
	if resp.BabyID == "" {
		t.Error("expected non-empty baby_id")
	}
}

func TestCreateFamilyRejects400OnMalformedJSON(t *testing.T) {
	t.Parallel()

	server := newTestServer(&stubFamilyRepository{})
	req := httptest.NewRequest(http.MethodPost, "/families", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testSessionToken)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateFamilyRejects401WhenUnauthenticated(t *testing.T) {
	t.Parallel()

	server := newTestServer(&stubFamilyRepository{})
	b, _ := json.Marshal(validCreateFamilyBody())
	req := httptest.NewRequest(http.MethodPost, "/families", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header.
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
