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

	"github.com/hjoeftung/keklik/internal/apperror"
	"github.com/hjoeftung/keklik/internal/auth"
	"github.com/hjoeftung/keklik/internal/family"
	"github.com/hjoeftung/keklik/internal/infrastructure"
	"github.com/hjoeftung/keklik/internal/sleep"
)

const testSessionToken = "test-session-token"

// stubFamilyRepository is a minimal FamilyRepository test double.
type stubFamilyRepository struct {
	family family.Family
	err    error
}

func (r *stubFamilyRepository) Save(_ context.Context, f family.Family) error {
	if r.err == nil {
		r.family = f
	}
	return r.err
}

func (r *stubFamilyRepository) FindByID(_ context.Context, _ family.FamilyID) (family.Family, error) {
	if r.err != nil {
		return family.Family{}, r.err
	}
	if r.family.ID() == "" {
		return family.Family{}, apperror.New(apperror.CodeNotFound, "family not found")
	}
	return r.family, nil
}

func (r *stubFamilyRepository) FindByMemberID(_ context.Context, _ family.FamilyMemberID) (family.Family, error) {
	if r.err != nil {
		return family.Family{}, r.err
	}
	if r.family.ID() == "" {
		return family.Family{}, apperror.New(apperror.CodeNotFound, "family not found")
	}
	return r.family, nil
}

func (r *stubFamilyRepository) FindByAccountID(_ context.Context, _ auth.AccountID) (family.Family, error) {
	if r.err != nil {
		return family.Family{}, r.err
	}
	if r.family.ID() == "" {
		return family.Family{}, apperror.New(apperror.CodeNotFound, "family not found")
	}
	return r.family, nil
}

func (r *stubFamilyRepository) FindByInviteToken(_ context.Context, token family.InviteToken) (family.Family, error) {
	if r.err != nil {
		return family.Family{}, r.err
	}
	if r.family.ID() == "" {
		return family.Family{}, apperror.New(apperror.CodeNotFound, "family not found")
	}
	for _, inviteLink := range r.family.InviteLinks() {
		if inviteLink.Token == token {
			return r.family, nil
		}
	}
	return family.Family{}, apperror.New(apperror.CodeNotFound, "family not found")
}

type stubFamilyMemberRepository struct {
	member family.FamilyMember
}

func (r *stubFamilyMemberRepository) Save(_ context.Context, member family.FamilyMember) error {
	r.member = member
	return nil
}

func (r *stubFamilyMemberRepository) FindByID(_ context.Context, _ family.FamilyMemberID) (family.FamilyMember, error) {
	if r.member.ID == "" {
		return family.FamilyMember{}, apperror.New(apperror.CodeNotFound, "family member not found")
	}
	return r.member, nil
}

func (r *stubFamilyMemberRepository) FindByAccountID(_ context.Context, accountID auth.AccountID) (family.FamilyMember, error) {
	if r.member.AccountID == accountID {
		return r.member, nil
	}
	return family.FamilyMember{}, apperror.New(apperror.CodeNotFound, "family member not found")
}

// stubAccountRepository is a minimal AccountRepository test double.
type stubAccountRepository struct {
	account auth.Account
	saved   []auth.Account
	err     error
}

func (r *stubAccountRepository) Save(_ context.Context, account auth.Account) error {
	if r.err != nil {
		return r.err
	}
	r.saved = append(r.saved, account)
	r.account = account
	return nil
}
func (r *stubAccountRepository) FindByID(_ context.Context, _ auth.AccountID) (auth.Account, error) {
	if r.err != nil {
		return auth.Account{}, r.err
	}
	return r.account, nil
}
func (r *stubAccountRepository) FindByGoogleSubjectID(_ context.Context, googleSubjectID string) (auth.Account, error) {
	if r.err != nil {
		return auth.Account{}, r.err
	}
	if r.account.GoogleSubjectID != googleSubjectID {
		return auth.Account{}, auth.ErrAccountNotFound
	}
	return r.account, nil
}

// stubRefreshTokenRepository is a minimal RefreshTokenRepository test double.
type stubRefreshTokenRepository struct{}

func (r *stubRefreshTokenRepository) Save(_ context.Context, _ auth.RefreshToken) error {
	return nil
}
func (r *stubRefreshTokenRepository) FindByToken(_ context.Context, _ string) (auth.RefreshToken, error) {
	return auth.RefreshToken{}, auth.ErrRefreshTokenNotFound
}
func (r *stubRefreshTokenRepository) Revoke(_ context.Context, _ string) error { return nil }
func (r *stubRefreshTokenRepository) RevokeAllForAccount(_ context.Context, _ auth.AccountID) error {
	return nil
}

// stubTokenValidator is a minimal TokenValidator test double.
// If validToken is non-empty, only that exact token is accepted.
type stubTokenValidator struct {
	validToken string
	identity   auth.Identity
	err        error
}

