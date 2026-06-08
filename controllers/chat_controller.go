package controllers

import (
	"marilancy/config"
	"marilancy/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SendMessage(c *gin.Context) {
	var input struct {
		ProjectID  uint   `json:"project_id" binding:"required"`
		ReceiverID uint   `json:"receiver_id" binding:"required"`
		Content    string `json:"content" binding:"required"`
		SenderRole string `json:"sender_role"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	userIDRaw, _ := c.Get("user_id")
	senderID := userIDRaw.(uint)

	var senderStatus, receiverStatus string

	if input.SenderRole == "client" {

		config.DB.Model(&models.Client{}).Where("id = ?", senderID).Pluck("status", &senderStatus)
		config.DB.Model(&models.Freelancer{}).Where("id = ?", input.ReceiverID).Pluck("status", &receiverStatus)
	} else {

		config.DB.Model(&models.Freelancer{}).Where("id = ?", senderID).Pluck("status", &senderStatus)
		config.DB.Model(&models.Client{}).Where("id = ?", input.ReceiverID).Pluck("status", &receiverStatus)
	}

	if senderStatus == "suspended" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Gagal mengirim pesan. Akun Anda sedang ditangguhkan."})
		return
	}
	if receiverStatus == "suspended" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Percakapan dikunci. Akun lawan bicara Anda telah dinonaktifkan."})
		return
	}

	msg := models.Message{
		ProjectID:  input.ProjectID,
		SenderID:   senderID,
		SenderRole: input.SenderRole,
		ReceiverID: input.ReceiverID,
		Content:    input.Content,
		CreatedAt:  time.Now(),
	}

	if err := config.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim pesan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pesan terkirim", "data": msg})
}

func GetChatHistory(c *gin.Context) {
	projectID := c.Query("project_id")
	targetID := c.Query("target_id")
	userRole := c.Query("role")

	userIDRaw, _ := c.Get("user_id")
	userID := userIDRaw.(uint)

	var messages []models.Message

	if userRole != "" {
		config.DB.Model(&models.Message{}).
			Where("project_id = ? AND receiver_id = ? AND LOWER(sender_role) != LOWER(?) AND is_read = ?", projectID, userID, userRole, false).
			UpdateColumn("is_read", true)
	}

	config.DB.Where("project_id = ? AND ((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?))",
		projectID, userID, targetID, targetID, userID).
		Order("created_at ASC").
		Find(&messages)

	c.JSON(http.StatusOK, messages)
}

func GetChatList(c *gin.Context) {
	userIDRaw, _ := c.Get("user_id")
	userID := userIDRaw.(uint)
	userRole := c.Query("role")

	type ChatSummary struct {
		ProjectID    uint      `json:"project_id"`
		TargetUserID uint      `json:"target_user_id"`
		Content      string    `json:"content"`
		CreatedAt    time.Time `json:"created_at"`
		UnreadCount  int       `json:"unread_count" gorm:"-"`
		TargetName   string    `json:"target_name"`
		TargetPhoto  string    `json:"target_photo"`
	}

	var summaries []ChatSummary

	query := `
	SELECT 
		m1.project_id,

		IF(m1.sender_id = ?, m1.receiver_id, m1.sender_id) as target_user_id,

		m1.content,
		m1.created_at,

		COALESCE(f.nama, c.nama_client) as target_name,
		COALESCE(f.foto_profil, c.foto_profil) as target_photo

	FROM messages m1

	INNER JOIN (
		SELECT 
			project_id,
			IF(sender_id = ?, receiver_id, sender_id) as target_id,
			MAX(created_at) as max_time
		FROM messages
		WHERE sender_id = ? OR receiver_id = ?
		GROUP BY project_id, target_id
	) m2

	ON m1.project_id = m2.project_id
	AND IF(m1.sender_id = ?, m1.receiver_id, m1.sender_id) = m2.target_id
	AND m1.created_at = m2.max_time

	LEFT JOIN freelancers f
	ON f.id = IF(m1.sender_id = ?, m1.receiver_id, m1.sender_id)

	LEFT JOIN clients c
	ON c.id = IF(m1.sender_id = ?, m1.receiver_id, m1.sender_id)

	ORDER BY m1.created_at DESC
	`

	config.DB.Raw(
		query,
		userID,
		userID,
		userID,
		userID,
		userID,
		userID,
		userID,
	).Scan(&summaries)

	for i := range summaries {
		var count int64
		config.DB.Model(&models.Message{}).
			Where("project_id = ? AND receiver_id = ? AND (sender_role IS NULL OR sender_role = '' OR LOWER(sender_role) != LOWER(?)) AND is_read = ?",
				summaries[i].ProjectID, userID, userRole, false).
			Count(&count)

		summaries[i].UnreadCount = int(count)
	}

	c.JSON(http.StatusOK, summaries)
}

func MarkAsRead(c *gin.Context) {
	var input struct {
		ProjectID uint `json:"project_id"`
		SenderID  uint `json:"sender_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	userIDRaw, _ := c.Get("user_id")
	userID := userIDRaw.(uint)

	result := config.DB.Model(&models.Message{}).
		Where("project_id = ? AND sender_id = ? AND receiver_id = ? AND is_read = ?",
			input.ProjectID, input.SenderID, userID, false).
		UpdateColumn("is_read", true)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update status read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pesan ditandai dibaca"})
}

func EditMessage(c *gin.Context) {
	msgID := c.Param("id")
	var input struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	userIDRaw, _ := c.Get("user_id")
	userID := userIDRaw.(uint)

	var msg models.Message
	if err := config.DB.Where("id = ? AND sender_id = ?", msgID, userID).First(&msg).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesan tidak ditemukan atau Anda tidak memiliki akses"})
		return
	}

	config.DB.Model(&msg).Updates(map[string]interface{}{
		"content":    input.Content,
		"updated_at": time.Now(),
	})
	c.JSON(http.StatusOK, gin.H{"message": "Pesan berhasil diedit", "data": msg})
}

func DeleteMessage(c *gin.Context) {
	msgID := c.Param("id")

	userIDRaw, _ := c.Get("user_id")
	userID := userIDRaw.(uint)

	var msg models.Message
	if err := config.DB.Where("id = ? AND sender_id = ?", msgID, userID).First(&msg).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesan tidak ditemukan atau Anda tidak memiliki akses"})
		return
	}

	config.DB.Model(&msg).Update("content", "Pesan ini telah dihapus")
	c.JSON(http.StatusOK, gin.H{"message": "Pesan berhasil dihapus"})
}
