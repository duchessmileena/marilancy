package models

import "time"

type Application struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Status        string    `json:"status"`
	TanggalDaftar time.Time `gorm:"autoCreateTime" json:"tanggal_daftar"`
	FreelancerID  uint      `gorm:"uniqueIndex:idx_freelancer_job_app" json:"freelancer_id"`
	JobID         uint      `gorm:"uniqueIndex:idx_freelancer_job_app" json:"job_id"`
	Job           Job       `gorm:"foreignKey:JobID;references:ID" json:"job"`

	Freelancer Freelancer `gorm:"foreignKey:FreelancerID;references:ID" json:"freelancer"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
