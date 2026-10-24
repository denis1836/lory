package routes

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"lory/internal/api"
	"lory/internal/config"
	"lory/internal/middleware"
	"lory/internal/model"
)

type ProjectHandler struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewProjectHandler(db *pgxpool.Pool, cfg *config.Config) *ProjectHandler {
	return &ProjectHandler{db: db, cfg: cfg}
}

// POST /api/project/create
// creates a new project
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var prj model.Project
	var user, ok = middleware.GetUser(r.Context())
	if !ok {
		api.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&prj); err != nil {
		api.Error(w, http.StatusBadRequest, "wrong JSON format")
		return
	}

	if prj.Name == "" {
		api.Error(w, http.StatusBadRequest, "missing fields")
		return
	}

	createProjectQuery := `
		INSERT INTO Projects(usrID, name, description)
		VALUES($1, $2, $3)
		RETURNING prjID, created_at;
	`
	err := h.db.QueryRow(r.Context(), createProjectQuery, user.ID, prj.Name, prj.Description).Scan(
		&prj.ID, &prj.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			api.Error(w, http.StatusBadRequest, "a project with name already exists")
		}

		log.Printf("INSERT DB ERROR: %v", err)

		api.Error(w, http.StatusInternalServerError, "database error")
		return
	}

	api.JSON(w, http.StatusCreated, map[string]interface{}{
		"message": "project created",
		"project": prj,
	})

	return
}
