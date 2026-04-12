package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/hjoeftung/keklik/internal/family"
	"github.com/hjoeftung/keklik/internal/infrastructure"
)

type healthResponse struct {
	Status string `json:"status"`
}

// NewServer wires the HTTP transport and returns a ready-to-start server.
func NewServer(config infrastructure.Config, createFamily *family.CreateFamilyHandler) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("POST /families", func(w http.ResponseWriter, r *http.Request) {
		createFamilyHandler(w, r, createFamily)
	})

	return &http.Server{
		Addr:    config.Address(),
		Handler: mux,
	}
}

func healthHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(writer).Encode(healthResponse{Status: "ok"})
}
