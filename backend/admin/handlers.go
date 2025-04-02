package admin

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/kujacorp/checklist/types"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func NewHandler(db *sql.DB, tmpl *template.Template) *Handler {
	return &Handler{DB: db, Tmpl: tmpl}
}

func (h *Handler) AdminHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Users       []types.User
		Message     string
		MessageType string
	}{}

	rows, err := h.DB.Query("SELECT username, created_at FROM users ORDER BY created_at")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user types.User
		if err := rows.Scan(&user.Username, &user.CreatedAt); err != nil {
			continue
		}
		data.Users = append(data.Users, user)
	}

	h.Tmpl.Execute(w, data)
}

func (h *Handler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Username and password required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	_, err = h.DB.Exec("INSERT INTO users (username, password_hash) VALUES ($1, $2)",
		username, string(hashedPassword))
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	if username == "admin" {
		http.Error(w, "Cannot delete admin user", http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec("DELETE FROM users WHERE username = $1", username)
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
