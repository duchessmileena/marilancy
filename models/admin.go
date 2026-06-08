package models

import "time"

type Admin struct {
	ID        uint   `gorm:"primaryKey"`
	NamaAdmin string `gorm:"type:varchar(100)"`
	Email     string `gorm:"unique"`
	Password  string
	Role      string
	CreatedAt time.Time
}
