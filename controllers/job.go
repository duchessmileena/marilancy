package controllers

import (
	"encoding/json"
	"marilancy/config"
	"marilancy/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getUserID(c *gin.Context) (uint, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	switch v := val.(type) {
	case float64:
		return uint(v), true
	case int:
		return uint(v), true
	case uint:
		return v, true
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, false
		}
		return uint(i), true
	default:
		return 0, false
	}
}

func CreateJob(c *gin.Context) {
	var job models.Job

	var input struct {
		Judul             string   `json:"judul"`
		JobNo             string   `json:"job_no"`
		JobDesc           string   `json:"job_desc"`
		KebutuhanProyek   string   `json:"kebutuhan_proyek"`
		KebutuhanSkill    string   `json:"kebutuhan_skill"`
		Status            string   `json:"status"`
		Kategori          string   `json:"kategori"`
		Budget            string   `json:"budget"`
		BatasPendidikan   string   `json:"batas_pendidikan"`
		PengalamanKerja   string   `json:"pengalaman_kerja"`
		Tipe              string   `json:"tipe"`
		LokasiPelaksanaan string   `json:"lokasi_pelaksanaan"`
		Tags              []string `json:"tags"`
		ShareJob          string   `json:"share_job"`
		Level             string   `json:"level"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if input.Judul == "" || input.JobDesc == "" || input.KebutuhanProyek == "" ||
		input.KebutuhanSkill == "" || input.Kategori == "" || input.Budget == "" ||
		input.BatasPendidikan == "" || input.PengalamanKerja == "" ||
		input.Tipe == "" || input.LokasiPelaksanaan == "" || input.Level == "" {
		c.JSON(400, gin.H{"error": "Semua field wajib diisi"})
		return
	}

	if input.Judul != "" {
		for _, char := range input.Judul {
			if !((char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				char == ' ' || char == '-' || char == '&' ||
				char == '(' || char == ')' || char == ',' ||
				char == '.' || char == '/') {
				c.JSON(400, gin.H{"error": "Judul hanya boleh berisi huruf, spasi, dan tanda baca ( - & , . / () ). Angka dan simbol lain tidak diperbolehkan."})
				return
			}
		}
	}
	if input.JobNo != "" {
		for _, char := range input.JobNo {
			if (char < '0' || char > '9') && char != '-' {
				c.JSON(400, gin.H{"error": "Nomor/ID Job hanya boleh berisi angka dan tanda strip (-)."})
				return
			}
		}
	}

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

	if client.NamaClient == "" || client.Kontak == "" || client.Lokasi == "" || client.JenisUsaha == "" {
		c.JSON(400, gin.H{"error": "Profil belum lengkap! Harap lengkapi Profil Anda sebelum mem-posting job."})
		return
	}

	job.Judul = input.Judul
	job.JobNo = input.JobNo
	job.JobDesc = input.JobDesc
	job.KebutuhanProyek = input.KebutuhanProyek
	job.KebutuhanSkill = input.KebutuhanSkill
	job.Status = input.Status
	job.Kategori = input.Kategori
	job.Budget = input.Budget
	job.BatasPendidikan = input.BatasPendidikan
	job.PengalamanKerja = input.PengalamanKerja
	job.Tipe = input.Tipe
	job.LokasiPelaksanaan = input.LokasiPelaksanaan
	job.ShareJob = input.ShareJob
	job.Level = input.Level
	job.ClientID = userID

	tagsJSON, _ := json.Marshal(input.Tags)
	job.Tags = string(tagsJSON)

	if err := config.DB.Create(&job).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Job created"})
}

func GetJobs(c *gin.Context) {
	var jobs []models.Job

	err := config.DB.
		Preload("Client").
		Joins("JOIN clients ON clients.id = jobs.client_id").
		Where("jobs.status != ? AND clients.status != ?", "dihapus", "suspended").
		Find(&jobs).Error

	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal mengambil jobs"})
		return
	}

	for i := range jobs {
		var count int64
		err := config.DB.Model(&models.Application{}).
			Where("job_id = ?", jobs[i].ID).
			Count(&count).Error

		if err != nil {
			count = 0
		}
		jobs[i].ApplicationsCount = count
	}

	var result []gin.H
	for _, job := range jobs {
		var tags []string
		if job.Tags != "" {
			_ = json.Unmarshal([]byte(job.Tags), &tags)
		}

		result = append(result, gin.H{
			"id":                 job.ID,
			"judul":              job.Judul,
			"job_no":             job.JobNo,
			"job_desc":           job.JobDesc,
			"kebutuhan_proyek":   job.KebutuhanProyek,
			"kebutuhan_skill":    job.KebutuhanSkill,
			"status":             job.Status,
			"kategori":           job.Kategori,
			"budget":             job.Budget,
			"batas_pendidikan":   job.BatasPendidikan,
			"pengalaman_kerja":   job.PengalamanKerja,
			"tipe":               job.Tipe,
			"lokasi_pelaksanaan": job.LokasiPelaksanaan,
			"tags":               tags,
			"share_job":          job.ShareJob,
			"level":              job.Level,
			"client_id":          job.ClientID,
			"client": gin.H{
				"id":          job.Client.ID,
				"nama_client": job.Client.NamaClient,
				"foto_profil": job.Client.FotoProfil,
			},
			"applications_count": job.ApplicationsCount,
			"created_at":         job.CreatedAt,
		})
	}

	c.JSON(200, result)
}

func GetJobDetail(c *gin.Context) {
	id := c.Param("id")

	var job models.Job
	if err := config.DB.Preload("Client").First(&job, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Job tidak ditemukan"})
		return
	}

	c.JSON(200, job)
}
func DeleteJob(c *gin.Context) {
	id := c.Param("id")

	userID, ok := getUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var job models.Job
	if err := config.DB.First(&job, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Job tidak ditemukan"})
		return
	}

	if job.ClientID != userID {
		c.JSON(403, gin.H{"error": "Bukan job milikmu"})
		return
	}

	tx := config.DB.Begin()

	if err := tx.Model(&job).Update("status", "dihapus").Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Gagal menghapus job"})
		return
	}

	if err := tx.Model(&models.Application{}).
		Where("job_id = ?", job.ID).
		Update("status", "dihapus").Error; err != nil {

		tx.Rollback()
		c.JSON(500, gin.H{"error": "Gagal update application"})
		return
	}

	if err := tx.Exec("DELETE FROM tasks WHERE project_id IN (SELECT id FROM projects WHERE job_id = ?)", job.ID).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Gagal hapus tasks terkait"})
		return
	}

	if err := tx.Where("job_id = ?", job.ID).
		Delete(&models.Project{}).Error; err != nil {

		tx.Rollback()
		c.JSON(500, gin.H{"error": "Gagal hapus project"})
		return
	}

	tx.Commit()

	c.JSON(200, gin.H{"message": "Job, tasks, project, & application berhasil dihapus"})
}

func UpdateJob(c *gin.Context) {
	id := c.Param("id")

	userID, ok := getUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var job models.Job
	if err := config.DB.First(&job, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Job tidak ditemukan"})
		return
	}

	if job.ClientID != userID {
		c.JSON(403, gin.H{"error": "Bukan job milikmu"})
		return
	}

	var input struct {
		Judul             string `json:"judul"`
		JobNo             string `json:"job_no"`
		JobDesc           string `json:"job_desc"`
		KebutuhanProyek   string `json:"kebutuhan_proyek"`
		KebutuhanSkill    string `json:"kebutuhan_skill"`
		Budget            string `json:"budget"`
		LokasiPelaksanaan string `json:"lokasi_pelaksanaan"`
		Status            string `json:"status"`
		Kategori          string `json:"kategori"`
		Tipe              string `json:"tipe"`
		Level             string `json:"level"`
		BatasPendidikan   string `json:"batas_pendidikan"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if input.Judul != "" {
		for _, char := range input.Judul {
			if !((char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				char == ' ' || char == '-' || char == '&' ||
				char == '(' || char == ')' || char == ',' ||
				char == '.' || char == '/') {
				c.JSON(400, gin.H{"error": "Judul hanya boleh berisi huruf, spasi, dan tanda baca ( - & , . / () ). Angka dan simbol lain tidak diperbolehkan."})
				return
			}
		}
	}

	if input.JobNo != "" {
		for _, char := range input.JobNo {
			if (char < '0' || char > '9') && char != '-' {
				c.JSON(400, gin.H{"error": "Nomor/ID Job hanya boleh berisi angka dan tanda strip (-)."})
				return
			}
		}
	}

	if job.Status == "dihapus" {
		c.JSON(400, gin.H{"error": "Job sudah dihapus"})
		return
	}

	updateData := map[string]interface{}{
		"judul":              input.Judul,
		"job_no":             input.JobNo,
		"job_desc":           input.JobDesc,
		"kebutuhan_proyek":   input.KebutuhanProyek,
		"kebutuhan_skill":    input.KebutuhanSkill,
		"budget":             input.Budget,
		"lokasi_pelaksanaan": input.LokasiPelaksanaan,
		"status":             input.Status,
		"kategori":           input.Kategori,
		"tipe":               input.Tipe,
		"level":              input.Level,
		"batas_pendidikan":   input.BatasPendidikan,
	}

	if err := config.DB.Model(&models.Job{}).
		Where("id = ?", id).
		Updates(updateData).Error; err != nil {

		c.JSON(500, gin.H{"error": "Gagal update job"})
		return
	}

	c.JSON(200, gin.H{"message": "Job updated"})
}
