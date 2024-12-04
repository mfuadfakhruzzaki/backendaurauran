// config/storage.go
package config

import (
	"log"
	"os"
)

// StorageConfig menyimpan konfigurasi untuk penyimpanan (storage) menggunakan Amazon S3
type StorageConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
}

// LoadStorageConfig memuat konfigurasi storage dari variabel lingkungan
func LoadStorageConfig() StorageConfig {
	storageConfig := StorageConfig{
		Region:          os.Getenv("AWS_REGION"),
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		BucketName:      os.Getenv("AWS_S3_BUCKET_NAME"),
	}

	// Validasi bahwa semua konfigurasi yang diperlukan telah diisi
	if storageConfig.Region == "" {
		log.Fatal("AWS_REGION environment variable is required")
	}
	if storageConfig.BucketName == "" {
		log.Fatal("AWS_S3_BUCKET_NAME environment variable is required")
	}

	// Jika AccessKeyID atau SecretAccessKey tidak diisi, asumsi menggunakan IAM Role atau metode autentikasi lain
	if storageConfig.AccessKeyID == "" || storageConfig.SecretAccessKey == "" {
		log.Println("AWS_ACCESS_KEY_ID atau AWS_SECRET_ACCESS_KEY tidak ditemukan. Menggunakan metode autentikasi default AWS SDK.")
	}

	return storageConfig
}
