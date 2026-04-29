package sleep

import (
	"context"
	"time"
)

type SleepSessionRepository interface {
	Save(ctx context.Context, session SleepSession) error
	SaveAll(ctx context.Context, sessions []SleepSession) error
	FindByID(ctx context.Context, id SleepSessionID) (SleepSession, error)
}

// Transactor executes a function within a database transaction, committing on
// success and rolling back on any error.
type Transactor interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type EditableSleepSessionRepository interface {
	SleepSessionRepository
	DeleteByID(ctx context.Context, id SleepSessionID) error
}

type ActiveSleepSessionRepository interface {
	FindActiveByBabyID(ctx context.Context, babyID BabyID) (SleepSession, error)
}

type SleepSessionHistoryRepository interface {
	FindByBabyIDAndDateRange(ctx context.Context, babyID BabyID, dateRange DateRange) ([]SleepSession, error)
}

type NightWindowRepository interface {
	Save(ctx context.Context, nw NightWindow) error
	DeleteByIDs(ctx context.Context, ids []NightWindowID) error
	FindByBabyID(ctx context.Context, babyID BabyID) ([]NightWindow, error)
}

// CompletedSleepSessionsSinceRepository finds completed (stopped) sessions from a given point in time.
type CompletedSleepSessionsSinceRepository interface {
	FindCompletedByBabyIDSince(ctx context.Context, babyID BabyID, since time.Time) ([]SleepSession, error)
}

// SleepElapsedTimeRepository finds the most recent sessions for elapsed-time calculations.
// Both methods return an AppError with CodeNotFound when no matching session exists.
type SleepElapsedTimeRepository interface {
	FindMostRecentByBabyID(ctx context.Context, babyID BabyID) (SleepSession, error)
	FindMostRecentCompletedByBabyID(ctx context.Context, babyID BabyID) (SleepSession, error)
}
