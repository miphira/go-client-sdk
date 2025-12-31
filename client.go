// Package sdk provides a Go client for the Miphira Object Storage API.
// It supports generating presigned URLs for secure, time-limited access to objects.
package sdk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// FileResponse represents the response returned by the API after uploading a file.
// The URL field contains the actual public URL with the server-generated UUID filename.
type FileResponse struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	OriginalName  string                 `json:"original_name"`
	Size          int64                  `json:"size"`
	SizeFormatted string                 `json:"size_formatted"`
	MimeType      string                 `json:"mime_type"`
	BucketID      string                 `json:"bucket_id"`
	URL           string                 `json:"url"` // This is the actual URL with server-generated UUID
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

// Client represents a Miphira Object Storage API client.
type Client struct {
	BaseURL    string
	ProjectID  string
	BucketName string
	AccessKey  string
	SecretKey  string
}

// NewClient creates a new Object Storage client with all required configuration.
//
// Example:
//
//	client := sdk.NewClient(
//	    "https://storage.miphira.com",
//	    "550e8400-e29b-41d4-a716-446655440000",  // Project ID
//	    "images",                                 // Bucket Name
//	    "MOS_YourAccessKey12345678",              // Access Key
//	    "your-secret-key-here",                   // Secret Key
//	)
func NewClient(baseURL, projectID, bucketName, accessKey, secretKey string) *Client {
	return &Client{
		BaseURL:    baseURL,
		ProjectID:  projectID,
		BucketName: bucketName,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
	}
}

