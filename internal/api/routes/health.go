package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"lory/internal/api"
)

type HealthResponse struct {
	Database bool `json:"dbhealth"`
}

func HealthHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err := db.Ping(ctx)

		resp := HealthResponse{
			Database: err == nil,
		}

		status := http.StatusOK
		if err != nil {
			status = http.StatusServiceUnavailable
		}

		api.JSON(w, status, resp)
	}
}
