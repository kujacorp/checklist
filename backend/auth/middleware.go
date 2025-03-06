package auth

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kujacorp/checklist/types"
	"golang.org/x/crypto/bcrypt"
)

type Middleware struct {
    DB     *sql.DB
    JWTKey []byte
}

func NewMiddleware(db *sql.DB, jwtKey []byte) *Middleware {
    return &Middleware{DB: db, JWTKey: jwtKey}
}

func (m *Middleware) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header required", http.StatusUnauthorized)
            return
        }

        tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
        claims := &types.Claims{}

        token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
            return m.JWTKey, nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), "username", claims.Username)
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}

func (m *Middleware) BasicAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        username, password, ok := r.BasicAuth()
        if !ok {
            log.Printf("Auth failed: No basic auth credentials provided (IP: %s)", r.RemoteAddr)
            w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        var storedHash string
        err := m.DB.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&storedHash)
        if err != nil {
            log.Printf("Auth failed: User '%s' not found (IP: %s)", username, r.RemoteAddr)
            w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
            log.Printf("Auth failed: Invalid password for user '%s' (IP: %s)", username, r.RemoteAddr)
            w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    }
}
