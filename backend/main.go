package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/kujacorp/checklist/admin"
	"github.com/kujacorp/checklist/api"
	"github.com/kujacorp/checklist/auth"
	_ "github.com/lib/pq"
)

var db *sql.DB
var tmpl *template.Template
var jwtKey = []byte("your-secret-key")

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

	http.HandleFunc("/admin", mw.BasicAuth(adminHandler.AdminHandler))
	http.HandleFunc("/admin/users", mw.BasicAuth(adminHandler.CreateUserHandler))
	http.HandleFunc("/admin/users/delete", mw.BasicAuth(adminHandler.DeleteUserHandler))

	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
