package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hjoeftung/keklik/internal/family"
	"github.com/hjoeftung/keklik/internal/infrastructure"
)

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

func newTestServer(repo family.FamilyRepository) *http.Server {
	h := family.NewCreateFamilyHandler(repo)
	return NewServer(infrastructure.Config{
		HTTP: infrastructure.HTTPConfig{Port: 8080},
	}, h)
}

func validCreateFamilyBody() map[string]any {
	return map[string]any{
		"family_name": "Smith Family",
		"baby_name":   "Emma",
		"timezone":    "Europe/Berlin",
		"night_window": map[string]any{
			"start_hour":   20,
			"start_minute": 30,
			"end_hour":     7,
			"end_minute":   0,
		},
		"creator_name":             "Alice",
		"creator_google_subject_id": "google-subject-123",
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

func TestCreateFamilyRejects400OnInvalidTimezone(t *testing.T) {
	t.Parallel()

	server := newTestServer(&stubFamilyRepository{})
	body := validCreateFamilyBody()
	body["timezone"] = "Not/ATimezone"

	rec := postJSON(t, server, "/families", body)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var resp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Code != "invalid_timezone" {
		t.Errorf("expected code %q, got %q", "invalid_timezone", resp.Code)
	}
}

func TestCreateFamilyRejects400OnInvalidNightWindow(t *testing.T) {
	t.Parallel()

	server := newTestServer(&stubFamilyRepository{})
	body := validCreateFamilyBody()
	body["night_window"] = map[string]any{
		"start_hour":   8,
		"start_minute": 0,
		"end_hour":     8,
		"end_minute":   0,
	}

	rec := postJSON(t, server, "/families", body)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateFamilyRejects400OnMalformedJSON(t *testing.T) {
	t.Parallel()

	server := newTestServer(&stubFamilyRepository{})
	req := httptest.NewRequest(http.MethodPost, "/families", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
