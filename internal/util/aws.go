package util

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	s3Client        *s3.Client
	s3BucketName    string
	s3BucketRegion  string
	s3ProductFolder string
	s3UserFolder    string
	awsAccessKey    string
	awsSecretKey    string
)

// init menginisialisasi konfigurasi S3 dari environment variables
// Environment variables yang digunakan:
//   - AWS_BUCKET_NAME: nama bucket S3
//   - AWS_BUCKET_REGION: region AWS (contoh: ap-southeast-1)
//   - AWS_ACCESS_KEY: AWS Access Key ID
//   - AWS_SECRET_KEY: AWS Secret Access Key
//   - AWS_S3_PRODUCT_FOLDER: folder untuk produk (default: "products")
//   - AWS_S3_USER_FOLDER: folder untuk user (default: "users")
func init() {
	s3BucketName = os.Getenv("AWS_BUCKET_NAME")
	s3BucketRegion = os.Getenv("AWS_BUCKET_REGION")
	awsAccessKey = os.Getenv("AWS_ACCESS_KEY")
	awsSecretKey = os.Getenv("AWS_SECRET_KEY")
	s3ProductFolder = os.Getenv("AWS_S3_PRODUCT_FOLDER")
	s3UserFolder = os.Getenv("AWS_S3_USER_FOLDER")

	if s3ProductFolder == "" {
		s3ProductFolder = "products"
	}
	if s3UserFolder == "" {
		s3UserFolder = "users"
	}
}

// InitS3Client menginisialisasi S3 client dengan konfigurasi dari environment
// Fungsi ini harus dipanggil sekali saat aplikasi dimulai
// Menggunakan explicit credentials dari environment variables
func InitS3Client() error {
	// Validasi konfigurasi yang wajib ada
	if s3BucketName == "" {
		return fmt.Errorf("AWS_BUCKET_NAME environment variable is required")
	}
	if s3BucketRegion == "" {
		return fmt.Errorf("AWS_BUCKET_REGION environment variable is required")
	}
	if awsAccessKey == "" {
		return fmt.Errorf("AWS_ACCESS_KEY environment variable is required")
	}
	if awsSecretKey == "" {
		return fmt.Errorf("AWS_SECRET_KEY environment variable is required")
	}

	// Buat credentials secara eksplisit
	creds := credentials.NewStaticCredentialsProvider(awsAccessKey, awsSecretKey, "")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(s3BucketRegion),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return fmt.Errorf("unable to load SDK config: %v", err)
	}

	s3Client = s3.NewFromConfig(cfg)
	return nil
}

// IsS3Configured mengembalikan true jika S3 client sudah diinisialisasi
func IsS3Configured() bool {
	return s3Client != nil && s3BucketName != ""
}

// UploadFile mengupload file ke S3 bucket
// Parameters:
//   - ctx: context untuk cancelable operation
//   - file: io.Reader yang berisi data file
//   - fileName: nama file termasuk ekstensi
//   - fileSize: ukuran file dalam bytes (opsional, bisa 0)
//   - folder: folder destination di S3 bucket
//
// Returns:
//   - string: URL lengkap ke file yang diupload
//   - error: error jika upload gagal
func UploadFile(ctx context.Context, file io.Reader, fileName string, fileSize int64, folder string) (string, error) {
	if !IsS3Configured() {
		if err := InitS3Client(); err != nil {
			return "", fmt.Errorf("failed to initialize S3 client: %v", err)
		}
	}

	// Sanitasi fileName untuk mencegah path traversal
	fileName = filepath.Base(fileName)
	key := filepath.Join(folder, fileName)

	uploader := manager.NewUploader(s3Client)

	uploadInput := &s3.PutObjectInput{
		Bucket:      aws.String(s3BucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(getContentType(fileName)),
	}

	if fileSize > 0 {
		uploadInput.ContentLength = aws.Int64(fileSize)
	}

	_, uploadErr := uploader.Upload(ctx, uploadInput)
	if uploadErr != nil {
		return "", fmt.Errorf("unable to upload file: %v", uploadErr)
	}

	// Build URL secara eksplisit agar konsisten
	fileURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s3BucketName, s3BucketRegion, key)
	return fileURL, nil
}

// UploadProductImage mengupload gambar produk ke folder yang sudah dikonfigurasi
func UploadProductImage(ctx context.Context, file io.Reader, fileName string, fileSize int64) (string, error) {
	return UploadFile(ctx, file, fileName, fileSize, s3ProductFolder)
}

// UploadUserImage mengupload gambar user/avatar ke folder yang sudah dikonfigurasi
func UploadUserImage(ctx context.Context, file io.Reader, fileName string, fileSize int64) (string, error) {
	return UploadFile(ctx, file, fileName, fileSize, s3UserFolder)
}

// DeleteFile menghapus file dari S3 berdasarkan URL
func DeleteFile(ctx context.Context, fileURL string) error {
	if !IsS3Configured() {
		if err := InitS3Client(); err != nil {
			return fmt.Errorf("failed to initialize S3 client: %v", err)
		}
	}

	key, err := extractKeyFromURL(fileURL)
	if err != nil {
		return err
	}

	_, err = s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("unable to delete file: %v", err)
	}

	return nil
}

// extractKeyFromURL mengekstrak object key dari URL S3
// Mendukung format:
//   - https://bucket-name.s3.region.amazonaws.com/key
//   - https://s3.region.amazonaws.com/bucket-name/key
//   - https://bucket-name.s3.amazonaws.com/key (legacy)
func extractKeyFromURL(fileURL string) (string, error) {
	// Hapus query parameter jika ada
	if idx := strings.Index(fileURL, "?"); idx != -1 {
		fileURL = fileURL[:idx]
	}

	// Coba berbagai format URL S3
	patterns := []string{
		fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", s3BucketName, s3BucketRegion),
		fmt.Sprintf("https://s3.%s.amazonaws.com/%s/", s3BucketRegion, s3BucketName),
		fmt.Sprintf("https://%s.s3.amazonaws.com/", s3BucketName),
	}

	for _, pattern := range patterns {
		if strings.HasPrefix(fileURL, pattern) {
			return strings.TrimPrefix(fileURL, pattern), nil
		}
	}

	return "", fmt.Errorf("invalid S3 URL format: expected URL to start with bucket URL")
}

// getContentType menentukan MIME type berdasarkan ekstensi file
func getContentType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}