func (v *stubTokenValidator) Validate(_ context.Context, token string) (auth.Identity, error) {
	if v.err != nil {
		return auth.Identity{}, v.err
	}
	if v.validToken != "" && token != v.validToken {
		return auth.Identity{}, auth.ErrInvalidToken
	}
	return v.identity, nil
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

type stubCompletedSessionsRepo struct{}

func (r *stubCompletedSessionsRepo) FindCompletedByBabyIDSince(_ context.Context, _ sleep.BabyID, _ time.Time) ([]sleep.SleepSession, error) {
	return nil, nil
}

type stubSleepSessionWriter struct{}

func (r *stubSleepSessionWriter) Save(_ context.Context, _ sleep.SleepSession) error {
	return nil
}

func (r *stubSleepSessionWriter) SaveAll(_ context.Context, _ []sleep.SleepSession) error {
	return nil
}

func (r *stubSleepSessionWriter) FindByID(_ context.Context, _ sleep.SleepSessionID) (sleep.SleepSession, error) {
	return sleep.SleepSession{}, errors.New("not implemented")
}

type stubTransactor struct{}

func (t *stubTransactor) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func newTestServer(familyRepo family.FamilyRepository, memberRepo family.FamilyMemberRepository) *http.Server {
	createFamily := family.NewCreateFamilyHandler(familyRepo)
	createInviteLink := family.NewCreateFamilyInviteLinkHandler(
		familyRepo,
		memberRepo,
		"http://localhost:8080",
		24*time.Hour,
	)
	joinByInvite := family.NewJoinFamilyByInviteLinkHandler(familyRepo, memberRepo)
	createSleepProfile := sleep.NewCreateSleepProfileHandler(&stubSleepProfileRepository{}, &stubCompletedSessionsRepo{}, &stubSleepSessionWriter{}, &stubTransactor{})
	account := auth.Account{
		ID:              "test-account-id",
		GoogleSubjectID: "google-subject-123",
		Email:           "alice@example.com",
	}
	return NewServer(
		infrastructure.Config{
			HTTP: infrastructure.HTTPConfig{Port: 8080},
			App:  infrastructure.AppConfig{BaseURL: "http://localhost:8080"},
		},
		Dependencies{
			Accounts:           &stubAccountRepository{account: account},
			Validator:          &stubTokenValidator{validToken: testSessionToken, identity: auth.Identity{AccountID: account.ID}},
			CreateFamily:       createFamily,
			CreateInviteLink:   createInviteLink,
			JoinFamilyByInvite: joinByInvite,
			CreateSleepProfile: createSleepProfile,
		},
	)
}

func validCreateFamilyBody() map[string]any {
	return map[string]any{
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

	server := newTestServer(&stubFamilyRepository{}, &stubFamilyMemberRepository{})
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

	server := newTestServer(&stubFamilyRepository{}, &stubFamilyMemberRepository{})
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

	server := newTestServer(&stubFamilyRepository{}, &stubFamilyMemberRepository{})
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

func TestCreateFamilyRejects401WithWrongToken(t *testing.T) {
	t.Parallel()

	server := newTestServer(&stubFamilyRepository{}, &stubFamilyMemberRepository{})
	b, _ := json.Marshal(validCreateFamilyBody())
	req := httptest.NewRequest(http.MethodPost, "/families", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCreateFamilyInviteLinkReturns201WithURL(t *testing.T) {
	t.Parallel()

	aggregate, err := family.NewFamily(
		family.FamilyID("family-1"),
		[]family.FamilyMember{{
			ID:        family.FamilyMemberID("member-1"),
			FamilyID:  family.FamilyID("family-1"),
			Name:      "Alice",
			AccountID: "test-account-id",
		}},
		[]family.Baby{{ID: family.BabyID("baby-1"), FamilyID: family.FamilyID("family-1"), Name: "Emma"}},
	)
	if err != nil {
		t.Fatalf("seed family: %v", err)
	}

	server := newTestServer(
		&stubFamilyRepository{family: aggregate},
		&stubFamilyMemberRepository{member: aggregate.Members()[0]},
	)
	rec := postJSON(t, server, "/families/invite-links", map[string]any{})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp createFamilyInviteLinkResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if resp.InviteURL == "" {
		t.Fatal("expected non-empty invite URL")
	}
}

func TestJoinFamilyByInviteLinkReturns201WithMembership(t *testing.T) {
	t.Parallel()

	aggregate, err := family.NewFamily(
		family.FamilyID("family-1"),
		[]family.FamilyMember{{
			ID:        family.FamilyMemberID("member-1"),
			FamilyID:  family.FamilyID("family-1"),
			Name:      "Alice",
			AccountID: "google-owner",
		}},
		[]family.Baby{{ID: family.BabyID("baby-1"), FamilyID: family.FamilyID("family-1"), Name: "Emma"}},
	)
	if err != nil {
		t.Fatalf("seed family: %v", err)
	}
	if err := aggregate.AddInviteLink(family.InviteLink{
		Token:             family.InviteToken("invite-1"),
		FamilyID:          aggregate.ID(),
		CreatedByMemberID: family.FamilyMemberID("member-1"),
		ExpiresAt:         time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("seed invite link: %v", err)
	}

	server := newTestServer(
		&stubFamilyRepository{family: aggregate},
		&stubFamilyMemberRepository{},
	)
	rec := postJSON(t, server, "/families/join-by-invite-link", map[string]any{
		"token": "invite-1",
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp joinFamilyByInviteLinkResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.FamilyID == "" || resp.MemberID == "" {
		t.Fatalf("expected non-empty ids, got %+v", resp)
	}
}

func TestJoinFamilyByInviteLinkRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	aggregate, err := family.NewFamily(
		family.FamilyID("family-1"),
		[]family.FamilyMember{{
			ID:        family.FamilyMemberID("member-1"),
			FamilyID:  family.FamilyID("family-1"),
			Name:      "Alice",
			AccountID: "google-owner",
		}},
		[]family.Baby{{ID: family.BabyID("baby-1"), FamilyID: family.FamilyID("family-1"), Name: "Emma"}},
	)
	if err != nil {
		t.Fatalf("seed family: %v", err)
	}

	server := newTestServer(
		&stubFamilyRepository{family: aggregate},
		&stubFamilyMemberRepository{},
	)
	rec := postJSON(t, server, "/families/join-by-invite-link", map[string]any{
		"token": "missing",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}
