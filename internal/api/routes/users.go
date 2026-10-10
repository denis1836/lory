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
		RETURNING usrID, name, email, created_at;
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

		log.Printf("REGISTER DB ERROR: %v", err)

		api.Error(w, http.StatusInternalServerError, "database registration error")
		return
	}

	api.JSON(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered!",
		"user":    user,
	})
}

// POST /api/user/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.Error(w, http.StatusBadRequest, "wrong JSON format")
		return
	}

	if req.Email == "" || req.Password == "" {
		api.Error(w, http.StatusBadRequest, "fill all missing fields")
		return
	}

	checkUserPasswordQuery := `SELECT usrID, password_hash FROM Users WHERE email=$1;`
	var lp model.LoginPasswordCheck
	err := h.db.QueryRow(r.Context(), checkUserPasswordQuery, req.Email).Scan(
		&lp.UserID, &lp.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			api.Error(w, http.StatusBadRequest, "user with this email does not exist")
			return
		}

		log.Printf("LOGIN DB ERROR: %v", err)

		api.Error(w, http.StatusInternalServerError, "database login error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(lp.PasswordHash), []byte(req.Password)); err != nil {
		api.Error(w, http.StatusBadRequest, "wrong password")
		return
	}

	sessionToken, err := utils.GenerateSessionToken()
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "failed to generate session token")
		return
	}
	expiresAt := time.Now().Add(16 * 24 * time.Hour)

	var sessionID int
	loginUserQuery := `
		INSERT INTO User_Sessions (usrID, session_token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING usr_sesSID; 
	`
	err = h.db.QueryRow(r.Context(), loginUserQuery, lp.UserID, sessionToken, expiresAt).Scan(&sessionID)
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "failed to store session")
		return
	}

	claims := jwt.MapClaims{
		"sessionToken": sessionToken,
		"exp":          expiresAt.Unix(),
	}
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := tokenObj.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		api.Error(w, http.StatusInternalServerError, "failed to generate JWT")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_payload",
		Value:    tokenString,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	api.JSON(w, http.StatusAccepted, map[string]interface{}{
		"message": "Logged In!",
	})
}
