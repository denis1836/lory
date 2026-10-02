package routes

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"lory/internal/api"
)

func InitRoutes(db *pgxpool.Pool) *api.Router {
	r := api.NewRouter()

	//TODO: add middleware layyers

	r.Get("/api/health", HealthHandler(db))

	r.Static("web-ui")

	return r
}
