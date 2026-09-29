package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"lory/internal/api"
	"lory/internal/config"
)

type contextKey string

const userContextKey contextKey = "user_context_key"

type UserContext struct {
	ID   int
	Role string
}

func Authenticate(db *pgxpool.Pool, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie("session_payload")
			if err != nil {
				if errors.Is(err, http.ErrNoCookie) {
					api.Error(w, http.StatusUnauthorized, "No session. Log in.")
					return
				}
				api.Error(w, http.StatusBadRequest, "Error reading cookie")
				return
			}

			tokenString := cookie.Value

			secret := cfg.JWTSecret

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

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				api.Error(w, http.StatusUnauthorized, "Invalid token structure")
				return
			}

			sessionIDFloat, ok := claims["sessionId"].(float64)
			if !ok {
				api.Error(w, http.StatusUnauthorized, "Missing sessionId")
				return
			}
			sessionID := int(sessionIDFloat)

			var user UserContext
			query := `
				SELECT us.user_id, u.role
				FROM User_Sessions us
				JOIN Users u ON us.user_id = u.user_id
				WHERE us.session_token = $1 AND us.expires_at > NOW();
			`
			err = db.QueryRow(r.Context(), query, sessionID).Scan(&user.ID, &user.Role)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					api.Error(w, http.StatusUnauthorized, "Session expired or invalid.")
					return
				}
				log.Printf("Auth DB error: %v", err)
				api.Error(w, http.StatusInternalServerError, "Server error")
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, &user)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
