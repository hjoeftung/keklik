package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/hjoeftung/keklik/internal/apperror"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, r *http.Request, err apperror.AppError) {
	status := apperror.HTTPStatus(err.Code)
	attrs := []any{
		"method", r.Method,
		"path", r.URL.Path,
		"error_code", string(err.Code),
		"status", status,
	}
	if status >= 500 {
		slog.ErrorContext(r.Context(), "request_failed", attrs...)
	} else {
		slog.WarnContext(r.Context(), "request_error", attrs...)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{
		Code:    string(err.Code),
		Message: err.Message,
	})
}
