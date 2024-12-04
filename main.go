package main

import (
	"context"
	"fmt"

	"github.com/mfuadfakhruzzaki/backendaurauran/config"
	"github.com/mfuadfakhruzzaki/backendaurauran/models"
	"github.com/mfuadfakhruzzaki/backendaurauran/routes"
	"github.com/mfuadfakhruzzaki/backendaurauran/storage"
	"github.com/mfuadfakhruzzaki/backendaurauran/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configurations from .env
	config.LoadConfig()

	// Initialize logger
	utils.InitLogger()

	// Initialize validator
	utils.InitValidator()

	// Connect to the database
	db, err := setupDatabase()
	if err != nil {
		utils.Logger.Fatalf("Failed to connect to the database: %v", err)
	}

	// Initialize models
	models.InitModels(db)

	// Run migrations
	if err := db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Task{},
		&models.Activity{},
		&models.Collaboration{},
		&models.File{},
		&models.Note{},
		&models.Notification{},
		&models.EmailVerificationToken{},
		&models.Token{},
	); err != nil {
		utils.Logger.Fatalf("Failed to run auto migrations: %v", err)
	}

	// Load storage configuration
	storageConfig := config.LoadStorageConfig()

	// Initialize the storage service with AWS S3 credentials
	storageService, err := initializeS3StorageService(storageConfig)
	if err != nil {
		utils.Logger.Fatalf("Failed to initialize storage service: %v", err)
	}

	// Setup router with all routes
	router := routes.SetupRouter(db, storageService, storageConfig.BucketName)

	// Run the server
	port := config.AppConfig.Server.Port
	if port == "" {
		port = "8080"
	}
	utils.Logger.Infof("Server is running on port %s", port)
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		utils.Logger.Fatalf("Failed to run server: %v", err)
	}
}

func setupDatabase() (*gorm.DB, error) {
	dbConfig := config.AppConfig.Database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.Name, dbConfig.SSLMode, dbConfig.Timezone)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	utils.Logger.Println("Database connected successfully")
	return db, nil
}

// initializeS3StorageService initializes the S3 storage service using StorageConfig
func initializeS3StorageService(storageConfig config.StorageConfig) (storage.StorageService, error) {
	// Initialize S3StorageService
	s3Service, err := storage.NewS3StorageService(
		context.Background(),
		storageConfig.Region,
		storageConfig.AccessKeyID,
		storageConfig.SecretAccessKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize S3 storage service: %w", err)
	}

	utils.Logger.Println("S3 Storage Service initialized successfully")
	return s3Service, nil
}
