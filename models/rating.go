package models

import "time"

type Rating struct {
	ID           uint `gorm:"primaryKey"`
	Nilai        float64
	Komentar     string
	ClientID     uint
	FreelancerID uint
	ProjectID    uint
	CreatedAt    time.Time
}
