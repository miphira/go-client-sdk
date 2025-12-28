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
	BaseURL   string
	AccessKey string
	SecretKey string
}

// NewClient creates a new Object Storage client.
//
// Example:
//
//	client := sdk.NewClient(
//	    "https://storage.miphira.com",
//	    "MOS_YourAccessKey12345678",
//	    "your-secret-key-here",
//	)
func NewClient(baseURL, accessKey, secretKey string) *Client {
	return &Client{
		BaseURL:   baseURL,
		AccessKey: accessKey,
		SecretKey: secretKey,
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
//	url := client.GetObjectURL("project-uuid", "images", "photo.jpg", time.Hour)
func (c *Client) GetObjectURL(projectID, bucketName, filename string, expiresIn time.Duration) string {
	path := fmt.Sprintf("/api/v1/projects/%s/buckets/%s/objects/%s", projectID, bucketName, filename)
	return c.GeneratePresignedURL("GET", path, expiresIn)
}

// UploadObjectURL generates a presigned URL for uploading an object.
//
// Example:
//
//	url := client.UploadObjectURL("project-uuid", "images", time.Hour)
//	// Use this URL with a multipart/form-data POST request
func (c *Client) UploadObjectURL(projectID, bucketName string, expiresIn time.Duration) string {
	path := fmt.Sprintf("/api/v1/projects/%s/buckets/%s/objects", projectID, bucketName)
	return c.GeneratePresignedURL("POST", path, expiresIn)
}

// DeleteObjectURL generates a presigned URL for deleting an object.
//
// Example:
//
//	url := client.DeleteObjectURL("project-uuid", "images", "photo.jpg", time.Hour)
func (c *Client) DeleteObjectURL(projectID, bucketName, filename string, expiresIn time.Duration) string {
	path := fmt.Sprintf("/api/v1/projects/%s/buckets/%s/objects/%s", projectID, bucketName, filename)
	return c.GeneratePresignedURL("DELETE", path, expiresIn)
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
