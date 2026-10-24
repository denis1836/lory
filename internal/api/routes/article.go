package routes

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"lory/internal/api"
	"lory/internal/config"
	"lory/internal/middleware"
	"lory/internal/model"
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

		api.Error(w, http.StatusInternalServerError, "database error")
		return
	}

	hashBytes := sha256.Sum256([]byte(art.Title + art.Content))
	revHash := hex.EncodeToString(hashBytes[:])

	var artRev model.ArticleRevision
	insertFirstRevisionQuery := `
		INSERT INTO Article_Revisions(artID, usrID, rev_num, title, content, content_hash, message)
		VALUES($1, $2, $3, $4, $5, $6, $7)
		RETURNING artRID, created_at;
	`
	err = h.db.QueryRow(r.Context(), insertFirstRevisionQuery, art.ID, user.ID, 1, art.Title, art.Content, revHash, "Create \""+art.Title+"\"").Scan(
		&artRev.ID, &artRev.CreatedAt,
	)

	if err != nil {
		log.Printf("INSERT DB ERROR: %v", err)
		api.Error(w, http.StatusInternalServerError, "database error")
		return
	}

	api.JSON(w, http.StatusCreated, map[string]interface{}{
		"message": "article created",
		"article": art,
		"art_rev": artRev,
	})

	return
}

// GET /api/article/<id>
// shows the current state of the article (latest rev)
func (h *ArticleHandler) GetArticleById(w http.ResponseWriter, r *http.Request) {
	idParam := r.PathValue("id")
	articleID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		api.Error(w, http.StatusBadRequest, "invalid article id format")
		return
	}

	var art model.Article
	art.ID = uint64(articleID)

	getArticleQuery := `
		SELECT slug, title, content, created_by, created_at, updated_at
		FROM Articles
		WHERE artID = $1 AND deleted_at IS NULL;
	`

	err = h.db.QueryRow(r.Context(), getArticleQuery, art.ID).Scan(
		&art.Slug, &art.Title, &art.Content, &art.CreatedBy, &art.CreatedAt, &art.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			api.Error(w, http.StatusNotFound, "article not found")
			return
		}

		log.Printf("SELECT DB ERROR: %v", err)
		api.Error(w, http.StatusInternalServerError, "database error")
		return
	}

	api.JSON(w, http.StatusOK, map[string]interface{}{
		"article": art,
	})
}

// GET /api/article/<id>/history/<args>
// gets a list off revs based on the given arguments
func (h *ArticleHandler) GetArticleRevisionsHistory(w http.ResponseWriter, r *http.Request) {
	return
}

// GET /api/article/<id>/<hash>
// get article by specific rev
func (h *ArticleHandler) GetArticleByIdAndHash(w http.ResponseWriter, r *http.Request) {
	return
}

// POST /api/article/edit/<id>
// edit article based on the latest rev
func (h *ArticleHandler) EditArticleById(w http.ResponseWriter, r *http.Request) {
	return
}

// POST /api/article/edit/<id>/info
// info to display in ui while editing
func (h *ArticleHandler) GetArticleEditInfo(w http.ResponseWriter, r *http.Request) {
	return
}

// GET /api/article/contributors
// users who contributed to the article and their connected stats
func (h *ArticleHandler) GetArticleContributors(w http.ResponseWriter, r *http.Request) {
	return
}

// GET /api/article/<id>/<hash>/diff
// shows the diff with the previous rev
func (h *ArticleHandler) GetArticleRevDiff(w http.ResponseWriter, r *http.Request) {
	return
}
