package handler

import (
	"encoding/json"
	"net/http"
	"time"
	"database/sql"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string           `json:"status"`
	Database  DatabaseHealth   `json:"database"`
	Timestamp time.Time        `json:"timestamp"`
}

type DatabaseHealth struct {
	Status          string `json:"status"`
	OpenConnections int    `json:"open_connections"`
	InUse           int    `json:"in_use"`
	Idle            int    `json:"idle"`
	MaxOpen         int    `json:"max_open"`
}

// Check handles GET /health
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbStats := h.db.Stats()

	resp := HealthResponse{
		Timestamp: time.Now().UTC(),
		Database: DatabaseHealth{
			OpenConnections: dbStats.OpenConnections,
			InUse:           dbStats.InUse,
			Idle:            dbStats.Idle,
			MaxOpen:         dbStats.MaxOpenConnections,
		},
	}

	if err := h.db.Ping(); err != nil {
		resp.Status = "degraded"
		resp.Database.Status = "disconnected"
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	resp.Status = "ok"
	resp.Database.Status = "connected"
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

/*
🎓 NOTES:

Refactored từ Buổi 1:
- Buổi 1: Health check logic trong main.go
- Buổi 2: Extracted to separate handler

Benefits:
- Consistent with other handlers
- Can add more health checks (database, etc.) in Buổi 3
- Reusable and testable
*/
