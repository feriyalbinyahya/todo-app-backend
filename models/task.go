package models

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
	Completed   bool      `json:"completed"`
	UserID      uint      `json:"user_id"` // Menyimpan ID User
	SubTasks    []SubTask `json:"sub_tasks" gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE;"`
	Overdue     bool      `json:"overdue" gorm:"-"`          // Tidak disimpan di database
	Progress    float64   `json:"progress" gorm:"default:0"` // Tidak disimpan di database
}

type SubTask struct {
	gorm.Model
	TaskID    uint   `json:"task_id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}
