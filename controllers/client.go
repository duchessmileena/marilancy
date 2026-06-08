package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"marilancy/config"
	"marilancy/models"

	"github.com/gin-gonic/gin"
)

func GetClientProfile(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var client models.Client
	if err := config.DB.First(&client, userID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Client tidak ditemukan"})
		return
	}

	c.JSON(200, client)
}

func UpdateClientProfile(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var client models.Client
	if err := config.DB.First(&client, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client tidak ditemukan"})
		return
	}

	var input struct {
		NamaClient       string `form:"nama_client"`
		Lokasi           string `form:"lokasi"`
		JenisUsaha       string `form:"jenis_usaha"`
		JumlahPegawai    int    `form:"jumlah_pegawai"`
		Kontak           string `form:"kontak"`
		Deskripsi        string `form:"deskripsi"`
		KulturPerusahaan string `form:"kultur_perusahaan"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid: " + err.Error()})
		return
	}

	if input.JumlahPegawai < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jumlah pegawai tidak boleh kurang dari 0!"})
		return
	}

	client.NamaClient = input.NamaClient
	client.Lokasi = input.Lokasi
	client.JenisUsaha = input.JenisUsaha
	client.JumlahPegawai = input.JumlahPegawai
	client.Kontak = input.Kontak
	client.Deskripsi = input.Deskripsi
	client.KulturPerusahaan = input.KulturPerusahaan

	fileFoto, err := c.FormFile("foto_profil")
	if err == nil {
		ext := strings.ToLower(filepath.Ext(fileFoto.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Foto profil wajib berupa file JPG, JPEG, atau PNG!"})
			return
		}

		fileName := fmt.Sprintf("fotoclient_%d_%d%s", userID, time.Now().Unix(), ext)
		savePath := filepath.Join("uploads", fileName)

		if err := c.SaveUploadedFile(fileFoto, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan foto profil"})
			return
		}
		client.FotoProfil = "/" + filepath.ToSlash(savePath)
	}

	file, err := c.FormFile("galeri")
	if err == nil {

		fileName := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		savePath := filepath.Join("uploads", fileName)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan gambar galeri"})
			return
		}

		client.Galeri = "/" + filepath.ToSlash(savePath)
	}

	if err := config.DB.Save(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update profil"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profil berhasil diupdate", "data": client})
}

func GetApplicationsByClient(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var jobs []models.Job
	if err := config.DB.Where("client_id = ?", userID).Find(&jobs).Error; err != nil {
		c.JSON(500, gin.H{"error": "Gagal ambil job"})
		return
	}

	var apps []models.Application

	for _, job := range jobs {
		var jobApps []models.Application
		config.DB.Where("lowongan_id = ?", job.ID).Preload("Freelancer").Find(&jobApps)
		apps = append(apps, jobApps...)
	}

	c.JSON(200, apps)
}

func GetClientByID(c *gin.Context) {
	id := c.Param("id")

	var client models.Client
	if err := config.DB.Where("id = ?", id).First(&client).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client tidak ditemukan"})
		return
	}
	if client.Status == "suspended" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akun klien ini telah dinonaktifkan oleh sistem karena indikasi pelanggaran."})
		return
	}

	c.JSON(http.StatusOK, client)
}
