// config/config.go
package config

import (
	"log"

	"github.com/joho/godotenv"
)

// Config menyimpan seluruh konfigurasi aplikasi
type Config struct {
    Database DatabaseConfig
    Server   ServerConfig
    JWT      JWTConfig
    Email    EmailConfig
    Logger   LoggerConfig
    Storage  StorageConfig // Tambahkan StorageConfig di sini
}

var AppConfig *Config

// LoadConfig memuat semua konfigurasi dari file .env
func LoadConfig() {
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    AppConfig = &Config{
        Database: LoadDatabaseConfig(),
        Server:   LoadServerConfig(),
        JWT:      LoadJWTConfig(),
        Email:    LoadEmailConfig(),
        Logger:   LoadLoggerConfig(),
        Storage:  LoadStorageConfig(), // Inisialisasi StorageConfig di sini
    }
}
