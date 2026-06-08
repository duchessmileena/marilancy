package controllers

import (
	"errors"
	"fmt"
	"marilancy/config"
	"marilancy/models"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var input struct {
		Nama       string `json:"nama"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		Role       string `json:"role"`
		JenisUsaha string `json:"jenis_usaha"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if !strings.HasSuffix(strings.ToLower(input.Email), "@gmail.com") {
		c.JSON(400, gin.H{"error": "Hanya email dengan format @gmail.com yang dapat didaftarkan"})
		return
	}

	if err := validateNameRules(input.Nama); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var countClient, countFreelancer, countAdmin int64

	config.DB.Model(&models.Client{}).Where("email = ?", input.Email).Count(&countClient)
	config.DB.Model(&models.Freelancer{}).Where("email = ?", input.Email).Count(&countFreelancer)
	config.DB.Model(&models.Admin{}).Where("email = ?", input.Email).Count(&countAdmin)

	if countClient > 0 || countFreelancer > 0 || countAdmin > 0 {
		c.JSON(400, gin.H{"error": "Email sudah terdaftar! Silakan gunakan email lain atau login."})
		return
	}

	if err := validatePasswordStrength(input.Password); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if input.Role == "client" && input.JenisUsaha == "" {
		c.JSON(400, gin.H{"error": "Klien wajib mengisi Jenis Usaha"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal hash password"})
		return
	}

	switch input.Role {
	case "freelancer":
		err = config.DB.Create(&models.Freelancer{
			Nama:     input.Nama,
			Email:    input.Email,
			Password: string(hash),
			Role:     "freelancer",
		}).Error

	case "client":
		err = config.DB.Create(&models.Client{
			NamaClient: input.Nama,
			Email:      input.Email,
			Password:   string(hash),
			Role:       "client",
			JenisUsaha: input.JenisUsaha,
		}).Error

	case "admin":
		err = config.DB.Create(&models.Admin{
			NamaAdmin: input.Nama,
			Email:     input.Email,
			Password:  string(hash),
			Role:      "admin",
		}).Error

	default:
		c.JSON(400, gin.H{"error": "Role tidak valid"})
		return
	}

	if err != nil {
		fmt.Println("❌ REGISTER ERROR:", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Register berhasil"})
}

func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if !strings.HasSuffix(strings.ToLower(input.Email), "@gmail.com") {
		c.JSON(400, gin.H{"error": "Gunakan email @gmail.com untuk login"})
		return
	}

	var userID uint
	var email, pass, role, userStatus string

	var client models.Client
	if err := config.DB.Where("email = ?", input.Email).First(&client).Error; err == nil {
		userID = client.ID
		email = client.Email
		pass = client.Password
		userStatus = client.Status
		role = "client"
	}

	if role == "" {
		var freelancer models.Freelancer
		if err := config.DB.Where("email = ?", input.Email).First(&freelancer).Error; err == nil {
			userID = freelancer.ID
			email = freelancer.Email
			pass = freelancer.Password
			userStatus = freelancer.Status
			role = "freelancer"
		}
	}

	if role == "" {
		var admin models.Admin
		if err := config.DB.Where("email = ?", input.Email).First(&admin).Error; err == nil {
			userID = admin.ID
			email = admin.Email
			pass = admin.Password
			role = "admin"
			userStatus = "active"
		}
	}

	if role == "" {
		c.JSON(400, gin.H{"error": "User tidak ditemukan"})
		return
	}
	if userStatus == "suspended" {
		c.JSON(403, gin.H{"error": "Akun ini telah di-suspend karena indikasi pelanggaran. Silakan hubungi admin."})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(pass), []byte(input.Password)); err != nil {
		c.JSON(400, gin.H{"error": "Password salah"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.JWT_SECRET))
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal generate token"})
		return
	}

	c.JSON(200, gin.H{
		"token": tokenString,
		"role":  role,
	})
}

func ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var count int64

	config.DB.Model(&models.Client{}).Where("email = ?", input.Email).Count(&count)
	if count == 0 {
		config.DB.Model(&models.Freelancer{}).Where("email = ?", input.Email).Count(&count)
	}
	if count == 0 {
		config.DB.Model(&models.Admin{}).Where("email = ?", input.Email).Count(&count)
	}

	if count == 0 {
		c.JSON(404, gin.H{"error": "Email tidak ditemukan di sistem"})
		return
	}

	c.JSON(200, gin.H{"message": "Email ditemukan, silakan reset password"})
}

func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("Password minimal 8 karakter")
	}

	hasUpper, _ := regexp.MatchString(`[A-Z]`, password)
	if !hasUpper {
		return errors.New("Password harus mengandung minimal 1 huruf besar")
	}

	hasNumber, _ := regexp.MatchString(`[0-9]`, password)
	if !hasNumber {
		return errors.New("Password harus mengandung minimal 1 angka")
	}

	hasSymbol, _ := regexp.MatchString(`[^a-zA-Z0-9]`, password)
	if !hasSymbol {
		return errors.New("Password harus mengandung minimal 1 simbol/tanda baca")
	}

	return nil
}

func validateNameRules(name string) error {
	if len(name) < 4 {
		return errors.New("Nama minimal terdiri dari 4 karakter")
	}

	isCapitalized, _ := regexp.MatchString(`^[A-Z]`, name)
	if !isCapitalized {
		return errors.New("Nama harus diawali dengan huruf kapital")
	}

	hasLetter, _ := regexp.MatchString(`[a-zA-Z]`, name)
	if !hasLetter {
		return errors.New("Nama tidak boleh hanya angka, harus mengandung huruf")
	}

	isValidChars, _ := regexp.MatchString(`^[a-zA-Z0-9\s.,'-]+$`, name)
	if !isValidChars {
		return errors.New("Nama mengandung karakter khusus yang tidak diizinkan")
	}

	return nil
}

func ResetPassword(c *gin.Context) {
	var input struct {
		Email       string `json:"email"`
		NewPassword string `json:"newPassword"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := validatePasswordStrength(input.NewPassword); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 10)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal memproses password baru"})
		return
	}

	updated := false

	var client models.Client
	if err := config.DB.Where("email = ?", input.Email).First(&client).Error; err == nil {
		client.Password = string(hash)
		config.DB.Save(&client)
		updated = true
	}

	if !updated {
		var freelancer models.Freelancer
		if err := config.DB.Where("email = ?", input.Email).First(&freelancer).Error; err == nil {
			freelancer.Password = string(hash)
			config.DB.Save(&freelancer)
			updated = true
		}
	}

	if !updated {
		var admin models.Admin
		if err := config.DB.Where("email = ?", input.Email).First(&admin).Error; err == nil {
			admin.Password = string(hash)
			config.DB.Save(&admin)
			updated = true
		}
	}

	if !updated {
		c.JSON(404, gin.H{"error": "User tidak ditemukan saat mereset password"})
		return
	}

	c.JSON(200, gin.H{"message": "Password berhasil diubah"})
}
