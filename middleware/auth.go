package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"marilancy/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(id uint, role string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	t, err := token.SignedString([]byte(config.JWT_SECRET))
	if err != nil {
		fmt.Println("❌ TOKEN SIGN ERROR:", err)
		return ""
	}

	return t
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("🔥 ===== JWT MIDDLEWARE HIT =====")

		authHeader := c.GetHeader("Authorization")
		fmt.Println("📩 AUTH HEADER:", authHeader)

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(config.JWT_SECRET), nil
		})

		if err != nil || !token.Valid {
			fmt.Println("❌ JWT ERROR:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
			c.Abort()
			return
		}

		var userID uint
		if idFloat, ok := claims["user_id"].(float64); ok {
			userID = uint(idFloat)
		} else if idInt, ok := claims["user_id"].(int); ok {
			userID = uint(idInt)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id"})
			c.Abort()
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid role"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("role", role)

		fmt.Println("✅ AUTH OK:", userID, role)

		c.Next()
	}
}

func RoleMiddleware(expectedRole string) gin.HandlerFunc {
	return func(c *gin.Context) {

		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Role not found"})
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid role type"})
			c.Abort()
			return
		}

		fmt.Println("👤 ROLE FROM TOKEN:", roleStr)
		fmt.Println("🎯 EXPECTED ROLE:", expectedRole)

		if roleStr == "admin" {
			fmt.Println("🟡 ADMIN ACCESS GRANTED")
			c.Next()
			return
		}

		if roleStr != expectedRole {
			fmt.Println("❌ FORBIDDEN ACCESS")
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}
