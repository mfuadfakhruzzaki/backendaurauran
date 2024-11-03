// config/logger.go
package config

import "os"

// LoggerConfig menyimpan konfigurasi logger
type LoggerConfig struct {
    Level string
}

// LoadLoggerConfig memuat konfigurasi logger dari variabel lingkungan
func LoadLoggerConfig() LoggerConfig {
    return LoggerConfig{
        Level: os.Getenv("LOG_LEVEL"),
    }
}
