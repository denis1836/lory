package routes

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"lory/internal/api"
	"lory/internal/config"
	"lory/internal/middleware"
)

func InitRoutes(db *pgxpool.Pool, cfg *config.Config) *api.Router {
	r := api.NewRouter()

	//MIDDLEWARE
	authMiddleware := middleware.Authenticate(db, cfg)

	//HANDLERS
	userHandler := NewUserHandler(db, cfg)
	projectHandler := NewProjectHandler(db, cfg)
	articleHandler := NewArticleHandler(db, cfg)

	//HTML
	r.Static("web-ui")

	//GET
	r.Get("/api/health", HealthHandler(db))

	//POST
	r.Post("/api/user/register", userHandler.Register)
	r.Post("/api/user/login", userHandler.Login)

	r.Post("/api/project/create", projectHandler.Create, authMiddleware)

	r.Post("/api/article/create", articleHandler.Create, authMiddleware)
	r.Get("/api/article/{id}", articleHandler.GetArticleById, authMiddleware)

	return r
}
