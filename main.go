package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var db *sql.DB

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

const apiKey = "secretkey"

func main() {
	// Initialize the database
	initDB()

	// Initialize the Gin router
	r := gin.Default()

	// Middleware for API key verification
	r.Use(authMiddleware)

	r.POST("/tasks", createTask)
	r.GET("/tasks", getTasks)
	r.GET("/tasks/:id/status", getTaskStatus)

	// Start the background task processing goroutine
	go processTasks()

	// Run the server
	r.Run(":8080")
}

func initDB() {
	connStr := "host=host.docker.internal user=postgres dbname=postgres password=postgres sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Create tasks table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id SERIAL PRIMARY KEY,
			title TEXT,
			description TEXT,
			status TEXT,
			created_at TIMESTAMPTZ
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func createTask(c *gin.Context) {
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.Status = "incomplete"
	task.CreatedAt = time.Now()

	// Insert the task into the database
	_, err := db.Exec("INSERT INTO tasks (title, description, status, created_at) VALUES ($1, $2, $3, $4)",
		task.Title, task.Description, task.Status, task.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func getTasks(c *gin.Context) {
	var tasks []Task
	rows, err := db.Query("SELECT * FROM tasks")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt)
		if err != nil {
			log.Println(err)
			continue
		}
		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, tasks)
}

func getTaskStatus(c *gin.Context) {
	id := c.Param("id")
	var status string
	err := db.QueryRow("SELECT status FROM tasks WHERE id = $1", id).Scan(&status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

func processTasks() {
	for {
		// Simulate task processing delay (5 minutes)
		time.Sleep(5 * time.Minute)
		log.Println("checking for incomplete tasks")
		_, err := db.Exec("UPDATE tasks SET status = 'completed' WHERE status = 'incomplete' AND created_at <= $1", time.Now().Add(-5*time.Minute))
		if err != nil {
			log.Println("Failed to update task status to completed")
		}
	}
}

func authMiddleware(c *gin.Context) {
	apiKeyHeader := c.GetHeader("API-Key")
	if apiKeyHeader != apiKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		c.Abort()
		return
	}
	c.Next()
}
