package database

import (
	"fmt"
	"log"
	"os"
	"todo-app/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)
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
