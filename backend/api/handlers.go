package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func (h *Handler) GetListsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	rows, err := h.DB.Query(`
		SELECT id, title, description, created_at, updated_at
		FROM todo_lists
		WHERE username = $1
		ORDER BY created_at DESC`, username)
	if err != nil {
		http.Error(w, "Failed to fetch todo lists", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var lists []types.TodoList
	for rows.Next() {
		var list types.TodoList
		if err := rows.Scan(&list.ID, &list.Title, &list.Description, &list.CreatedAt, &list.UpdatedAt); err != nil {
			continue
		}
		list.Username = username
		lists = append(lists, list)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lists)
}

func (h *Handler) GetListHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	// Extract list ID from the path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	listID, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	var list types.TodoList
	err = h.DB.QueryRow(`
		SELECT id, title, description, created_at, updated_at
		FROM todo_lists
		WHERE id = $1 AND username = $2`, listID, username).Scan(
		&list.ID, &list.Title, &list.Description, &list.CreatedAt, &list.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo list not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	list.Username = username

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) CreateListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.Context().Value("username").(string)

	var req types.TodoListCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	var list types.TodoList
	err := h.DB.QueryRow(`
		INSERT INTO todo_lists (username, title, description)
		VALUES ($1, $2, $3)
		RETURNING id, title, description, created_at, updated_at`,
		username, req.Title, req.Description).Scan(
		&list.ID, &list.Title, &list.Description, &list.CreatedAt, &list.UpdatedAt)

	if err != nil {
		http.Error(w, "Failed to create todo list", http.StatusInternalServerError)
		return
	}

	list.Username = username

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) UpdateListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.Context().Value("username").(string)

	// Extract list ID from the path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	listID, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	var req types.TodoListUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	now := time.Now()

	var list types.TodoList
	err = h.DB.QueryRow(`
		UPDATE todo_lists
		SET title = $1, description = $2, updated_at = $3
		WHERE id = $4 AND username = $5
		RETURNING id, title, description, created_at, updated_at`,
		req.Title, req.Description, now, listID, username).Scan(
		&list.ID, &list.Title, &list.Description, &list.CreatedAt, &list.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo list not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update todo list", http.StatusInternalServerError)
		}
		return
	}

	list.Username = username

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) DeleteListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.Context().Value("username").(string)

	// Extract list ID from the path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	listID, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	// Check if the list exists and belongs to the user
	var exists bool
	err = h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM todo_lists WHERE id = $1 AND username = $2)",
		listID, username).Scan(&exists)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Todo list not found", http.StatusNotFound)
		return
	}

	// Begin transaction
	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Delete all todos in the list - this should cascade automatically due to foreign key,
	// but we're being explicit here
	_, err = tx.Exec("DELETE FROM todos WHERE list_id = $1", listID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to delete todos", http.StatusInternalServerError)
		return
	}

	// Delete the list itself
	_, err = tx.Exec("DELETE FROM todo_lists WHERE id = $1 AND username = $2", listID, username)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to delete todo list", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetTodosForListHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	// Extract list ID from the path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 || pathParts[len(pathParts)-2] != "todos" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	listID, err := strconv.Atoi(pathParts[len(pathParts)-3])
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	// Check if the list exists and belongs to the user
	var exists bool
	err = h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM todo_lists WHERE id = $1 AND username = $2)",
		listID, username).Scan(&exists)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Todo list not found", http.StatusNotFound)
		return
	}

	rows, err := h.DB.Query(`
		SELECT id, list_id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE list_id = $1 AND username = $2
		ORDER BY created_at DESC`, listID, username)
	if err != nil {
		http.Error(w, "Failed to fetch todos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var todos []types.Todo
	for rows.Next() {
		var todo types.Todo
		if err := rows.Scan(&todo.ID, &todo.ListID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt); err != nil {
			continue
		}
		todos = append(todos, todo)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func (h *Handler) GetTodoHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	// Extract todo ID from the path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	todoID, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	var todo types.Todo
	err = h.DB.QueryRow(`
		SELECT id, list_id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE id = $1 AND username = $2`, todoID, username).Scan(
		&todo.ID, &todo.ListID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func (h *Handler) CreateTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.Context().Value("username").(string)

	// Extract list ID from the path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 || pathParts[len(pathParts)-2] != "todos" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	listID, err := strconv.Atoi(pathParts[len(pathParts)-3])
	if err != nil {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}

	// Check if list exists and belongs to the user
	var exists bool
	err = h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM todo_lists WHERE id = $1 AND username = $2)",
		listID, username).Scan(&exists)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Todo list not found", http.StatusNotFound)
		return
	}

	var req types.TodoCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	var todo types.Todo
	err = h.DB.QueryRow(`
		INSERT INTO todos (list_id, username, title, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, list_id, title, description, completed, created_at, updated_at`,
		listID, username, req.Title, req.Description).Scan(
		&todo.ID, &todo.ListID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)

	if err != nil {
		http.Error(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

func (h *Handler) UpdateTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.Context().Value("username").(string)

	// Extract todo ID from the path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	todoID, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	var req types.TodoUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	now := time.Now()

	var todo types.Todo
	err = h.DB.QueryRow(`
		UPDATE todos
		SET title = $1, description = $2, completed = $3, updated_at = $4
		WHERE id = $5 AND username = $6
		RETURNING id, list_id, title, description, completed, created_at, updated_at`,
		req.Title, req.Description, req.Completed, now, todoID, username).Scan(
		&todo.ID, &todo.ListID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update todo", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func (h *Handler) ToggleTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.Context().Value("username").(string)

	// Extract todo ID from the path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	todoID, err := strconv.Atoi(pathParts[len(pathParts)-2])
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	now := time.Now()

	var todo types.Todo
	err = h.DB.QueryRow(`
		UPDATE todos
		SET completed = NOT completed, updated_at = $1
		WHERE id = $2 AND username = $3
		RETURNING id, list_id, title, description, completed, created_at, updated_at`,
		now, todoID, username).Scan(
		&todo.ID, &todo.ListID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to toggle todo", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func (h *Handler) DeleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.Context().Value("username").(string)

	// Extract todo ID from the path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	todoID, err := strconv.Atoi(pathParts[len(pathParts)-1])
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	// Check if the todo exists and belongs to the user
	var exists bool
	err = h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM todos WHERE id = $1 AND username = $2)",
		todoID, username).Scan(&exists)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	_, err = h.DB.Exec("DELETE FROM todos WHERE id = $1 AND username = $2", todoID, username)
	if err != nil {
		http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
