package routes

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"lory/internal/api"
	"lory/internal/config"
)

func InitRoutes(db *pgxpool.Pool, cfg *config.Config) *api.Router {
	r := api.NewRouter()

	//MIDDLEWARE
	//TODO: add middleware layyers

	//HANDLERS
	userHandler := NewUserHandler(db, cfg)

	//HTML
	r.Static("web-ui")

	//GET
	r.Get("/api/health", HealthHandler(db))

	//POST
	r.Post("/api/user/register", userHandler.Register)
	r.Post("/api/user/login", userHandler.Login)

	return r
}
