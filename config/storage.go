// config/storage.go
package config

import "os"

// StorageConfig menyimpan konfigurasi untuk penyimpanan (storage)
type StorageConfig struct {
    CredentialsPath string
    BucketName      string
}

// LoadStorageConfig memuat konfigurasi storage dari variabel lingkungan
func LoadStorageConfig() StorageConfig {
    return StorageConfig{
        CredentialsPath: os.Getenv("STORAGE_CREDENTIALS_PATH"),
        BucketName:      os.Getenv("STORAGE_BUCKET_NAME"),
    }
}
