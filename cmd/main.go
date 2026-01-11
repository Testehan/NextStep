package main

import (
	"log"
	"os"
	"productivity-app/internal/db"
	"productivity-app/internal/handlers"
	"productivity-app/internal/repositories"
	"productivity-app/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("DATABASE_NAME")
	if dbName == "" {
		dbName = "productivity"
	}

	client, err := db.Connect(mongoURI)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(nil)

	database := client.Database(dbName)

	// Repositories
	goalRepo := repositories.NewGoalRepository(database)
	projectRepo := repositories.NewProjectRepository(database)
	actionRepo := repositories.NewActionRepository(database)

	// Service
	service := services.NewProductivityService(goalRepo, projectRepo, actionRepo)

	// Handlers
	handler := handlers.NewProductivityHandler(service)

	// Router
	r := gin.Default()

	// CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://testdan.com", "http://testdan.com", "https://www.testdan.com", "http://www.testdan.com", "http://localhost:5173", "http://localhost:3000", "http://127.0.0.1:5173", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Auth Middleware
	r.Use(handlers.AuthMiddleware())

	r.GET("/dashboard", handler.GetDashboard)
	r.POST("/goals", handler.CreateGoal)
	r.GET("/goals", handler.GetGoals)
	r.PATCH("/goals/:id", handler.UpdateGoal)
	r.DELETE("/goals/:id", handler.DeleteGoal)
	r.POST("/capture", handler.Capture)
	r.GET("/actions", handler.GetActions)
	r.POST("/actions", handler.CreateAction)
	r.PATCH("/actions/:id", handler.UpdateAction)
	r.DELETE("/actions/:id", handler.DeleteAction)
	r.POST("/actions/:id/complete", handler.CompleteAction)
	r.POST("/projects", handler.CreateProject)
	r.GET("/projects", handler.GetProjects)
	r.PATCH("/projects/:id", handler.UpdateProject)
	r.DELETE("/projects/:id", handler.DeleteProject)
	r.POST("/projects/:id/promote", handler.PromoteProject)
	r.GET("/weekly-review", handler.GetWeeklyReview)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := "0.0.0.0:" + port
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
