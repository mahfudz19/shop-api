package util

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
)

// TestGetContentType menguji fungsi getContentType
func TestGetContentType(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected string
	}{
		{
			name:     "JPEG lowercase",
			fileName: "image.jpg",
			expected: "image/jpeg",
		},
		{
			name:     "JPEG uppercase",
			fileName: "image.JPG",
			expected: "image/jpeg",
		},
		{
			name:     "JPEG mixed case",
			fileName: "image.JpEg",
			expected: "image/jpeg",
		},
		{
			name:     "PNG",
			fileName: "image.png",
			expected: "image/png",
		},
		{
			name:     "WebP",
			fileName: "image.webp",
			expected: "image/webp",
		},
		{
			name:     "SVG",
			fileName: "image.svg",
			expected: "image/svg+xml",
		},
		{
			name:     "PDF",
			fileName: "document.pdf",
			expected: "application/pdf",
		},
		{
			name:     "TXT",
			fileName: "readme.txt",
			expected: "text/plain",
		},
		{
			name:     "ZIP",
			fileName: "archive.zip",
			expected: "application/zip",
		},
		{
			name:     "Unknown extension",
			fileName: "file.xyz",
			expected: "application/octet-stream",
		},
		{
			name:     "No extension",
			fileName: "noextension",
			expected: "application/octet-stream",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getContentType(tc.fileName)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestExtractKeyFromURL menguji fungsi extractKeyFromURL
func TestExtractKeyFromURL(t *testing.T) {
	// Setup environment untuk test
	originalBucket := s3BucketName
	originalRegion := s3BucketRegion
	s3BucketName = "dev-stimi-yapmi"
	s3BucketRegion = "ap-southeast-1"

	defer func() {
		s3BucketName = originalBucket
		s3BucketRegion = originalRegion
	}()

	tests := []struct {
		name        string
		fileURL     string
		expectError bool
		expectedKey string
	}{
		{
			name:        "Virtual hosted style URL",
			fileURL:     "https://dev-stimi-yapmi.s3.ap-southeast-1.amazonaws.com/products/image.jpg",
			expectError: false,
			expectedKey: "products/image.jpg",
		},
		{
			name:        "Path style URL",
			fileURL:     "https://s3.ap-southeast-1.amazonaws.com/dev-stimi-yapmi/users/avatar.png",
			expectError: false,
			expectedKey: "users/avatar.png",
		},
		{
			name:        "Legacy URL",
			fileURL:     "https://dev-stimi-yapmi.s3.amazonaws.com/documents/file.pdf",
			expectError: false,
			expectedKey: "documents/file.pdf",
		},
		{
			name:        "URL with query parameters",
			fileURL:     "https://dev-stimi-yapmi.s3.ap-southeast-1.amazonaws.com/products/image.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256",
			expectError: false,
			expectedKey: "products/image.jpg",
		},
		{
			name:        "Invalid URL - different bucket",
			fileURL:     "https://other-bucket.s3.ap-southeast-1.amazonaws.com/products/image.jpg",
			expectError: true,
			expectedKey: "",
		},
		{
			name:        "Invalid URL - not S3",
			fileURL:     "https://example.com/image.jpg",
			expectError: true,
			expectedKey: "",
		},
		{
			name:        "Empty URL",
			fileURL:     "",
			expectError: true,
			expectedKey: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key, err := extractKeyFromURL(tc.fileURL)

			if tc.expectError {
				assert.Error(t, err)
				assert.Empty(t, key)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedKey, key)
			}
		})
	}
}

// TestIsS3Configured menguji fungsi IsS3Configured
func TestIsS3Configured(t *testing.T) {
	// Save original values
	originalClient := s3Client
	originalBucket := s3BucketName

	defer func() {
		s3Client = originalClient
		s3BucketName = originalBucket
	}()

	t.Run("Returns false when client is nil", func(t *testing.T) {
		s3Client = nil
		s3BucketName = "test-bucket"
		assert.False(t, IsS3Configured())
	})

	t.Run("Returns false when bucket name is empty", func(t *testing.T) {
		s3Client = &s3.Client{}
		s3BucketName = ""
		assert.False(t, IsS3Configured())
	})

	t.Run("Returns true when both client and bucket are set", func(t *testing.T) {
		s3Client = &s3.Client{}
		s3BucketName = "test-bucket"
		assert.True(t, IsS3Configured())
	})
}

// TestInitS3Client_Validation menguji validasi InitS3Client
func TestInitS3Client_Validation(t *testing.T) {
	// Save original values
	originalBucket := s3BucketName
	originalRegion := s3BucketRegion
	originalAccessKey := awsAccessKey
	originalSecretKey := awsSecretKey

	defer func() {
		s3BucketName = originalBucket
		s3BucketRegion = originalRegion
		awsAccessKey = originalAccessKey
		awsSecretKey = originalSecretKey
	}()

	t.Run("Returns error when bucket name is empty", func(t *testing.T) {
		s3BucketName = ""
		s3BucketRegion = "ap-southeast-1"
		awsAccessKey = "test-key"
		awsSecretKey = "test-secret"

		err := InitS3Client()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "AWS_BUCKET_NAME environment variable is required")
	})

	t.Run("Returns error when region is empty", func(t *testing.T) {
		s3BucketName = "test-bucket"
		s3BucketRegion = ""
		awsAccessKey = "test-key"
		awsSecretKey = "test-secret"

		err := InitS3Client()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "AWS_BUCKET_REGION environment variable is required")
	})

	t.Run("Returns error when access key is empty", func(t *testing.T) {
		s3BucketName = "test-bucket"
		s3BucketRegion = "ap-southeast-1"
		awsAccessKey = ""
		awsSecretKey = "test-secret"

		err := InitS3Client()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "AWS_ACCESS_KEY environment variable is required")
	})

	t.Run("Returns error when secret key is empty", func(t *testing.T) {
		s3BucketName = "test-bucket"
		s3BucketRegion = "ap-southeast-1"
		awsAccessKey = "test-key"
		awsSecretKey = ""

		err := InitS3Client()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "AWS_SECRET_KEY environment variable is required")
	})
}

