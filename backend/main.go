package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var tmpl *template.Template

type User struct {
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User User `json:"user"`
}

func init() {
	var err error
	dsn := fmt.Sprintf("host=postgres user=postgres password=postgres dbname=postgres sslmode=disable")
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	tmpl = template.Must(template.ParseFiles("templates/admin.html"))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var storedHash string
	var user User
	err := db.QueryRow(
		"SELECT password_hash, username, created_at FROM users WHERE username = $1",
		req.Username,
	).Scan(&storedHash, &user.Username, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{User: user})
}

func viewCountHandler(w http.ResponseWriter, r *http.Request) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM visits").Scan(&count)
	if err != nil {
		http.Error(w, "Failed to get visit count", http.StatusInternalServerError)
		return
	}
	count++
	_, _ = db.Exec("INSERT INTO visits DEFAULT VALUES")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			log.Printf("Auth failed: No basic auth credentials provided (IP: %s)", r.RemoteAddr)
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var storedHash string

		err := db.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&storedHash)
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

func adminHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Users       []User
		Message     string
		MessageType string
	}{}

	rows, err := db.Query("SELECT username, created_at FROM users ORDER BY created_at")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Username, &user.CreatedAt); err != nil {
			continue
		}
		data.Users = append(data.Users, user)
	}

	tmpl.Execute(w, data)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
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

	_, err = db.Exec("INSERT INTO users (username, password_hash) VALUES ($1, $2)", username, string(hashedPassword))
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	if username == "admin" {
		http.Error(w, "Cannot delete admin user", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM users WHERE username = $1", username)
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func main() {
	http.HandleFunc("/", viewCountHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/admin", basicAuth(adminHandler))
	http.HandleFunc("/admin/users", basicAuth(createUserHandler))
	http.HandleFunc("/admin/users/delete", basicAuth(deleteUserHandler))

	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
