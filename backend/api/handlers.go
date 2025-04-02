package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/kujacorp/checklist/auth"
	"github.com/kujacorp/checklist/types"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	DB     *sql.DB
	JWTKey []byte
}

func NewHandler(db *sql.DB, jwtKey []byte) *Handler {
	return &Handler{DB: db, JWTKey: jwtKey}
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req types.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var storedHash string
	var user types.User
	err := h.DB.QueryRow(
		"SELECT password_hash, username, created_at FROM users WHERE username = $1",
		req.Username,
	).Scan(&storedHash, &user.Username, &user.CreatedAt)

	if err != nil || bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.Username, h.JWTKey)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.LoginResponse{
		Token: token,
		User:  user,
	})
}

func (h *Handler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	var req types.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var exists bool
	err := h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)",
		req.Username).Scan(&exists)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	var user types.User
	err = h.DB.QueryRow(
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING username, created_at",
		req.Username, string(hashedPassword),
	).Scan(&user.Username, &user.CreatedAt)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateToken(user.Username, h.JWTKey)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.LoginResponse{
		Token: token,
		User:  user,
	})
}

func (h *Handler) VerifyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ViewCountHandler(w http.ResponseWriter, r *http.Request) {
	var count int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM visits").Scan(&count)
	if err != nil {
		http.Error(w, "Failed to get visit count", http.StatusInternalServerError)
		return
	}
	count++
	_, _ = h.DB.Exec("INSERT INTO visits DEFAULT VALUES")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}
