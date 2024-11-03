// storage/storage.go
package storage

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// StorageService adalah interface untuk operasi penyimpanan file
type StorageService interface {
    UploadFile(ctx context.Context, bucketName, objectName string, data io.Reader, contentType string) (string, error)
    DeleteFile(ctx context.Context, bucketName, objectName string) error
    GeneratePresignedURL(ctx context.Context, bucketName, objectName string, expiration time.Duration) (string, error)
}

// GCSStorageService adalah implementasi StorageService menggunakan Google Cloud Storage
type GCSStorageService struct {
    Client              *storage.Client
    ServiceAccountEmail string
    PrivateKey          *rsa.PrivateKey
}

// NewGCSStorageService membuat instance baru dari GCSStorageService
func NewGCSStorageService(ctx context.Context, credentialsPath string) (*GCSStorageService, error) {
    // Membaca kredensial dari file JSON
    data, err := ioutil.ReadFile(credentialsPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read credentials file: %v", err)
    }

    // Menginisialisasi credentials
    creds, err := google.CredentialsFromJSON(ctx, data, storage.ScopeFullControl)
    if err != nil {
        return nil, fmt.Errorf("failed to parse credentials: %v", err)
    }

    // Parsing JSON untuk mendapatkan email dan private key
    var serviceAccountInfo struct {
        Type                    string `json:"type"`
        ProjectID               string `json:"project_id"`
        PrivateKeyID            string `json:"private_key_id"`
        PrivateKey              string `json:"private_key"`
        ClientEmail             string `json:"client_email"`
        ClientID                string `json:"client_id"`
        AuthURI                 string `json:"auth_uri"`
        TokenURI                string `json:"token_uri"`
        AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
        ClientX509CertURL       string `json:"client_x509_cert_url"`
    }

    if err := json.Unmarshal(data, &serviceAccountInfo); err != nil {
        return nil, fmt.Errorf("failed to unmarshal service account info: %v", err)
    }

    // Decode private key
    block, _ := pem.Decode([]byte(serviceAccountInfo.PrivateKey))
    if block == nil {
        return nil, fmt.Errorf("failed to parse private key PEM")
    }

    privateKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
    privateKey, ok := privateKeyInterface.(*rsa.PrivateKey)
    if !ok {
        return nil, fmt.Errorf("not an RSA private key")
    }
    if err != nil {
        return nil, fmt.Errorf("failed to parse RSA private key: %v", err)
    }

    client, err := storage.NewClient(ctx, option.WithCredentials(creds))
    if err != nil {
        return nil, fmt.Errorf("failed to create GCS client: %v", err)
    }

    return &GCSStorageService{
        Client:              client,
        ServiceAccountEmail: serviceAccountInfo.ClientEmail,
        PrivateKey:          privateKey,
    }, nil
}

// UploadFile mengupload file ke bucket GCS
// Mengembalikan URL file yang diupload
func (s *GCSStorageService) UploadFile(ctx context.Context, bucketName, objectName string, data io.Reader, contentType string) (string, error) {
    bkt := s.Client.Bucket(bucketName)
    obj := bkt.Object(objectName)
    writer := obj.NewWriter(ctx)
    writer.ContentType = contentType

    // Set public access jika diperlukan
    // obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader)

    if _, err := io.Copy(writer, data); err != nil {
        writer.Close()
        return "", fmt.Errorf("failed to write to GCS: %v", err)
    }

    if err := writer.Close(); err != nil {
        return "", fmt.Errorf("failed to close GCS writer: %v", err)
    }

    // Membuat URL file yang diupload
    fileURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)
    return fileURL, nil
}

// DeleteFile menghapus file dari bucket GCS
func (s *GCSStorageService) DeleteFile(ctx context.Context, bucketName, objectName string) error {
    bkt := s.Client.Bucket(bucketName)
    obj := bkt.Object(objectName)

    if err := obj.Delete(ctx); err != nil {
        return fmt.Errorf("failed to delete file from GCS: %v", err)
    }

    return nil
}

// GeneratePresignedURL membuat signed URL untuk mengakses objek dalam GCS
func (s *GCSStorageService) GeneratePresignedURL(ctx context.Context, bucketName, objectName string, expiration time.Duration) (string, error) {
    opts := &storage.SignedURLOptions{
        Scheme:         storage.SigningSchemeV4,
        Method:         "GET",
        Expires:        time.Now().Add(expiration),
        GoogleAccessID: s.ServiceAccountEmail,
        PrivateKey:     x509.MarshalPKCS1PrivateKey(s.PrivateKey),
    }

    signedURL, err := storage.SignedURL(bucketName, objectName, opts)
    if err != nil {
        return "", fmt.Errorf("failed to generate signed URL: %v", err)
    }

    return signedURL, nil
}

// GenerateUniqueObjectName menghasilkan nama objek unik menggunakan UUID dan timestamp
func GenerateUniqueObjectName(originalName string) string {
    ext := ""
    if dot := strings.LastIndex(originalName, "."); dot != -1 {
        ext = originalName[dot:]
    }
    return fmt.Sprintf("%s%s", uuid.New().String(), ext)
}
