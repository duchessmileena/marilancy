package models

import "time"

type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ProjectID  uint      `json:"project_id"`
	SenderID   uint      `json:"sender_id"`
	SenderRole string    `json:"sender_role"`
	ReceiverID uint      `json:"receiver_id"`
	Content    string    `json:"content"`
	IsRead     bool      `json:"is_read" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
