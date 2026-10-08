package routes

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"lory/internal/api"
	"lory/internal/config"
	"lory/internal/model"
)

type UserHandler struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewUserHandler(db *pgxpool.Pool, cfg *config.Config) *UserHandler {
	return &UserHandler{db: db, cfg: cfg}
}

// POST /api/user/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, http.StatusBadRequest, "wrong JSON format")
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		api.Error(w, http.StatusBadRequest, "fill all missing fields")
		return
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "server hashing error")
		return
	}

	var user model.User
	//TODO: email verification (smtp)
	createUserQuery := `
		INSERT INTO Users(name, email, password_hash, created_at, last_online, is_verified)
		VALUES ($1, $2, $3, NOW(), NOW(), TRUE)
		RETURNING user_id, name, email, created_at;
	`
	err = h.db.QueryRow(r.Context(), createUserQuery, req.Name, req.Email, string(hashedBytes)).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			api.Error(w, http.StatusBadRequest, "this email is already used")
			return
		}
		api.Error(w, http.StatusInternalServerError, "database registration error")
		return
	}

	api.JSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered!",
		"user":    user,
	})
}
