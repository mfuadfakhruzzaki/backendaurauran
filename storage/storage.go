// storage/storage.go
package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

// StorageService adalah interface untuk operasi penyimpanan file
type StorageService interface {
	UploadFile(ctx context.Context, bucketName, objectName string, data io.Reader, contentType string) (string, error)
	DeleteFile(ctx context.Context, bucketName, objectName string) error
	GeneratePresignedURL(ctx context.Context, bucketName, objectName string, expiration time.Duration) (string, error)
}

// S3StorageService adalah implementasi StorageService menggunakan Amazon S3
type S3StorageService struct {
	Client       *s3.Client
	PresignClient *s3.PresignClient
	Region       string
}

// NewS3StorageService membuat instance baru dari S3StorageService
func NewS3StorageService(ctx context.Context, region, accessKey, secretKey string) (*S3StorageService, error) {
	// Konfigurasi AWS SDK
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %v", err)
	}

	// Membuat client S3
	client := s3.NewFromConfig(cfg)

	// Membuat Presign Client
	presignClient := s3.NewPresignClient(client)

	return &S3StorageService{
		Client:       client,
		PresignClient: presignClient,
		Region:       region,
	}, nil
}

// UploadFile mengupload file ke bucket S3
// Mengembalikan URL file yang diupload
func (s *S3StorageService) UploadFile(ctx context.Context, bucketName, objectName string, data io.Reader, contentType string) (string, error) {
	// Mengunggah objek ke S3
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectName),
		Body:        data, // Menggunakan io.Reader langsung tanpa membaca seluruh data
		ContentType: aws.String(contentType),
		ACL:         s3types.ObjectCannedACLPrivate, // Atur ACL sesuai kebutuhan
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %v", err)
	}

	// Membuat URL file yang diupload
	fileURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, s.Region, objectName)
	return fileURL, nil
}

// DeleteFile menghapus file dari bucket S3
func (s *S3StorageService) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %v", err)
	}

	// Opsional: Tunggu hingga objek benar-benar terhapus
	waiter := s3.NewObjectNotExistsWaiter(s.Client)
	err = waiter.Wait(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("error while waiting for object deletion: %v", err)
	}

	return nil
}

// GeneratePresignedURL membuat signed URL untuk mengakses objek dalam S3
func (s *S3StorageService) GeneratePresignedURL(ctx context.Context, bucketName, objectName string, expiration time.Duration) (string, error) {
	presignedReq, err := s.PresignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return presignedReq.URL, nil
}

// GenerateUniqueObjectName menghasilkan nama objek unik menggunakan UUID dan timestamp
func GenerateUniqueObjectName(originalName string) string {
	ext := path.Ext(originalName)
	return fmt.Sprintf("%s%s", uuid.New().String(), ext)
}
