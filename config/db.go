package config

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {

	dsn := "root:@tcp(127.0.0.1:3306)/?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB ERROR (initial):", err)
	}

	db.Exec("CREATE DATABASE IF NOT EXISTS marilancy")

	dsn2 := "root:@tcp(127.0.0.1:3306)/marilancy?charset=utf8mb4&parseTime=True&loc=Local"

	db2, err := gorm.Open(mysql.Open(dsn2), &gorm.Config{})
	if err != nil {
		log.Fatal("DB ERROR (final):", err)
	}

	DB = db2
}
