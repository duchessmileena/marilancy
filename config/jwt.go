package config

import (
	"log"
	"os"
)

var JWT_SECRET []byte

func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET not set")
	}
	JWT_SECRET = []byte(secret)
}
