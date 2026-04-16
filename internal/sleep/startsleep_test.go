package sleep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hjoeftung/keklik/internal/apperror"
)

func TestStartSleepHappyPathPersistsSession(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, time.April, 16, 10, 0, 0, 0, time.UTC)
	repo := &stubSleepSessionRepo{}
	h := NewStartSleepHandler(repo)

	result, err := h.Handle(context.Background(), StartSleepCommand{
		BabyID:            BabyID("baby-1"),
		CreatedByMemberID: FamilyMemberID("member-1"),
		StartedAt:         startedAt,
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if result.ID == "" {
		t.Error("expected non-empty ID")
	}
	if !result.StartedAt.Equal(startedAt) {
		t.Errorf("expected StartedAt %v, got %v", startedAt, result.StartedAt)
	}
	if repo.saved == nil {
		t.Fatal("expected session to be saved")
	}
	if !repo.saved.IsActive() {
		t.Error("expected saved session to be active")
	}
}

func TestStartSleepDefaultsStartedAtToNowWhenZero(t *testing.T) {
	t.Parallel()

	fixed := time.Date(2026, time.April, 16, 10, 0, 0, 0, time.UTC)
	repo := &stubSleepSessionRepo{}
	h := NewStartSleepHandler(repo)
	h.now = func() time.Time { return fixed }

	result, err := h.Handle(context.Background(), StartSleepCommand{
		BabyID:            BabyID("baby-1"),
		CreatedByMemberID: FamilyMemberID("member-1"),
		// StartedAt is zero
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if !result.StartedAt.Equal(fixed.UTC()) {
		t.Errorf("expected StartedAt %v, got %v", fixed.UTC(), result.StartedAt)
	}
}

func TestStartSleepEmptyBabyIDReturnsError(t *testing.T) {
	t.Parallel()

	repo := &stubSleepSessionRepo{}
	h := NewStartSleepHandler(repo)

	_, err := h.Handle(context.Background(), StartSleepCommand{
		BabyID:            BabyID(""),
		CreatedByMemberID: FamilyMemberID("member-1"),
		StartedAt:         time.Now().UTC(),
	})

	if !errors.Is(err, ErrEmptyBabyID) {
		t.Fatalf("expected ErrEmptyBabyID, got %v", err)
	}
}

func TestStartSleepEmptyMemberIDReturnsError(t *testing.T) {
	t.Parallel()

	repo := &stubSleepSessionRepo{}
	h := NewStartSleepHandler(repo)

	_, err := h.Handle(context.Background(), StartSleepCommand{
		BabyID:            BabyID("baby-1"),
		CreatedByMemberID: FamilyMemberID(""),
		StartedAt:         time.Now().UTC(),
	})

	if !errors.Is(err, ErrEmptyFamilyMemberID) {
		t.Fatalf("expected ErrEmptyFamilyMemberID, got %v", err)
	}
}

func TestStartSleepSaveErrorPropagates(t *testing.T) {
	t.Parallel()

	saveErr := apperror.New(apperror.CodeActiveSleepExists, "active session exists")
	repo := &stubSleepSessionRepo{saveErr: saveErr}
	h := NewStartSleepHandler(repo)

	_, err := h.Handle(context.Background(), StartSleepCommand{
		BabyID:            BabyID("baby-1"),
		CreatedByMemberID: FamilyMemberID("member-1"),
		StartedAt:         time.Now().UTC(),
	})

	var appErr apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Code != apperror.CodeActiveSleepExists {
		t.Errorf("expected CodeActiveSleepExists, got %q", appErr.Code)
	}
}
