package routes

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"lory/internal/api"
	"lory/internal/config"
	"lory/internal/model"
	"lory/internal/utils"
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

	if err := json.NewDecoder(r.Body).Decode(&art); err != nil {
		api.Error(w, http.StatusBadRequest, "wrong JSON format")
		return
	}

	//TODO: db logic for article contribution

	return
}
