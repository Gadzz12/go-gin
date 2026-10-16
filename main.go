package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

type Todo struct {
	ID     int    `json:"id"`
	Title  string `json:"title" binding:"required"`
	IsDone bool   `json:"is_done"`
}

type Handler struct {
	DB *sql.DB
}

func respondError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application terminated: %v", err)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}

	db, err := connectDatabase(os.Getenv("DATABASE_URL"))
	if err != nil {
		return databaseStartupError(err)
	}
	defer db.Close()

	mux := newMux(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("HTTP server listening on :%s", port)
	return server.ListenAndServe()
}

func connectDatabase(dbURL string) (*sql.DB, error) {
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is empty")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func databaseStartupError(err error) error {
	return fmt.Errorf("database startup error: %w", err)
}

func newMux(db *sql.DB) http.Handler {
	h := &Handler{DB: db}
	r := gin.Default()

	// Endpoint dasar (tanpa prefix /api/v1)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	r.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
	})
	r.GET("/hello/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.JSON(http.StatusOK, gin.H{"message": "Hello, " + name + "!"})
	})
	r.GET("/search", func(c *gin.Context) {
		q := c.DefaultQuery("q", "")
		limit := c.DefaultQuery("limit", "10")
		c.JSON(http.StatusOK, gin.H{
			"query": q,
			"limit": limit,
		})
	})

	// Grouping Rute dengan Prefix /api/v1
	v1 := r.Group("/api/v1")
	{
		v1.GET("/todos", h.List)
		v1.GET("/todos/:id", h.GetByID)
		v1.POST("/todos", h.Create)
		v1.PUT("/todos/:id", h.Update)
		v1.PATCH("/todos/:id/done", h.ToggleDone) // Tantangan PATCH /api/v1/todos/:id/done
		v1.DELETE("/todos/:id", h.Delete)
	}

	return r
}

// GET /api/v1/todos - Mengambil semua todo (mendukung query parameter ?done=true/false)
func (h *Handler) List(c *gin.Context) {
	query := "SELECT id, title, is_done FROM todos"
	args := []any{}

	if done := c.Query("done"); done != "" {
		isDone, err := strconv.ParseBool(done)
		if err != nil {
			respondError(c, http.StatusBadRequest, "Parameter done harus true atau false")
			return
		}
		query += " WHERE is_done = $1"
		args = append(args, isDone)
	}

	rows, err := h.DB.QueryContext(c.Request.Context(), query+" ORDER BY id", args...)
	if err != nil {
		log.Println("List:", err)
		respondError(c, http.StatusInternalServerError, "Gagal mengambil data")
		return
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.IsDone); err != nil {
			log.Println("List scan:", err)
			respondError(c, http.StatusInternalServerError, "Gagal membaca data")
			return
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		log.Println("List rows:", err)
		respondError(c, http.StatusInternalServerError, "Gagal membaca data")
		return
	}

	c.JSON(http.StatusOK, todos)
}

// GET /api/v1/todos/:id - Mengambil detail satu todo berdasarkan ID
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID harus berupa angka")
		return
	}

	query := "SELECT id, title, is_done FROM todos WHERE id = $1"
	var t Todo
	err = h.DB.QueryRowContext(c.Request.Context(), query, id).Scan(&t.ID, &t.Title, &t.IsDone)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "Todo tidak ditemukan")
			return
		}
		log.Println("GetByID:", err)
		respondError(c, http.StatusInternalServerError, "Gagal mengambil data")
		return
	}

	c.JSON(http.StatusOK, t)
}

// POST /api/v1/todos - Menambahkan todo baru
func (h *Handler) Create(c *gin.Context) {
	var newTodo Todo
	if err := c.ShouldBindJSON(&newTodo); err != nil {
		respondError(c, http.StatusBadRequest, "Format JSON tidak valid atau Title kosong")
		return
	}

	query := "INSERT INTO todos (title, is_done) VALUES ($1, $2) RETURNING id, title, is_done"
	err := h.DB.QueryRowContext(c.Request.Context(), query, newTodo.Title, newTodo.IsDone).
		Scan(&newTodo.ID, &newTodo.Title, &newTodo.IsDone)
	if err != nil {
		log.Println("Create:", err)
		respondError(c, http.StatusInternalServerError, "Gagal menyimpan data")
		return
	}

	c.JSON(http.StatusCreated, newTodo)
}

// PUT /api/v1/todos/:id - Memperbarui data todo berdasarkan ID
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID harus berupa angka")
		return
	}

	var updatedData Todo
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		respondError(c, http.StatusBadRequest, "Format JSON tidak valid atau Title kosong")
		return
	}

	query := "UPDATE todos SET title = $1, is_done = $2 WHERE id = $3 RETURNING id, title, is_done"
	var t Todo
	err = h.DB.QueryRowContext(c.Request.Context(), query, updatedData.Title, updatedData.IsDone, id).
		Scan(&t.ID, &t.Title, &t.IsDone)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "Todo tidak ditemukan")
			return
		}
		log.Println("Update:", err)
		respondError(c, http.StatusInternalServerError, "Gagal memperbarui data")
		return
	}

	c.JSON(http.StatusOK, t)
}

// PATCH /api/v1/todos/:id/done - Mengubah status is_done menjadi true
func (h *Handler) ToggleDone(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID harus berupa angka")
		return
	}

	query := "UPDATE todos SET is_done = true WHERE id = $1 RETURNING id, title, is_done"
	var t Todo
	err = h.DB.QueryRowContext(c.Request.Context(), query, id).Scan(&t.ID, &t.Title, &t.IsDone)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(c, http.StatusNotFound, "Todo tidak ditemukan")
			return
		}
		log.Println("ToggleDone:", err)
		respondError(c, http.StatusInternalServerError, "Gagal memperbarui status todo")
		return
	}

	c.JSON(http.StatusOK, t)
}

// DELETE /api/v1/todos/:id - Menghapus todo berdasarkan ID
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID harus berupa angka")
		return
	}

	query := "DELETE FROM todos WHERE id = $1"
	result, err := h.DB.ExecContext(c.Request.Context(), query, id)
	if err != nil {
		log.Println("Delete:", err)
		respondError(c, http.StatusInternalServerError, "Gagal menghapus data")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		respondError(c, http.StatusNotFound, "Todo tidak ditemukan")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo berhasil dihapus"})
}
