package controllers

import (
	"net/http"
	"time"
	"todo-app/database"
	"todo-app/models"

	"github.com/gin-gonic/gin"
)

// GET ALL TASKS WITH FILTER
func GetTasks(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var tasks []models.Task
	database.DB.Where("user_id = ?", userID).Preload("SubTasks").Find(&tasks)

	// Hitung progress & overdue
	currentTime := time.Now()
	ongoingTasks := make([]models.Task, 0) // Pastikan ini tidak nil
	completedTasks := make([]models.Task, 0)
	overdueTasks := make([]models.Task, 0)

	for i := range tasks {
		// Cek apakah Task Overdue
		tasks[i].Overdue = !tasks[i].Completed && currentTime.After(tasks[i].Deadline)

		// Hitung Progress Task berdasarkan SubTask
		totalSubTasks := len(tasks[i].SubTasks)
		completedCount := 0
		for _, sub := range tasks[i].SubTasks {
			if sub.Completed {
				completedCount++
			}
		}
		if totalSubTasks > 0 {
			tasks[i].Progress = (float64(completedCount) / float64(totalSubTasks)) * 100
		} else {
			tasks[i].Progress = 100 // Jika tidak ada sub-task, default 100%
		}

		// Kategorisasi Task
		if tasks[i].Completed {
			completedTasks = append(completedTasks, tasks[i])
		} else if tasks[i].Overdue {
			overdueTasks = append(overdueTasks, tasks[i])
		} else {
			ongoingTasks = append(ongoingTasks, tasks[i])
		}
	}

	// Ambil query filter dari URL
	filter := c.Query("filter") // Bisa "ongoing", "completed", "overdue"

	var result []models.Task
	var message string

	switch filter {
	case "ongoing":
		result = ongoingTasks
		message = "List of ongoing tasks"
	case "completed":
		result = completedTasks
		message = "List of completed tasks"
	case "overdue":
		result = overdueTasks
		message = "List of overdue tasks"
	default:
		result = tasks
		message = "List of all tasks"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": message,
		"tasks":   result,
	})
}

// GET Single Task by ID
func GetTask(c *gin.Context) {
	var task models.Task
	id := c.Param("id")

	if err := database.DB.Preload("SubTasks").First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Hitung apakah task overdue
	currentTime := time.Now()
	task.Overdue = !task.Completed && currentTime.After(task.Deadline)

	// Hitung progress berdasarkan jumlah subtask yang selesai
	totalSubTasks := len(task.SubTasks)
	completedCount := 0
	for _, sub := range task.SubTasks {
		if sub.Completed {
			completedCount++
		}
	}
	if totalSubTasks > 0 {
		task.Progress = (float64(completedCount) / float64(totalSubTasks)) * 100
	} else {
		// Jika tidak ada sub-task, progress langsung 100% jika task completed
		if task.Completed {
			task.Progress = 100
		} else {
			task.Progress = 0
		}
	}

	c.JSON(http.StatusOK, task)
}

// CREATE Task with SubTasks
func CreateTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.UserID = userID.(uint)
	task.Progress = 0
	// Save Task with SubTasks
	database.DB.Create(&task)
	c.JSON(http.StatusCreated, task)
}

