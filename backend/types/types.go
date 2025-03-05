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
