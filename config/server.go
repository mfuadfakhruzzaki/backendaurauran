// config/server.go
package config

import "os"

// ServerConfig menyimpan konfigurasi server
type ServerConfig struct {
    Port string
    Env  string
}

// LoadServerConfig memuat konfigurasi server dari variabel lingkungan
func LoadServerConfig() ServerConfig {
    return ServerConfig{
        Port: os.Getenv("PORT"),
        Env:  os.Getenv("ENV"),
    }
}