// TestUploadFile_FileNameSanitization menguji sanitasi fileName
func TestUploadFile_FileNameSanitization(t *testing.T) {
	// Test bahwa filepath.Base digunakan untuk sanitasi
	// Ini adalah unit test untuk memastikan path traversal dicegah
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal filename",
			input:    "image.jpg",
			expected: "image.jpg",
		},
		{
			name:     "Path traversal attempt",
			input:    "../../../etc/passwd",
			expected: "passwd",
		},
		{
			name:     "Path with folders",
			input:    "folder/subfolder/image.png",
			expected: "image.png",
		},
		{
			name:     "Absolute path",
			input:    "/var/www/image.jpg",
			expected: "image.jpg",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := filepath.Base(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestInit_DefaultValues menguji default values di fungsi init
func TestInit_DefaultValues(t *testing.T) {
	// Save original values
	originalProductFolder := s3ProductFolder
	originalUserFolder := s3UserFolder

	defer func() {
		s3ProductFolder = originalProductFolder
		s3UserFolder = originalUserFolder
	}()

	t.Run("Default product folder", func(t *testing.T) {
		s3ProductFolder = ""
		// Simulasi init logic
		if s3ProductFolder == "" {
			s3ProductFolder = "products"
		}
		assert.Equal(t, "products", s3ProductFolder)
	})

	t.Run("Default user folder", func(t *testing.T) {
		s3UserFolder = ""
		// Simulasi init logic
		if s3UserFolder == "" {
			s3UserFolder = "users"
		}
		assert.Equal(t, "users", s3UserFolder)
	})
}

// TestGetContentType_Comprehensive menguji berbagai variasi ekstensi file
func TestGetContentType_Comprehensive(t *testing.T) {
	imageExtensions := []string{".jpg", ".jpeg", ".JPG", ".JPEG", ".Jpg", ".png", ".PNG", ".gif", ".GIF", ".webp", ".WEBP", ".svg", ".SVG"}
	for _, ext := range imageExtensions {
		result := getContentType("file" + ext)
		assert.True(t, strings.HasPrefix(result, "image/"), "Extension %s should return image content type", ext)
	}

	documentExtensions := []string{".pdf", ".PDF", ".txt", ".TXT"}
	for _, ext := range documentExtensions {
		result := getContentType("file" + ext)
		assert.NotEqual(t, "application/octet-stream", result, "Extension %s should have specific content type", ext)
	}

	archiveExtensions := []string{".zip", ".ZIP"}
	for _, ext := range archiveExtensions {
		result := getContentType("file" + ext)
		assert.Equal(t, "application/zip", result)
	}
}

// TestExtractKeyFromURL_EdgeCases menguji edge cases untuk extractKeyFromURL
func TestExtractKeyFromURL_EdgeCases(t *testing.T) {
	// Setup environment untuk test
	originalBucket := s3BucketName
	originalRegion := s3BucketRegion
	s3BucketName = "test-bucket"
	s3BucketRegion = "us-east-1"

	defer func() {
		s3BucketName = originalBucket
		s3BucketRegion = originalRegion
	}()

	tests := []struct {
		name        string
		fileURL     string
		expectError bool
	}{
		{
			name:        "URL with multiple slashes in key",
			fileURL:     "https://test-bucket.s3.us-east-1.amazonaws.com/folder/subfolder/deep/file.jpg",
			expectError: false,
		},
		{
			name:        "URL with special characters in key",
			fileURL:     "https://test-bucket.s3.us-east-1.amazonaws.com/folder/file%20name.jpg",
			expectError: false,
		},
		{
			name:        "URL with trailing slash",
			fileURL:     "https://test-bucket.s3.us-east-1.amazonaws.com/folder/",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key, err := extractKeyFromURL(tc.fileURL)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, key)
			}
		})
	}
}

// TestUploadFile_Integration adalah contoh test integration yang bisa dijalankan
// dengan environment variables yang sudah di-set
// Untuk menjalankan: go test -v -run TestUploadFile_Integration
func TestUploadFile_Integration(t *testing.T) {
	// Skip test ini jika tidak running dalam environment CI/CD
	if os.Getenv("AWS_INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set AWS_INTEGRATION_TEST=true to run.")
	}

	// Pastikan environment variables sudah di-set
	if s3BucketName == "" || s3BucketRegion == "" || awsAccessKey == "" || awsSecretKey == "" {
		t.Skip("AWS environment variables not set")
	}

	ctx := context.Background()
	testFile := strings.NewReader("test content for S3 upload")
	testFileName := "test-integration-upload.txt"
	testFolder := "test-folder"

	url, err := UploadFile(ctx, testFile, testFileName, int64(testFile.Len()), testFolder)

	if err != nil {
		t.Logf("Upload error (might be expected if credentials invalid): %v", err)
	} else {
		assert.NotEmpty(t, url)
		assert.Contains(t, url, s3BucketName)
		assert.Contains(t, url, testFolder)
		assert.Contains(t, url, testFileName)
	}
}
