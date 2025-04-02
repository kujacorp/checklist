package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kujacorp/checklist/admin"
	"github.com/kujacorp/checklist/api"
	"github.com/kujacorp/checklist/auth"
	_ "github.com/lib/pq"
)

var db *sql.DB
var tmpl *template.Template
var jwtKey []byte

func init() {
	var err error
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "host=postgres user=postgres password=postgres dbname=postgres sslmode=disable"
	}
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	jwtKeyEnv := os.Getenv("JWT_SECRET_KEY")
	if jwtKeyEnv == "" {
		log.Fatal("Environment variable JWT_SECRET_KEY is required")
	}
	jwtKey = []byte(jwtKeyEnv)

	tmpl = template.Must(template.ParseFiles("templates/admin.html"))
}

func main() {
	apiHandler := api.NewHandler(db, jwtKey)
	adminHandler := admin.NewHandler(db, tmpl)
	mw := auth.NewMiddleware(db, jwtKey)

	http.HandleFunc("/", mw.AuthMiddleware(apiHandler.ViewCountHandler))
	http.HandleFunc("/login", apiHandler.LoginHandler)
	http.HandleFunc("/signup", apiHandler.SignupHandler)
	http.HandleFunc("/verify", mw.AuthMiddleware(apiHandler.VerifyHandler))

	http.HandleFunc("/lists", mw.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			apiHandler.GetListsHandler(w, r)
		case http.MethodPost:
			apiHandler.CreateListHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/lists/", mw.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/lists/" {
			http.NotFound(w, r)
			return
		}

		// Check if it's a request for todos in a list
		if strings.Contains(r.URL.Path, "/todos") {
			// If it ends with /todos, get all todos for the list
			if strings.HasSuffix(r.URL.Path, "/todos") {
				apiHandler.GetTodosForListHandler(w, r)
				return
			}

			// If it contains /todos/ with something after it, it's for adding a todo
			if strings.Contains(r.URL.Path, "/todos/") && r.Method == http.MethodPost {
				apiHandler.CreateTodoHandler(w, r)
				return
			}
		}

		// Handle specific list operations
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			apiHandler.GetListHandler(w, r)
		case http.MethodPut, http.MethodPatch:
			apiHandler.UpdateListHandler(w, r)
		case http.MethodDelete:
			apiHandler.DeleteListHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/todos/", mw.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/todos/" {
			http.NotFound(w, r)
			return
		}

		// Check if it's a toggle request
		if strings.HasSuffix(r.URL.Path, "/toggle") {
			apiHandler.ToggleTodoHandler(w, r)
			return
		}

		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			apiHandler.GetTodoHandler(w, r)
		case http.MethodPut, http.MethodPatch:
			apiHandler.UpdateTodoHandler(w, r)
		case http.MethodDelete:
			apiHandler.DeleteTodoHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/admin", mw.BasicAuth(adminHandler.AdminHandler))
	http.HandleFunc("/admin/users", mw.BasicAuth(adminHandler.CreateUserHandler))
	http.HandleFunc("/admin/users/delete", mw.BasicAuth(adminHandler.DeleteUserHandler))

	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
