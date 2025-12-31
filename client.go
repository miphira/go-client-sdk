// Package sdk provides a Go client for the Miphira Object Storage API.
// It supports generating presigned URLs for secure, time-limited access to objects.
package sdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"time"
)

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
