package main

import (
	"fmt"
	"log"
	"todo-app/controllers"
	"todo-app/database"
	"todo-app/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Init database
	database.ConnectDatabase()
	database.MigrateDatabase()

	r := gin.Default()

	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	// Middleware untuk endpoint yang perlu autentikasi
	authRoutes := r.Group("/")
	authRoutes.Use(middleware.AuthMiddleware())
	{
		// Routes
		authRoutes.GET("/tasks", controllers.GetTasks)
		authRoutes.GET("/tasks/:id", controllers.GetTask)
		authRoutes.POST("/tasks", controllers.CreateTask)
		authRoutes.PUT("/tasks/:id", controllers.UpdateTask)
		authRoutes.DELETE("/tasks/:id", controllers.DeleteTask)

		authRoutes.PUT("/tasks/:id/checklist", controllers.ChecklistTask)

		// Endpoint untuk SubTask
		authRoutes.POST("/subtasks", controllers.CreateSubTask)
		authRoutes.DELETE("/subtasks/:id", controllers.DeleteSubTask)
		authRoutes.PUT("/subtasks/:id/checklist", controllers.ChecklistSubTask)
	}

	fmt.Println("Server berjalan di port 8080...")
	r.Run(":8080")
}
