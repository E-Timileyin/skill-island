package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/E-Timileyin/skill-island/services/api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// APIError is the standard error response format.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Handler holds dependencies for HTTP route handlers.
type Handler struct {
	DB  *pgxpool.Pool
	Cfg config.Config
}

// Health returns the service health status.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
