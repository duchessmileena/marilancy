package config

import "os"

func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		secret = "marilancy-dev-secret"
	}

	return []byte(secret)
}
