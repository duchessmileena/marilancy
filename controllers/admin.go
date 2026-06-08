package controllers

import (
	"marilancy/config"
	"marilancy/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminDashboardData(c *gin.Context) {
	var totalFreelancers, totalClients, totalJobs int64
	config.DB.Model(&models.Freelancer{}).Count(&totalFreelancers)
	config.DB.Model(&models.Client{}).Count(&totalClients)
	config.DB.Model(&models.Job{}).Where("status != ?", "dihapus").Count(&totalJobs)

	c.JSON(http.StatusOK, gin.H{
		"freelancers": totalFreelancers,
		"clients":     totalClients,
		"jobs":        totalJobs,
	})
}

func GetFreelancers(c *gin.Context) {
	var users []models.Freelancer
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal ambil freelancer"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func GetClients(c *gin.Context) {
	var clients []models.Client
	if err := config.DB.Find(&clients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal ambil client"})
		return
	}
	c.JSON(http.StatusOK, clients)
}

func SuspendFreelancer(c *gin.Context) {
	id := c.Param("id")
	var user models.Freelancer

	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Freelancer tidak ditemukan"})
		return
	}

	newStatus := "suspended"
	if user.Status == "suspended" {
		newStatus = "active"
	}

	config.DB.Model(&user).Update("status", newStatus)
	c.JSON(http.StatusOK, gin.H{"msg": "Status akun berhasil diubah menjadi " + newStatus})
}

func SuspendClient(c *gin.Context) {
	id := c.Param("id")
	var user models.Client

	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client tidak ditemukan"})
		return
	}

	newStatus := "suspended"
	if user.Status == "suspended" {
		newStatus = "active"
	}

	config.DB.Model(&user).Update("status", newStatus)
	c.JSON(http.StatusOK, gin.H{"msg": "Status akun berhasil diubah menjadi " + newStatus})
}

func AdminGetJobs(c *gin.Context) {
	var jobs []models.Job
	if err := config.DB.Preload("Client").Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal ambil job"})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func DeleteJobs(c *gin.Context) {
	id := c.Param("id")
	if err := config.DB.Model(&models.Job{}).Where("id = ?", id).Update("status", "dihapus").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyembunyikan job"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Job berhasil disembunyikan"})
}

func RestoreJobs(c *gin.Context) {
	id := c.Param("id")
	if err := config.DB.Model(&models.Job{}).Where("id = ?", id).Update("status", "ditutup").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memulihkan job"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Job berhasil dipulihkan"})
}

func AdminGetTransactions(c *gin.Context) {
	var results []struct {
		ProjectID      uint    `json:"project_id"`
		FreelancerName string  `json:"freelancer_name"`
		ClientName     string  `json:"client_name"`
		Nominal        float64 `json:"nominal"`
		Status         string  `json:"status"`
	}

	err := config.DB.Table("transactions").
		Select("transactions.project_id, freelancers.nama as freelancer_name, clients.nama_client as client_name, transactions.nominal, transactions.status").
		Joins("join jobs on jobs.id = transactions.project_id").
		Joins("join freelancers on freelancers.id = transactions.freelancer_id").
		Joins("join clients on clients.id = transactions.client_id").
		Scan(&results).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data transaksi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}
