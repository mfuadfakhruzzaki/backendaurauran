// config/jwt.go
package config

import "os"

// JWTConfig menyimpan konfigurasi JWT
type JWTConfig struct {
    Secret          string
    ExpiresIn       string
    RefreshExpiresIn string
}

// LoadJWTConfig memuat konfigurasi JWT dari variabel lingkungan
func LoadJWTConfig() JWTConfig {
    return JWTConfig{
        Secret:          os.Getenv("JWT_SECRET"),
        ExpiresIn:       os.Getenv("JWT_EXPIRES_IN"),
        RefreshExpiresIn: os.Getenv("JWT_REFRESH_EXPIRES_IN"),
    }
}
