package controllers

import (
	"marilancy/config"
	"marilancy/models"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

func CreateSupportTicket(c *gin.Context) {
	kontakRaw := c.PostForm("kontak")
	kontakClean := CleanPhoneNumber(kontakRaw)

	ticket := models.SupportTicket{
		Nama:     c.PostForm("nama"),
		Email:    c.PostForm("email"),
		Kontak:   kontakClean,
		Kategori: c.PostForm("kategori"),
		Judul:    c.PostForm("judul"),
		Detail:   c.PostForm("detail"),
	}

	file, err := c.FormFile("bukti")
	if err == nil {
		filename := "uploads/" + file.Filename
		c.SaveUploadedFile(file, filename)
		ticket.BuktiPath = filename
	}
	if err := config.DB.Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menyimpan laporan ke database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Laporan berhasil dikirim!"})
}

func GetSupportTickets(c *gin.Context) {
	var tickets []models.SupportTicket

	if err := config.DB.Order("created_at desc").Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data laporan"})
		return
	}

	c.JSON(http.StatusOK, tickets)
}


func CleanPhoneNumber(phone string) string {
	re := regexp.MustCompile(`[^\d]`)
	return re.ReplaceAllString(phone, "")
}