// UPDATE Task
func UpdateTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var task models.Task
	id := c.Param("id")

	// Cek apakah task ada dan dimiliki oleh user yang login
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).
		Preload("SubTasks").First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found or unauthorized"})
		return
	}

	// Bind data baru ke task
	var updatedTask struct {
		Title       string           `json:"title"`
		Description string           `json:"description"`
		Deadline    time.Time        `json:"deadline"`
		SubTasks    []models.SubTask `json:"sub_tasks"`
	}

	if err := c.ShouldBindJSON(&updatedTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Perbarui field Title, Description, dan Deadline
	task.Title = updatedTask.Title
	task.Description = updatedTask.Description
	task.Deadline = updatedTask.Deadline

	// **Mengelola SubTasks**
	var newSubTaskIDs = make(map[uint]bool)
	for _, subTask := range updatedTask.SubTasks {
		if subTask.TaskID == 0 { // TaskID harus diisi dengan ID task ini
			subTask.TaskID = task.ID
		}
		newSubTaskIDs[subTask.ID] = true
	}

	var updatedSubTasks []models.SubTask
	for _, subTask := range task.SubTasks {
		if newSubTaskIDs[subTask.ID] { // Subtask masih ada di data baru
			updatedSubTasks = append(updatedSubTasks, subTask)
		} else {
			// Hapus SubTask yang tidak ada di request
			database.DB.Delete(&subTask)
		}
	}

	// Tambahkan SubTasks baru yang belum ada
	for _, subTask := range updatedTask.SubTasks {
		if subTask.ID == 0 {
			updatedSubTasks = append(updatedSubTasks, subTask)
		}
	}

	// Simpan perubahan SubTasks
	task.SubTasks = updatedSubTasks
	database.DB.Save(&task)

	// **Hitung Progress Secara Otomatis**
	var progress float64
	if len(task.SubTasks) > 0 {
		completedSubTasks := 0
		for _, subTask := range task.SubTasks {
			if subTask.Completed {
				completedSubTasks++
			}
		}
		progress = (float64(completedSubTasks) / float64(len(task.SubTasks))) * 100
	} else {
		progress = 0
	}

	task.Progress = progress

	// Simpan perubahan ke database
	database.DB.Save(&task)

	// Return response dengan nilai progress baru
	c.JSON(http.StatusOK, task)
}

// DELETE Task and its SubTasks
func DeleteTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var task models.Task
	id := c.Param("id")

	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).
		Preload("SubTasks").First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found or unauthorized"})
		return
	}

	// Delete Task (SubTasks akan ikut terhapus karena CASCADE)
	database.DB.Delete(&task)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Task deleted successfully",
	})
}

func ChecklistTask(c *gin.Context) {
	var task models.Task
	id := c.Param("id")

	if err := database.DB.Preload("SubTasks").First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Toggle status completed
	task.Completed = !task.Completed

	// Jika task di-checklist, tandai semua sub-task sebagai selesai
	if task.Completed {
		task.Progress = 100
		for i := range task.SubTasks {
			task.SubTasks[i].Completed = true
		}
	} else {
		task.Progress = 0
		// Jika task batal checklist, semua subtask harus kembali tidak selesai
		for i := range task.SubTasks {
			task.SubTasks[i].Completed = false
		}
	}

	// Simpan perubahan
	database.DB.Save(&task)
	database.DB.Save(&task.SubTasks)

	c.JSON(http.StatusOK, task)
}

func ChecklistSubTask(c *gin.Context) {
	var subTask models.SubTask
	id := c.Param("id")

	if err := database.DB.First(&subTask, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SubTask not found"})
		return
	}

	// Toggle status completed subtask
	subTask.Completed = !subTask.Completed
	database.DB.Save(&subTask)

	// Update status task utama
	var task models.Task
	if err := database.DB.Preload("SubTasks").First(&task, subTask.TaskID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load task"})
		return
	}

	// Cek apakah semua subtasks sudah selesai
	allCompleted := true
	completedCount := 0
	for _, st := range task.SubTasks {
		if st.Completed {
			completedCount++
		} else {
			allCompleted = false
		}
	}

	// Jika semua sub-task selesai, task utama otomatis selesai
	task.Completed = allCompleted

	// Hitung progress (persentase sub-task yang selesai)
	totalSubTasks := len(task.SubTasks)
	if totalSubTasks > 0 {
		task.Progress = (float64(completedCount) / float64(totalSubTasks)) * 100
	} else {
		// Jika tidak ada sub-task, progress langsung 100% jika completed
		task.Progress = 100
	}

	// Simpan perubahan
	database.DB.Save(&task)

	c.JSON(http.StatusOK, gin.H{
		"task":    task,
		"subtask": subTask,
	})
}

// CREATE SubTask for a Task
func CreateSubTask(c *gin.Context) {
	var subTask models.SubTask
	if err := c.ShouldBindJSON(&subTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Create(&subTask)
	c.JSON(http.StatusCreated, subTask)
}

// DELETE SubTask
func DeleteSubTask(c *gin.Context) {
	var subTask models.SubTask
	id := c.Param("id")

	if err := database.DB.First(&subTask, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SubTask not found"})
		return
	}

	database.DB.Delete(&subTask)
	c.JSON(http.StatusOK, gin.H{"message": "SubTask deleted successfully"})
}
