// routes/routes.go
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mfuadfakhruzzaki/backendaurauran/controllers"
	"github.com/mfuadfakhruzzaki/backendaurauran/middlewares"
	"github.com/mfuadfakhruzzaki/backendaurauran/storage"
	"gorm.io/gorm"
)

// SetupRouter sets up all routes for the application
func SetupRouter(db *gorm.DB, storageService storage.StorageService, bucketName string) *gin.Engine {
	router := gin.Default()

	// Initialize controllers with dependencies
	fileController := controllers.NewFileController(db, storageService, bucketName)
	notificationController := controllers.NewNotificationController(db)
	// Inisialisasi controller lain jika diperlukan, misalnya:
	// userController := controllers.NewUserController(db)
	// teamController := controllers.NewTeamController(db)
	// projectController := controllers.NewProjectController(db)
	// ... dan seterusnya

	// Global middleware
	router.Use(middlewares.CORSMiddleware())
	router.Use(middlewares.LoggingMiddleware())
	router.Use(middlewares.RecoveryMiddleware())
	router.Use(middlewares.RateLimitMiddleware())

	// Auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
		auth.POST("/logout", controllers.Logout)
		auth.GET("/verify-email", controllers.VerifyEmail)
		auth.POST("/request-password-reset", controllers.RequestPasswordReset)
		auth.POST("/reset-password", controllers.ResetPassword)
		auth.GET("/reset-password", controllers.ResetPasswordForm)
		auth.POST("/reset-password-api", controllers.ResetPasswordAPI)
	}

	// Protected routes (requires authentication)
	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware())
	{
		// User routes
		user := protected.Group("/users")
		{
			user.GET("/profile", controllers.GetProfile)
			user.PUT("/profile", controllers.UpdateProfile)
			user.DELETE("/profile", controllers.DeleteProfile)
		}

		// Team routes
		team := protected.Group("/teams")
		{
			team.POST("/", controllers.CreateTeam)
			team.GET("/", controllers.ListTeams)
			team.GET("/:team_id", controllers.GetTeam)
			team.PUT("/:team_id", controllers.UpdateTeam)
			team.DELETE("/:team_id", controllers.DeleteTeam)

			// Team Members routes
			members := team.Group("/:team_id/members")
			{
				members.POST("/", controllers.AddTeamMember)
				members.GET("/", controllers.ListTeamMembers)
				members.DELETE("/:user_id", controllers.RemoveTeamMember)
			}
		}

		// Project routes
		project := protected.Group("/projects")
		{
			project.POST("/", controllers.CreateProject)
			project.GET("/", controllers.ListProjects)
			project.GET("/:project_id", controllers.GetProject)
			project.PUT("/:project_id", controllers.UpdateProject)
			project.DELETE("/:project_id", controllers.DeleteProject)

			// Collaborators routes
			collab := project.Group("/:project_id/collaborators")
			{
				collab.POST("/", controllers.AddCollaborator)
				collab.GET("/", controllers.ListCollaborators)
				collab.PUT("/:collaborator_id", controllers.UpdateCollaboratorRole)
				collab.DELETE("/:collaborator_id", controllers.RemoveCollaborator)
			}

			// Project Teams routes
			projectTeams := project.Group("/:project_id/teams")
			{
				projectTeams.POST("/", controllers.AddProjectTeam)
				projectTeams.GET("/", controllers.ListProjectTeams)
				projectTeams.DELETE("/:team_id", controllers.RemoveProjectTeam)
			}

			// Activity routes
			activity := project.Group("/:project_id/activities")
			{
				activity.POST("/", controllers.CreateActivity)
				activity.GET("/", controllers.ListActivities)
				activity.GET("/:activity_id", controllers.GetActivity)
				activity.PUT("/:activity_id", controllers.UpdateActivity)
				activity.DELETE("/:activity_id", controllers.DeleteActivity)
			}

			// Task routes
			task := project.Group("/:project_id/tasks")
			{
				task.POST("/", controllers.CreateTask)
				task.GET("/", controllers.ListTasks)
				task.GET("/:task_id", controllers.GetTask)
				task.PUT("/:task_id", controllers.UpdateTask)
				task.DELETE("/:task_id", controllers.DeleteTask)
			}

			// Note routes
			note := project.Group("/:project_id/notes")
			{
				note.POST("/", controllers.CreateNote)
				note.GET("/", controllers.ListNotes)
				note.GET("/:id", controllers.GetNote)
				note.PUT("/:id", controllers.UpdateNote)
				note.DELETE("/:id", controllers.DeleteNote)
			}

			// File routes (using fileController instance methods)
			file := project.Group("/:project_id/files")
			{
				file.POST("/", fileController.UploadFile)
				file.GET("/", fileController.ListFiles)
				file.GET("/:file_id", fileController.DownloadFile)
				file.DELETE("/:file_id", fileController.DeleteFile)
			}

			// Notification routes (using notificationController instance methods)
			notification := project.Group("/:project_id/notifications")
			{
				notification.POST("/", notificationController.CreateNotification)
				notification.GET("/", notificationController.ListNotifications)
				notification.GET("/:notification_id", notificationController.GetNotification)
				notification.PUT("/:notification_id", notificationController.UpdateNotification)
				notification.DELETE("/:notification_id", notificationController.DeleteNotification)
			}
		}
	}

	return router
}
