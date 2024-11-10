// routes/routes.go
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mfuadfakhruzzaki/backendaurauran/controllers"
	"github.com/mfuadfakhruzzaki/backendaurauran/middlewares"
	"github.com/mfuadfakhruzzaki/backendaurauran/models"
	"github.com/mfuadfakhruzzaki/backendaurauran/storage"
	"gorm.io/gorm"
)

// SetupRouter sets up all routes for the application
func SetupRouter(db *gorm.DB, storageService storage.StorageService, bucketName string) *gin.Engine {
	router := gin.Default()

	// Initialize controllers with dependencies
	fileController := controllers.NewFileController(db, storageService, bucketName)
	// You can initialize other controllers similarly if they require dependencies

	// Middleware global
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
		auth.POST("/auth/reset-password-api", controllers.ResetPasswordAPI)
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

		// Project routes
		project := protected.Group("/projects")
		{
			project.POST("/", middlewares.RoleMiddleware(models.RoleAdmin, models.RoleManager), controllers.CreateProject)
			project.GET("/", controllers.ListProjects)
			project.GET("/:project_id", controllers.GetProject)
			project.PUT("/:project_id", controllers.UpdateProject)
			project.DELETE("/:project_id", controllers.DeleteProject)

			// Collaboration routes
			collab := project.Group("/:project_id/collaborators")
			{
				collab.POST("/", controllers.AddCollaborator)
				collab.GET("/", controllers.ListCollaborators)
				collab.PUT("/:collaborator_id", controllers.UpdateCollaboratorRole)
				collab.DELETE("/:collaborator_id", controllers.RemoveCollaborator)
			}

			// Activity routes
			activity := project.Group("/:project_id/activities")
			{
				activity.POST("/", controllers.CreateActivity)
				activity.GET("/", controllers.ListActivities)
				activity.GET("/:id", controllers.GetActivity)
				activity.PUT("/:id", controllers.UpdateActivity)
				activity.DELETE("/:id", controllers.DeleteActivity)
			}

			// Task routes
			task := project.Group("/:project_id/tasks")
			{
				task.POST("/", controllers.CreateTask)
				task.GET("/", controllers.ListTasks)
				task.GET("/:id", controllers.GetTask)
				task.PUT("/:id", controllers.UpdateTask)
				task.DELETE("/:id", controllers.DeleteTask)
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
				file.GET("/:id", fileController.DownloadFile)
				file.DELETE("/:id", fileController.DeleteFile)
			}

			// Notification routes
			notification := project.Group("/:project_id/notifications")
			{
				notification.POST("/", controllers.CreateNotification)
				notification.GET("/", controllers.ListNotifications)
				notification.GET("/:id", controllers.GetNotification)
				notification.PUT("/:id", controllers.UpdateNotification)
				notification.DELETE("/:id", controllers.DeleteNotification)
			}
		}
	}

	return router
}
