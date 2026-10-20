package routes

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"lory/internal/api"
	"lory/internal/config"
	"lory/internal/model"
	"lory/internal/middleware"
)

type ArticleHandler struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewArticleHandler(db *pgxpool.Pool, cfg *config.Config) *ArticleHandler {
	return &ArticleHandler{db: db, cfg: cfg}
}

// POST /api/article/create
func (h *ArticleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var art model.Article
	var user, ok = middleware.GetUser(r.Context())
	if !ok {
		api.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&art); err != nil {
		api.Error(w, http.StatusBadRequest, "wrong JSON format")
		return
	}

	if art.Title == "" || art.ProjectID != 0 {
		api.Error(w, http.StatusBadRequest, "missing fields")
		return
	}

	if art.Content == "" {
		art.Content = "Article content"
	}
	if art.Slug == "" {
		art.Slug = strings.ToLower(art.Title)
	}

	createArticleQuery := `
		INSERT INTO Articles(prjID, slug, title, content, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING artID, created_at, updated_at;
	`
	err := h.db.QueryRow(r.Context(), createArticleQuery, art.ProjectID, art.Slug, art.Title, art.Content, user.ID).Scan(
		&art.ID, &art.CreatedAt, &art.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			api.Error(w, http.StatusBadRequest, "an article with this slug already exists")
		}

		log.Printf("INSERT DB ERROR: %v", err)

		api.Error(w, http.StatusInternalServerError, "database insert error")
		return
	}

	//TODO: add article revision logic

	api.JSON(w, http.StatusCreated, map[string]interface{}{
		"message": "article created",
		"article": art,
	})

	return
}
