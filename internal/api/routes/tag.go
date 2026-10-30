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
	"lory/internal/model"
)

type TagsHandler struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewTagsHandler(db *pgxpool.Pool, cfg *config.Config) *TagsHandler {
	return &TagsHandler{db: db, cfg: cfg}
}

// POST /api/tag/create
// creates a tag
func (h *TagsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var tag model.Tag

	if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
		api.Error(w, http.StatusBadRequest, "wront JSON format")
		return
	}

	if tag.Name == "" {
		api.Error(w, http.StatusBadRequest, "missing fields")
		return
	}
	if tag.Color == "" {
		tag.Color = "#000000"
	}

	createTagQuery := `
		INSERT INTO Tag(prjID, name, description, color)
		VALUES($1, $2, $3, $4)
		RETURNING tgID, created_at;
	`
	err := h.db.QueryRow(r.Context(), createTagQuery, tag.ProjectID, tag.Name, tag.Description, tag.Color).Scan(
		&tag.ID, &tag.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == " tags_name_lower_uniq" {
				api.Error(w, http.StatusBadRequest, "a tag with this slug already exists")
			} else {
				api.Error(w, http.StatusBadRequest, "duplicate key value violation")
			}
			return
		}

		log.Printf("INSERT DB ERROR: %v", err)
		api.Error(w, http.StatusInternalServerError, "database error")
		return
	}

	api.JSON(w, http.StatusCreated, map[string]interface{}{
		"message": "tag created",
		"tag":     tag,
	})

	return
}