// GenerateSignature creates an HMAC-SHA256 signature for the given parameters.
func (c *Client) GenerateSignature(method, path string, expires int64) string {
	stringToSign := fmt.Sprintf("%s\n%s\n%d", method, path, expires)

	h := hmac.New(sha256.New, []byte(c.SecretKey))
	h.Write([]byte(stringToSign))

	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// GeneratePresignedURL creates a presigned URL for the specified HTTP method and path.
//
// Parameters:
//   - method: HTTP method (GET, POST, DELETE)
//   - path: API path (e.g., /api/v1/projects/{projectId}/buckets/{bucketName}/objects/{filename})
//   - expiresIn: Duration until the URL expires
//
// Returns a fully-formed presigned URL with authentication parameters.
func (c *Client) GeneratePresignedURL(method, path string, expiresIn time.Duration) string {
	expires := time.Now().Add(expiresIn).Unix()
	signature := c.GenerateSignature(method, path, expires)

	return fmt.Sprintf("%s%s?X-Mos-AccessKey=%s&X-Mos-Expires=%d&X-Mos-Signature=%s",
		c.BaseURL,
		path,
		c.AccessKey,
		expires,
		url.QueryEscape(signature),
	)
}

// GetObjectURL generates a presigned URL for downloading/viewing an object.
//
// Example:
//
//	url := client.GetObjectURL("photo.jpg", time.Hour)
func (c *Client) GetObjectURL(filename string, expiresIn time.Duration) string {
	path := fmt.Sprintf("/api/v1/projects/%s/buckets/%s/objects/%s", c.ProjectID, c.BucketName, filename)
	return c.GeneratePresignedURL("GET", path, expiresIn)
}

// UploadObjectURL generates a presigned URL for uploading an object.
//
// Example:
//
//	url := client.UploadObjectURL(time.Hour)
//	// Use this URL with a multipart/form-data POST request
func (c *Client) UploadObjectURL(expiresIn time.Duration) string {
	path := fmt.Sprintf("/api/v1/projects/%s/buckets/%s/objects", c.ProjectID, c.BucketName)
	return c.GeneratePresignedURL("POST", path, expiresIn)
}

// DeleteObjectURL generates a presigned URL for deleting an object.
//
// Example:
//
//	url := client.DeleteObjectURL("photo.jpg", time.Hour)
func (c *Client) DeleteObjectURL(filename string, expiresIn time.Duration) string {
	path := fmt.Sprintf("/api/v1/projects/%s/buckets/%s/objects/%s", c.ProjectID, c.BucketName, filename)
	return c.GeneratePresignedURL("DELETE", path, expiresIn)
}

// GetPublicObjectURL generates a public URL for accessing an object without authentication.
// This URL does not require signatures or expiration time and is accessible to anyone.
//
// Note: Public URLs only work in beta mode where all buckets are public.
// In production, you should use presigned URLs with GetObjectURL() for secure access.
//
// Example:
//
//	url := client.GetPublicObjectURL("photo.jpg")
//	// Returns: https://storage.example.com/api/v1/public/projects/{projectId}/buckets/{bucket}/photo.jpg
func (c *Client) GetPublicObjectURL(filename string) string {
	return fmt.Sprintf("%s/api/v1/public/projects/%s/buckets/%s/%s",
		c.BaseURL,
		c.ProjectID,
		c.BucketName,
		filename,
	)
}

// PresignedURLOptions provides additional options for URL generation.
type PresignedURLOptions struct {
	ExpiresIn time.Duration
}

// DefaultOptions returns default presigned URL options (1 hour expiry).
func DefaultOptions() PresignedURLOptions {
	return PresignedURLOptions{
		ExpiresIn: time.Hour,
	}
}

// UploadOptions provides options for file upload operations.
type UploadOptions struct {
	Metadata  map[string]interface{} // Optional metadata to attach to the file
	ExpiresIn time.Duration          // URL expiration time (default: 1 hour)
}

// Upload uploads a file from the local filesystem and returns the server response.
// The returned FileResponse contains the actual URL with server-generated UUID filename.
//
// Example:
//
//	resp, err := client.Upload("photo.jpg", &sdk.UploadOptions{
//	    Metadata: map[string]interface{}{"category": "profile"},
//	    ExpiresIn: time.Hour,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("File uploaded: %s\n", resp.URL)
//	// Use resp.URL or extract filename from it
func (c *Client) Upload(filePath string, opts *UploadOptions) (*FileResponse, error) {
	// Set defaults
	if opts == nil {
		opts = &UploadOptions{ExpiresIn: time.Hour}
	}
	if opts.ExpiresIn == 0 {
		opts.ExpiresIn = time.Hour
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	// Add metadata if provided
	if opts.Metadata != nil {
		metadataJSON, err := json.Marshal(opts.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		writer.WriteField("metadata", string(metadataJSON))
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Generate presigned URL
	uploadURL := c.UploadObjectURL(opts.ExpiresIn)

	// Create request
	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var fileResp FileResponse
	if err := json.NewDecoder(resp.Body).Decode(&fileResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &fileResp, nil
}

// UploadBytes uploads file content from memory (byte slice) and returns the server response.
// Useful for uploading generated content, images from memory, or data from other sources.
//
// Example:
//
//	data := []byte("Hello, World!")
//	resp, err := client.UploadBytes("hello.txt", data, &sdk.UploadOptions{
//	    Metadata: map[string]interface{}{"type": "text"},
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("File uploaded: %s\n", resp.URL)
func (c *Client) UploadBytes(filename string, data []byte, opts *UploadOptions) (*FileResponse, error) {
	// Set defaults
	if opts == nil {
		opts = &UploadOptions{ExpiresIn: time.Hour}
	}
	if opts.ExpiresIn == 0 {
		opts.ExpiresIn = time.Hour
	}

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}

	// Add metadata if provided
	if opts.Metadata != nil {
		metadataJSON, err := json.Marshal(opts.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		writer.WriteField("metadata", string(metadataJSON))
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Generate presigned URL
	uploadURL := c.UploadObjectURL(opts.ExpiresIn)

	// Create request
	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var fileResp FileResponse
	if err := json.NewDecoder(resp.Body).Decode(&fileResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &fileResp, nil
}

// Download downloads a file and saves it to the specified local path.
// Uses the server-generated UUID filename to fetch the file.
//
// Example:
//
//	err := client.Download("8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg", "local_photo.jpg", time.Hour)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (c *Client) Download(filename string, localPath string, expiresIn time.Duration) error {
	// Generate URL (use public URL in beta mode)
	url := c.GetPublicObjectURL(filename)

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Copy data
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

// Delete deletes a file from storage using presigned URL.
//
// Example:
//
//	err := client.Delete("8aabd7f7-1dbf-4ea4-8918-db66069746e7.jpg", time.Hour)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (c *Client) Delete(filename string, expiresIn time.Duration) error {
	// Generate presigned delete URL
	url := c.DeleteObjectURL(filename, expiresIn)

	// Create DELETE request
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
