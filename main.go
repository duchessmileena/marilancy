package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"marilancy/config"
	"marilancy/models"
	"marilancy/routes"

	"github.com/gin-gonic/gin"
)

func main() {

	config.ConnectDB()

	config.DB.AutoMigrate(
		&models.Freelancer{},
		&models.Client{},
		&models.Admin{},
		&models.Job{},
		&models.Application{},
		&models.Project{},
		&models.Task{},
		&models.Rating{},
		&models.Message{},
		&models.Transaction{},
		&models.SupportTicket{},
	)
	config.SeedAdmin()

	go func() {
		for {
			time.Sleep(1 * time.Hour)

			batasWaktu := time.Now().Add(-48 * time.Hour)

			var txs []models.Transaction
			err := config.DB.Where("status = ? AND created_at <= ?", "pending", batasWaktu).Find(&txs).Error
			if err == nil && len(txs) > 0 {
				for _, tx := range txs {
					config.DB.Model(&tx).Update("status", "success")

					config.DB.Model(&models.Project{}).Where("id = ?", tx.ProjectID).Update("payment_status", "paid")

					fmt.Printf("[AUTO-APPROVE] Transaksi ID %d otomatis sukses karena sudah melewati 2 hari tanpa rejeksi freelancer.\n", tx.ID)
				}
			}
		}
	}()

	r := gin.Default()

	r.Static("/static", "./static")
	r.Static("/uploads", "./uploads")
	r.LoadHTMLGlob("templates/*")

	routes.SetupRoutes(r)

	r.GET("/guest", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard_guest.html", nil)
	})

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/guest")
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})

	r.GET("/freelancer", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard_freelancer.html", nil)
	})

	r.GET("/client", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard_client.html", nil)
	})

	port := os.Getenv("PORT")
	fmt.Println("PORT =", port)
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
