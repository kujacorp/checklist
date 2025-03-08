package types

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type SignupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TodoList struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TodoListCreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type TodoListUpdateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Todo struct {
	ID          int       `json:"id"`
	ListID      int       `json:"list_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TodoCreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type TodoUpdateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}
