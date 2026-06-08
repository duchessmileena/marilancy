package config

import "os"

func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		secret = "marilancy-secret-dev"
	}

	return []byte(secret)
}
