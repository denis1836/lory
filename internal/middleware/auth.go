package middleware

import (
	"errors"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"lory/internal/api"
)

//TODO: veryfing token and role in middleware

type contextKey string

const userContextKey contextKey = "user_context_key"

type UserContext struct {
	ID   int
	Role string
}

func Authenticate(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookies("session_payload")
			if err != nil {
				if errors.Is(err, http.ErrNoCookie) {
					api.Error(w, http.StatusUnauthorized, "No session. Log in.")
					return
				}
				api.Error(w, http.StatusBadRequest, "Error reading cookie")
			}

			tokenString := cookie.Value

			//TODO: fetch jwt secret from config object
			secret := os.Getenv("JWT_SECRET")

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				if errors.Is(err, jwt.ErrTokenExpired) {
					api.Error(w, http.StatusUnauthorized, "Session expired.")
					return
				}
				api.Error(w, http.StatusUnauthorized, "Invalid token.")
				return
			}

		})
	}
}
