package infrastructure

import (
	"context"
	"log/slog"
	"os"

	"github.com/hjoeftung/keklik/internal/reqctx"
)

// SetupLogger configures the global slog logger to emit JSON with request IDs
// automatically extracted from the context on every log call.
func SetupLogger() {
	base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(contextHandler{base}))
}

type contextHandler struct{ slog.Handler }

func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id, ok := reqctx.RequestIDFromContext(ctx); ok {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, r)
}
