package models

import (
	"time"
)

type SupportTicket struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Nama      string    `gorm:"size:255;not null" json:"nama"`
	Email     string    `gorm:"size:255;not null" json:"email"`
	Kontak    string    `gorm:"size:50" json:"kontak"`
	Kategori  string    `gorm:"size:100;not null" json:"kategori"`
	Judul     string    `gorm:"size:255;not null" json:"judul"`
	Detail    string    `gorm:"type:text;not null" json:"detail"`
	BuktiPath string    `gorm:"size:255" json:"bukti_path"`
	Status    string    `gorm:"size:50;default:'Pending'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
