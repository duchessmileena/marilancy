package routes

import (
	"marilancy/controllers"
	"marilancy/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {

	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)
	r.POST("/forgot-password", controllers.ForgotPassword)
	r.POST("/reset-password", controllers.ResetPassword)
	r.POST("/api/support/lapor", middleware.AuthMiddleware(), controllers.CreateSupportTicket)
	r.GET("/jobs", controllers.GetJobs)
	r.GET("/jobs/:id", controllers.GetJobDetail)
	r.GET("/clients/:id", controllers.GetClientByID)

	r.GET("/freelancer/:id/rating-summary", controllers.GetFreelancerRatingSummary)

	r.GET("/client/rating", func(c *gin.Context) {
		c.HTML(200, "rating.html", nil)
	})
	r.GET("/freelancer/projects", func(c *gin.Context) {
		c.HTML(200, "my-projects.html", nil)
	})
	r.GET("/freelancer/verifikasi-pembayaran.html", func(c *gin.Context) {
		c.HTML(200, "verifikasi-pembayaran.html", nil)
	})

	r.GET("/client/projects", func(c *gin.Context) {
		c.HTML(200, "my-projects-client.html", nil)
	})
	r.GET("/lihatprofileclient.html", func(c *gin.Context) {
		c.HTML(200, "lihatprofileclient.html", nil)
	})
	r.GET("/admin", func(c *gin.Context) {
		c.HTML(200, "dashboard_admin.html", nil)
	})

	r.GET("/job/detail", func(c *gin.Context) {
		c.HTML(200, "job_detail.html", nil)
	})

	r.GET("/kategori", func(c *gin.Context) {
		c.HTML(200, "kategori.html", nil)
	})

	r.GET("/legal", func(c *gin.Context) {
		c.HTML(200, "legal.html", nil)
	})

	r.GET("/tentang", func(c *gin.Context) {
		c.HTML(200, "tentang.html", nil)
	})

	r.GET("/client/profile/view", func(c *gin.Context) {
		c.HTML(200, "profile.html", nil)
	})

	r.GET("/freelancer/profile/view", func(c *gin.Context) {
		c.HTML(200, "profilefree.html", nil)
	})
	r.GET("/client/applicants", func(c *gin.Context) {
		c.HTML(200, "pendaftar.html", nil)
	})
	r.GET("/client/saved-applicants", func(c *gin.Context) {
		c.HTML(200, "saved-applicants.html", nil)
	})
	r.GET("/freelancer/notification", func(c *gin.Context) {
		c.HTML(200, "notification.html", nil)
	})

	r.GET("/freelancer/workspace", func(c *gin.Context) {
		c.HTML(200, "workspace-freelancer.html", nil)
	})
	r.GET("/client/workspace", func(c *gin.Context) {
		c.HTML(200, "workspace-client.html", nil)
	})

	r.GET("/chat", func(c *gin.Context) {
		c.HTML(200, "chat-list.html", nil)
	})
	r.GET("/chat/room", func(c *gin.Context) {
		c.HTML(200, "chat-room.html", nil)
	})
	r.GET("/client/transactions", func(c *gin.Context) {
		c.HTML(200, "transactions.html", nil)
	})

	apiChat := r.Group("/api/chat")
	apiChat.Use(middleware.AuthMiddleware())
	{
		apiChat.POST("/send", controllers.SendMessage)
		apiChat.GET("/history", controllers.GetChatHistory)
		apiChat.GET("/list", controllers.GetChatList)
		apiChat.POST("/read", controllers.MarkAsRead)
		apiChat.PUT("/:id", controllers.EditMessage)
		apiChat.DELETE("/:id", controllers.DeleteMessage)
	}

	apiProjects := r.Group("/api/projects")
	apiProjects.Use(middleware.AuthMiddleware())
	{
		apiProjects.GET("/:id/rating", controllers.GetProjectRating)
		apiProjects.GET("/:id", controllers.GetProjectDetail)
		apiProjects.POST("/task", controllers.CreateTask)
		apiProjects.PUT("/task/:task_id", controllers.UpdateTaskStatus)
		apiProjects.PUT("/:id/complete", controllers.CompleteProject)
		apiProjects.PUT("/:id/revision", controllers.RequestRevision)
		apiProjects.PUT("/:id/cancel", controllers.CancelProject)
		apiProjects.DELETE("/task/:task_id", controllers.DeleteTask)
		apiProjects.PATCH("/task/:task_id/title", controllers.UpdateTaskTitle)
		apiProjects.PUT("/task/:task_id/priority", controllers.UpdateTaskPriority)
		apiProjects.PUT("/:id/deadline", controllers.UpdateProjectDeadline)
		apiProjects.POST("/:id/pay", controllers.ConfirmPayment)
		apiProjects.GET("/transactions", controllers.GetClientTransactions)
		apiProjects.PUT("/transactions/:id/status", controllers.UpdateTransactionStatus)
	}

	freelancer := r.Group("/freelancer")
	freelancer.Use(
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("freelancer"),
	)
	{
		freelancer.GET("/profile", controllers.GetFreelancerProfile)
		freelancer.PUT("/profile", controllers.UpdateFreelancerProfile)
		freelancer.POST("/apply", controllers.ApplyJob)
		freelancer.GET("/applications", controllers.GetMyApplications)
		freelancer.POST("/withdraw", controllers.WithdrawApplication)
		freelancer.GET("/my-projects", controllers.GetMyProjects)
	}

	client := r.Group("/client")
	client.Use(
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("client"),
	)
	{
		client.GET("/profile", controllers.GetClientProfile)
		client.PUT("/profile", controllers.UpdateClientProfile)
		client.POST("/jobs", controllers.CreateJob)
		client.PUT("/jobs/:id", controllers.UpdateJob)
		client.GET("/jobs/:id/applicants", controllers.GetJobApplicants)
		client.PUT("/application/status", controllers.UpdateApplicationStatus)
		client.DELETE("/jobs/:id", controllers.DeleteJob)
		client.GET("/my-projects", controllers.GetClientProjects)
		client.POST("/rating", controllers.CreateRating)
		client.GET("/check-rating", controllers.CheckRating)
	}

	admin := r.Group("/admin")
	admin.Use(
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("admin"),
	)
	{
		admin.GET("/data", controllers.AdminDashboardData)

		admin.GET("/freelancers", controllers.GetFreelancers)
		admin.GET("/clients", controllers.GetClients)

		admin.PUT("/freelancers/:id/suspend", controllers.SuspendFreelancer)
		admin.PUT("/clients/:id/suspend", controllers.SuspendClient)

		admin.GET("/jobs", controllers.AdminGetJobs)
		admin.DELETE("/jobs/:id", controllers.DeleteJobs)
		admin.PUT("/jobs/:id/restore", controllers.RestoreJobs)
		admin.GET("/transactions", controllers.AdminGetTransactions)
		admin.GET("/support-tickets", controllers.GetSupportTickets)

	}
}
