package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

var db *pg.DB

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

const (
	apiKey          = "secretkey"
	dbHost          = "db"
	dbPort          = "5432"
	dbUser          = "postgres"
	dbPassword      = "postgres"
	dbName          = "postgres"
	processingDelay = 1 * time.Minute
)

var taskChannel = make(chan Task)

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
	go startTaskProcessor()

	// Run the server
	r.Run(":8080")
}

func initDB() {
	db = pg.Connect(&pg.Options{
		Addr:     fmt.Sprintf("%s:%s", dbHost, dbPort),
		User:     dbUser,
		Password: dbPassword,
		Database: dbName,
	})

	// Create tasks table if it doesn't exist
	err := db.Model(&Task{}).CreateTable(&orm.CreateTableOptions{IfNotExists: true})
	if err != nil {
		log.Fatalf("Error creating database table: %v", err)
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
	_, err := db.Model(&task).Insert()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// Send the task to the processing channel
	taskChannel <- task

	c.JSON(http.StatusCreated, task)
}

func getTasks(c *gin.Context) {
	var tasks []Task
	err := db.Model(&tasks).Select()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func getTaskStatus(c *gin.Context) {
	taskID := c.Param("id")
	var task Task
	err := db.Model(&task).Where("id = ?", taskID).Select()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func processTask(task Task) {
	// Simulate task processing delay (5 minutes)
	time.Sleep(processingDelay)
	task.Status = "completed"
	_, err := db.Model(&task).WherePK().Update()
	if err != nil {
		log.Println("Failed to update task status to completed")
	}
	log.Printf("Task %d marked as completed", task.ID)
}

func startTaskProcessor() {
	for task := range taskChannel {
		go processTask(task)
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
