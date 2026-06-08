package controllers

import (
	"fmt"
	"marilancy/config"
	"marilancy/models"
	"strings"

	"github.com/gin-gonic/gin"
)

func getUserIDFromContext(c *gin.Context) (uint, bool) {
	idRaw, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := idRaw.(uint)
	if !ok {
		return 0, false
	}

	return id, true
}

func GetFreelancerProfile(c *gin.Context) {
	id, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	fmt.Println("🔥 HIT GET /freelancer/profile, ID:", id)

	var user models.Freelancer
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	c.JSON(200, user)
}

func UpdateFreelancerProfile(c *gin.Context) {
	id, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var user models.Freelancer
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	user.Nama = c.PostForm("nama")
	user.Gender = c.PostForm("gender")
	user.Location = c.PostForm("location")
	user.EducationLevel = c.PostForm("education_level")
	user.JobInterest = c.PostForm("job_interest")
	user.Bio = c.PostForm("bio")
	user.Skill = c.PostForm("skill")
	user.WorkPre = c.PostForm("work_pre")

	user.NamaBank = c.PostForm("nama_bank")
	user.NoRekening = c.PostForm("no_rekening")

	fileFoto, err := c.FormFile("foto_profil")
	if err == nil {
		ext := strings.ToLower(fileFoto.Filename)
		if !strings.HasSuffix(ext, ".jpg") && !strings.HasSuffix(ext, ".jpeg") && !strings.HasSuffix(ext, ".png") {
			c.JSON(400, gin.H{"error": "Foto profil wajib berupa file JPG, JPEG, atau PNG!"})
			return
		}
		path := fmt.Sprintf("uploads/foto_%d_%s", id, fileFoto.Filename)
		c.SaveUploadedFile(fileFoto, path)
		user.FotoProfil = "/" + path
	}

	var age, exp int
	fmt.Sscanf(c.PostForm("age"), "%d", &age)
	fmt.Sscanf(c.PostForm("years_of_experience"), "%d", &exp)

	if age < 0 || exp < 0 {
		c.JSON(400, gin.H{"error": "Umur dan Pengalaman tidak boleh bernilai minus!"})
		return
	}
	user.Age = age
	user.YearsOfExperience = exp

	user.MonthlySalaryExp = c.PostForm("monthly_salary_exp")

	file, err := c.FormFile("resume")
	if err == nil {
		if !strings.HasSuffix(strings.ToLower(file.Filename), ".pdf") {
			c.JSON(400, gin.H{"error": "Resume wajib berupa file PDF!"})
			return
		}
		path := "uploads/resume_" + file.Filename
		c.SaveUploadedFile(file, path)
		user.Resume = "/" + path
	}

	fileCert, err := c.FormFile("certificates")
	if err == nil {
		if !strings.HasSuffix(strings.ToLower(fileCert.Filename), ".pdf") {
			c.JSON(400, gin.H{"error": "Sertifikat wajib berupa file PDF!"})
			return
		}
		path := "uploads/cert_" + fileCert.Filename
		c.SaveUploadedFile(fileCert, path)
		user.Certificates = "/" + path
	}

	fileAttach, err := c.FormFile("attachments")
	if err == nil {
		path := "uploads/attach_" + fileAttach.Filename
		c.SaveUploadedFile(fileAttach, path)
		user.Attachments = "/" + path
	}

	config.DB.Save(&user)

	c.JSON(200, gin.H{"msg": "Profile updated"})
}

func GetMyApplications(c *gin.Context) {

	userID, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var apps []models.Application

	err := config.DB.
		Joins("JOIN jobs ON jobs.id = applications.job_id").
		Preload("Job").
		Preload("Job").Preload("Job.Client").
		Where("freelancer_id = ?", userID).
		Where("applications.status != ?", "dihapus").
		Where("jobs.status != ?", "dihapus").
		Order("tanggal_daftar desc").
		Find(&apps).Error

	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal ambil data"})
		return
	}

	for i := range apps {

		if apps[i].Job.ID == 0 && apps[i].JobID != 0 {
			var job models.Job
			err := config.DB.First(&job, apps[i].JobID).Error
			if err == nil {
				apps[i].Job = job
			}
		}

		if apps[i].Status != "pending" {
			continue
		}

		if apps[i].JobID != 0 && apps[i].Job.ID == 0 {
			apps[i].Status = "ditolak (job tidak tersedia)"
			continue
		}

		if apps[i].Job.Status == "dihapus" {
			apps[i].Status = "ditolak (job dihapus)"
		} else if apps[i].Job.Status == "ditutup" {
			apps[i].Status = "ditolak (job ditutup)"
		}
	}

	c.JSON(200, apps)
}

func GetMyCompletedJobs(c *gin.Context) {
	id, ok := getUserIDFromContext(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var apps []models.Application

	err := config.DB.
		Joins("JOIN jobs ON jobs.id = applications.job_id").
		Preload("Job").
		Where("applications.freelancer_id = ?", id).
		Where("applications.status = ?", "accepted").
		Where("applications.status != ?", "dihapus").
		Where("jobs.status != ?", "dihapus").
		Find(&apps).Error

	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal ambil completed jobs"})
		return
	}

	c.JSON(200, apps)
}
