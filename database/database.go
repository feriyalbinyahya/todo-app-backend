package database

import (
	"fmt"
	"log"
	"todo-app/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := "host=localhost user=postgres password=admin123 dbname=todo_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed connect to database:", err)
	}

	fmt.Println("Success connect to database!")
	DB = db
}

func MigrateDatabase() {
	DB.AutoMigrate(&models.Task{}, &models.SubTask{}, &models.User{}, &models.BlacklistedToken{})
	fmt.Println("Migrasi Database Berhasil!")
}
