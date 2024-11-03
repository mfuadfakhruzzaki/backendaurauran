// config/database.go
package config

import "os"

// DatabaseConfig menyimpan konfigurasi database
type DatabaseConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    Name     string
    SSLMode  string
    Timezone string
}

// LoadDatabaseConfig memuat konfigurasi database dari variabel lingkungan
func LoadDatabaseConfig() DatabaseConfig {
    return DatabaseConfig{
        Host:     os.Getenv("DB_HOST"),
        Port:     os.Getenv("DB_PORT"),
        User:     os.Getenv("DB_USER"),
        Password: os.Getenv("DB_PASSWORD"),
        Name:     os.Getenv("DB_NAME"),
        SSLMode:  os.Getenv("DB_SSLMODE"),
        Timezone: os.Getenv("DB_TIMEZONE"),
    }
}
