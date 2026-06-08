package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"marilancy/config"
	"marilancy/models"

	"github.com/gin-gonic/gin"
)

func GetProjectDetail(c *gin.Context) {
	projectID := c.Param("id")
	var project models.Project

	if err := config.DB.Preload("Job").Preload("Client").Preload("Freelancer").Preload("Tasks").Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
		return
	}

	totalTasks := len(project.Tasks)
	completedTasks := 0
	for _, task := range project.Tasks {
		if task.Status == "done" {
			completedTasks++
		}
	}

	progress := 0
	if totalTasks > 0 {
		progress = (completedTasks * 100) / totalTasks
	}

	c.JSON(http.StatusOK, gin.H{
		"project":  project,
		"progress": progress,
	})
}

func CreateTask(c *gin.Context) {
	var input struct {
		ProjectID uint   `json:"project_id"`
		Title     string `json:"title"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := models.Task{
		ProjectID: input.ProjectID,
		Title:     input.Title,
		Status:    "todo",
	}

	if err := config.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task ditambahkan", "task": task})
}

func UpdateTaskStatus(c *gin.Context) {
	taskID := c.Param("task_id")
	var input struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Model(&models.Task{}).Where("id = ?", taskID).Update("status", input.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update task"})
		return
	}

	var task models.Task
	if err := config.DB.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Status task diperbarui"})
		return
	}

	var project models.Project
	if err := config.DB.Preload("Tasks").First(&project, task.ProjectID).Error; err == nil {

		totalTasks := len(project.Tasks)
		completedTasks := 0
		for _, t := range project.Tasks {
			if t.Status == "done" {
				completedTasks++
			}
		}

		newProgress := 0
		if totalTasks > 0 {
			newProgress = (completedTasks * 100) / totalTasks
		}

		newStatus := project.Status
		if newProgress > 0 && newProgress < 100 && project.Status == "active" {
			newStatus = "inprogress"
		} else if newProgress == 0 && project.Status == "inprogress" {
			newStatus = "active"
		}

		config.DB.Model(&models.Project{}).Where("id = ?", project.ID).Updates(map[string]interface{}{
			"progress": newProgress,
			"status":   newStatus,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status diperbarui"})
}

func CompleteProject(c *gin.Context) {
	projectID := c.Param("id")

	var project models.Project
	if err := config.DB.Preload("Tasks").Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
		return
	}

	totalTasks := len(project.Tasks)
	completedTasks := 0
	for _, task := range project.Tasks {
		if task.Status == "done" {
			completedTasks++
		}
	}

	if totalTasks == 0 || completedTasks < totalTasks {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Semua task harus selesai 100% terlebih dahulu!"})
		return
	}

	link := c.PostForm("submission_link")
	file, errFile := c.FormFile("submission_file")

	if link == "" && errFile != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Harap sertakan link atau upload file hasil kerja!"})
		return
	}

	if errFile == nil {
		fileName := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		savePath := filepath.Join("uploads", fileName)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file hasil kerja"})
			return
		}
		project.SubmissionFile = "/" + filepath.ToSlash(savePath)
	}

	project.SubmissionLink = link
	project.Status = "completed"

	if err := config.DB.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyelesaikan proyek"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proyek berhasil diselesaikan"})
}

func RequestRevision(c *gin.Context) {
	projectID := c.Param("id")

	var project models.Project
	if err := config.DB.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
		return
	}

	project.Status = "active"
	project.SubmissionLink = ""
	project.SubmissionFile = ""

	if err := config.DB.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal meminta pengiriman ulang"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proyek dibuka kembali untuk revisi"})
}

func ConfirmPayment(c *gin.Context) {
	projectID := c.Param("id")

	var project models.Project
	if err := config.DB.Preload("Job").Preload("Transactions").Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
		return
	}

	if len(project.Transactions) > 0 {
		lastTx := project.Transactions[len(project.Transactions)-1]

		if lastTx.Status == "success" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Project ini sudah lunas, tidak perlu membayar lagi."})
			return
		}
		if lastTx.Status == "pending" || lastTx.Status == "process" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Pembayaran sebelumnya sedang diproses. Harap tunggu konfirmasi."})
			return
		}
	}

	nominalStr := c.PostForm("nominal")
	nominal, errParse := strconv.ParseFloat(nominalStr, 64)
	if errParse != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nominal pembayaran tidak valid!"})
		return
	}

	budgetStr := project.Job.Budget
	parts := strings.Split(budgetStr, "-")
	minBudgetStr := strings.TrimSpace(parts[0])

	cleanBudgetStr := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, minBudgetStr)

	minBudget, _ := strconv.ParseFloat(cleanBudgetStr, 64)

	if nominal < minBudget {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Pembayaran minimal adalah Rp %.0f", minBudget),
		})
		return
	}

	file, errFile := c.FormFile("bukti_transfer")
	if errFile != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Harap unggah bukti transfer!",
		})
		return
	}

	fileName := fmt.Sprintf("pay_%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
	savePath := filepath.Join("uploads", fileName)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal simpan file",
		})
		return
	}

	fileURL := "/" + filepath.ToSlash(savePath)

	newTransaction := models.Transaction{
		ProjectID:     project.ID,
		ClientID:      project.ClientID,
		FreelancerID:  project.FreelancerID,
		Nominal:       nominal,
		BuktiTransfer: fileURL,
		Status:        "pending",
	}

	if err := config.DB.Create(&newTransaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal simpan transaksi",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pembayaran berhasil dikirim ulang dan sedang diproses.",
	})
}

func GetMyProjects(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var projects []models.Project
	if err := config.DB.Preload("Job").Preload("Client").Preload("Tasks").Preload("Transactions").Where("freelancer_id = ?", userID).Order("created_at desc").Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data proyek"})
		return
	}

	var result []gin.H
	for _, p := range projects {
		totalTasks := len(p.Tasks)
		completedTasks := 0
		for _, t := range p.Tasks {
			if t.Status == "done" {
				completedTasks++
			}
		}

		progress := 0
		if totalTasks > 0 {
			progress = (completedTasks * 100) / totalTasks
		}

		statusPembayaran := "belum dibayar"
		if len(p.Transactions) > 0 {
			var latestTx models.Transaction
			for _, tx := range p.Transactions {
				if tx.ID > latestTx.ID {
					latestTx = tx
				}
			}

			switch latestTx.Status {
			case "pending":
				statusPembayaran = "sedang diproses"
			case "process":
				statusPembayaran = "sedang diproses"
			case "success":
				statusPembayaran = "sudah bayar"
			case "rejected":
				statusPembayaran = "ditolak"
			default:
				statusPembayaran = "belum dibayar"
			}
		}

		result = append(result, gin.H{
			"id":                p.ID,
			"status":            p.Status,
			"job":               p.Job,
			"client":            p.Client,
			"progress":          progress,
			"tasks":             p.Tasks,
			"start_date":        p.StartDate,
			"end_date":          p.EndDate,
			"status_pembayaran": statusPembayaran,
			"transactions":      p.Transactions,
		})
	}

	c.JSON(http.StatusOK, result)
}

func GetClientTransactions(c *gin.Context) {
	userID, _ := getUserID(c)
	var txs []models.Transaction
	config.DB.Preload("Project.Job").Preload("Project.Freelancer").Where("client_id = ?", userID).Find(&txs)
	c.JSON(200, txs)
}

func UpdateTransactionStatus(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Status string `json:"status"`
	}
	c.ShouldBindJSON(&input)
	config.DB.Model(&models.Transaction{}).Where("id = ?", id).Update("status", input.Status)
	c.JSON(200, gin.H{"message": "Status updated"})
}

func GetClientProjects(c *gin.Context) {
	userID, _ := getUserID(c)
	var projects []models.Project

	if err := config.DB.Preload("Job").Preload("Freelancer").Preload("Tasks").Preload("Transactions").Where("client_id = ?", userID).Find(&projects).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal ambil data"})
		return
	}

	var result []gin.H
	for _, p := range projects {
		totalTasks := len(p.Tasks)
		completedTasks := 0
		for _, t := range p.Tasks {
			if t.Status == "done" {
				completedTasks++
			}
		}

		progress := 0
		if totalTasks > 0 {
			progress = (completedTasks * 100) / totalTasks
		}

		statusPembayaran := "belum dibayar"
		if len(p.Transactions) > 0 {
			lastTx := p.Transactions[len(p.Transactions)-1]
			statusPembayaran = lastTx.Status
		}

		result = append(result, gin.H{
			"id":                p.ID,
			"status":            p.Status,
			"job":               p.Job,
			"freelancer":        p.Freelancer,
			"progress":          progress,
			"start_date":        p.StartDate,
			"end_date":          p.EndDate,
			"status_pembayaran": statusPembayaran,
			"transactions":      p.Transactions,
		})
	}

	c.JSON(http.StatusOK, result)
}

func DeleteTask(c *gin.Context) {
	taskID := c.Param("task_id")

	if err := config.DB.Delete(&models.Task{}, taskID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task berhasil dihapus"})
}

func UpdateTaskTitle(c *gin.Context) {
	taskID := c.Param("task_id")
	var input struct {
		Title string `json:"title"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}

	if err := config.DB.Model(&models.Task{}).Where("id = ?", taskID).Update("title", input.Title).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate judul task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Judul task diperbarui"})
}
func UpdateTaskPriority(c *gin.Context) {
	taskID := c.Param("task_id")
	var input struct {
		Priority string `json:"priority"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	if err := config.DB.Model(&models.Task{}).Where("id = ?", taskID).Update("priority", input.Priority).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal update prioritas"})
		return
	}

	c.JSON(200, gin.H{"message": "Prioritas diperbarui"})
}

func UpdateProjectDeadline(c *gin.Context) {
	projectID := c.Param("id")
	var input struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid"})
		return
	}

	start, _ := time.Parse("2006-01-02", input.StartDate)
	end, _ := time.Parse("2006-01-02", input.EndDate)

	if err := config.DB.Model(&models.Project{}).Where("id = ?", projectID).Updates(map[string]interface{}{
		"start_date": start,
		"end_date":   end,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui deadline"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deadline proyek berhasil diatur"})
}

func CancelProject(c *gin.Context) {
	projectID := c.Param("id")

	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var project models.Project
	if err := config.DB.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project tidak ditemukan"})
		return
	}

	if project.ClientID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki hak untuk membatalkan proyek ini"})
		return
	}

	if project.Status == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proyek sudah selesai, tidak dapat dibatalkan"})
		return
	}
	if project.Status == "dibatalkan" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proyek ini sudah dibatalkan sebelumnya"})
		return
	}

	if project.EndDate == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Batas waktu (due date) proyek belum diatur."})
		return
	}

	if time.Now().Before(*project.EndDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Proyek belum melewati batas waktu. Batas waktu pengerjaan proyek adalah hingga %s.", (*project.EndDate).Format("02-01-2006")),
		})
		return
	}

	project.Status = "dibatalkan"
	if err := config.DB.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui status pembatalan proyek"})
		return
	}

	if project.JobID != 0 {
		if err := config.DB.Model(&models.Job{}).Where("id = ?", project.JobID).Update("status", "dibuka").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Proyek dibatalkan, tetapi gagal membuka kembali lowongan kerja"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Proyek berhasil dibatalkan karena melewati batas waktu pengerjaan. Status lowongan telah dibuka kembali.",
		"data":    project,
	})
}
