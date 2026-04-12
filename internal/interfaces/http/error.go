package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/hjoeftung/keklik/internal/apperror"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, err apperror.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apperror.HTTPStatus(err.Code))
	_ = json.NewEncoder(w).Encode(errorResponse{
		Code:    string(err.Code),
		Message: err.Message,
	})
}
