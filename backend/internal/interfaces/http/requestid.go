package httpapi

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/hjoeftung/keklik/internal/reqctx"
)

const headerRequestID = "X-Request-ID"

// withRequestID is middleware that generates a unique request ID, injects it
// into the context, and echoes it in the X-Request-ID response header.
func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set(headerRequestID, id)
		next.ServeHTTP(w, r.WithContext(reqctx.WithRequestID(r.Context(), id)))
	})
}
